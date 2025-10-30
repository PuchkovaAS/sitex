package resources

import (
	"fmt"
	"net/http"
	"sitex/internal/user"
	"sitex/views/components"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	templeadapter "sitex/pkg/temple_adapter"
)

type ResourceHandlerDeps struct {
	CustomLogger       *zerolog.Logger
	UserRepository     *user.UserRepository
	ResourceRepository *ResourceRepository
}

type ResourceHandler struct {
	router         fiber.Router
	customLogger   *zerolog.Logger
	userRepository *user.UserRepository
	repository     *ResourceRepository
}

func NewHandler(router fiber.Router, deps ResourceHandlerDeps) *ResourceHandler {
	h := &ResourceHandler{
		router:         router,
		customLogger:   deps.CustomLogger,
		userRepository: deps.UserRepository,
		repository:     deps.ResourceRepository,
	}

	return h
}

func (h *ResourceHandler) SetupAdminRoutes(adminGroup fiber.Router) {
	adminGroup.Delete("/user/delete_resource/:id", h.deleteResource)
}

func (h *ResourceHandler) SetupPrivateRoutes(privetGroup fiber.Router) {
	privetGroup.Post("/user/add_resource", h.addResource)
}

func (h *ResourceHandler) deleteResource(c *fiber.Ctx) error {
	resourceId, err := c.ParamsInt("id")
	if err != nil {
		return templeadapter.Render(c,
			components.Notification(
				"Неверный id статуса",
				components.NotificationFail,
			),
			fiber.StatusOK,
		)
	}

	// Проверяем, что пользователь удаляет только свой статус
	emailAdmin := c.Locals("email").(string)
	isAdmin := h.userRepository.IsAdmin(emailAdmin)

	email := c.Query("email", emailAdmin)

	if !isAdmin {
		return templeadapter.Render(c,
			components.Notification(
				"Ошибка при удаление статуса, не хватает прав доступа",
				components.NotificationFail,
			),
			fiber.StatusOK,
		)
	}

	err = h.repository.DeleteResource(resourceId, email, emailAdmin)
	if err != nil {
		return templeadapter.Render(c,
			components.Notification(
				"Ошибка при удаление статуса",
				components.NotificationFail,
			),
			fiber.StatusOK,
		)
	}
	redirectURL := fmt.Sprintf("/resources/?email=%s", email)
	c.Response().Header.Add("Hx-Redirect", redirectURL)
	return c.Redirect(redirectURL, http.StatusOK)
}

func (h *ResourceHandler) addResource(c *fiber.Ctx) error {
	form := resourceAddForm{}
	if err := c.BodyParser(&form); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат данных",
		})
	}

	emailAdmin := c.Locals("email").(string)
	isAdmin := h.userRepository.IsAdmin(emailAdmin)

	email := c.Query("email", emailAdmin)

	if emailAdmin != email && !isAdmin {
		return templeadapter.Render(c,
			components.Notification(
				"Ошибка при добавление статуса",
				components.NotificationFail,
			),
			fiber.StatusOK,
		)
	}

	err := h.repository.AddResource(resourcesAddInfo{
		Email:         email,
		Quantity:      form.Quantity,
		Name:          form.Name,
		Status:        form.Status,
		Date:          form.Date,
		Description:   form.Description,
		WhoAddedEmail: emailAdmin,
	})
	if err != nil {
		return templeadapter.Render(c,
			components.Notification(
				err.Error(),
				components.NotificationFail,
			),
			fiber.StatusOK,
		)
	}

	redirectURL := fmt.Sprintf("/resources/?email=%s", email)
	c.Response().Header.Add("Hx-Redirect", redirectURL)
	return c.Redirect(redirectURL, http.StatusOK)
}
