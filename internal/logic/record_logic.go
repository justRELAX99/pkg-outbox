package logic

import (
	"broker_transaction_outbox/internal/entity"
	"broker_transaction_outbox/pkg/kafka"
	"broker_transaction_outbox/pkg/logger"
	"broker_transaction_outbox/pkg/postgres"
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	defaultLimit                   = 1
	defaultProcessRecordsSleepTime = time.Second
)

type recordsLogic struct {
	storeRepository entity.Store
	transactor      postgres.Transactor
	broker          kafka.BrokerClient

	done chan struct{}
	wg   sync.WaitGroup
}

func NewRecordsLogic(
	storeRepository entity.Store,
	transactor postgres.Transactor,
	broker kafka.BrokerClient,
) entity.RecordLogic {
	r := &recordsLogic{
		storeRepository: storeRepository,
		transactor:      transactor,
		broker:          broker,
		done:            make(chan struct{}),
		wg:              sync.WaitGroup{},
	}
	return r
}

func (r *recordsLogic) StartProcessRecords(countGoroutines int) {
	r.wg.Add(countGoroutines)
	for i := 0; i < countGoroutines; i++ {
		go r.processRecords()
	}
}

func (r *recordsLogic) StopProcessRecords() {
	close(r.done)
	r.wg.Wait()
}

func (r *recordsLogic) processRecords() {
	defer r.wg.Done()
	ctx := context.Background()
	log := logger.GetLogger()
	for {
		time.Sleep(defaultProcessRecordsSleepTime)
		select {
		case <-r.done:
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
		err := r.broker.Publish(ctx, record.Message)
		if err != nil {
			errorRecords = append(errorRecords, record)
		}
		successfulRecords = append(successfulRecords, record)
	}
	return successfulRecords, errorRecords
}
