package main

import (
	"sitex/config"
	"sitex/internal/pages"
	"sitex/pkg/database"
	"sitex/pkg/logger"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func main() {
	config.Init()

	logConfig := config.NewLogConfig()
	customLogger := logger.NewLogger(logConfig)

	dbConfig := config.NewDatabaseConfig()
	database.NewDb(dbConfig, customLogger)

	app := fiber.New()

	store := session.New()

	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: customLogger,
	}))
	app.Use(recover.New())

	app.Static("/public", "./public")
	pages.NewHandler(app, store)

	app.Listen(":3000")
}
