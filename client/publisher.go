package client

import (
	"context"
)

type (
	MessageHandler func(ctx context.Context, message CustomMessage) error
	MiddlewareFunc func(next MessageHandler) MessageHandler
)

type Publisher interface {
	Pre(mw ...MiddlewareFunc)
	Publish(context.Context, string, interface{}, ...Header) error
}
