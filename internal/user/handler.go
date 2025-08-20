package user

import (
	"sitex/views/components"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog"

	templeadapter "sitex/pkg/temple_adapter"
)

type UserHandler struct {
	router       fiber.Router
	customLogger *zerolog.Logger
	store        *session.Store
	repository   *UserRepository
}

func NewHandler(
	router fiber.Router,
	customLogger *zerolog.Logger,
	store *session.Store,
	repository *UserRepository,
) {
	h := &UserHandler{
		router:       router,
		customLogger: customLogger,
		store:        store,
		repository:   repository,
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
		Email:       email,
		Status:      form.Status,
		Date:        form.Date,
		Description: form.Description,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return templeadapter.Render(c,
		components.Notification(
			"Новость успешно создана",
			components.NotificationSuccess,
		),
		fiber.StatusCreated,
	)
}
