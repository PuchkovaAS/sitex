package pages

import (
	"bytes"
	"fmt"
	"io/fs"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sitex/internal/dt"
	"sitex/internal/resources"
	"sitex/internal/user"
	"sitex/views"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/rs/zerolog"

	templeadapter "sitex/pkg/temple_adapter"
)

type DocTreeNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []DocTreeNode
}

type cachedMarkdown struct {
	htmlContent  string
	lastModified time.Time
}

type PagesHandlerDeps struct {
	Store              *session.Store
	Repository         *user.UserRepository
	CustomLogger       *zerolog.Logger
	UserService        *user.UserService
	ResourceRepository *resources.ResourceRepository
}

type PagesHandler struct {
	router             fiber.Router
	store              *session.Store
	repository         *user.UserRepository
	customLogger       *zerolog.Logger
	userService        *user.UserService
	resourceRepository *resources.ResourceRepository

	// Кэш для скомпилированных документов
	markdownCache map[string]cachedMarkdown
	cacheMutex    sync.RWMutex
}

func NewHandler(router fiber.Router, deps PagesHandlerDeps) *PagesHandler {
	h := &PagesHandler{
		router:             router,
		store:              deps.Store,
		repository:         deps.Repository,
		customLogger:       deps.CustomLogger,
		userService:        deps.UserService,
		resourceRepository: deps.ResourceRepository,
		markdownCache:      make(map[string]cachedMarkdown),
		cacheMutex:         sync.RWMutex{},
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

func (h *PagesHandler) convertMarkdownToHTML(content []byte) string {
	extensions := parser.CommonExtensions |
		parser.AutoHeadingIDs |
		parser.NoEmptyLineBeforeBlock |
		parser.Tables |
		parser.FencedCode |
		parser.Strikethrough

	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(content)

	opts := html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}

// Получение списка всех md файлов
func (h *PagesHandler) getMarkdownFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			relPath, _ := filepath.Rel(dir, path)
			// Заменяем обратные слеши на прямые для URL
			relPath = strings.ReplaceAll(relPath, "\\", "/")
			files = append(files, relPath)
		}
		return nil
	})

	return files, err
}

// Получение и кэширование контента
func (h *PagesHandler) getCachedMarkdown(filePath string) (string, error) {
	h.cacheMutex.RLock()
	cached, exists := h.markdownCache[filePath]
	h.cacheMutex.RUnlock()

	if exists {
		// Проверяем актуальность кэша
		fileInfo, err := os.Stat(filePath)
		if err == nil && fileInfo.ModTime().Equal(cached.lastModified) {
			return cached.htmlContent, nil
		}
	}

	// Читаем файл
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Конвертируем в HTML
	htmlContent := h.convertMarkdownToHTML(content)

	// Сохраняем в кэш
	fileInfo, _ := os.Stat(filePath)
	h.cacheMutex.Lock()
	h.markdownCache[filePath] = cachedMarkdown{
		htmlContent:  htmlContent,
		lastModified: fileInfo.ModTime(),
	}
	h.cacheMutex.Unlock()

	return htmlContent, nil
}

