package routers

import (
	"sitex/internal/auth"
	"sitex/internal/pages"
	"sitex/internal/user"
	"sitex/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog"
)

type RouterHandlerDeps struct {
	CustomLogger *zerolog.Logger
	Store        *session.Store
	Repository   *user.UserRepository
	AuthService  *auth.AuthService
	UserService  *user.UserService
}

type RouterHandler struct {
	router       fiber.Router
	customLogger *zerolog.Logger
	store        *session.Store
	repository   *user.UserHandler
}

func NewHandler(app *fiber.App, deps RouterHandlerDeps) {
	// Создаём ОДНУ группу /api
	api := app.Group("/api")

	// Публичные API
	authHandler := auth.NewHandler(api, auth.AuthHandlerDeps{
		CustomLogger: deps.CustomLogger,
		Store:        deps.Store,
		Repository:   deps.Repository,
		Service:      deps.AuthService,
	})
	authHandler.SetupPublicRoutes()

	// Приватные API — AuthMiddleware
	privateAPI := api.Group("", middleware.AuthMiddleware(deps.Store))
	userHandler := user.NewHandler(api, user.UserHandlerDeps{
		CustomLogger: deps.CustomLogger,
		Store:        deps.Store,
		Repository:   deps.Repository,
	})
	userHandler.SetupPrivateRoutes(privateAPI)
	authHandler.SetupPrivateRoutes(privateAPI)

	// Админские API — IsAdminMiddleware
	adminAPI := api.Group("", middleware.IsAdminMiddleware(deps.Store, deps.Repository))
	authHandler.SetupAdminRoutes(adminAPI)

	pagesHandler := pages.NewHandler(app, pages.PagesHandlerDeps{
		Store:        deps.Store,
		Repository:   deps.Repository,
		CustomLogger: deps.CustomLogger,
		UserService:  deps.UserService,
	})
	pagesHandler.SetupPublicRoutes()

	privateAPP := app.Group("", middleware.AuthMiddleware(deps.Store))
	pagesHandler.SetupPrivateRoutes(privateAPP)

	adminAPP := app.Group("", middleware.IsAdminMiddleware(deps.Store, deps.Repository))
	pagesHandler.SetupAdminRoutes(adminAPP)
}
