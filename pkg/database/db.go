package database

import (
	"sitex/config"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Db struct {
	*gorm.DB
}

func NewDb(
	config *config.DatabaseConfig,
	logger *zerolog.Logger,
) *Db {
	db, err := gorm.Open(postgres.Open(config.Url), &gorm.Config{})
	if err != nil {
		logger.Error().Msg("Не удалось подключиться к БД")
		panic(err)
	}

	logger.Info().Msg("Подключились к БД")
	return &Db{db}
}
