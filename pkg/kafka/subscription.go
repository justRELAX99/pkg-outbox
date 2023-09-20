package kafka

import (
	"broker_transaction_outbox/pkg/logger"
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"sync"
)

type subscriptions struct {
	mwFuncs       []MiddlewareFunc
	subscriptions []subscription
}

func newSubscriptions() subscriptions {
	return subscriptions{
		mwFuncs:       make([]MiddlewareFunc, 0),
		subscriptions: make([]subscription, 0),
	}
}

type subscription struct {
	topic      string
	checkError bool
	goroutines int
	spec       *TopicSpecifications
	handler    Handler
}

func (s *subscriptions) initConsumers(config kafka.ConfigMap, serviceName string, reconnectFunc func()) (consumers consumers, err error) {
	once := &sync.Once{}
	consumers, err = newConsumersBySubscriptions(config, *s, serviceName)
	if err != nil {
		return consumers, err
	}

	// Запускаем каждого консумера в отдельной горутине
	for _, c := range consumers.consumers {
		consumers.wg.Add(1)
		go func(consumer *consumer, done <-chan struct{}) {
			err := consumer.startConsume(done, s.mwFuncs)
			consumers.wg.Done()
			if err != nil {
				once.Do(func() {
					reconnectFunc()
				})
			}
		}(c, consumers.done)
	}

	logger.GetLogger().Info("KAFKA CONSUMERS IS READY")
	return consumers, nil
}

func (s *subscriptions) createTopics(producer producer) (err error) {
	// Создаём админский клиент через настройки подключения продусера
	adminClient, err := kafka.NewAdminClientFromProducer(producer.kafkaProducer)
	if err != nil {
		return errors.Wrap(err, "cant init kafka admin client")
	}
	defer adminClient.Close()
	log := logger.GetLogger()
	specifications := make([]kafka.TopicSpecification, 0, len(s.subscriptions))
	for _, subscriber := range s.subscriptions {
		specification := kafka.TopicSpecification{
			Topic:             subscriber.topic,
			ReplicationFactor: defaultReplicationFactor,
			NumPartitions:     defaultNumPartitions,
		}
		// Если нет настроек топика, то при создании будут подставляться дефолтные
		if subscriber.spec != nil {
			// Костылик на проверку, чтобы фактор репликации не был больше числа брокеров
			if subscriber.spec.ReplicationFactor > maxReplicationFactor {
				subscriber.spec.ReplicationFactor = maxReplicationFactor
				log.Warnf("Number of replicas cannot be more than %v, set the maximum value", maxReplicationFactor)
			}
			specification.NumPartitions = subscriber.spec.NumPartitions
			specification.ReplicationFactor = subscriber.spec.ReplicationFactor
		}
		specifications = append(specifications, specification)
	}
	result, err := adminClient.CreateTopics(context.Background(), specifications)
	if err != nil {
		return errors.Wrapf(err, "%v: cant create topics", err.Error())
	}
	for _, v := range result {
		// Если такой топик уже есть, то будет ошибка внутри структуры, если ошибки нет, то в структуре будет "Success"
		log.Infof("%v: %v", v.Topic, v.Error.String())
	}
	return nil
}
