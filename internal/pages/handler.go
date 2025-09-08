package pages

import (
	"math"
	"net/http"
	"sitex/internal/dt"
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
	h.setupAdminRoutes()
}

func (h *PagesHandler) setupPublicRoutes() {
	h.router.Get("/login", h.login)
	h.router.Get("/errors/403", h.error403)
}

func (h *PagesHandler) setupAdminRoutes() {
	admin := h.router.Group("/", middleware.IsAdminMiddleware(h.store))
	admin.Get("/create_user", h.createUser)
	admin.Get("/users_activity", h.usersActivity)
}

func (h *PagesHandler) setupPrivateRoutes() {
	private := h.router.Group("/", middleware.AuthMiddleware(h.store))

	private.Get("/", h.home)
	private.Get("/history_status", h.historyStatus)
	private.Get("/year_statistics", h.yearStatistic)
	private.Get("/profile", h.profile)
	private.Get("/profile_update", h.updateUser)
	private.Get("/change_password", h.changePassword)
}

func (h *PagesHandler) createUser(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	departments, _ := h.repository.GetAllDepartments()

	component := views.CreateUserPage(views.CreateUserPageProps{
		Departments: departments,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) error403(c *fiber.Ctx) error {
	component := views.Errors403Page()
	return templeadapter.Render(c, component, http.StatusForbidden)
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

func (h *PagesHandler) getEmailForChangeUser(email string, c *fiber.Ctx) string {
	userInfo := c.Locals("user_info").(dt.UserInfo)
	var emailUser string
	if userInfo.IsAdmin {
		emailUser = c.Query("email", email)
	} else {
		emailUser = email
	}
	return emailUser
}

func (h *PagesHandler) changePassword(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)
	emailUser := h.getEmailForChangeUser(email, c)

	component := views.ChangePasswordPage(emailUser)
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) updateUser(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)
	emailUser := h.getEmailForChangeUser(email, c)

	employee, err := h.repository.GetEmployeeInfo(emailUser)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	departments, _ := h.repository.GetAllDepartments()
	component := views.UpdateProfilePage(views.UpdateProfileProps{
		Departments: departments,
		Employee:    employee,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) profile(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)
	emailUser := h.getEmailForChangeUser(email, c)

	employee, err := h.repository.GetEmployeeInfo(emailUser)
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
	emailUser := h.getEmailForChangeUser(email, c)

	yearHistory, statusCount, err := h.userService.GetYearHistory(emailUser)
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
	emailUser := h.getEmailForChangeUser(email, c)

	lastAddStatus, err := h.repository.GetLastAddStatus(emailUser)
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
	emailUser := h.getEmailForChangeUser(email, c)

	month := c.QueryInt("month", int(time.Now().Month()))
	monthHistory, statusCount, err := h.userService.GetMonthHistory(month, emailUser, 2)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	lastAddStatus, err := h.repository.GetLastAddStatus(emailUser, 6)
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

func (h *PagesHandler) usersActivity(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	month := c.QueryInt("month", int(time.Now().Month()))
	departmentID := c.QueryInt("department", 0)

	PAGE_ITEMS := 3
	page := c.QueryInt("page", 1)

	TotalUsers, err := h.repository.GetCountUsersByDepartment(departmentID)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	TotalPages := int(math.Ceil(float64(TotalUsers) / float64(PAGE_ITEMS)))

	monthHistory, statusCount, err := h.userService.GetMonthHistory(month, email, 3)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	departments, _ := h.repository.GetAllDepartments()

	component := views.MultiUserActivityPage(views.MultiUserActivityPageProps{
		Department:        departments,
		CurrentDepartment: departmentID,
		CurrentPage:       page,
		TotalPages:        TotalPages,
		TotalUsers:        int(TotalUsers),
		Users: []user.ActivityInfo{
			{
				StatusCount:  statusCount,
				MonthHistory: monthHistory,
				CurrentMonth: month,
			},
			{
				StatusCount:  statusCount,
				MonthHistory: monthHistory,
				CurrentMonth: month,
			},
		},
		CurrentMonth: month,
		UsersOnPage:  PAGE_ITEMS,
		QueryParams:  c.Queries(), // Передаем параметры запроса
	})
	return templeadapter.Render(c, component, http.StatusOK)
}
