package database

import (
	"context"
	"sitex/config"

	"github.com/jackc/pgx/v5/pgxpool"
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

func NewDbPool(
	config *config.DatabaseConfig,
	logger *zerolog.Logger,
) *pgxpool.Pool {
	dbpool, err := pgxpool.New(context.Background(), config.Url)
	if err != nil {
		logger.Error().Msg("Не удалось подключиться к БД")
		panic(err)
	}
	logger.Info().Msg("Подключились к БД")
	return dbpool
}
