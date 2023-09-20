package app

import (
	"broker_transaction_outbox/internal/logic"
	"broker_transaction_outbox/internal/repository"
	"broker_transaction_outbox/pkg"
	configEntity "broker_transaction_outbox/pkg/config/entity"
	"broker_transaction_outbox/pkg/kafka"
	"broker_transaction_outbox/pkg/logger"
	"broker_transaction_outbox/pkg/postgres"
	"broker_transaction_outbox/pkg/signal"
	"context"
	"fmt"
)

func Run(configSettings configEntity.Settings, serviceName string) {
	logger.SetDefaultLogger("")
	//broker clients
	var (
		k = kafka.NewBrokerClient(configSettings.KafkaProducer, configSettings.KafkaConsumer, serviceName)
	)
	kafka.Start(k)

	//db clients
	var (
		pg, transactor = postgres.NewClient(configSettings.PostgresConfigs, serviceName)
	)

	//db repositories
	var (
		storeRepository = repository.NewStoreRepository(pg)
	)

	//logic
	var (
		_            = logic.NewPublisherLogic(storeRepository)
		recordsLogic = logic.NewRecordsLogic(storeRepository, transactor, k)
	)

	testConsumer("test_topic", k)
	recordsLogic.StartProcessRecords(1)
	signal.NewSignalHandler(k)
	select {}
}

func testConsumer(topic string, k kafka.BrokerClient) {
	k.Subscribe(topic, testHandler, 1, &kafka.TopicSpecifications{
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
}

func testProducer(topic string, k kafka.BrokerClient) {
	err := k.Publish(context.Background(), pkg.Message{
		Topic: topic,
		Body:  []byte("test"),
	})
	if err != nil {
		fmt.Printf("produce err: %s\n", err.Error())
	}
}

func testHandler(ctx context.Context, msg []byte) error {
	fmt.Printf(string(msg))
	return nil
}
