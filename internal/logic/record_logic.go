package logic

import (
	"broker_transaction_outbox/internal/entity"
	"broker_transaction_outbox/pkg/logger"
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	defaultLimit                   = 1
	defaultProcessRecordsSleepTime = time.Second
)

type recordsLogic struct {
	storeRepository entity.Store
	transactor      entity.Transactor
	broker          entity.Publisher

	syncGroup *entity.SyncGroup
}

func NewRecordsLogic(
	storeRepository entity.Store,
	transactor entity.Transactor,
	broker entity.Publisher,
) entity.RecordLogic {
	r := &recordsLogic{
		storeRepository: storeRepository,
		transactor:      transactor,
		broker:          broker,
		syncGroup:       entity.NewSyncGroup(),
	}
	return r
}

func (r *recordsLogic) StartProcessRecords(countGoroutines int) {
	r.syncGroup.Add(countGoroutines)
	for i := 0; i < countGoroutines; i++ {
		go r.processRecords()
	}
}

func (r *recordsLogic) StopProcessRecords() {
	r.syncGroup.Close()
}

func (r *recordsLogic) processRecords() {
	defer r.syncGroup.Done()
	ctx := context.Background()
	log := logger.GetLogger()
	for {
		time.Sleep(defaultProcessRecordsSleepTime)
		select {
		case <-r.syncGroup.IsDone():
			return
		default:
			err := r.processRecordsWork(ctx)
			if err != nil {
				log.WithError(err).Error("cant process records")
			}
		}
	}
}

func (r *recordsLogic) processRecordsWork(ctx context.Context) error {
	err := r.transactor.Begin(&ctx)
	if err != nil {
		return errors.Wrap(err, "cant begin tx for process records")
	}
	defer r.transactor.Rollback(&ctx)
	records, err := r.storeRepository.GetPendingRecords(ctx, entity.Filter{
		Limit: defaultLimit,
	})
	if err != nil {
		return errors.Wrap(err, "cant get pending records")
	}
	if len(records) == 0 {
		return nil
	}

	successfulRecords, errorRecords := r.publishRecords(ctx, records)
	//TODO maybe we shouldn’t update the records status to err and stay pending
	if len(errorRecords) > 0 {
		err = r.storeRepository.UpdateRecordsStatus(ctx, errorRecords, entity.DeliveredErr)
		if err != nil {
			log.WithError(err).Error("cant update status for records with error")
		}
	}
	if len(successfulRecords) > 0 {
		err = r.storeRepository.DeleteRecords(ctx, successfulRecords)
		if err != nil {
			log.WithError(err).Error("cant delete successful delivered records")
		}
	}

	return r.transactor.Commit(&ctx)
}

func (r *recordsLogic) publishRecords(ctx context.Context, records []entity.Record) (entity.Records, entity.Records) {
	successfulRecords := make([]entity.Record, 0, len(records))
	errorRecords := make([]entity.Record, 0, len(records))
	for _, record := range records {
		err := r.broker.Publish(ctx, record.Message.Topic, record.Message.Body, record.Message.Headers...)
		if err != nil {
			errorRecords = append(errorRecords, record)
		}
		successfulRecords = append(successfulRecords, record)
	}
	return successfulRecords, errorRecords
}
