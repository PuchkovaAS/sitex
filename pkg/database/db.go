package database

import (
	"context"
	"sitex/config"
	"time"

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
	var db *gorm.DB
	var err error

	maxRetries := 3
	retryDelay := 10 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = gorm.Open(postgres.Open(config.Url), &gorm.Config{})
		if err == nil {
			logger.Info().Msg("Успешно подключились к БД через GORM")
			return &Db{db}
		}

		logger.Warn().
			Int("attempt", attempt).
			Err(err).
			Msg("Не удалось подключиться к БД, повторная попытка через 10 секунд...")

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	// Если все попытки провалились
	logger.Error().Msg("Не удалось подключиться к БД после 3 попыток")
	panic(err)
}

func NewDbPool(
	config *config.DatabaseConfig,
	logger *zerolog.Logger,
) *pgxpool.Pool {
	var dbpool *pgxpool.Pool
	var err error

	maxRetries := 3
	retryDelay := 10 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		dbpool, err = pgxpool.New(context.Background(), config.Url)
		if err == nil {
			// Проверим соединение (опционально, но рекомендуется)
			if pingErr := dbpool.Ping(context.Background()); pingErr != nil {
				logger.Warn().
					Int("attempt", attempt).
					Err(pingErr).
					Msg("Подключение создано, но ping не удался, повтор...")
				dbpool.Close()
				if attempt < maxRetries {
					time.Sleep(retryDelay)
				}
				continue
			}

			logger.Info().Msg("Успешно подключились к БД через pgxpool")
			return dbpool
		}

		logger.Warn().
			Int("attempt", attempt).
			Err(err).
			Msg("Не удалось подключиться к БД, повторная попытка через 10 секунд...")

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	// Если все попытки провалились
	logger.Error().Msg("Не удалось подключиться к БД после 3 попыток")
	panic(err)
}