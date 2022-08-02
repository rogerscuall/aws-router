package auth

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func GetClient() (client *ec2.Client, err error) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, ErrNoDefaultAuthentication
	}

	client = ec2.NewFromConfig(cfg)

	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		appCreds := aws.NewCredentialsCache(ec2rolecreds.New())
		value, err := appCreds.Retrieve(context.TODO())
		if err != nil {
			return nil, ErrNoEC2ProfileRole
		}
		log.Printf("[INFO] Using EC2 Profile Role: %s", value.AccessKeyID)
	}
	log.Printf("Account: %s, Arn: %s", aws.ToString(identity.Account), aws.ToString(identity.Arn))
	return client, nil
}
