package mdfiles

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"sitex/config"
	"sitex/internal/dt"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type cachedMarkdown struct {
	htmlContent  string
	lastModified time.Time
}

type MdFilesService struct {
	// Кэш для скомпилированных документов
	markdownCache  map[string]cachedMarkdown
	cacheMutex     sync.RWMutex
	MdFilesBaseDir string
}

func NewMdFilesService(mdFilesConfig *config.MdFilesConfig) *MdFilesService {
	return &MdFilesService{
		markdownCache:  make(map[string]cachedMarkdown),
		cacheMutex:     sync.RWMutex{},
		MdFilesBaseDir: mdFilesConfig.Path,
	}
}

// Генерация HTML дерева документов с поддержкой папок
func (s *MdFilesService) RenderFileTreeHTML(entries []dt.FileEntry, currentFolder string) string {
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

// Строим дерево только для указанной директории (без рекурсии вглубь)
func (s *MdFilesService) BuildFileTreeForDirectory(dir, folderPrefix string) ([]dt.FileEntry, error) {
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

func (s *MdFilesService) convertMarkdownToHTML(content []byte) string {
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

// Получение и кэширование контента
func (s *MdFilesService) GetCachedMarkdown(filePath string) (string, error) {
	s.cacheMutex.RLock()
	cached, exists := s.markdownCache[filePath]
	s.cacheMutex.RUnlock()

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
	htmlContent := s.convertMarkdownToHTML(content)

	// Сохраняем в кэш
	fileInfo, _ := os.Stat(filePath)
	s.cacheMutex.Lock()
	s.markdownCache[filePath] = cachedMarkdown{
		htmlContent:  htmlContent,
		lastModified: fileInfo.ModTime(),
	}
	s.cacheMutex.Unlock()

	return htmlContent, nil
}
