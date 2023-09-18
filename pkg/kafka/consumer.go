package kafka

import (
	"broker_transaction_outbox/pkg"
	"broker_transaction_outbox/pkg/logger"
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"sync"
)

type consumers struct {
	consumers []*consumer
	mwFuncs   []MiddlewareFunc
}

type consumer struct {
	topic      string
	checkError bool
	handler    Handler
	*kafka.Consumer

	gCtx   context.Context
	cancel context.CancelFunc

	wg *sync.WaitGroup
}

func newConsumers(size int) consumers {
	return consumers{
		consumers: make([]*consumer, 0, size),
		mwFuncs:   make([]MiddlewareFunc, 0),
	}
}

func newConsumer(
	topic string,
	checkError bool,
	handler Handler,
	kafkaConsumer *kafka.Consumer,
) *consumer {
	gCtx, cancel := context.WithCancel(context.Background())
	return &consumer{
		gCtx:       gCtx,
		cancel:     cancel,
		wg:         new(sync.WaitGroup),
		topic:      topic,
		checkError: checkError,
		handler:    handler,
		Consumer:   kafkaConsumer,
	}
}

func newConsumersBySubscriptions(config kafka.ConfigMap, subscriptions subscriptions, serviceName string) (consumers consumers, err error) {
	consumers = newConsumers(len(subscriptions))
	config["group.id"] = serviceName

	for i := 0; i < len(subscriptions); i++ {
		for j := 0; j < subscriptions[i].goroutines; j++ {
			newConsumer, err := newConsumerBySubscription(config, subscriptions[i])
			if err != nil {
				return consumers, err
			}
			consumers.consumers = append(consumers.consumers, newConsumer)
		}
	}
	return consumers, nil
}

func newConsumerBySubscription(config kafka.ConfigMap, subscription subscription) (*consumer, error) {
	config["client.id"] = uuid.New().String()
	// Создаём консумера
	kafkaConsumer, err := kafka.NewConsumer(&config)
	if err != nil {
		return &consumer{}, errors.Wrap(err, "cant create kafka consumer")
	}
	// Подписываем консумера на топик
	err = kafkaConsumer.Subscribe(subscription.topic, nil)
	if err != nil {
		return &consumer{}, errors.Wrap(err, "cant subscribe kafka consumer")
	}
	return newConsumer(subscription.topic, subscription.checkError, subscription.handler, kafkaConsumer), nil
}

func (c *consumer) startConsume(consumer consumer) error {
	c.wg.Add(1)
	defer c.wg.Done()
	log := logger.GetLogger()
	// Прогоняем хендлер через миддлверы
	var handler MessageHandler = func(ctx context.Context, message pkg.Message) error {
		return consumer.handler(ctx, message.Body)
	}
	for j := len(c.mwFuncs) - 1; j >= 0; j-- {
		handler = c.mwFuncs[j](handler)
	}
	for {
		select {
		case <-c.gCtx.Done():
			return nil
		default:
			msg, err := consumer.ReadMessage(readTimeout)
			if err != nil {
				if kafkaErr, ok := err.(kafka.Error); ok {
					// Если retriable (но со стороны консумера вроде бы такого нет), то пробуем снова
					if kafkaErr.IsRetriable() || kafkaErr.Code() == kafka.ErrTimedOut {
						continue
					}
				}
				return errors.Wrap(err, "cant read kafka message")
			}
			err = handler(context.Background(), pkg.NewByKafkaMessage(msg))
			if err != nil && consumer.checkError {
				log.WithError(err).Debug("try to read message again")
				consumer.rollbackConsumerTransaction(msg.TopicPartition)
			}
		}
	}
}

func (c *consumers) stopConsumers() {
	log := logger.GetLogger()
	c.cancel()
	c.wg.Wait()
	for i := range c.consumers {
		_, err := c.consumers[i].Commit()
		if err != nil {
			log.WithError(err).Errorf("cant commit offset for topic: %s", err.Error())
		}
		// Отписка от назначенных топиков
		err = c.consumers[i].Unsubscribe()
		if err != nil {
			log.WithError(err).Errorf("cant unsubscribe connection: %s", err.Error())
		}
		// Закрытие соединения
		err = c.consumers[i].Close()
		if err != nil {
			log.WithError(err).Errorf("cant close consumer connection: %s", err.Error())
		}
	}
}

func (c *consumer) rollbackConsumerTransaction(topicPartition kafka.TopicPartition) {
	// В committed лежит массив из одного элемента, потому что передаём одну партицию, которую нужно сбросить
	committed, err := c.Committed([]kafka.TopicPartition{{Topic: &c.topic, Partition: topicPartition.Partition}}, -1)
	log := logger.GetLogger()
	if err != nil {
		log.Error(err)
		return
	}
	if committed[0].Offset < 0 {
		committed[0].Offset = kafka.OffsetBeginning
	} else {
		committed[0].Offset = topicPartition.Offset
	}
	err = c.Seek(committed[0], 0)
	if err != nil {
		log.Error(err)
		return
	}
	return
}
