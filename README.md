# ZoneUp

ZoneUp is a command-line tool that sets up domains on AWS Route53 using a predefined delegation set.

## Prerequisites

- Go 1.21 or later
- AWS credentials configured (either through environment variables or AWS credentials file)
- Route53 Delegation Set ID
- SOA Webmaster email address

## Environment Variables

The following environment variables must be set:

- `ZONEUP_DELEGATION_SET_ID`: Your AWS Route53 delegation set ID
- `ZONEUP_SOA_EMAIL`: Email address for the SOA record (e.g., webmaster@example.com)
- `ZONEUP_NAMESERVERS`: Comma-separated list of nameservers (e.g., ns1.example.com,ns2.example.com)
- `AWS_ACCESS_KEY_ID`: Your AWS access key ID
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret access key
- `AWS_REGION`: Your AWS region (e.g., us-east-1)

## Installation

```bash
go install github.com/nellcorp/zoneup@latest
```

## Usage

```bash
zoneup example.com
```

This will create a new hosted zone for example.com using the configured delegation set and SOA email.
