package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/aws-sdk-go/aws"
)

const (
	envDelegationSetID = "ZONEUP_DELEGATION_SET_ID"
	envSOAEmail        = "ZONEUP_SOA_EMAIL"
	envNameServers     = "ZONEUP_NAMESERVERS"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s domain.com", os.Args[0])
	}

	domain := os.Args[1]
	if err := validateDomain(domain); err != nil {
		log.Fatal(err)
	}

	delegationSetID := os.Getenv(envDelegationSetID)
	if delegationSetID == "" {
		log.Fatalf("%s environment variable is required", envDelegationSetID)
	}

	soaEmail := os.Getenv(envSOAEmail)
	if soaEmail == "" {
		log.Fatalf("%s environment variable is required", envSOAEmail)
	}

	nameServers := strings.Split(os.Getenv(envNameServers), ",")
	if len(nameServers) == 0 {
		log.Fatalf("%s environment variable is required", envNameServers)
	}

	if err := createHostedZone(domain, delegationSetID, soaEmail, nameServers); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully created hosted zone for domain: %s\n", domain)
}

func validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("invalid domain format: %s", domain)
	}
	return nil
}

func createHostedZone(domain, delegationSetID, soaEmail string, nameServers []string) error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %v", err)
	}

	client := route53.NewFromConfig(cfg)

	input := &route53.CreateHostedZoneInput{
		Name:            &domain,
		DelegationSetId: &delegationSetID,
		CallerReference: aws.String(fmt.Sprintf("zoneup-%s-%d", domain, time.Now().Unix())),
		HostedZoneConfig: &types.HostedZoneConfig{
			Comment:     aws.String(fmt.Sprintf("Created by zoneup for %s", domain)),
			PrivateZone: false,
		},
	}

	result, err := client.CreateHostedZone(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create hosted zone: %v", err)
	}

	nsRecords := make([]types.ResourceRecord, len(nameServers))
	for i, ns := range nameServers {
		nsRecords[i] = types.ResourceRecord{
			Value: aws.String(ns),
		}
	}

	// Update SOA record with custom email
	soaUpdateInput := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: result.HostedZone.Id,
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name:            aws.String(domain),
						Type:            types.RRTypeNs,
						TTL:             aws.Int64(60),
						ResourceRecords: nsRecords,
					},
				},
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(domain),
						Type: types.RRTypeSoa,
						TTL:  aws.Int64(900),
						ResourceRecords: []types.ResourceRecord{
							{
								Value: aws.String(fmt.Sprintf("%s. %s. 1 7200 900 1209600 86400",
									result.DelegationSet.NameServers[0],
									soaEmail)),
							},
						},
					},
				},
			},
		},
	}

	_, err = client.ChangeResourceRecordSets(ctx, soaUpdateInput)
	if err != nil {
		return fmt.Errorf("failed to update SOA record: %v", err)
	}

	return nil
}
