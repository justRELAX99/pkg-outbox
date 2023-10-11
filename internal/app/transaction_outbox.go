package app

import (
	"context"
	kafkaClient "gitlab.enkod.tech/pkg/kafka/client"
	postgres "gitlab.enkod.tech/pkg/postgres/client"
	"gitlab.enkod.tech/pkg/transactionoutbox/client"
	configEntity "gitlab.enkod.tech/pkg/transactionoutbox/pkg/config/entity"
	"gitlab.enkod.tech/pkg/transactionoutbox/pkg/logger"
	"time"
)

func Run(configSettings configEntity.Settings, serviceName string) {
	logger.SetDefaultLogger("debug")
	const (
		testTopic = "test_topic"
	)
	var (
		pgClient, transactor = postgres.NewClient(configSettings.PostgresConfigs, serviceName, nil)
		k                    = kafkaClient.NewClient(
			configSettings.KafkaProducer,
			configSettings.KafkaConsumer,
			serviceName,
			nil,
			"local_")
		transactionOutboxAdapter     = client.NewKafkaAdapter(k)
		transactionOutbox, publisher = client.NewOutbox(pgClient, transactor, transactionOutboxAdapter, serviceName, nil)
	)
	transactionOutbox.StartProcessRecords(1)
	testConsumer(testTopic, k)
	testProducer(testTopic, publisher)
	time.Sleep(time.Second * 10)
	transactionOutbox.StopProcessRecords()
}

func testConsumer(topic string, k kafkaClient.Client) {
	k.Subscribe(testHandler, 1, &kafkaClient.TopicSpecifications{
		NumPartitions:     1,
		ReplicationFactor: 1,
		Topic:             topic,
	})
}

func testProducer(topic string, k client.Publisher) {
	err := k.Publish(context.Background(), topic, "test", kafkaClient.MessageHeader{
		Key:   "test",
		Value: []byte("test"),
	})
	if err != nil {
		logger.GetLogger().WithError(err).Error("produce err")
	}
}

func testHandler(ctx context.Context, msg []byte) error {
	logger.GetLogger().Info(string(msg))
	return nil
}
