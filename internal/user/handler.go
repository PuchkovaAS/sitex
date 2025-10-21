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

func NewHandler(router fiber.Router, deps UserHandlerDeps) *UserHandler {
	h := &UserHandler{
		router:       router,
		customLogger: deps.CustomLogger,
		store:        deps.Store,
		repository:   deps.Repository,
	}

	return h
}

func (h *UserHandler) SetupAdminRoutes(adminGroup fiber.Router) {
	adminGroup.Post("/user/add_time_event", h.addTimeEvent)
	adminGroup.Delete("/user/delete_time_event/:id", h.deleteTimeEvent)
}

func (h *UserHandler) SetupPrivateRoutes(privetGroup fiber.Router) {
	privetGroup.Post("/user/add_status", h.addStatus)
	privetGroup.Delete("/user/delete_status/:id", h.deleteStatus)
}

func (h *UserHandler) deleteTimeEvent(c *fiber.Ctx) error {
	timeEventID, err := c.ParamsInt("id")
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
	isAdmin := h.repository.IsAdmin(emailAdmin)

	email := c.Query("email", emailAdmin)

	if emailAdmin != email && !isAdmin {
		return templeadapter.Render(c,
			components.Notification(
				"Ошибка при удаление статуса",
				components.NotificationFail,
			),
			fiber.StatusOK,
		)
	}

	err = h.repository.DeleteTimeEvent(timeEventID, email, emailAdmin)
	if err != nil {
		return templeadapter.Render(c,
			components.Notification(
				"Ошибка при удаление статуса",
				components.NotificationFail,
			),
			fiber.StatusOK,
		)
	}
	redirectURL := fmt.Sprintf("/?email=%s", email)
	c.Response().Header.Add("Hx-Redirect", redirectURL)
	return c.Redirect(redirectURL, http.StatusOK)
}

func (h *UserHandler) deleteStatus(c *fiber.Ctx) error {
	statusID, err := c.ParamsInt("id")
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
	isAdmin := h.repository.IsAdmin(emailAdmin)

	email := c.Query("email", emailAdmin)

	if emailAdmin != email && !isAdmin {
		return templeadapter.Render(c,
			components.Notification(
				"Ошибка при удаление статуса",
				components.NotificationFail,
			),
			fiber.StatusOK,
		)
	}

	err = h.repository.DeleteStatus(statusID, email)
	if err != nil {
		return templeadapter.Render(c,
			components.Notification(
				"Ошибка при удаление статуса",
				components.NotificationFail,
			),
			fiber.StatusOK,
		)
	}
	redirectURL := fmt.Sprintf("/?email=%s", email)
	c.Response().Header.Add("Hx-Redirect", redirectURL)
	return c.Redirect(redirectURL, http.StatusOK)
}

func (h *UserHandler) addTimeEvent(c *fiber.Ctx) error {
	form := timeEventAddForm{
		EventType:     c.FormValue("event_type"),
		Date:          c.FormValue("date"),
		ScheduledTime: c.FormValue("scheduled_time"),
		ActualTime:    c.FormValue("actual_time"),
		Description:   c.FormValue("description"),
	}
	if err := c.BodyParser(&form); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат данных",
		})
	}

	// Валидация
	if form.EventType == "" || form.Date == "" || form.ScheduledTime == "" || form.ActualTime == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Все поля обязательны",
		})
	}

	emailAdmin := c.Locals("email").(string)

	email := c.Query("email", emailAdmin)

	err := h.repository.AddTimeEvent(timeEventAddInfo{
		WhoAddEmail:   emailAdmin,
		Email:         email,
		EventType:     form.EventType,
		Date:          form.Date,
		ScheduledTime: form.ScheduledTime,
		ActualTime:    form.ActualTime,
		Description:   form.Description,
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

	// Парсим дату и получаем месяц
	date, err := time.Parse("2006-01-02", form.Date)
	if err != nil {
		date = time.Now()
	}
	month := date.Month()

	redirectURL := fmt.Sprintf("/?email=%s&month=%d", email, month)
	c.Response().Header.Add("Hx-Redirect", redirectURL)
	return c.Redirect(redirectURL, http.StatusOK)
}

func (h *UserHandler) addStatus(c *fiber.Ctx) error {
	form := statusAddForm{}
	if err := c.BodyParser(&form); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат данных",
		})
	}

	emailAdmin := c.Locals("email").(string)
	isAdmin := h.repository.IsAdmin(emailAdmin)

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

	err := h.repository.AddStatus(statusAddInfo{
		Email:        email,
		Status:       form.Status,
		Date:         form.Date,
		Description:  form.Description,
		OneTimeEvent: form.OneTimeEvent,
		WhoAddEmail:  emailAdmin,
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

	// Парсим дату и получаем месяц
	date, err := time.Parse("2006-01-02", form.Date)
	if err != nil {
		date = time.Now()
	}
	month := date.Month()

	redirectURL := fmt.Sprintf("/?email=%s&month=%d", email, month)
	c.Response().Header.Add("Hx-Redirect", redirectURL)
	return c.Redirect(redirectURL, http.StatusOK)
}
