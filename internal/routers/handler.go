package routers

import (
	"sitex/internal/auth"
	"sitex/internal/pages"
	"sitex/internal/resources"
	"sitex/internal/user"
	"sitex/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog"

	mdfiles "sitex/internal/md_files"
)

type RouterHandlerDeps struct {
	CustomLogger       *zerolog.Logger
	Store              *session.Store
	Repository         *user.UserRepository
	ResourceRepository *resources.ResourceRepository
	AuthService        *auth.AuthService
	UserService        *user.UserService
	MdFilesService     *mdfiles.MdFilesService
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
	resourcesHandler := resources.NewHandler(api, resources.ResourceHandlerDeps{
		CustomLogger:       deps.CustomLogger,
		UserRepository:     deps.Repository,
		ResourceRepository: deps.ResourceRepository,
	})
	userHandler.SetupPrivateRoutes(privateAPI)
	authHandler.SetupPrivateRoutes(privateAPI)
	resourcesHandler.SetupPrivateRoutes(privateAPI)

	// Админские API — IsAdminMiddleware
	adminAPI := api.Group("", middleware.IsAdminMiddleware(deps.Store, deps.Repository))
	authHandler.SetupAdminRoutes(adminAPI)
	userHandler.SetupAdminRoutes(adminAPI)
	resourcesHandler.SetupAdminRoutes(adminAPI)

	pagesHandler := pages.NewHandler(app, pages.PagesHandlerDeps{
		Store:              deps.Store,
		Repository:         deps.Repository,
		CustomLogger:       deps.CustomLogger,
		UserService:        deps.UserService,
		ResourceRepository: deps.ResourceRepository,
		MdFilesService:     deps.MdFilesService,
	})
	pagesHandler.SetupPublicRoutes()

	privateAPP := app.Group("", middleware.AuthMiddleware(deps.Store))
	pagesHandler.SetupPrivateRoutes(privateAPP)

	adminAPP := app.Group("", middleware.IsAdminMiddleware(deps.Store, deps.Repository))
	pagesHandler.SetupAdminRoutes(adminAPP)
}
