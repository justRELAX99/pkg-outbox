package main

import (
	"broker_transaction_outbox/migration/app"
	"broker_transaction_outbox/pkg/config"
	"database/sql"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	serviceName = "broker_transaction_outbox_migration"
)

func main() {
	configSettings, err := config.LoadConfigSettingsByPath("config")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	db, err := sql.Open("pgx", configSettings.PostgresConfigs.GetDSN(serviceName))
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	app.Run(db, serviceName, nil)
}
