package parser

import (
	"bytes"
	"errors"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

var linkAttrs = map[string][]string{
	"a":      {"href"},
	"link":   {"href"},
	"img":    {"src", "srcset"},
	"script": {"src"},
	"iframe": {"src"},
	"source": {"src", "srcset"},
	"video":  {"src", "poster"},
	"audio":  {"src"},
	"embed":  {"src"},
}

var (
	ErrUnsupported = errors.New("unsupported scheme or empty")
	ErrExternal    = errors.New("external link")
)

type Parser struct {
	Base *url.URL
}

func New(base *url.URL) *Parser {
	return &Parser{Base: base}
}

// ExtractLinks парсит htmlContent и возвращает список нормализованных URL
// которые принадлежат тому же хосту (или субдоменам).
func (p *Parser) ExtractLinks(htmlContent []byte) ([]*url.URL, error) {
	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var links []*url.URL

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if attrs, ok := linkAttrs[n.Data]; ok {
				for _, a := range attrs {
					val := getAttr(n, a)
					if val == "" {
						continue
					}
					if a == "srcset" {
						for _, raw := range parseSrcset(val) {
							if u, err := p.Normalize(raw); err == nil {
								if _, ok := seen[u.String()]; !ok {
									seen[u.String()] = struct{}{}
									links = append(links, u)
								}
							}
						}
						continue
					}
					if u, err := p.Normalize(val); err == nil {
						if _, ok := seen[u.String()]; !ok {
							seen[u.String()] = struct{}{}
							links = append(links, u)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return links, nil
}

// Normalize берет raw URL из атрибута и преобразует в абсолютный URL,
// проверяет схему и принадлежность домену (субдомены допускаются).
func (p *Parser) Normalize(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrUnsupported
	}
	if strings.HasPrefix(raw, "mailto:") || strings.HasPrefix(raw, "javascript:") || strings.HasPrefix(raw, "tel:") {
		return nil, ErrUnsupported
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	// protocol-relative
	if u.Scheme == "" && strings.HasPrefix(raw, "//") {
		u.Scheme = p.Base.Scheme
	}
	// relative -> resolve
	if u.Hostname() == "" {
		u = p.Base.ResolveReference(u)
	}
	// check same domain or subdomain
	if !isSameDomain(u.Hostname(), p.Base.Hostname()) {
		return nil, ErrExternal
	}
	// normalize: remove fragment and query for deduplication
	u.Fragment = ""
	u.RawQuery = ""
	return u, nil
}

func getAttr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

func parseSrcset(s string) []string {
	var out []string
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// URL возможно идет первым, затем дескриптор (1x, 2x и тп)
		if fields := strings.Fields(p); len(fields) > 0 {
			out = append(out, fields[0])
		}
	}
	return out
}

func isSameDomain(host, base string) bool {
	host = strings.ToLower(host)
	base = strings.ToLower(base)
	if host == base {
		return true
	}
	return strings.HasSuffix(host, "."+base)
}

// LinkAttrsFor возвращает интересующие атрибуты для тега (используется при переписывании ссылок)
func LinkAttrsFor(tag string) ([]string, bool) {
	a, ok := linkAttrs[tag]
	return a, ok
}
