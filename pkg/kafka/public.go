package kafka

import (
	"broker_transaction_outbox/pkg"
	"context"
)

type Handler func(ctx context.Context, message []byte) error

type TopicSpecifications struct {
	NumPartitions     int
	ReplicationFactor int
}

type BrokerClient interface {
	Start() error
	Pre(mw ...interface{})
	StopSubscribe()
	StopProduce()
	Publish(context.Context, pkg.Message) error
	Subscribe(string, Handler, int, *TopicSpecifications, ...bool)
}
