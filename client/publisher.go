package client

import (
	"context"
)

type Pre func(ctx context.Context, message CustomMessage)

type ReceivedPublisher interface {
	Publish(context.Context, string, interface{}, ...map[string][]byte) error
}

type GivenPublisher interface {
	PrePublish(Pre)
	Publish(context.Context, string, interface{}, ...map[string][]byte) error
}
