package logger

import (
	"os"
	"sitex/config"

	"github.com/rs/zerolog"
)

func NewLogger(config *config.LogConfig) *zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.Level(config.Level))
	var logger zerolog.Logger

	if config.Format == "json" {
		logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	} else {

		consolerWriter := zerolog.ConsoleWriter{Out: os.Stdout}
		logger = zerolog.New(consolerWriter).With().Timestamp().Logger()
	}

	return &logger
}
