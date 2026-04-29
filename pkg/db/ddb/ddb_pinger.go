package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Pinger struct {
	client *dynamodb.Client
	table  string
}

func NewPinger(c *dynamodb.Client, table string) *Pinger {
	return &Pinger{client: c, table: table}
}

func (p *Pinger) PingContext(ctx context.Context) error {
	_, err := p.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(p.table),
	})
	return err
}
