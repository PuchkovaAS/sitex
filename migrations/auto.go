package main

import (
	"sitex/config"
	"sitex/internal/user"
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
		&user.Employee{},
		&user.StatusType{},
		&user.StatusPeriod{},
	)
	customLogger.Info().Msg("Мигрция прошла успешно")
}
