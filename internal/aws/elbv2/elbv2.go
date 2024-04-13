package elbv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/jaehong21/hibiscus/internal/aws"
)

// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2
var client *elasticloadbalancingv2.Client

func DescribeLoadBalancers() ([]types.LoadBalancer, error) {
	if err := setupClient(); err != nil {
		return nil, err
	}
	elbs, err := client.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, err
	}

	return elbs.LoadBalancers, nil
}

func DescribeListeners(loadBalancerArn *string) ([]types.Listener, error) {
	if err := setupClient(); err != nil {
		return nil, err
	}
	listeners, err := client.DescribeListeners(context.TODO(), &elasticloadbalancingv2.DescribeListenersInput{
		LoadBalancerArn: loadBalancerArn,
	})
	if err != nil {
		return nil, err
	}

	return listeners.Listeners, nil
}

func DescribeRules(listenerArn *string) ([]types.Rule, error) {
	if err := setupClient(); err != nil {
		return nil, err
	}
	rules, err := client.DescribeRules(context.TODO(), &elasticloadbalancingv2.DescribeRulesInput{
		ListenerArn: listenerArn,
	})
	if err != nil {
		return nil, err
	}

	return rules.Rules, nil
}

func setupClient() error {
	if client != nil {
		return nil
	}
	cfg, err := aws.GetAWSConfig(context.Background())
	if err != nil {
		return err
	}
	client = elasticloadbalancingv2.NewFromConfig(cfg)
	return nil
}
