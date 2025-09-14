package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/GkadyrG/L2/L2.16/internal/models"
	"github.com/GkadyrG/L2/L2.16/internal/storage"
)

type Downloader struct {
	client  *http.Client
	storage *storage.Storage
}

func New(s *storage.Storage, timeout time.Duration) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: timeout,
		},
		storage: s,
	}
}

func (d *Downloader) Storage() *storage.Storage {
	return d.storage
}

// Download пытается вернуть ресурс из storage (кеш), иначе скачивает и сохраняет.
func (d *Downloader) Download(ctx context.Context, u *url.URL) (*models.Resource, bool, error) {
	// проверяем кеш
	if r, ok := d.storage.Get(u.String()); ok {
		return r, true, nil
	}
	// загрузка
	b, ctype, err := d.GetContent(ctx, u)
	if err != nil {
		return nil, false, err
	}
	res := storage.NewResource(u, b, ctype)
	if err := d.storage.Save(res); err != nil {
		return nil, false, err
	}
	return res, false, nil
}

func (d *Downloader) GetContent(ctx context.Context, u *url.URL) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("http status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return body, resp.Header.Get("Content-Type"), nil
}
