package repository

import (
	"context"
	"fmt"
	entity2 "github.com/enkodio/pkg-outbox/internal/outbox/entity"
	"github.com/enkodio/pkg-outbox/outbox"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/pkg/errors"
)

type storeRepository struct {
	client outbox.RepositoryClient
}

func NewStoreRepository(client outbox.RepositoryClient) entity2.Store {
	return &storeRepository{
		client: client,
	}
}

func (s *storeRepository) AddRecord(ctx context.Context, record entity2.Record) error {
	query := fmt.Sprintf(`INSERT INTO %s
	(uuid, message, state, created_on,service_name)
	VALUES($1, $2, $3, $4,$5);`,
		outboxTable)
	_, err := s.client.Exec(ctx, query, record.Uuid, record.Message, record.State, record.CreatedOn, record.ServiceName)
	if err != nil {
		return errors.Wrap(err, sqlErr)
	}
	return nil
}

func (s *storeRepository) GetPendingRecords(ctx context.Context, filter entity2.Filter) (records []entity2.Record, err error) {
	query := fmt.Sprintf(`SELECT uuid,message 
	FROM %s
	WHERE state=$1 ORDER BY created_on asc`,
		outboxTable)
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	query += " FOR UPDATE"
	rows, err := s.client.Query(ctx, query, entity2.PendingDelivery)
	if err != nil {
		return nil, errors.Wrap(err, sqlErr)
	}
	defer rows.Close()
	if err = pgxscan.ScanAll(&records, rows); err != nil {
		return records, errors.Wrap(err, scanningErr)
	}
	return records, nil
}

func (s *storeRepository) UpdateRecordsStatus(ctx context.Context, records entity2.Records, status entity2.RecordState) error {
	query := fmt.Sprintf(`UPDATE %s SET state = $1 WHERE uuid = any($2)`,
		outboxTable)
	_, err := s.client.Exec(ctx, query, status, records.GetUuids())
	if err != nil {
		return errors.Wrap(err, sqlErr)
	}
	return err
}

func (s *storeRepository) DeleteRecords(ctx context.Context, records entity2.Records) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE uuid = any($1)",
		outboxTable)
	_, err := s.client.Exec(ctx, query, records.GetUuids())
	if err != nil {
		return errors.Wrap(err, sqlErr)
	}
	return nil
}
