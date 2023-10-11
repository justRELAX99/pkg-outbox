package main

import (
	log "github.com/sirupsen/logrus"
	"gitlab.enkod.tech/pkg/transactionoutbox/internal/app"
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
		os.Exit(2)
	}
	app.Run(configSettings, serviceName)
}
