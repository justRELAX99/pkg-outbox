package app

import (
	"database/sql"
	"flag"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
	"gitlab.enkod.tech/pkg/transactionoutbox/pkg/environment"
	"gitlab.enkod.tech/pkg/transactionoutbox/pkg/logger"
)

const (
	defaultMigrationPath = "./migration/migrations"
	defaultCommand       = "status"
)

func setLogger(withLog bool) {
	if withLog {
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

func Run(db *sql.DB, serviceName string, migrationArgs map[string]string) {
	dir, command, args, withLog := getMigrationArgs(migrationArgs)
	setLogger(withLog)

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
