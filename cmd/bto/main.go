package main

import (
	"broker_transaction_outbox/internal/app"
	"broker_transaction_outbox/pkg/config"
	"github.com/labstack/gommon/log"
	"os"
)

const (
	serviceName = "broker_transaction_outbox"
)

func main() {
	configSettings, err := config.LoadConfigSettingsByPath("config")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	app.Run(configSettings, serviceName)
}
