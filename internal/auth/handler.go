package auth

import (
	"errors"
	"fmt"
	"net/http"
	"sitex/internal/user"
	"sitex/pkg/validator"
	"sitex/views/components"
	"strings"

	"github.com/a-h/templ"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog"

	templeadapter "sitex/pkg/temple_adapter"
)

type AuthHandlerDeps struct {
	CustomLogger *zerolog.Logger
	Store        *session.Store
	Repository   *user.UserRepository
	Service      *AuthService
}

type AuthHandler struct {
	router       fiber.Router
	customLogger *zerolog.Logger
	store        *session.Store
	repository   *user.UserRepository
	service      *AuthService
}

func NewHandler(router fiber.Router, deps AuthHandlerDeps) *AuthHandler {
	h := &AuthHandler{
		router:       router,
		customLogger: deps.CustomLogger,
		store:        deps.Store,
		repository:   deps.Repository,
		service:      deps.Service,
	}
	return h
}

// SetupPublicRoutes регистрирует публичные API-эндпоинты (без middleware).
func (h *AuthHandler) SetupPublicRoutes() {
	h.router.Post("/login", h.apiLogin)
}

func (h *AuthHandler) SetupAdminRoutes(adminGroup fiber.Router) {
	adminGroup.Post("/create_user", h.apiCreateUser)
	adminGroup.Put("/profile_update", h.apiUpdateUser)
}

func (h *AuthHandler) SetupPrivateRoutes(privetGroup fiber.Router) {
	privetGroup.Put("/change_password", h.apiChangePassword)
	privetGroup.Get("/logout", h.apiLogout)
}

func validationPassword(password, confirmPassword string) error {
	// Валидация
	if len(password) < 6 {
		return errors.New(ErrPasswordIsLess)
	}

	if password != confirmPassword {
		return errors.New(ErrPasswordNotEq)
	}

	return nil
}

func (h *AuthHandler) apiChangePassword(c *fiber.Ctx) error {
	var req changePasswordForm
	if err := c.BodyParser(&req); err != nil {
		component := components.Notification(
			"Неверный формат данных",
			components.NotificationFail,
		)
		return templeadapter.Render(c, component, http.StatusBadRequest)
	}

	err := validationPassword(req.NewPassword, req.ConfirmPassword)
	if err != nil {

		component := components.Notification(
			err.Error(),
			components.NotificationFail,
		)
		return templeadapter.Render(c, component, http.StatusOK)
	}

	err = h.service.ChangePassword(req)
	if err != nil {
		component := components.Notification(
			"Возникла ошибка на сервере"+err.Error(),
			components.NotificationFail,
		)
		return templeadapter.Render(c, component, http.StatusOK)

	}
	redirectURL := fmt.Sprintf("/profile?email=%s", req.Email)
	c.Response().Header.Add("Hx-Redirect", redirectURL)
	return c.Redirect(redirectURL, http.StatusOK)
}

func (h *AuthHandler) apiCreateUser(c *fiber.Ctx) error {
	form := userCreateForm{}

	// Парсим форму
	if err := c.BodyParser(&form); err != nil {
		component := components.Notification(
			err.Error(),
			components.NotificationFail,
		)
		return templeadapter.Render(c, component, http.StatusOK)
	}

	var departmentName string
	if form.NewDepartment != "" {
		departmentName = strings.TrimSpace(form.NewDepartment)
	} else {
		departmentName = form.Department
	}

	// Валидация
	error := validate.Validate(
		&validators.StringIsPresent{
			Name:    "Имя",
			Field:   form.FirstName,
			Message: "Имя не задано",
		},
		&validators.StringIsPresent{
			Name:    "Фамилия",
			Field:   form.LastName,
			Message: "Фамилия не задана",
		},
		&validators.StringIsPresent{
			Name:    "Должность",
			Field:   form.Position,
			Message: "Должность не задана",
		},
		&validators.StringIsPresent{
			Name:    "Отдел",
			Field:   departmentName,
			Message: "Отдел не задан",
		},
		&validators.EmailIsPresent{
			Name:    "Email",
			Field:   form.Email,
			Message: "email не верный",
		},
	)

	var component templ.Component
	if len(error.Errors) > 0 {
		component = components.Notification(
			validator.FormatErrors(error),
			components.NotificationFail,
		)
		return templeadapter.Render(c, component, http.StatusOK)
	}

	err := validationPassword(form.Password, form.ConfirmPassword)
	if err != nil {

		component = components.Notification(
			err.Error(),
			components.NotificationFail,
		)
		return templeadapter.Render(c, component, http.StatusOK)
	}

	err = h.service.Register(form)
	if err != nil {
		component := components.Notification(
			"Возникла ошибка на сервере"+err.Error(),
			components.NotificationFail,
		)
		return templeadapter.Render(c, component, http.StatusOK)

	}

	redirectURL := fmt.Sprintf("/profile?email=%s", form.Email)
	c.Response().Header.Add("Hx-Redirect", redirectURL)
	return c.Redirect(redirectURL, http.StatusOK)
}

