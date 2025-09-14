package internal_test

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/GkadyrG/L2/L2.16/internal/storage"
)

func TestMakeLocalPathAndSave(t *testing.T) {
	tmp := t.TempDir()
	s := storage.NewStorage(tmp)
	u, _ := url.Parse("https://example.com/path")
	r := storage.NewResource(u, []byte("<html></html>"), "text/html")
	if err := s.Save(r); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	expected := filepath.Join(tmp, r.LocalPath)
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("file not found: %v", err)
	}
	// get by url
	if got, ok := s.Get(u.String()); !ok || got.LocalPath != r.LocalPath {
		t.Fatalf("get failed")
	}
}
