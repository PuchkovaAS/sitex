package user

import (
	"fmt"
	"net/http"
	"sitex/views/components"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog"

	templeadapter "sitex/pkg/temple_adapter"
)

type UserHandlerDeps struct {
	CustomLogger *zerolog.Logger
	Store        *session.Store
	Repository   *UserRepository
}

type UserHandler struct {
	router       fiber.Router
	customLogger *zerolog.Logger
	store        *session.Store
	repository   *UserRepository
}

func NewHandler(router fiber.Router, deps UserHandlerDeps) {
	h := &UserHandler{
		router:       router,
		customLogger: deps.CustomLogger,
		store:        deps.Store,
		repository:   deps.Repository,
	}

	authGroup := router.Group("/api")
	authGroup.Post("/user/add_status", h.addStatus)
}

func (h *UserHandler) addStatus(c *fiber.Ctx) error {
	form := statusAddForm{}
	if err := c.BodyParser(&form); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат данных",
		})
	}

	email := c.Locals("email").(string)

	err := h.repository.AddStatus(statusAddInfo{
		Email:        email,
		Status:       form.Status,
		Date:         form.Date,
		Description:  form.Description,
		OneTimeEvent: form.OneTimeEvent,
	})
	if err != nil {
		return templeadapter.Render(c,
			components.Notification(
				err.Error(),
				components.NotificationFail,
			),
			fiber.StatusInternalServerError,
		)
	}

	// Парсим дату и получаем месяц
	date, err := time.Parse("2006-01-02", form.Date)
	if err != nil {
		date = time.Now()
	}
	month := date.Month()

	redirectURL := fmt.Sprintf("/?month=%d", month)
	c.Response().Header.Add("Hx-Redirect", redirectURL)
	return c.Redirect(redirectURL, http.StatusOK)
}
