package storage

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/GkadyrG/L2/L2.16/internal/models"
)

// Storage отвечает за сохранение ресурсов на диск и за карту скачанных URL'ов.
type Storage struct {
	baseDir   string
	mu        sync.Mutex
	resources map[string]*models.Resource // ключ: URL.String()
}

// NewStorage создаёт storage и директорию базового каталога.
func NewStorage(baseDir string) *Storage {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		log.Fatalf("failed to create output dir: %v", err)
	}
	return &Storage{
		baseDir:   baseDir,
		resources: make(map[string]*models.Resource),
	}
}

// MakeLocalPath переводит u в относительный локальный путь (host/path).
// Пример: https://example.com/ -> example.com/index.html
func MakeLocalPath(u *url.URL) string {
	p := u.Path
	if p == "" || strings.HasSuffix(p, "/") {
		p = filepath.Join(p, "index.html")
	} else if filepath.Ext(p) == "" {
		// если нет расширения, считаем HTML
		p = p + ".html"
	}
	// корректируем: убрать ведущий '/'
	p = strings.TrimPrefix(p, "/")
	rel := filepath.Join(u.Host, p)
	rel = filepath.ToSlash(rel)
	// ограничение длины (предотвратить проблемы с очень длинными именами)
	if len(rel) > 300 {
		rel = rel[:300]
	}
	return rel
}

func NewResource(u *url.URL, content []byte, contentType string) *models.Resource {
	return &models.Resource{
		URL:         u,
		LocalPath:   MakeLocalPath(u),
		ContentType: contentType,
		Content:     content,
		IsHTML:      strings.Contains(contentType, "text/html"),
	}
}

// Save сохраняет ресурс на диск и регистрирует его.
// Если ресурс уже сохранён — просто возвращает nil.
func (s *Storage) Save(r *models.Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := r.URL.String()
	if _, ok := s.resources[key]; ok {
		return nil
	}
	full := filepath.Join(s.baseDir, filepath.FromSlash(r.LocalPath))
	if _, err := os.Stat(full); err == nil {
		// файл уже есть на диске — всё равно зарегистрируем
		s.resources[key] = r
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(full, r.Content, 0o644); err != nil {
		return err
	}
	s.resources[key] = r
	// also register trimmed slash version
	if !strings.HasSuffix(key, "/") {
		s.resources[strings.TrimSuffix(key, "/")] = r
	}
	return nil
}

// Get возвращает ресурс по URL (если есть в map).
func (s *Storage) Get(urlKey string) (*models.Resource, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.resources[urlKey]; ok {
		return r, true
	}
	if r, ok := s.resources[strings.TrimSuffix(urlKey, "/")]; ok {
		return r, true
	}
	return nil, false
}
