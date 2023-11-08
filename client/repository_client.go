package client

import (
	"context"
	"database/sql"
	pg "github.com/enkodio/pkg-postgres/client"
)

type RepositoryClient interface {
	Query(context.Context, string, ...interface{}) (pg.Rows, error)
	GetSqlDB() *sql.DB
}

type Transactor interface {
	Begin(*context.Context) error
	Rollback(ctx *context.Context)
	Commit(ctx *context.Context) error
}
