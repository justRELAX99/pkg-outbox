package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

const (
	defaultMaxDelay    = 5
	defaultMaxAttempts = 5

	maxConnIdleTime = time.Second * 5

	logTickerSec = 60
)

type client struct {
	pool        *pgxpool.Pool
	serviceName string
}

func NewClient(cfg Config, serviceName string) (Client, Transactor) {
	var pool *pgxpool.Pool
	dsn := "host=" + cfg.Host +
		" port=" + cfg.Port +
		" user=" + cfg.User +
		" dbname=" + cfg.DBName +
		" sslmode=disable password=" + cfg.Password +
		" application_name=" + serviceName
	err := DoWithAttempts(
		func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			pgxCfg, err := pgxpool.ParseConfig(dsn)
			if err != nil {
				log.Fatalf("Unable to parse config: %v\n", err)
			}
			pgxCfg.MaxConns = int32(cfg.MaxOpenConns)
			if pgxCfg.MaxConns == 0 {
				pgxCfg.MaxConns = 4
			}
			pgxCfg.MaxConnIdleTime = maxConnIdleTime
			pool, err = pgxpool.NewWithConfig(ctx, pgxCfg)
			if err != nil {
				log.Println("Failed to connect to postgres... Going to do the next attempt")
				return err
			}

			return nil
		}, cfg.MaxAttempts, cfg.MaxDelay,
	)

	if err != nil {
		log.Fatal("All attempts are exceeded. Unable to connect to postgres")
	}

	pg := client{
		pool:        pool,
		serviceName: serviceName,
	}
	return &pg, &pg
}

func (c *client) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	tx := c.getTx(ctx)
	if tx != nil {
		return tx.Query(ctx, query, args...)
	}
	return c.pool.Query(ctx, query, args...)
}

func (c *client) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	tx := c.getTx(ctx)
	if tx != nil {
		return tx.QueryRow(ctx, query, args...)
	}
	return c.pool.QueryRow(ctx, query, args...)
}

func (c *client) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	tx := c.getTx(ctx)
	if tx != nil {
		return tx.Exec(ctx, query, args...)
	}
	return c.pool.Exec(ctx, query, args...)
}

func DoWithAttempts(fn func() error, maxAttempts int, delay int) error {
	var err error
	if maxAttempts == 0 {
		maxAttempts = defaultMaxAttempts
	}
	if delay == 0 {
		delay = defaultMaxDelay
	}
	for maxAttempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(time.Second * time.Duration(delay))
			maxAttempts--

			continue
		}

		return nil
	}

	return err
}
