package client

type Message struct {
	Headers MessageHeaders `json:"headers"`
	Body    interface{}    `json:"body"`
	Topic   string         `json:"topic"`
}

func NewMessage(topic string, body interface{}, headers Headers) Message {
	return Message{
		Topic:   topic,
		Body:    body,
		Headers: NewMessageHeaders(headers),
	}
}
