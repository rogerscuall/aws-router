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
		return nil, err
	}
	client = ec2.NewFromConfig(cfg)
	return client, nil
	// clientSTS := sts.NewFromConfig(cfg)
	// 	identity, err := clientSTS.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	//fmt.Printf("Account: %s, Arn: %s", aws.ToString(identity.Account), aws.ToString(identity.Arn))
}