func (h *AuthHandler) apiUpdateUser(c *fiber.Ctx) error {
	form := userUpdateForm{}

	// Парсим форму
	if err := c.BodyParser(&form); err != nil {
		component := components.Notification(
			err.Error(),
			components.NotificationFail,
		)
		return templeadapter.Render(c, component, http.StatusOK)
	}

	var departmentName string
	if form.NewDepartment != "" {
		departmentName = strings.TrimSpace(form.NewDepartment)
	} else {
		departmentName = form.Department
	}

	// Конвертируем строки в boolean
	isActive := form.IsActive == "true"
	isAdmin := form.IsAdmin == "true"
	showTimeEvents := form.ShowTimeEvent == "true"

	// Валидация
	error := validate.Validate(
		&validators.StringIsPresent{
			Name:    "Имя",
			Field:   form.FirstName,
			Message: "Имя не задано",
		},
		&validators.StringIsPresent{
			Name:    "Фамилия",
			Field:   form.LastName,
			Message: "Фамилия не задана",
		},
		&validators.StringIsPresent{
			Name:    "Должность",
			Field:   form.Position,
			Message: "Должность не задана",
		},
		&validators.StringIsPresent{
			Name:    "Отдел",
			Field:   departmentName,
			Message: "Отдел не задан",
		},
	)
	var component templ.Component
	if len(error.Errors) > 0 {
		component = components.Notification(
			validator.FormatErrors(error),
			components.NotificationFail,
		)
		return templeadapter.Render(c, component, http.StatusOK)
	}

	if err := h.repository.UpdateUserProfile(form.Email, user.UserUpdateData{
		FirstName:      strings.TrimSpace(form.FirstName),
		LastName:       strings.TrimSpace(form.LastName),
		Position:       strings.TrimSpace(form.Position),
		Department:     departmentName,
		IsActive:       isActive,
		IsAdmin:        isAdmin,
		ShowTimeEvents: showTimeEvents,
	}); err != nil {
		component = components.Notification(
			err.Error(),
			components.NotificationFail,
		)
		return templeadapter.Render(c, component, http.StatusOK)
	}
	redirectURL := fmt.Sprintf("/profile?email=%s", form.Email)
	c.Response().Header.Add("Hx-Redirect", redirectURL)
	return c.Redirect(redirectURL, http.StatusOK)
}

func (h *AuthHandler) apiLogout(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		panic(err)
	}
	sess.Delete("email")
	if err := sess.Save(); err != nil {
		panic(err)
	}
	return c.Redirect("/login", http.StatusFound)
}

func (h *AuthHandler) apiLogin(c *fiber.Ctx) error {
	form := LoginForm{
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
	}

	error := validate.Validate(
		&validators.EmailIsPresent{
			Name:    "Email",
			Field:   form.Email,
			Message: "Email не задан или не верный",
		},
		&validators.StringIsPresent{
			Name:    "Password",
			Field:   form.Password,
			Message: "Пароль не задан",
		},
	)
	var component templ.Component
	if len(error.Errors) > 0 {
		component = components.Notification(
			validator.FormatErrors(error),
			components.NotificationFail,
		)
		c.Set("Content-Type", "text/html")
		return templeadapter.Render(c, component, http.StatusOK)
	}

	if err := h.service.Login(form); err != nil {

		component = components.Notification(
			"Пароль или логин неверен",
			components.NotificationFail,
		)
		return templeadapter.Render(c, component, http.StatusOK)
	}

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
