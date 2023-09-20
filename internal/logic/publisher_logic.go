package logic

import (
	"broker_transaction_outbox/internal/entity"
	"broker_transaction_outbox/pkg"
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type publisherLogic struct {
	storeRepository entity.Store
}

func NewPublisherLogic(storeRepository entity.Store) entity.Publisher {
	return &publisherLogic{
		storeRepository: storeRepository,
	}
}

func (p *publisherLogic) Publish(ctx context.Context, message pkg.Message) error {
	if message.Topic == "" {
		return errors.New("invalid message, topic is empty")
	}
	if len(message.Body) == 0 {
		return errors.New("invalid message, body is empty")
	}
	newID := uuid.New()
	record := entity.Record{
		Uuid:      newID,
		Message:   message,
		State:     entity.PendingDelivery,
		CreatedOn: time.Now().UTC().Unix(),
	}
	return p.storeRepository.AddRecord(ctx, record)
}
