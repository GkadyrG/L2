package internal_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/GkadyrG/L2/L2.16/internal/downloader"
	"github.com/GkadyrG/L2/L2.16/internal/storage"
)

func TestDownloaderDownload(t *testing.T) {
	// test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ok" {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html>OK</html>"))
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	tmp := t.TempDir()
	s := storage.NewStorage(tmp)
	d := downloader.New(s, 5*time.Second)

	u, _ := url.Parse(ts.URL + "/ok")
	ctx := context.Background()
	res, fromCache, err := d.Download(ctx, u)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}
	if fromCache {
		t.Fatalf("should not be from cache first time")
	}
	if res == nil || res.Content == nil {
		t.Fatalf("invalid resource")
	}
	// second download should be from cache
	res2, fromCache2, err := d.Download(ctx, u)
	if err != nil {
		t.Fatalf("download 2 failed: %v", err)
	}
	if !fromCache2 {
		t.Fatalf("expected from cache")
	}
	if res2.LocalPath != res.LocalPath {
		t.Fatalf("local path mismatch")
	}
}
