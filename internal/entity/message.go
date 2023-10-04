package entity

type Message struct {
	Headers Headers     `json:"headers"`
	Body    interface{} `json:"body"`
	Topic   string      `json:"topic"`
}

func NewMessage(topic string, body interface{}, headers Headers) Message {
	return Message{
		Topic:   topic,
		Body:    body,
		Headers: headers.GetValidHeaders(),
	}
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
