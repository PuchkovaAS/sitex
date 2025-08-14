package pages

import (
	"sitex/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type PagesHandler struct {
	router fiber.Router
	store  *session.Store
}

func NewHandler(router fiber.Router, store *session.Store) {
	h := &PagesHandler{
		router: router,
		store:  store,
	}
	h.setupPublicRoutes()
	h.setupPrivateRoutes()
}

func (h *PagesHandler) setupPublicRoutes() {
	h.router.Get("/login", h.login)
}

func (h *PagesHandler) setupPrivateRoutes() {
	private := h.router.Group("/", middleware.AuthMiddleware(h.store))

	private.Post("/", h.home)
}

func (h *PagesHandler) login(c *fiber.Ctx) error {
	return c.SendString("login")
}

func (h *PagesHandler) home(c *fiber.Ctx) error {
	return c.SendString("ho")
}
