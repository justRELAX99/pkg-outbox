package broker_transaction_outbox

import (
	"broker_transaction_outbox/internal/entity"
	"broker_transaction_outbox/internal/logic"
	"broker_transaction_outbox/internal/repository"
	"broker_transaction_outbox/migration/app"
)

func NewOutbox(pgClient entity.Client, tx entity.Transactor, publisher entity.Publisher, serviceName string) (entity.RecordLogic, entity.Publisher) {
	app.Run(pgClient.GetSqlDB(), serviceName, nil)
	var (
		storeRepository = repository.NewStoreRepository(pgClient)
	)

	var (
		recordLogic    = logic.NewRecordsLogic(storeRepository, tx, publisher)
		publisherLogic = logic.NewPublisherLogic(storeRepository, serviceName)
	)
	return recordLogic, publisherLogic
}
