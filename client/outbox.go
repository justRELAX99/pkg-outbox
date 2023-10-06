package client

import (
	"github.com/sirupsen/logrus"
	"gitlab.enkod.tech/pkg/transactionoutbox/internal/logic"
	"gitlab.enkod.tech/pkg/transactionoutbox/internal/repository"
	"gitlab.enkod.tech/pkg/transactionoutbox/migration/app"
	"gitlab.enkod.tech/pkg/transactionoutbox/pkg/logger"
)

func NewOutbox(
	pgClient Client,
	tx Transactor,
	publisher Publisher,
	serviceName string,
	log *logrus.Logger,
) (RecordLogic, Publisher) {
	if log != nil {
		logger.SetLogger(log)
	} else {
		logger.SetDefaultLogger("debug")
	}

	app.Run(pgClient.GetSqlDB(), serviceName, map[string]string{
		"c": "up",
	})

	var (
		storeRepository = repository.NewStoreRepository(pgClient)
	)

	var (
		recordLogic    = logic.NewRecordsLogic(storeRepository, tx, publisher)
		publisherLogic = logic.NewPublisherLogic(storeRepository, serviceName)
	)
	return recordLogic, publisherLogic
}
