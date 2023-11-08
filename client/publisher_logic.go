package client

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.enkod.tech/pkg/transactionoutbox/internal/entity"
	"time"
)

type publisherLogic struct {
	serviceName     string
	storeRepository Store
	prePublish      []Pre
}

func newPublisherLogic(storeRepository Store, serviceName string) GivenPublisher {
	return &publisherLogic{
		storeRepository: storeRepository,
		serviceName:     serviceName,
	}
}

func (p *publisherLogic) validateMessage(message Message) error {
	if message.Topic == "" {
		return errors.New("invalid message, topic is empty")
	}
	if message.Body == nil {
		return errors.New("invalid message, body is empty")
	}
	return nil
}

func (p *publisherLogic) Publish(ctx context.Context, topic string, data interface{}, headers ...map[string][]byte) error {
	message := NewMessage(topic, data, headers...)
	err := p.validateMessage(message)
	if err != nil {
		return err
	}
	for _, pre := range p.prePublish {
		pre(ctx, &message)
	}
	record := Record{
		ServiceName: p.serviceName,
		Uuid:        uuid.New(),
		Message:     message,
		State:       PendingDelivery,
		CreatedOn:   time.Now().UTC().Unix(),
	}
	return p.storeRepository.AddRecord(ctx, record)
}

func (p *publisherLogic) PrePublish(pre Pre) {
	p.prePublish = append(p.prePublish, pre)
}
