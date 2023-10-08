package client

import (
	"context"
	kafka "gitlab.enkod.tech/pkg/kafka/client"
)

type kafkaAdapter struct {
	kafka kafka.Client
}

func NewKafkaAdapter(kafka kafka.Client) Publisher {
	return &kafkaAdapter{
		kafka: kafka,
	}
}

func (k *kafkaAdapter) Publish(ctx context.Context, topic string, data interface{}, headers ...Header) error {
	kafkaHeaders := make([]kafka.Header, len(headers))
	for i, header := range headers {
		kafkaHeaders[i] = header
	}
	return k.kafka.Publish(ctx, topic, data, kafkaHeaders...)
}
