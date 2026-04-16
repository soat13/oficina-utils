package sqs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/rs/zerolog/log"
	"github.com/soat13/oficina-utils/pkg/messaging"
)

type Publisher struct {
	client  *sns.Client
	baseARN string
}

func NewPublisher(ctx context.Context, awsEndpoint string, snsBaseARN string) (messaging.Publisher, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := sns.NewFromConfig(cfg, func(o *sns.Options) {
		if awsEndpoint != "" {
			o.BaseEndpoint = &awsEndpoint
		}
	})

	baseARN := strings.TrimSuffix(snsBaseARN, ":")

	return &Publisher{
		client:  client,
		baseARN: baseARN,
	}, nil
}

func (b *Publisher) Publish(ctx context.Context, topic string, payload []byte) error {
	if b.client == nil {
		return errors.New("sns client is nil")
	}

	if topic == "" {
		return errors.New("topic is required")
	}

	if len(payload) == 0 {
		return errors.New("payload is required")
	}

	message := string(payload)

	_, err := b.client.Publish(ctx, &sns.PublishInput{
		TopicArn: &topic,
		Message:  &message,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("topic", topic).
			Str("topic_arn", topic).
			Msg("failed to publish message to SNS")

		return fmt.Errorf("sns publish to topic %s: %w", topic, err)
	}

	log.Debug().
		Str("topic", topic).
		Str("topic_arn", topic).
		Msg("message published to SNS")

	return nil
}
