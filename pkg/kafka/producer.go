package kafka

import (
	"broker_transaction_outbox/pkg/logger"
	"context"
	"github.com/CossackPyra/pyraconv"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type producer struct {
	kafkaProducer *kafka.Producer
}

func (p *producer) initProducer(config kafka.ConfigMap) (err error) {
	log := logger.GetLogger()
	config["client.id"] = uuid.New().String()

	// FIXME Два костыля, нужно подумать, что делать с тем, что с консула числа маршлятся во float64
	config["queue.buffering.max.messages"] = int(pyraconv.ToInt64(config["queue.buffering.max.messages"]))
	config["linger.ms"] = int(pyraconv.ToInt64(config["linger.ms"]))

	p.kafkaProducer, err = kafka.NewProducer(&config)
	if err != nil {
		return errors.Wrap(err, "cant create kafka producer")
	}

	log.Info("KAFKA PRODUCER IS READY")
	return nil
}

func (p *producer) stop() {
	p.kafkaProducer.Flush(flushTimeout)
	p.kafkaProducer.Close()
}

func (p *producer) produce(ctx context.Context, message *kafka.Message, deliveryChannel chan kafka.Event) error {
	log := logger.FromContext(ctx)
	for {
		err := p.kafkaProducer.Produce(message, deliveryChannel)
		if err != nil {
			if err.(kafka.Error).Code() == kafka.ErrQueueFull {
				// Если очередь забита, пробуем отправить снова через 5 секунд
				log.WithError(err).
					Warnf("kafka queue full, try again after %v second", queueFullWaitTime.Seconds())
				time.Sleep(queueFullWaitTime)
				continue
			} else {
				return err
			}
		}
		break
	}
	return nil
}
