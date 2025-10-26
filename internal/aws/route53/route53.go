package route53

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/jaehong21/hibiscus/internal/aws"
)

var client *route53.Client

// Map of AWS region to ELB hosted zone ID
// including Application Load Balancer and Classic Load Balancer
// https://docs.aws.amazon.com/general/latest/gr/elb.html
var ELB_HOSTED_ZONE_IDS = map[string]string{
	"us-east-1":      "Z35SXDOTRQ7X7K",
	"us-east-2":      "Z3AADJGX6KTTL2",
	"us-west-1":      "Z368ELLRRE2KJ0",
	"us-west-2":      "Z1H1FL5HABSF5",
	"af-south-1":     "Z268VQBMOI5EKX",
	"ap-east-1":      "Z3DQVH9N71FHZ0",
	"ap-south-1":     "ZP97RAFLXTNZK",
	"ap-northeast-3": "Z5LXEXXYW11ES",
	"ap-northeast-2": "ZWKZPGTI48KDX",
	"ap-southeast-1": "Z1LMS91P8CMLE5",
	"ap-southeast-2": "Z1GM3OXH4ZPM65",
	"ap-northeast-1": "Z14GRHDCWA56QT",
	"ca-central-1":   "ZQSVJUPU6J1EY",
	"eu-central-1":   "Z215JYRZR1TBD5",
	"eu-west-1":      "Z32O12XQLNTSW2",
	"eu-west-2":      "ZHURV8PSTC4K8",
	"eu-south-1":     "Z3ULH7SSC9OV64",
	"eu-west-3":      "Z3Q77PNBQS71R4",
	"eu-north-1":     "Z23TAZ6LKFMNIO",
	"me-south-1":     "ZS929ML54UICD",
	"sa-east-1":      "Z2P70J7HTTTPLU",
}

// Map of AWS region to NLB (Network Load Balancer) hosted zone ID
// https://docs.aws.amazon.com/general/latest/gr/elb.html
var NLB_HOSTED_ZONE_IDS = map[string]string{
	"us-east-1":      "Z26RNL4JYFTOTI",
	"us-east-2":      "ZLMOA37VPKANP",
	"us-west-1":      "Z24FKFUX50B4VW",
	"us-west-2":      "Z18D5FSROUN65G",
	"af-south-1":     "Z203XCE67M25HM",
	"ap-east-1":      "Z12Y7K3UBGUAD1",
	"ap-south-1":     "ZVDDRBQ08TROA",
	"ap-northeast-3": "Z1GWIQ4HH19I5X",
	"ap-northeast-2": "ZIBE1TIR4HY56",
	"ap-southeast-1": "ZKVM4W9LS7TM",
	"ap-southeast-2": "ZCT6FZBF4DROD",
	"ap-northeast-1": "Z31USIVHYNEOWT",
	"ca-central-1":   "Z2EPGBW3API2WT",
	"eu-central-1":   "Z3F0SRJ5LGBH90",
	"eu-west-1":      "Z2IFOLAFXWLO4F",
	"eu-west-2":      "ZD4D7Y8KGAS4G",
	"eu-south-1":     "Z23146JA1KNAFP",
	"eu-west-3":      "Z1CMS0P5QUZ6D5",
	"eu-north-1":     "Z1UDT6IFJ4EJM",
	"me-south-1":     "Z3QSRYVP46NYYV",
	"sa-east-1":      "ZTK26PT1VY4CU",
}

const (
	// CloudFront hosted zone ID is global
	// https://docs.aws.amazon.com/general/latest/gr/cf_region.html
	CLOUDFRONT_HOSTED_ZONE_ID = "Z2FDTNDATAQYW2"
)

// IsCloudFrontAlias checks if the alias target's hosted zone ID matches CloudFront's ID
func IsCloudFrontAlias(hostedZoneId *string) bool {
	if hostedZoneId == nil {
		return false
	}
	return *hostedZoneId == CLOUDFRONT_HOSTED_ZONE_ID
}

// IsELBAlias checks if the alias target's hosted zone ID matches any of the ELB IDs
func IsELBAlias(hostedZoneId *string) bool {
	if hostedZoneId == nil {
		return false
	}

	// Check ALB and Classic Load Balancer IDs
	for _, id := range ELB_HOSTED_ZONE_IDS {
		if *hostedZoneId == id {
			return true
		}
	}
	// Check Network Load Balancer IDs
	for _, id := range NLB_HOSTED_ZONE_IDS {
		if *hostedZoneId == id {
			return true
		}
	}

	return false
}

func ListHostedZones() ([]types.HostedZone, error) {
	if err := setupClient(); err != nil {
		return nil, err
	}

	zones, err := client.ListHostedZones(context.TODO(), &route53.ListHostedZonesInput{})
	if err != nil {
		return nil, err
	}

	return zones.HostedZones, nil
}

func ListRecords(hostedZoneID *string) ([]types.ResourceRecordSet, error) {
	if err := setupClient(); err != nil {
		return nil, err
	}

	records, err := client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
		HostedZoneId: hostedZoneID,
	})
	if err != nil {
		return nil, err
	}

	return records.ResourceRecordSets, nil
}

func UpsertRecord(hostedZoneID *string, record types.ResourceRecordSet) error {
	if err := setupClient(); err != nil {
		return err
	}

	_, err := client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: hostedZoneID,
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action:            types.ChangeActionUpsert,
					ResourceRecordSet: &record,
				},
			},
		},
	})

	return err
}

func setupClient() error {
	if client != nil {
		return nil
	}

	cfg, err := aws.GetAWSConfig(context.Background())
	if err != nil {
		return err
	}

	client = route53.NewFromConfig(cfg)
	return nil
}
