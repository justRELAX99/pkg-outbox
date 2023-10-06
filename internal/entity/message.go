package entity

type Message struct {
	Body  interface{} `json:"body"`
	Topic string      `json:"topic"`
}

func NewMessage(topic string, body interface{}) Message {
	return Message{
		Topic: topic,
		Body:  body,
	}
}
