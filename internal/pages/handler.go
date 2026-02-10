package pages

import (
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sitex/internal/dt"
	"sitex/internal/resources"
	"sitex/internal/user"
	"sitex/views"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog"

	mdfiles "sitex/internal/md_files"

	templeadapter "sitex/pkg/temple_adapter"
)

type PagesHandlerDeps struct {
	Store              *session.Store
	Repository         *user.UserRepository
	CustomLogger       *zerolog.Logger
	UserService        *user.UserService
	ResourceRepository *resources.ResourceRepository
	MdFilesService     *mdfiles.MdFilesService
}

type PagesHandler struct {
	router             fiber.Router
	store              *session.Store
	repository         *user.UserRepository
	customLogger       *zerolog.Logger
	userService        *user.UserService
	resourceRepository *resources.ResourceRepository
	mdFilesService     *mdfiles.MdFilesService
}

func NewHandler(router fiber.Router, deps PagesHandlerDeps) *PagesHandler {
	h := &PagesHandler{
		router:             router,
		store:              deps.Store,
		repository:         deps.Repository,
		customLogger:       deps.CustomLogger,
		userService:        deps.UserService,
		resourceRepository: deps.ResourceRepository,
		mdFilesService:     deps.MdFilesService,
	}
	return h
}

func (h *PagesHandler) SetupPublicRoutes() {
	h.router.Get("/login", h.login)
	h.router.Get("/errors/403", h.error403)
}

func (h *PagesHandler) SetupAdminRoutes(adminGroup fiber.Router) {
	adminGroup.Get("/create_user", h.createUser)
	adminGroup.Get("/users_activity", h.usersActivity)
	adminGroup.Get("/profile_update", h.updateUser)
	adminGroup.Get("/users_status_history", h.usersStatusHistory)
	adminGroup.Get("/users_time_event_history", h.usersTimeEventHistory)
	adminGroup.Get("/users_resources", h.usersResources)
}

