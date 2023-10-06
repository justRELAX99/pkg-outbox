package broker_transaction_outbox

import (
	"github.com/sirupsen/logrus"
	"gitlab.enkod.tech/pkg/transactionoutbox/internal/entity"
	"gitlab.enkod.tech/pkg/transactionoutbox/internal/logic"
	"gitlab.enkod.tech/pkg/transactionoutbox/internal/repository"
	"gitlab.enkod.tech/pkg/transactionoutbox/migration/app"
	"gitlab.enkod.tech/pkg/transactionoutbox/pkg/logger"
)

func NewOutbox(
	pgClient entity.Client,
	tx entity.Transactor,
	publisher entity.Publisher,
	serviceName string,
	log *logrus.Logger,
) (entity.RecordLogic, entity.Publisher) {
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
