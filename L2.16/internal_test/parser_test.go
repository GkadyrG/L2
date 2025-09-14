package internal_test

import (
	"net/url"
	"testing"

	"github.com/GkadyrG/L2/L2.16/internal/parser"
)

func TestExtractLinksBasic(t *testing.T) {
	base, _ := url.Parse("https://example.com")
	p := parser.New(base)
	html := `<html><head><link href="/style.css"></head><body>
	<a href="/a">A</a>
	<img src="/img.png" srcset="/img-1x.png 1x, /img-2x.png 2x">
	<a href="https://other.com/x">external</a>
	</body></html>`
	links, err := p.ExtractLinks([]byte(html))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]bool{
		"https://example.com/style.css":  true,
		"https://example.com/a":          true,
		"https://example.com/img.png":    true,
		"https://example.com/img-1x.png": true,
		"https://example.com/img-2x.png": true,
	}
	if len(links) != len(want) {
		t.Fatalf("expected %d links, got %d", len(want), len(links))
	}
	for _, u := range links {
		if !want[u.String()] {
			t.Errorf("unexpected link: %s", u.String())
		}
	}
}
