package client

import (
	"github.com/enkodio/pkg-outbox/migration/app"
	"github.com/enkodio/pkg-outbox/pkg/logger"
	"github.com/sirupsen/logrus"
)

func NewOutbox(
	pgClient RepositoryClient,
	tx Transactor,
	publisher ReceivedPublisher,
	serviceName string,
	log *logrus.Logger,
) (RecordLogic, GivenPublisher) {
	if log != nil {
		logger.SetLogger(log)
	} else {
		logger.SetDefaultLogger("debug")
	}

	app.Run(pgClient.GetSqlDB(), serviceName, map[string]string{
		"c":   "up",
		"dir": "/migration/migrations",
	})

	var (
		storeRepository = newStoreRepository(pgClient)
	)

	var (
		recordLogic    = newRecordsLogic(storeRepository, tx, publisher)
		publisherLogic = newPublisherLogic(storeRepository, serviceName)
	)
	return recordLogic, publisherLogic
}
