package main

import (
	"sitex/config"
	"sitex/internal/pages"
	"sitex/pkg/database"
	"sitex/pkg/logger"
	"time"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/postgres/v3"
)

func main() {
	config.Init()

	logConfig := config.NewLogConfig()
	customLogger := logger.NewLogger(logConfig)

	dbConfig := config.NewDatabaseConfig()
	database.NewDb(dbConfig, customLogger)

	app := fiber.New()

	app.Static("/public", "./public")

	dbpool := database.NewDbPool(dbConfig, customLogger)
	defer dbpool.Close()

	storage := postgres.New(postgres.Config{
		DB:         dbpool,
		Table:      "sessions",
		Reset:      false,
		GCInterval: 10 * time.Second,
	})

	store := session.New(session.Config{
		Storage: storage,
	})

	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: customLogger,
	}))
	app.Use(recover.New())

	pages.NewHandler(app, store)

	app.Listen(":3000")
}
