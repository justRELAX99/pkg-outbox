package main

import (
	"broker_transaction_outbox/migration/app"
	"broker_transaction_outbox/migration/entity"
	"broker_transaction_outbox/pkg/config"
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
	app.Run(entity.NewDataBaseSettingsByPgConfig(configSettings.PostgresConfigs), serviceName, nil)
}
