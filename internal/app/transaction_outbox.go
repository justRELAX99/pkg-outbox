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
		transactionOutbox, publisher = client.NewOutbox(pgClient, transactor, k, serviceName, nil)
	)
	testConsumer(testTopic, k)
	testProducer(testTopic, publisher)
	k.Start()

	transactionOutbox.StartProcessRecords()
	time.Sleep(time.Second * 10)
	transactionOutbox.StopProcessRecords()
	select {}
}

func testConsumer(topic string, k kafkaClient.Client) {
	k.Subscribe(testHandler, 1, &kafkaClient.TopicSpecifications{
		NumPartitions:     1,
		ReplicationFactor: 1,
		Topic:             topic,
	})
}

func testProducer(topic string, k client.GivenPublisher) {
	err := k.Publish(context.Background(), topic, "test", map[string][]byte{
		"test": []byte("test"),
	})
	if err != nil {
		logger.GetLogger().WithError(err).Error("produce err")
	}
}

func testHandler(ctx context.Context, msg []byte) error {
	logger.GetLogger().Info(string(msg))
	return nil
}
