package models

import "net/url"

// Resource представляет скачанный ресурс и путь для локального сохранения.
type Resource struct {
	URL         *url.URL // исходный URL
	LocalPath   string   // относительный путь внутри output dir (host/...)
	ContentType string
	Content     []byte
	IsHTML      bool
}