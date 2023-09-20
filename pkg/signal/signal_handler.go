package signal

import (
	"broker_transaction_outbox/pkg/kafka"
	"broker_transaction_outbox/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const stopTimeout = time.Minute * 1

func NewSignalHandler(b kafka.BrokerClient) {
	var (
		c        = make(chan os.Signal, 1)
		stopChan = make(chan bool, 1)
		log      = logger.FromContext(nil)
	)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stopChan

		// Через stopTimeout сервис завершит работу
		go func() {
			time.Sleep(stopTimeout)
			log.Warningf("Service done: close producers and exit after %v minute", stopTimeout.Minutes())

			b.StopProduce()

			os.Exit(1)
		}()

		if b != nil {
			log.Warning("STOP KAFKA client")
			b.StopSubscribe()
		}

		// Если все текущие процессы успели завершиться раньше - завершаем
		go func() {
			log.Warning("Service done: close producers and exit after waiting group")
			b.StopProduce()
			os.Exit(0)
		}()
	}()
	go func() {
		sig := <-c
		log.Warnf("STOP TRIGGER: %s signal!!!", sig.String())
		stopChan <- true
	}()
}
