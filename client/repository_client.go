package client

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Client interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	GetSqlDB() *sql.DB
}

type Transactor interface {
	Begin(*context.Context) error
	Rollback(*context.Context)
	Commit(*context.Context) error
}
