package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName(".env")
	viper.SetConfigType(
		"env",
	)
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No .env file found, using environment variables only")
		} else {
			log.Printf("Error reading config file: %v\n", err)
		}
	} else {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}
}

type DatabaseConfig struct {
	Url string
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Url: fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			viper.GetString("DB_USER"),
			viper.GetString("DB_PASSWORD"),
			viper.GetString("DB_HOST"),
			viper.GetString("DB_PORT"),
			viper.GetString("DB_NAME"),
			viper.GetString("DB_SSLMODE"),
		),
	}
}

type LogConfig struct {
	Level  int
	Format string
}

func NewLogConfig() *LogConfig {
	return &LogConfig{
		Level:  GetEnv("LOG_LEVEL", 0),
		Format: GetEnv("LOG_FORMAT", "json"),
	}
}

func GetEnv[T any](key string, defaultValue T) T {
	switch any(defaultValue).(type) {
	case string:
		return any(viper.GetString(key)).(T)
	case int:
		return any(viper.GetInt(key)).(T)
	case bool:
		return any(viper.GetBool(key)).(T)
	case time.Duration:
		return any(viper.GetDuration(key)).(T)
	case float64:
		return any(viper.GetFloat64(key)).(T)
	default:
		// For unsupported types, try to get the value and cast it
		val := viper.Get(key)
		if val == nil {
			return defaultValue
		}
		if typedVal, ok := val.(T); ok {
			return typedVal
		}
		return defaultValue
	}
}
