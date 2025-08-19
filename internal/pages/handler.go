package pages

import (
	"net/http"
	"sitex/pkg/middleware"
	"sitex/views"
	"sitex/views/components"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	templeadapter "sitex/pkg/temple_adapter"
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

	h.router.Post("/api/login", h.apiLogin)
}

func (h *PagesHandler) setupPrivateRoutes() {
	private := h.router.Group("/", middleware.AuthMiddleware(h.store))

	private.Get("/", h.home)
}

func (h *PagesHandler) login(c *fiber.Ctx) error {
	component := views.Login()
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) home(c *fiber.Ctx) error {
	component := views.ActivityPage()
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) apiLogin(c *fiber.Ctx) error {
	form := LoginForm{
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
	}
	if form.Email == "a@a.ru" && form.Password == "1" {
		sess, err := h.store.Get(c)
		if err != nil {
			panic(err)
		}
		sess.Set("email", form.Email)
		if err := sess.Save(); err != nil {
			panic(err)
		}
		c.Response().Header.Add("Hx-Redirect", "/")
		return c.Redirect("/", http.StatusOK)
	}

	component := components.Notification(
		"Пароль или логин неверен",
		components.NotificationFail,
	)
	return templeadapter.Render(c, component, http.StatusBadRequest)
}
