package dt

type FileEntry struct {
	Name     string
	Path     string
	IsDir    bool
	Children []FileEntry
	IsLast   bool // ← ключевой флаг для CSS
}
