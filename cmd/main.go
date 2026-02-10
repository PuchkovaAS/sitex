package main

import (
	"fmt"
	"sitex/config"
	"sitex/internal/auth"
	"sitex/internal/resources"
	"sitex/internal/routers"
	"sitex/internal/user"
	"sitex/pkg/database"
	"sitex/pkg/logger"
	"sitex/pkg/middleware"
	"sitex/views"
	"time"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/postgres/v3"

	mdfiles "sitex/internal/md_files"

	templeadapter "sitex/pkg/temple_adapter"
)

func main() {
	config.Init()

	workCalendar := config.InitCalendar(
		fmt.Sprintf("%d/calendar.json", time.Now().Year()),
	)

	mdFilesConfig := config.NewMdFilesConfig()

	logConfig := config.NewLogConfig()
	customLogger := logger.NewLogger(logConfig)

	dbConfig := config.NewDatabaseConfig()
	db := database.NewDb(dbConfig, customLogger)

	app := config.ApiInit()

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

	// Middleware
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: customLogger,
	}))
	app.Use(recover.New())
	app.Use(middleware.CurrentURLMiddleware())

	// Repository
	userRepository := user.NewUserRepository(db)
	resourceRepository := resources.NewRepository(db)

	// Service
	userService := user.NewUserService(&user.UserServiceDeps{
		UserRepository: *userRepository,
		CustomLogger:   customLogger,
		WorkCalendar:   workCalendar,
	})

	authService := auth.NewAuthService(userRepository)

	mdFilesService := mdfiles.NewMdFilesService(mdFilesConfig)

	// Handler

	routers.NewHandler(app, routers.RouterHandlerDeps{
		CustomLogger:       customLogger,
		Store:              store,
		Repository:         userRepository,
		AuthService:        authService,
		UserService:        userService,
		ResourceRepository: resourceRepository,
		MdFilesService:     mdFilesService,
	})

	// Обработчик 404 — должен быть ПОСЛЕ всех других маршрутов
	app.Use(func(c *fiber.Ctx) error {
		c.Status(fiber.StatusNotFound) // Устанавливаем статус 404
		return templeadapter.Render(c, views.Errors404Page(), fiber.StatusOK)
	})
	app.Listen(":3000")
}
