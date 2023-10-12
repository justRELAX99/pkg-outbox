package client

import (
	"context"
)

type Publisher interface {
	Publish(context.Context, string, interface{}, ...Header) error
}
