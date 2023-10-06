package client

import (
	"context"
)

type Headers []Header

type Header interface {
	GetKey() string
	GetValue() []byte
}

func (h Headers) GetValidHeaders() Headers {
	validHeaders := make([]Header, 0, len(h))
	for _, header := range h {
		if header == nil {
			continue
		}
		validHeaders = append(validHeaders, header)
	}
	return validHeaders
}

type Publisher interface {
	Publish(context.Context, string, interface{}) error
}
