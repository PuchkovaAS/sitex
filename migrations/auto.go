package main

import (
	"os/user"
	"sitex/config"
	"sitex/pkg/database"
	"sitex/pkg/logger"
)

func main() {
	config.Init()

	logConfig := config.NewLogConfig()
	customLogger := logger.NewLogger(logConfig)
	dbConfig := config.NewDatabaseConfig()
	db := database.NewDb(dbConfig, customLogger)
	db.AutoMigrate(
		&user.User{},
	)
}
