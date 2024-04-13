package route53

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/jaehong21/hibiscus/internal/aws"
)

var client *route53.Client

const (
	// TODO: add ELB hosted zone id for each region
	ELB_HOSTED_ZONE           = ""
	CLOUDFRONT_HOSTED_ZONE_ID = "Z2FDTNDATAQYW2"
)

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

// TODO: need to add for Cloudfront and ELB
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
