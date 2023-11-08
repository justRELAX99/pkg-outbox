package client

import (
	"context"
	pg "github.com/enkodio/pkg-postgres/client"
)

type RepositoryClient interface {
	Exec(context.Context, string, ...interface{}) (pg.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pg.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pg.Row
}

type Transactor interface {
	Begin(*context.Context) error
	Rollback(ctx *context.Context)
	Commit(ctx *context.Context) error
}
