package main

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
	"gitlab.enkod.tech/pkg/transactionoutbox/migration/app"
	"gitlab.enkod.tech/pkg/transactionoutbox/pkg/config"
	"os"
)

const (
	serviceName = "transaction_outbox_migration"
)

func main() {
	configSettings, err := config.LoadConfigSettingsByPath("configs")
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
