package entity

import "broker_transaction_outbox/pkg/postgres"

const (
	pgDriverName = "pgx"
)

type DataBaseSettings struct {
	User         string
	Password     string
	Host         string
	Port         string
	DataBaseName string
	DriverName   string
}

func NewDataBaseSettingsByPgConfig(pgConf postgres.Config) DataBaseSettings {
	return DataBaseSettings{
		User:         pgConf.User,
		Password:     pgConf.Password,
		Host:         pgConf.Host,
		Port:         pgConf.Port,
		DataBaseName: pgConf.DBName,
		DriverName:   pgDriverName,
	}
}
