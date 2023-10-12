package client

import (
	"context"
)

type Pre func(ctx context.Context, message *Message)

type Publisher interface {
	PrePublish(Pre)
	Publish(context.Context, string, interface{}, ...Header) error
}