func (h *PagesHandler) SetupPrivateRoutes(privetGroup fiber.Router) {
	privetGroup.Get("/", h.home)

	privetGroup.Get("/change_password", h.changePassword)
	privetGroup.Get("/history_status", h.historyStatus)
	privetGroup.Get("/history_time_event", h.historyTimeEvent)
	privetGroup.Get("/year_statistics", h.yearStatistic)
	privetGroup.Get("/profile", h.profile)
	privetGroup.Get("/resources", h.resources)

	// Документация для всех пользователей
	privetGroup.Get("/docs", h.docsList)
	privetGroup.Get("/docs/view/*", h.viewMarkdown)
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

func (h *PagesHandler) resources(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	status := c.Query("status", "")

	page := c.QueryInt("page", 1)
	PAGE_ITEMS := 20
	h.UpdateUserInfo(email, c)
	emailUser := h.getEmailForChangeUser(email, c)

	isAdmin := h.repository.IsAdmin(email)
	lastResources, totalResources, err := h.resourceRepository.GetLastResources(resources.SearchParam{
		Email:        emailUser,
		DepartmentID: 0,
		SearchQuery:  "",
		Offset:       (page - 1) * PAGE_ITEMS,
		Limit:        PAGE_ITEMS,
		Status:       status,
	})
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	TotalPages := int(math.Ceil(float64(totalResources) / float64(PAGE_ITEMS)))
	userName, err := h.repository.GetEmployeeName(emailUser)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	component := views.MaterailResources(views.MaterailResourcesProps{
		Email:       emailUser,
		TotalPage:   TotalPages,
		CurrentPage: page,
		IsAdmin:     isAdmin,
		TotalItems:  int(totalResources),
		Resources:   lastResources,
		Status:      status,
		QueryParams: c.Queries(),
		UserName:    userName,
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
	isAdmin := h.repository.IsAdmin(email)

	component := views.ProfilePage(views.ProfileProps{
		Employee: employee,
		IsAdmin:  isAdmin,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

type EmployeeData struct {
	Email          string
	StatusText     string
	FirstName      string
	LastName       string
	Position       string
	Department     string
	IsAdmin        bool
	IsActive       bool
	ShowTimeEvents bool
}

func (h *PagesHandler) getEmployeeData(employeerEmail string) (EmployeeData, error) {
	today := time.Now().Truncate(24 * time.Hour)
	status, err := h.repository.GetCurrentStatus(employeerEmail, today)
	if err != nil {
		return EmployeeData{}, err
	}

	employeerInfo, err := h.repository.GetUserInfo(employeerEmail)
	if err != nil {
		return EmployeeData{}, err
	}

	return EmployeeData{
		Email:          employeerEmail,
		StatusText:     status,
		FirstName:      employeerInfo.FirstName,
		LastName:       employeerInfo.LastName,
		Position:       employeerInfo.Position,
		IsAdmin:        employeerInfo.IsAdmin,
		IsActive:       employeerInfo.IsActive,
		Department:     employeerInfo.Department,
		ShowTimeEvents: employeerInfo.ShowTimeEvents,
	}, nil
}

func (h *PagesHandler) yearStatistic(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)
	emailUser := h.getEmailForChangeUser(email, c)
	showTimeEvents := h.repository.GetShowTimeEvents(emailUser)

	yearHistory, statusCount, err := h.userService.GetYearHistory(emailUser, showTimeEvents)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	timeEventStat, err := h.repository.GetYearTimeEventStat(emailUser)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	employeeData, err := h.getEmployeeData(emailUser)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	component := views.YearStatisticPage(views.YearStatisticProps{
		StatusCount:     statusCount,
		YearHistory:     yearHistory,
		Email:           emailUser,
		StatusText:      employeeData.StatusText,
		FirstName:       employeeData.FirstName,
		LastName:        employeeData.LastName,
		Position:        employeeData.Position,
		IsAdmin:         employeeData.IsAdmin,
		IsActive:        employeeData.IsActive,
		Department:      employeeData.Department,
		LatelyMin:       timeEventStat.LatelyMin,
		LatelyCount:     timeEventStat.LatelyCount,
		EarlyLeaveMin:   timeEventStat.EarlyLeaveMin,
		EarlyLeaveCount: timeEventStat.EarlyLeaveCount,
		ShowTimeEvents:  employeeData.ShowTimeEvents,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) usersTimeEventHistory(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	PAGE_ITEMS := 20
	departmentID := c.QueryInt("department", 0)

	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	searchQuery := c.Query("search", "")
	departments, err := h.repository.GetAllDepartments()
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	lastAddStatus, TotalEvents, err := h.repository.GetLastAddTimeEvents(user.SearchParam{
		Email:        "",
		DepartmentID: uint(departmentID),
		SearchQuery:  searchQuery,
		Offset:       (page - 1) * PAGE_ITEMS,
		Limit:        PAGE_ITEMS,
	})
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	TotalPages := int(math.Ceil(float64(TotalEvents) / float64(PAGE_ITEMS)))

	isAdmin := h.repository.IsAdmin(email)

	component := views.UsersHistoryTimeEventPage(views.UsersHistoryTimeEventProps{
		TotalPage:      TotalPages,
		CurrentPage:    page,
		LastTimeEvents: lastAddStatus,
		Depatrments:    departments,
		DepartmentId:   departmentID,
		QueryParams:    c.Queries(),
		TotalItems:     int(TotalEvents),
		IsAdmin:        isAdmin,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) usersResources(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	PAGE_ITEMS := 20
	departmentID := c.QueryInt("department", 0)
	status := c.Query("status", "")

	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	searchQuery := c.Query("search", "")
	departments, err := h.repository.GetAllDepartments()
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	lastResources, totalResources, err := h.resourceRepository.GetLastResources(resources.SearchParam{
		Email:        "",
		DepartmentID: uint(departmentID),
		SearchQuery:  searchQuery,
		Offset:       (page - 1) * PAGE_ITEMS,
		Limit:        PAGE_ITEMS,
		Status:       status,
	})
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	TotalPages := int(math.Ceil(float64(totalResources) / float64(PAGE_ITEMS)))

	component := views.UsersReources(views.UsersResourcesProps{
		Resources:    lastResources,
		TotalPage:    TotalPages,
		CurrentPage:  page,
		Depatrments:  departments,
		DepartmentId: departmentID,
		QueryParams:  c.Queries(),
		TotalItems:   int(totalResources),
		Status:       status,
	})

	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) usersStatusHistory(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	PAGE_ITEMS := 20
	departmentID := c.QueryInt("department", 0)

	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	searchQuery := c.Query("search", "")
	departments, err := h.repository.GetAllDepartments()
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	lastAddStatus, TotalEvents, err := h.repository.GetLastAddEvents(user.SearchParam{
		Email:        "",
		DepartmentID: uint(departmentID),
		SearchQuery:  searchQuery,
		Offset:       (page - 1) * PAGE_ITEMS,
		Limit:        PAGE_ITEMS,
	})
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	TotalPages := int(math.Ceil(float64(TotalEvents) / float64(PAGE_ITEMS)))

	component := views.UsersHistoryStatusPage(views.UsersHistoryStatusProps{
		TotalPage:     TotalPages,
		CurrentPage:   page,
		LastAddStatus: lastAddStatus,
		Depatrments:   departments,
		DepartmentId:  departmentID,
		QueryParams:   c.Queries(),
		TotalItems:    int(TotalEvents),
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) historyStatus(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	PAGE_ITEMS := 20

	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)
	emailUser := h.getEmailForChangeUser(email, c)

	lastAddStatus, TotalEvents, err := h.repository.GetLastAddEvents(user.SearchParam{
		Email:        emailUser,
		DepartmentID: 0,
		SearchQuery:  "",
		Offset:       (page - 1) * PAGE_ITEMS,
		Limit:        PAGE_ITEMS,
	})
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	TotalPages := int(math.Ceil(float64(TotalEvents) / float64(PAGE_ITEMS)))

	isAdmin := h.repository.IsAdmin(email)

	showTimeEvents := h.repository.GetShowTimeEvents(email)
	userName, err := h.repository.GetEmployeeName(emailUser)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	component := views.HistoryStatusPage(views.HistoryStatusProps{
		TotalPage:      TotalPages,
		CurrentPage:    page,
		Email:          emailUser,
		LastAddStatus:  lastAddStatus,
		TotalItems:     int(TotalEvents),
		IsAdmin:        isAdmin,
		ShowTimeEvents: showTimeEvents,

		UserName: userName,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) historyTimeEvent(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	PAGE_ITEMS := 20

	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)
	emailUser := h.getEmailForChangeUser(email, c)

	lastTimeEvents, TotalEvents, err := h.repository.GetLastAddTimeEvents(user.SearchParam{
		Email:        emailUser,
		DepartmentID: 0,
		SearchQuery:  "",
		Offset:       (page - 1) * PAGE_ITEMS,
		Limit:        PAGE_ITEMS,
	})
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	TotalPages := int(math.Ceil(float64(TotalEvents) / float64(PAGE_ITEMS)))

	isAdmin := h.repository.IsAdmin(email)

	userName, err := h.repository.GetEmployeeName(emailUser)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	component := views.HistoryTimeEventPage(views.HistoryTimeEventProps{
		TotalPage:      TotalPages,
		CurrentPage:    page,
		Email:          emailUser,
		LastTimeEvents: lastTimeEvents,
		TotalItems:     int(TotalEvents),
		IsAdmin:        isAdmin,
		UserName:       userName,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) home(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)
	emailUser := h.getEmailForChangeUser(email, c)
	showTimeEvents := h.repository.GetShowTimeEvents(emailUser)

	month := c.QueryInt("month", int(time.Now().Month()))
	monthHistory, statusCount, err := h.userService.GetMonthHistory(month, emailUser, 2, showTimeEvents)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	lastAddStatus, err := h.repository.GetLastAddStatus(emailUser, 6)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	lastAddTimeEvent, err := h.repository.GetLastTimeEvents(emailUser, 6)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	employeeData, err := h.getEmployeeData(emailUser)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	timeEventStat, err := h.repository.GetTimeEventStat(month, emailUser)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	userIsAdmin := h.repository.IsAdmin(email)

	component := views.ActivityPage(views.ActivityPageProps{
		StatusCount:     statusCount,
		MonthHistory:    monthHistory,
		CurrentMonth:    month,
		LastAddStatus:   lastAddStatus,
		LastTimeEvents:  lastAddTimeEvent,
		Email:           emailUser,
		StatusText:      employeeData.StatusText,
		FirstName:       employeeData.FirstName,
		LastName:        employeeData.LastName,
		Position:        employeeData.Position,
		IsAdmin:         employeeData.IsAdmin,
		IsActive:        employeeData.IsActive,
		Department:      employeeData.Department,
		LatelyMin:       timeEventStat.LatelyMin,
		LatelyCount:     timeEventStat.LatelyCount,
		EarlyLeaveMin:   timeEventStat.EarlyLeaveMin,
		EarlyLeaveCount: timeEventStat.EarlyLeaveCount,
		UserIsAdmin:     userIsAdmin,
		ShowTimeEvents:  employeeData.ShowTimeEvents,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) usersActivity(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	month := c.QueryInt("month", int(time.Now().Month()))
	departmentID := c.QueryInt("department", 0)
	searchQuery := c.Query("search", "")

	departments, err := h.repository.GetAllDepartments()
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}

	PAGE_ITEMS := 5
	page := c.QueryInt("page", 1)

	employees, TotalUsers, err := h.repository.GetUsersByParam(
		user.SearchParam{
			DepartmentID: uint(departmentID),
			SearchQuery:  searchQuery,
			Offset:       (page - 1) * PAGE_ITEMS,
			Limit:        PAGE_ITEMS,
		},
	)
	if err != nil {
		h.customLogger.Error().Msg(err.Error())
		return c.SendStatus(500)
	}
	TotalPages := int(math.Ceil(float64(TotalUsers) / float64(PAGE_ITEMS)))

	var usersInfo []user.ActivityInfo

	today := time.Now().Truncate(24 * time.Hour)
	for _, employee := range employees {

		showTimeEvents := h.repository.GetShowTimeEvents(employee.Email)

		monthHistory, statusCount, err := h.userService.GetMonthHistory(month, employee.Email, 3, showTimeEvents)
		if err != nil {
			h.customLogger.Error().Msg(err.Error())
			return c.SendStatus(500)
		}

		timeEventStat, err := h.repository.GetTimeEventStat(month, employee.Email)
		if err != nil {
			h.customLogger.Error().Msg(err.Error())
			return c.SendStatus(500)
		}

		status, err := h.repository.GetCurrentStatus(employee.Email, today)

		// Создаем и добавляем ActivityInfo в срез
		usersInfo = append(usersInfo, user.ActivityInfo{
			Employee:        employee,
			StatusCount:     statusCount,
			MonthHistory:    monthHistory,
			CurrentMonth:    month,
			StatusText:      status,
			LatelyMin:       timeEventStat.LatelyMin,
			LatelyCount:     timeEventStat.LatelyCount,
			EarlyLeaveMin:   timeEventStat.EarlyLeaveMin,
			EarlyLeaveCount: timeEventStat.EarlyLeaveCount,
		},
		)
	}

	component := views.MultiUserActivityPage(views.MultiUserActivityPageProps{
		Department:        departments,
		CurrentDepartment: departmentID,
		CurrentPage:       page,
		TotalPages:        TotalPages,
		TotalUsers:        int(TotalUsers),
		Users:             usersInfo,
		CurrentMonth:      month,
		QueryParams:       c.Queries(), // Передаем параметры запроса
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

func (h *PagesHandler) docsList(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	folder := c.Query("folder", "")

	baseDir := h.mdFilesService.MdFilesBaseDir

	var currentDir string

	if folder == "" {
		currentDir = baseDir
	} else {
		currentDir = filepath.Join(baseDir, folder)
		absBase, _ := filepath.Abs(baseDir)
		absCurrent, _ := filepath.Abs(currentDir)
		if !strings.HasPrefix(absCurrent, absBase) {
			return c.Status(fiber.StatusForbidden).SendString("Access denied")
		}
	}

	entries, err := h.mdFilesService.BuildFileTreeForDirectory(currentDir, folder)
	if err != nil {
		h.customLogger.Error().Msgf("Error building file tree: %v", err)
		entries = []dt.FileEntry{}
	}

	treeHTML := h.mdFilesService.RenderFileTreeHTML(entries, folder)

	component := views.DocsTreePage(views.DocsTreePageProps{
		TreeHTML: treeHTML,
		Folder:   folder,
	})
	return templeadapter.Render(c, component, http.StatusOK)
}

// Страница просмотра конкретного документа
func (h *PagesHandler) viewMarkdown(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	filename := c.Params("*")
	if filename == "" {
		return c.Redirect("/docs")
	}

	mdFilesBaseDir := h.mdFilesService.MdFilesBaseDir

	filePath := filepath.Join(mdFilesBaseDir, filename)

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		h.customLogger.Error().Msgf("Invalid file path: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid file path")
	}

	contentDir, _ := filepath.Abs(mdFilesBaseDir)
	if !strings.HasPrefix(absPath, contentDir) {
		h.customLogger.Warn().Msgf("Attempted directory traversal: %s", filePath)
		return c.Status(fiber.StatusForbidden).SendString("Access denied")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.customLogger.Warn().Msgf("File not found: %s", filePath)
		return c.Status(fiber.StatusNotFound).SendString("Document not found")
	}

	htmlContent, err := h.mdFilesService.GetCachedMarkdown(filePath)
	if err != nil {
		h.customLogger.Error().Msgf("Error reading markdown file: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error reading document")
	}

	title := strings.TrimSuffix(filepath.Base(filename), ".md")
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.ReplaceAll(title, "-", " ")
	if len(title) > 0 {
		title = strings.ToUpper(string(title[0])) + title[1:]
	}

	component := views.MarkdownPage(views.MarkdownPageProps{
		Title:    title,
		Content:  htmlContent,
		FilePath: filename,
	})

	return templeadapter.Render(c, component, http.StatusOK)
}
