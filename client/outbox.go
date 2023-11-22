package client

import (
	"github.com/enkodio/pkg-outbox/internal/migration/app"
	logic2 "github.com/enkodio/pkg-outbox/internal/outbox/logic"
	"github.com/enkodio/pkg-outbox/internal/outbox/repository"
	"github.com/enkodio/pkg-outbox/internal/pkg/logger"
	"github.com/enkodio/pkg-outbox/outbox"
	"github.com/sirupsen/logrus"
)

func NewOutbox(
	pgClient outbox.RepositoryClient,
	tx outbox.Transactor,
	publisher outbox.Publisher,
	serviceName string,
	log *logrus.Logger,
) (outbox.RecordLogic, outbox.OutboxPublisher) {
	if log != nil {
		logger.SetLogger(log)
	} else {
		logger.SetDefaultLogger("debug")
	}

	app.Run(pgClient.GetSqlDB(), map[string]string{
		"c":   "up",
		"dir": "/migration/migrations",
	})

	var (
		storeRepository = repository.NewStoreRepository(pgClient)
	)

	var (
		recordLogic    = logic2.NewRecordsLogic(storeRepository, tx, publisher)
		publisherLogic = logic2.NewPublisherLogic(storeRepository, serviceName)
	)
	return recordLogic, publisherLogic
}