// Строим дерево только для указанной директории (без рекурсии вглубь)
func (h *PagesHandler) buildFileTreeForDirectory(dir, folderPrefix string) ([]dt.FileEntry, error) {
	var entries []dt.FileEntry

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		entry := dt.FileEntry{
			Name: file.Name(),
		}

		if folderPrefix == "" {
			entry.Path = file.Name()
		} else {
			entry.Path = folderPrefix + "/" + file.Name()
		}

		if file.IsDir() {
			entry.IsDir = true
		} else if strings.HasSuffix(strings.ToLower(file.Name()), ".md") {
			entry.IsDir = false
		} else {
			continue // Пропускаем не-MD файлы
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// Генерация HTML дерева документов с поддержкой папок
func (h *PagesHandler) renderFileTreeHTML(entries []dt.FileEntry, currentFolder string) string {
	if len(entries) == 0 {
		return "<p class=\"text-gray-500\">Нет доступных документов</p>"
	}

	var buf bytes.Buffer

	// Кнопка "Назад"
	if currentFolder != "" {
		parentFolder := ""
		parts := strings.Split(currentFolder, "/")
		if len(parts) > 1 {
			parentFolder = strings.Join(parts[:len(parts)-1], "/")
		}
		backURL := "/docs"
		if parentFolder != "" {
			backURL += "?folder=" + url.QueryEscape(parentFolder)
		}
		buf.WriteString(
			fmt.Sprintf("<div class=\"mb-4\"><a href=\"%s\" class=\"inline-flex items-center text-blue-600 hover:text-blue-800\">", backURL),
		)
		buf.WriteString(
			"<svg class=\"w-4 h-4 mr-1\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M15 19l-7-7 7-7\"></path></svg>",
		)
		buf.WriteString("Назад</a></div>\n")
	}

	buf.WriteString("<ul class=\"file-tree space-y-1\">\n")
	for _, entry := range entries {
		indent := "  "
		liClass := "relative"

		buf.WriteString(fmt.Sprintf("%s<li class=\"%s\">\n", indent, liClass))

		if entry.IsDir {
			// Для папок используем QueryEscape (нужно для URL с параметром folder)
			folderURL := "/docs?folder=" + url.QueryEscape(entry.Path)
			buf.WriteString(
				fmt.Sprintf(
					"%s  <a href=\"%s\" class=\"flex items-center text-gray-700 hover:text-blue-600 transition-colors py-1\">\n",
					indent,
					folderURL,
				),
			)
			buf.WriteString(fmt.Sprintf("%s    <svg class=\"mr-2 w-4 h-4\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\">\n", indent))
			buf.WriteString(
				fmt.Sprintf(
					"%s      <path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z\"></path>\n",
					indent,
				),
			)
			buf.WriteString(fmt.Sprintf("%s    </svg>\n", indent))
			buf.WriteString(fmt.Sprintf("%s    <span>%s</span>\n", indent, templ.EscapeString(entry.Name)))
			buf.WriteString(fmt.Sprintf("%s  </a>\n", indent))
		} else {
			// Для файлов НЕ используем url.QueryEscape — Fiber сам обработает путь
			fileURL := "/docs/view/" + entry.Path // ← УБРАЛИ url.QueryEscape!
			buf.WriteString(fmt.Sprintf("%s  <a href=\"%s\" class=\"flex items-center text-gray-700 hover:text-blue-600 transition-colors py-1\">\n", indent, fileURL))
			buf.WriteString(fmt.Sprintf("%s    <svg class=\"mr-2 w-4 h-4\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\">\n", indent))
			buf.WriteString(fmt.Sprintf("%s      <path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z\"></path>\n", indent))
			buf.WriteString(fmt.Sprintf("%s    </svg>\n", indent))
			buf.WriteString(fmt.Sprintf("%s    <span>%s</span>\n", indent, templ.EscapeString(entry.Name)))
			buf.WriteString(fmt.Sprintf("%s  </a>\n", indent))
		}

		buf.WriteString(fmt.Sprintf("%s</li>\n", indent))
	}
	buf.WriteString("</ul>\n")

	return buf.String()
}

func (h *PagesHandler) docsList(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	h.UpdateUserInfo(email, c)

	folder := c.Query("folder", "")
	baseDir := "./md_files"
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

	entries, err := h.buildFileTreeForDirectory(currentDir, folder)
	if err != nil {
		h.customLogger.Error().Msgf("Error building file tree: %v", err)
		entries = []dt.FileEntry{}
	}

	treeHTML := h.renderFileTreeHTML(entries, folder)

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

	// ИСПРАВЛЕНО: правильный путь к файлам
	filePath := filepath.Join("./md_files", filename)

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		h.customLogger.Error().Msgf("Invalid file path: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid file path")
	}

	contentDir, _ := filepath.Abs("./md_files")
	if !strings.HasPrefix(absPath, contentDir) {
		h.customLogger.Warn().Msgf("Attempted directory traversal: %s", filePath)
		return c.Status(fiber.StatusForbidden).SendString("Access denied")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.customLogger.Warn().Msgf("File not found: %s", filePath)
		return c.Status(fiber.StatusNotFound).SendString("Document not found")
	}

	htmlContent, err := h.getCachedMarkdown(filePath)
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

	isAdmin := h.repository.IsAdmin(email)

	component := views.MarkdownPage(views.MarkdownPageProps{
		Title:    title,
		Content:  htmlContent,
		FilePath: filename,
		IsAdmin:  isAdmin,
	})

	return templeadapter.Render(c, component, http.StatusOK)
}
