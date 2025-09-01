package pages

import (
	"net/http"
	"sitex/internal/user"
	"sitex/pkg/middleware"
	"sitex/views"
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
}

func (h *PagesHandler) setupPrivateRoutes() {
	private := h.router.Group("/", middleware.AuthMiddleware(h.store))

	private.Get("/", h.home)
	private.Get("/history_status", h.historyStatus)
	private.Get("/year_statistics", h.yearStatistic)
	private.Get("/profile", h.profile)
	private.Get("/profile_update", h.updateUser)
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

func (h *PagesHandler) updateUser(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	emailUser := c.Query("email", email)
	employee, err := h.repository.GetEmployeeInfo(emailUser)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	component := views.UpdateProfilePage(views.UpdateProfileProps{
		Employee: employee,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) profile(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)
	employee, err := h.repository.GetEmployeeInfo(email)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	component := views.ProfilePage(views.ProfileProps{
		Employee: employee,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) yearStatistic(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	yearHistory, statusCount, err := h.userService.GetYearHistory(email)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	component := views.YearStatisticPage(views.YearStatisticProps{
		StatusCount: statusCount,
		YearHistory: yearHistory,
	})
	return templeadapter.Render(c, component, http.StatusOK)
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
