package broker_transaction_outbox

import (
	"broker_transaction_outbox/internal/entity"
	"broker_transaction_outbox/internal/logic"
	"broker_transaction_outbox/internal/repository"
	"broker_transaction_outbox/migration/app"
	"broker_transaction_outbox/pkg/logger"
	"github.com/sirupsen/logrus"
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
