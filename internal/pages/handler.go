package pages

import (
	"net/http"
	"sitex/internal/user"
	"sitex/pkg/middleware"
	"sitex/views"
	"sitex/views/components"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog"

	templeadapter "sitex/pkg/temple_adapter"
)

type PagesHandlerDeps struct {
	Store        *session.Store
	Repository   *user.UserRepository
	CustomLogger *zerolog.Logger
	UserService  *user.UserService
}

type PagesHandler struct {
	router       fiber.Router
	store        *session.Store
	repository   *user.UserRepository
	customLogger *zerolog.Logger
	userService  *user.UserService
}

func NewHandler(router fiber.Router, deps PagesHandlerDeps) {
	h := &PagesHandler{
		router:       router,
		store:        deps.Store,
		repository:   deps.Repository,
		customLogger: deps.CustomLogger,
		userService:  deps.UserService,
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
	private.Get("/history_status", h.historyStatus)
	private.Get("/api/logout", h.apiLogout)
}

func (h *PagesHandler) login(c *fiber.Ctx) error {
	component := views.Login()
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) UpdateUserInfo(email string, c *fiber.Ctx) {
	today := time.Now().Truncate(24 * time.Hour)
	status, err := h.repository.GetCurrentStatus(email, today)

	if err != nil {
		c.Locals("user_status", "office")
	} else {
		c.Locals("user_status", status)
	}
	userInfo, _ := h.repository.GetUserInfo(email)
	c.Locals("user_info", userInfo)
}

func (h *PagesHandler) historyStatus(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	lastAddStatus, err := h.repository.GetLastAddStatus(email)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	component := views.HistoryStatusPage(views.HistoryStatusProps{
		LastAddStatus: lastAddStatus,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) home(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	month := c.QueryInt("month", int(time.Now().Month()))
	monthHistory, statusCount, err := h.userService.GetMonthHistory(month, email, 2)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	lastAddStatus, err := h.repository.GetLastAddStatus(email, 6)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	component := views.ActivityPage(views.ActivityPageProps{
		StatusCount:   statusCount,
		MonthHistory:  monthHistory,
		CurrentMonth:  month,
		LastAddStatus: lastAddStatus,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) apiLogout(c *fiber.Ctx) error {
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
