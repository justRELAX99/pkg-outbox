package kafka

import (
	"broker_transaction_outbox/pkg"
	"broker_transaction_outbox/pkg/logger"
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"time"
)

const (
	// Максимальное количество реплик каждой партиции (равно количеству брокеров в кластере)
	maxReplicationFactor = 3

	// Значение реплик каждой партиции по умолчанию
	defaultReplicationFactor = 1
	// Значение партиций для топика по умолчанию
	defaultNumPartitions = 3

	// Время ожидания, пока очередь в буфере продусера переполнена
	queueFullWaitTime = time.Second * 5

	reconnectTime = time.Second * 10

	flushTimeout = 5000

	readTimeout = time.Second
)

type (
	MessageHandler func(ctx context.Context, message pkg.Message) error
	MiddlewareFunc func(next MessageHandler) MessageHandler
)

type client struct {
	serviceName    string
	topicPrefix    string
	consumers      consumers
	subscriptions  subscriptions
	producer       producer
	producerConfig kafka.ConfigMap
	consumerConfig kafka.ConfigMap
}

func NewBrokerClient(
	producerConfig kafka.ConfigMap,
	consumerConfig kafka.ConfigMap,
	serviceName string,
) BrokerClient {
	return &client{
		subscriptions:  newSubscriptions(),
		serviceName:    serviceName,
		producerConfig: producerConfig,
		consumerConfig: consumerConfig,
	}
}

func Start(client BrokerClient) {
	log := logger.GetLogger()
	log.Info("START CONNECTING TO KAFKA")
	err := client.Start()
	if err != nil {
		log.Fatal(err, "can't start kafka client")
	}
}

func (c *client) Start() (err error) {
	err = c.producer.initProducer(c.producerConfig)
	if err != nil {
		return
	}
	if len(c.subscriptions.subscriptions) != 0 {
		err = c.subscriptions.createTopics(c.producer)
		if err != nil {
			return
		}
	}
	c.consumers, err = c.subscriptions.initConsumers(c.consumerConfig, c.serviceName, c.reconnect)
	if err != nil {
		return
	}
	return
}

func (c *client) Pre(mw ...interface{}) {
	for _, v := range mw {
		c.subscriptions.mwFuncs = append(c.subscriptions.mwFuncs, v.(MiddlewareFunc))
	}
}

func (c *client) StopSubscribe() {
	c.consumers.stopConsumers()
}

func (c *client) StopProduce() {
	c.producer.stop()
}

func (c *client) Publish(ctx context.Context, message pkg.Message) (err error) {
	return c.publish(ctx, message)
}

func (c *client) Subscribe(topic string, h Handler, goroutines int, spec *TopicSpecifications, checkError ...bool) {
	s := subscription{
		topic:      c.topicPrefix + topic,
		handler:    h,
		spec:       spec,
		goroutines: goroutines,
	}
	if len(checkError) != 0 {
		s.checkError = checkError[0]
	}
	c.subscriptions.subscriptions = append(c.subscriptions.subscriptions, s)
}

func (c *client) publish(ctx context.Context, message pkg.Message) (err error) {
	message.Topic = c.topicPrefix + message.Topic
	deliveryChannel := make(chan kafka.Event)

	go c.handleDelivery(ctx, message, deliveryChannel)

	err = c.producer.produce(
		ctx,
		message.ToKafkaMessage(),
		deliveryChannel,
	)
	if err != nil {
		return err
	}
	return
}

func (c *client) handleDelivery(ctx context.Context, message pkg.Message, deliveryChannel chan kafka.Event) {
	log := logger.FromContext(ctx)
	e := <-deliveryChannel
	close(deliveryChannel)
	switch event := e.(type) {
	case *kafka.Message:
		if event.TopicPartition.Error != nil {
			kafkaErr := event.TopicPartition.Error.(kafka.Error)
			// Если retriable, то ошибка временная, нужно пытаться переотправить снова, если нет, то ошибка nonretriable, просто логируем
			if kafkaErr.IsRetriable() {
				log.WithError(kafkaErr).
					Errorf("kafka produce retriable error, try again send topic: %v, message: %v",
						message.Topic, string(message.Body))
				err := c.publish(ctx, message)
				if err != nil {
					log.WithError(err).
						Errorf("Cant publish by kafka, topic: %v, message: %v",
							message.Topic, string(message.Body))
				}
			} else {
				log.WithError(kafkaErr).
					Errorf("kafka produce nonretriable error, can't send topic: %v, message: %v. Is fatal: %v",
						message.Topic, string(message.Body), kafkaErr.IsFatal())
			}
		}
	case kafka.Error:
		// Общие пользовательские ошибки, клиент сам пытается переотправить, просто логируем
		log.WithError(event).
			Errorf("publish error, topic: %v, message: %v. client tries to send again", message.Topic, string(message.Body))
	}
}

func (c *client) reconnect() {
	log := logger.GetLogger()
	log.Debugf("start reconnecting consumers")
	// Стопаем консумеры
	c.StopSubscribe()
	log.Debugf("consumers stopped")
	// Чистим массив остановленных консумеров
	c.consumers.consumers = make([]*consumer, 0)

	// Ждём 10 секунд для реконнекта
	time.Sleep(reconnectTime)

	// Запускаем новые консумеры
	for {
		var err error
		c.consumers, err = c.subscriptions.initConsumers(c.consumerConfig, c.serviceName, c.reconnect)
		if err != nil {
			log.WithError(err).Error("cant init consumers")
			time.Sleep(reconnectTime)
			continue
		}
		log.Debugf("new consumers initiated")
		break
	}
}
