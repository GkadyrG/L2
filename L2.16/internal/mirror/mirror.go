package mirror

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GkadyrG/L2/L2.16/internal/downloader"
	"github.com/GkadyrG/L2/L2.16/internal/models"
	"github.com/GkadyrG/L2/L2.16/internal/parser"
	"github.com/GkadyrG/L2/L2.16/internal/storage"
	"golang.org/x/net/html"
)

// Options — параметры запуска
type Options struct {
	RootURL   string
	MaxDepth  int
	OutputDir string
	Workers   int
	Timeout   time.Duration
}

// Run — основной запуск процесса зеркалирования
func Run(opts Options) error {
	parsed, err := url.Parse(opts.RootURL)
	if err != nil {
		return fmt.Errorf("invalid root url: %w", err)
	}
	// ensure scheme
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("only http/https supported")
	}

	store := storage.NewStorage(opts.OutputDir)
	down := downloader.New(store, opts.Timeout)
	parserObj := parser.New(parsed)

	// channels and worker pool
	tasks := make(chan *task, 1000)
	errs := make(chan error, 100)
	var wg sync.WaitGroup

	visited := make(map[string]struct{})
	var vmu sync.Mutex

	enqueue := func(u *url.URL, depth int) bool {
		key := u.String()
		vmu.Lock()
		_, ok := visited[key]
		if ok {
			vmu.Unlock()
			return false
		}
		visited[key] = struct{}{}
		vmu.Unlock()
		wg.Add(1)
		tasks <- &task{U: u, Depth: depth}
		return true
	}

	// seed
	enqueue(parsed, 0)

	// context to allow cancellation later if need
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start workers
	nWorkers := opts.Workers
	if nWorkers <= 0 {
		nWorkers = 1
	}
	for i := 0; i < nWorkers; i++ {
		go worker(ctx, &wg, tasks, errs, down, parserObj, enqueue, opts.MaxDepth)
	}

	// wait for completion (wg)
	go func() {
		wg.Wait()
		close(tasks)
	}()

	// collect errors and block until all done or fatal
	var firstErr error
	fin := make(chan struct{})
	go func() {
		for e := range errs {
			if firstErr == nil {
				firstErr = e
				// не прерываем сразу, но можно
			}
		}
		close(fin)
	}()

	// wait for tasks to end: when wg done, tasks closed, and errs closed
	wg.Wait()
	// workers may still be sending to errs, wait a brief moment then close
	close(errs)
	<-fin

	return firstErr
}

type task struct {
	U     *url.URL
	Depth int
}

func worker(ctx context.Context, wg *sync.WaitGroup, tasks <-chan *task, errs chan<- error, dl *downloader.Downloader, p *parser.Parser, enqueue func(*url.URL, int) bool, maxDepth int) {
	for t := range tasks {
		processOne(ctx, t, errs, dl, p, enqueue, maxDepth)
		wg.Done()
	}
}

func processOne(ctx context.Context, t *task, errs chan<- error, dl *downloader.Downloader, p *parser.Parser, enqueue func(*url.URL, int) bool, maxDepth int) {
	// скачиваем
	res, fromCache, err := dl.Download(ctx, t.U)
	if err != nil {
		// если 404 на вложенных — можно пропустить, здесь пишем в errs и возвращаемся
		errs <- fmt.Errorf("failed download %s: %w", t.U.String(), err)
		return
	}
	// если из кеша — можно не парсить повторно
	if fromCache && (t.Depth >= maxDepth || !res.IsHTML) {
		return
	}

	// если HTML и можем углубляться — парсим и ставим задачи
	if res.IsHTML && t.Depth < maxDepth {
		links, err := p.ExtractLinks(res.Content)
		if err != nil {
			errs <- fmt.Errorf("parse %s: %w", t.U.String(), err)
			return
		}
		// перед тем как ставить задачи — переписываем HTML ссылки локальными путями
		if err := rewriteHTML(res, links, p, dl, enqueue); err != nil {
			errs <- fmt.Errorf("rewrite %s: %w", t.U.String(), err)
			// продолжим после ошибки
		}
		// теперь ставим найденные ссылки
		for _, l := range links {
			enqueue(l, t.Depth+1)
		}
	}
}

// rewriteHTML переписывает ссылки в HTML-файле res.Content на локальные пути.
// local path вычисляется по storage.MakeLocalPath; relative ссылки формируются относительно текущего HTML.
func rewriteHTML(res *models.Resource, links []*url.URL, p *parser.Parser, dl *downloader.Downloader, enqueue func(*url.URL, int) bool) error {
	// создаём map from abs URL string -> local path
	mapping := map[string]string{}
	for _, u := range links {
		// если ресурс уже сохранён — возьмём его LocalPath
		if r, ok := dl.Storage().Get(u.String()); ok { // need access to storage; implement getter
			mapping[u.String()] = r.LocalPath
		} else {
			// рассчитаем путь заранее (не сохраняем, просто составляем путь)
			mapping[u.String()] = storage.MakeLocalPath(u)
		}
	}
	// парсим HTML и заменяем атрибуты
	doc, err := html.Parse(bytes.NewReader(res.Content))
	if err != nil {
		return err
	}
	var changed bool
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if attrs, ok := parser.LinkAttrsFor(n.Data); ok {
				for i, a := range n.Attr {
					for _, interest := range attrs {
						if a.Key != interest {
							continue
						}
						orig := strings.TrimSpace(a.Val)
						if orig == "" {
							continue
						}
						if interest == "srcset" {
							newVal := replaceSrcset(orig, mapping, p)
							if newVal != orig {
								n.Attr[i].Val = newVal
								changed = true
							}
							continue
						}
						// Resolve orig to absolute URL (use p.Normalize but it rejects external => use Parse+Resolve)
						u, err := url.Parse(orig)
						if err != nil {
							continue
						}
						if u.Hostname() == "" {
							u = p.Base.ResolveReference(u)
						}
						if local, ok := mapping[u.String()]; ok {
							// calculate relative path from current HTML to target local path
							rel := relativePath(res.LocalPath, local)
							n.Attr[i].Val = rel
							changed = true
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
	if !changed {
		return nil
	}
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return err
	}
	res.Content = buf.Bytes()
	// перезаписываем модифицированный HTML на диск
	return dl.Storage().Save(res)
}

// helper: relativePath from current local path (like "example.com/path/index.html") to target local path
func relativePath(current, target string) string {
	// use filepath.Rel then convert to slash
	curDir := filepath.Dir(current)
	rel, err := filepath.Rel(curDir, target)
	if err != nil {
		// fallback to target
		return filepath.ToSlash(target)
	}
	return filepath.ToSlash(rel)
}

// replaceSrcset заменяет все URL в srcset если они есть в mapping.
func replaceSrcset(orig string, mapping map[string]string, p *parser.Parser) string {
	parts := strings.Split(orig, ",")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Fields(part)
		if len(fields) == 0 {
			continue
		}
		rawURL := fields[0]
		u, err := url.Parse(rawURL)
		if err != nil {
			continue
		}
		if u.Hostname() == "" {
			u = p.Base.ResolveReference(u)
		}
		if local, ok := mapping[u.String()]; ok {
			fields[0] = local
			parts[i] = strings.Join(fields, " ")
		}
	}
	return strings.Join(parts, ", ")
}
