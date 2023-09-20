package app

import (
	"broker_transaction_outbox/migration/entity"
	"broker_transaction_outbox/pkg/environment"
	"broker_transaction_outbox/pkg/logger"
	"flag"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
)

const (
	defaultMigrationPath = "./migration/migrations"
	defaultCommand       = "status"
)

func setLogger(withLog bool) {
	if withLog {
		logger.SetDefaultLogger("")
		goose.SetLogger(logger.GetLogger())
	} else {
		goose.SetLogger(&logrus.Logger{})
	}
}

func getMigrationArgs(migrationArgs map[string]string) (dir string, command string, args string, withLog bool) {
	//получаем как переменный окружения при деплое через кубер
	command = environment.GetGooseAction()
	args = environment.GetGooseArgs()
	if command != "" || args != "" {
		flagDir := flag.String("dir", defaultMigrationPath, "migration dir")
		flag.Parse()
		return *flagDir, command, args, true
	}

	//получаем как флаги
	if migrationArgs == nil {
		flagDir := flag.String("dir", defaultMigrationPath, "migration dir")
		flagCommand := flag.String("c", defaultCommand, "command")
		flagArgs := flag.String("args", "", "args")
		flagWithLog := flag.Bool("l", false, "with log")
		flag.Parse()
		return *flagDir, *flagCommand, *flagArgs, *flagWithLog
	}
	//получаем из переданной мапы(для тестов)
	var ok bool
	if dir, ok = migrationArgs["dir"]; !ok {
		dir = defaultMigrationPath
	}
	if command, ok = migrationArgs["c"]; !ok {
		command = defaultCommand
	}
	args = migrationArgs["args"]

	if _, ok = migrationArgs["l"]; ok {
		withLog = true
	}
	return
}

func Run(dbSettings entity.DataBaseSettings, serviceName string, migrationArgs map[string]string) {
	dir, command, args, withLog := getMigrationArgs(migrationArgs)
	setLogger(withLog)

	dsn := "host=" + dbSettings.Host +
		" port=" + dbSettings.Port +
		" user=" + dbSettings.User +
		" dbname=" + dbSettings.DataBaseName +
		" sslmode=disable password=" + dbSettings.Password +
		" application_name=" + serviceName

	db, err := goose.OpenDBWithDriver(dbSettings.DriverName, dsn)
	if err != nil {
		panic(errors.Wrap(err, "goose: failed to open DB"))
	}
	defer func() {
		if err := db.Close(); err != nil {
			panic(errors.Wrap(err, "goose: failed to close DB"))
		}
	}()
	goose.SetTableName(serviceName)
	if err := goose.Run(command, db, dir, args); err != nil {
		panic(errors.Wrap(err, "goose: failed run migrations"))
	}
}
