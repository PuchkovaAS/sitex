package main

import (
	"sitex/config"
	"sitex/internal/pages"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func main() {
	config.Init()

	app := fiber.New()

	store := session.New()
	app.Use(recover.New())

	app.Static("/public", "./public")
	pages.NewHandler(app, store)

	app.Listen(":3000")
}
