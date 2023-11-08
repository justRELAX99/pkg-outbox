package client

type CustomMessage interface {
	GetBody() []byte
	Headers
}

type Message struct {
	Headers MessageHeaders `json:"headers"`
	Body    interface{}    `json:"body"`
	Topic   string         `json:"topic"`
}

func (m Message) GetBody() []byte {
	//TODO implement me
	panic("implement me")
}

func (m *Message) SetHeader(key string, value []byte) {
	//TODO implement me
	panic("implement me")
}

func (m Message) GetValueByKey(key string) []byte {
	//TODO implement me
	panic("implement me")
}

func NewMessage(topic string, body interface{}, headers ...map[string][]byte) Message {
	return Message{
		Topic:   topic,
		Body:    body,
		Headers: NewMessageHeaders(headers...),
	}
}
