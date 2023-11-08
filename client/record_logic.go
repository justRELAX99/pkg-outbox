package client

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.enkod.tech/pkg/postgres/client"
	"gitlab.enkod.tech/pkg/transactionoutbox/pkg/logger"
	"time"
)

const (
	defaultLimit                   = 100
	defaultCountGoroutines         = 1
	defaultProcessRecordsSleepTime = time.Second * 5
)

type recordsLogic struct {
	storeRepository Store
	transactor      Transactor
	broker          ReceivedPublisher
	RecordSettings
	syncGroup *SyncGroup
}

func newRecordsLogic(
	storeRepository Store,
	transactor client.Transactor,
	broker ReceivedPublisher,
) RecordLogic {
	r := &recordsLogic{
		storeRepository: storeRepository,
		transactor:      transactor,
		broker:          broker,
		syncGroup:       NewSyncGroup(),
		RecordSettings: RecordSettings{
			selectLimit:     defaultLimit,
			countGoroutines: defaultCountGoroutines,
			sleepTime:       defaultProcessRecordsSleepTime,
		},
	}
	return r
}

func (r *recordsLogic) StartProcessRecords() {
	r.syncGroup.Add(r.countGoroutines)
	for i := 0; i < r.countGoroutines; i++ {
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
		time.Sleep(r.sleepTime)
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
	records, err := r.storeRepository.GetPendingRecords(ctx, Filter{
		Limit: r.selectLimit,
	})
	if err != nil {
		return errors.Wrap(err, "cant get pending records")
	}
	if len(records) == 0 {
		return nil
	}

	log := logger.GetLogger()
	successfulRecords, errorRecords := r.publishRecords(ctx, records)
	//TODO maybe we shouldn’t update the records status to err and stay pending
	if len(errorRecords) > 0 {
		err = r.storeRepository.UpdateRecordsStatus(ctx, errorRecords, DeliveredErr)
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

func (r *recordsLogic) publishRecords(ctx context.Context, records []Record) (Records, Records) {
	successfulRecords := make([]Record, 0, len(records))
	errorRecords := make([]Record, 0, len(records))
	for _, record := range records {
		err := r.broker.Publish(ctx, record.Message.Topic, record.Message.Body, record.Message.Headers.ToMap())
		if err != nil {
			errorRecords = append(errorRecords, record)
		}
		successfulRecords = append(successfulRecords, record)
	}
	return successfulRecords, errorRecords
}
