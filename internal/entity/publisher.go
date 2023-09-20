package entity

import (
	"broker_transaction_outbox/pkg"
	"context"
)

type Publisher interface {
	Publish(context.Context, pkg.Message) error
}
