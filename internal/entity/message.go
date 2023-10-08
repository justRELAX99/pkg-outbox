package entity

import "gitlab.enkod.tech/pkg/transactionoutbox/client"

type Message struct {
	Headers client.Headers `json:"headers"`
	Body    interface{}    `json:"body"`
	Topic   string         `json:"topic"`
}

func NewMessage(topic string, body interface{}, headers client.Headers) Message {
	return Message{
		Topic:   topic,
		Body:    body,
		Headers: headers.GetValidHeaders(),
	}
}
