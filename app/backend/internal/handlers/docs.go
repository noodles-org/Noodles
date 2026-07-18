package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/mephalrith/noodles/backend/internal/config"
	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/model"
	"github.com/mephalrith/noodles/backend/internal/respond"
	"github.com/mephalrith/noodles/backend/internal/services"
)

var (
	sectionRE = regexp.MustCompile(`^## (.+)`)
	itemRE    = regexp.MustCompile(`^- \[(.+?)]\((.+?)(?:#.+?)?\)`)
)

type tocCache struct {
	mu       sync.RWMutex
	sections []model.DocTocSection
	loaded   bool
}

var toc tocCache

func parseTocMarkdown(md string) []model.DocTocSection {
	var sections []model.DocTocSection
	var current *model.DocTocSection

	for _, line := range strings.Split(md, "\n") {
		if m := sectionRE.FindStringSubmatch(line); m != nil {
			sections = append(sections, model.DocTocSection{Title: m[1]})
			current = &sections[len(sections)-1]
			continue
		}
		if m := itemRE.FindStringSubmatch(line); m != nil && current != nil {
			duplicate := false
			for _, item := range current.Items {
				if item.Path == m[2] {
					duplicate = true
					break
				}
			}
			if !duplicate {
				current.Items = append(current.Items, model.DocTocItem{Title: m[1], Path: m[2]})
			}
		}
	}

	return sections
}

func ResetTocCache() {
	toc.mu.Lock()
	defer toc.mu.Unlock()
	toc.loaded = false
	toc.sections = nil
}

func HandleDocsToc(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		toc.mu.RLock()
		if toc.loaded {
			sections := toc.sections
			toc.mu.RUnlock()
			respond.OK(w, map[string]any{"sections": sections})
			return
		}
		toc.mu.RUnlock()

		toc.mu.Lock()
		defer toc.mu.Unlock()

		if toc.loaded {
			respond.OK(w, map[string]any{"sections": toc.sections})
			return
		}

		data, err := os.ReadFile(filepath.Join(cfg.DocsPath, "toc.md"))
		if err != nil {
			services.Logger.Error("Failed to load table of contents", "error", err)
			respond.Error(w, errs.Internal("Failed to load table of contents"))
			return
		}

		toc.sections = parseTocMarkdown(string(data))
		toc.loaded = true

		respond.OK(w, map[string]any{"sections": toc.sections})
	}
}

func HandleDocsContent(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docPath := r.URL.Query().Get("path")
		if docPath == "" {
			respond.Error(w, errs.MissingPath)
			return
		}

		normalized := filepath.Clean(docPath)
		if strings.Contains(normalized, "..") || filepath.IsAbs(normalized) {
			respond.Error(w, errs.InvalidPath)
			return
		}

		full := filepath.Join(cfg.DocsPath, normalized)
		data, err := os.ReadFile(full)
		if err != nil {
			if os.IsNotExist(err) {
				respond.Error(w, errs.NotFound)
				return
			}
			respond.Error(w, errs.Internal("Failed to read document"))
			return
		}

		respond.OK(w, map[string]string{"content": string(data)})
	}
}
