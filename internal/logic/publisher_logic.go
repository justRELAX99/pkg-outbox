package logic

import (
	"broker_transaction_outbox/internal/entity"
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type publisherLogic struct {
	serviceName     string
	storeRepository entity.Store
}

func NewPublisherLogic(storeRepository entity.Store, serviceName string) entity.Publisher {
	return &publisherLogic{
		storeRepository: storeRepository,
		serviceName:     serviceName,
	}
}

func (p *publisherLogic) validateMessage(message entity.Message) error {
	if message.Topic == "" {
		return errors.New("invalid message, topic is empty")
	}
	if message.Body == nil {
		return errors.New("invalid message, body is empty")
	}
	return nil
}

func (p *publisherLogic) Publish(ctx context.Context, topic string, data interface{}, headers ...entity.Header) error {
	message := entity.NewMessage(topic, data, headers)
	err := p.validateMessage(message)
	if err != nil {
		return err
	}
	record := entity.Record{
		ServiceName: p.serviceName,
		Uuid:        uuid.New(),
		Message:     message,
		State:       entity.PendingDelivery,
		CreatedOn:   time.Now().UTC().Unix(),
	}
	return p.storeRepository.AddRecord(ctx, record)
}
