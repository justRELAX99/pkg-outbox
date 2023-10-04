package entity

import (
	"context"
)

type Headers []Header

type Header interface {
	GetKey() string
	GetValue() []byte
}

type Publisher interface {
	Publish(context.Context, string, interface{}, ...Header) error
}
