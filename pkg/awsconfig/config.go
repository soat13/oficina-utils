package awsconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type Config struct {
	Region          string
	EndpointURL     string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func New(ctx context.Context, c Config) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	if c.Region != "" {
		opts = append(opts, config.WithRegion(c.Region))
	}

	if c.AccessKeyID != "" && c.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				c.AccessKeyID,
				c.SecretAccessKey,
				c.SessionToken,
			),
		))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}
