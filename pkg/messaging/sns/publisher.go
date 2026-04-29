package sns

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/rs/zerolog/log"
	"github.com/soat13/oficina-utils/pkg/awsconfig"
	"github.com/soat13/oficina-utils/pkg/messaging"
)

type Publisher struct {
	client  *sns.Client
	baseARN string
}

func NewPublisher(ctx context.Context, cfg awsconfig.Config, baseARN string) (messaging.TopicPublisher, error) {
	awsCfg, err := awsconfig.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := sns.NewFromConfig(awsCfg, func(o *sns.Options) {
		if cfg.EndpointURL != "" {
			o.BaseEndpoint = aws.String(cfg.EndpointURL)
		}
	})

	return &Publisher{
		client:  client,
		baseARN: strings.TrimSuffix(baseARN, ":"),
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, message messaging.TopicMessage) error {
	payload, err := json.Marshal(message.Payload)
	if err != nil {
		return err
	}

	topicARN := p.topicARN(message.EventName)

	_, err = p.client.Publish(ctx, &sns.PublishInput{
		TopicArn:       aws.String(topicARN),
		Message:        aws.String(string(payload)),
		MessageGroupId: message.GroupID,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("topic", message.EventName).
			Str("topic_arn", topicARN).
			Msg("failed to publish message to SNS")

		return fmt.Errorf("sns publish to topic %s: %w", message.EventName, err)
	}

	log.Debug().
		Str("topic", message.EventName).
		Str("topic_arn", topicARN).
		Msg("message published to SNS")

	return nil
}

func (p *Publisher) topicARN(topic string) string {
	return fmt.Sprintf("%s:%s", p.baseARN, topic)
}
