package auth

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func GetClient() (client *ec2.Client, err error) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, ErrNoDefaultAuthentication
	}

	client = ec2.NewFromConfig(cfg)
	return client, nil
}
