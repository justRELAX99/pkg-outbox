package entity

import (
	"broker_transaction_outbox/pkg/postgres"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Settings struct {
	PostgresConfigs postgres.Config `json:"postgres"`
	RedisConfig     RedisConfig     `json:"redis"`
	KafkaProducer   kafka.ConfigMap `json:"kafkaProducer"`
	KafkaConsumer   kafka.ConfigMap `json:"kafkaConsumer"`
}
