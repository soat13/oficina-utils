package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go-v2/aws"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/soat13/oficina-utils/pkg/awsconfig"
	"github.com/soat13/oficina-utils/pkg/messaging"
)

type (
	Broker struct {
		client    *sqs.Client
		baseURL   string
		consumers []*consumer
		cancel    context.CancelFunc
		wg        sync.WaitGroup
		syncMode  bool
	}

	consumer struct {
		client   *sqs.Client
		queueURL string
		topic    string
		handler  messaging.Handler
	}

	snsEnvelope struct {
		Type    string `json:"Type"`
		Message string `json:"Message"`
	}
)

func NewSyncBroker() messaging.QueueBroker {
	return &Broker{syncMode: true}
}

func NewBroker(ctx context.Context, cfg awsconfig.Config, sqsBaseURL string) (messaging.QueueBroker, error) {
	awsCfg, err := awsconfig.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	awstrace.AppendMiddleware(&awsCfg)

	client := sqs.NewFromConfig(awsCfg, func(o *sqs.Options) {
		if cfg.EndpointURL != "" {
			o.BaseEndpoint = aws.String(cfg.EndpointURL)
		}
	})

	baseURL := strings.TrimSuffix(sqsBaseURL, "/")

	return &Broker{
		client:  client,
		baseURL: baseURL,
	}, nil
}

func (b *Broker) Send(ctx context.Context, message messaging.QueueMessage) error {
	payload, err := json.Marshal(message.Payload)
	if err != nil {
		return err
	}

	if b.syncMode {
		for _, c := range b.consumers {
			if c.topic == message.EventName {
				if err := c.handler(ctx, messaging.Message{Payload: payload}); err != nil {
					log.Error().Err(err).Str("topic", message.EventName).Msg("sync handler failed")
					return err
				}
			}
		}
		return nil
	}

	if b.client == nil {
		return nil
	}

	queueURL := b.queueURL(message.EventName)

	input := &sqs.SendMessageInput{
		QueueUrl:       aws.String(queueURL),
		MessageBody:    aws.String(string(payload)),
		MessageGroupId: message.GroupID,
	}

	_, err = b.client.SendMessage(ctx, input)
	if err != nil {
		log.Error().
			Err(err).
			Str("topic", message.EventName).
			Str("queue", queueURL).
			Msg("failed to publish message to SQS")

		return fmt.Errorf("sqs publish to %s: %w", message.EventName, err)
	}

	log.Debug().
		Str("topic", message.EventName).
		Str("queue", queueURL).
		Msg("message published to SQS")

	return nil
}

func (b *Broker) Subscribe(topic string, handler messaging.Handler) {
	b.consumers = append(b.consumers, &consumer{
		client:   b.client,
		queueURL: b.queueURL(topic),
		topic:    topic,
		handler:  handler,
	})
}

func (b *Broker) Listen(ctx context.Context) {
	if b.syncMode {
		return
	}

	ctx, b.cancel = context.WithCancel(ctx)

	for _, c := range b.consumers {
		b.wg.Add(1)
		go func(c *consumer) {
			defer b.wg.Done()
			c.poll(ctx)
		}(c)
	}

	log.Info().Int("count", len(b.consumers)).Msg("all SQS consumers started")
}

func (b *Broker) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
	b.wg.Wait()
	log.Info().Msg("all SQS consumers stopped")
}

func (b *Broker) queueURL(topic string) string {
	return fmt.Sprintf("%s/%s", b.baseURL, normalizeQueueName(topic))
}

func (c *consumer) poll(ctx context.Context) {
	logger := log.With().Str("queue", c.queueURL).Str("topic", c.topic).Logger()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("SQS consumer stopped")
			return
		default:
		}

		messages, err := c.receiveMessages(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Error().Err(err).Msg("failed to receive messages from SQS")
			continue
		}

		for _, msg := range messages {
			c.processMessage(ctx, logger, msg)
		}
	}
}

func (c *consumer) receiveMessages(ctx context.Context) ([]types.Message, error) {
	result, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(c.queueURL),
		MaxNumberOfMessages:   10,
		WaitTimeSeconds:       20,
		MessageAttributeNames: []string{"_datadog"},
	})
	if err != nil {
		return nil, err
	}

	return result.Messages, nil
}

func (c *consumer) processMessage(ctx context.Context, logger zerolog.Logger, msg types.Message) {
	var spanOpts []ddtrace.StartSpanOption
	if ddAttr, ok := msg.MessageAttributes["_datadog"]; ok && ddAttr.StringValue != nil {
		var headers map[string]string
		if json.Unmarshal([]byte(*ddAttr.StringValue), &headers) == nil {
			if parentCtx, err := tracer.Extract(tracer.TextMapCarrier(headers)); err == nil {
				spanOpts = append(spanOpts, tracer.ChildOf(parentCtx))
			}
		}
	}
	span := tracer.StartSpan("sqs.process",
		append(spanOpts,
			tracer.ResourceName(c.topic),
			tracer.Tag("messaging.system", "sqs"),
			tracer.Tag("messaging.destination", c.topic),
			tracer.Tag("messaging.message_id", aws.ToString(msg.MessageId)),
		)...,
	)
	ctx = tracer.ContextWithSpan(ctx, span)

	payload := []byte(aws.ToString(msg.Body))

	unwrappedPayload, err := unwrapSNSMessage(payload)
	if err != nil {
		span.Finish(tracer.WithError(err))
		logger.Error().
			Err(err).
			Str("message_id", aws.ToString(msg.MessageId)).
			Msg("failed to unwrap message")
		return
	}

	message := messaging.Message{
		Payload: unwrappedPayload,
	}

	if err := c.handler(ctx, message); err != nil {
		span.Finish(tracer.WithError(err))
		logger.Error().
			Err(err).
			Str("message_id", aws.ToString(msg.MessageId)).
			Msg("failed to handle message")
		return
	}
	span.Finish()

	_, err = c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
	if err != nil {
		logger.Error().
			Err(err).
			Str("message_id", aws.ToString(msg.MessageId)).
			Msg("failed to delete message from SQS")
	}
}

func unwrapSNSMessage(body []byte) ([]byte, error) {
	var env snsEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return body, nil
	}

	if env.Type == "Notification" && env.Message != "" {
		return []byte(env.Message), nil
	}

	return body, nil
}

func normalizeQueueName(topic string) string {
	const fifoSuffix = ".fifo"

	isFifo := strings.HasSuffix(topic, fifoSuffix)
	base := strings.TrimSuffix(topic, fifoSuffix)
	name := strings.ToLower(strings.ReplaceAll(base, ".", "-"))

	if isFifo {
		name += fifoSuffix
	}
	return name
}
