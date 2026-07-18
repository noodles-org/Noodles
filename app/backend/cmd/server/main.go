package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	chilib "github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mephalrith/noodles/backend/internal/config"
	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/handlers"
	"github.com/mephalrith/noodles/backend/internal/middleware"
	"github.com/mephalrith/noodles/backend/internal/model"
	"github.com/mephalrith/noodles/backend/internal/respond"
	"github.com/mephalrith/noodles/backend/internal/services"
)

func main() {
	cfg := config.Load()
	services.InitLogger(cfg)
	reg := services.InitMetrics()

	k8s := services.NewK8sService(cfg)
	argo := services.NewArgoCDService(cfg)

	r := chilib.NewRouter()

	// Global middleware
	r.Use(chimw.RealIP)
	r.Use(securityHeaders)
	if !cfg.IsProduction {
		r.Use(middleware.CORS(cfg.CORSOrigin))
	}
	r.Use(chimw.Recoverer)
	r.Use(middleware.RequestLogger)

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		respond.OK(w, map[string]string{"status": "ok"})
	})

	// Auth routes (with rate limiting handled by middleware)
	r.Route("/api/auth", func(sub chilib.Router) {
		sub.Get("/login", handlers.HandleLogin(cfg))
		sub.Get("/callback", handlers.HandleCallback(cfg))
		sub.With(middleware.RequireAuth(cfg)).Get("/me", handlers.HandleMe)
		sub.With(middleware.RequireAuth(cfg)).Post("/logout", handlers.HandleLogout(cfg))
	})

	// Deployment routes
	r.Route("/api/deployments", func(sub chilib.Router) {
		sub.Use(middleware.RequireAuth(cfg))
		sub.Get("/", handlers.HandleListDeployments(k8s, argo))
		sub.With(middleware.RequireRole(model.RoleAdmin)).Post("/{namespace}/{name}/restart", handlers.HandleRestartDeployment(k8s))
		sub.With(middleware.RequireRole(model.RoleAdmin)).Post("/{namespace}/{name}/pause", handlers.HandlePauseDeployment(k8s))
		sub.With(middleware.RequireRole(model.RoleAdmin)).Post("/{namespace}/{name}/resume", handlers.HandleResumeDeployment(k8s))
	})

	// Docs routes
	r.Route("/api/docs", func(sub chilib.Router) {
		sub.Use(middleware.RequireAuth(cfg))
		sub.Get("/toc", handlers.HandleDocsToc(cfg))
		sub.Get("/content", handlers.HandleDocsContent(cfg))
	})

	// Services routes
	r.Route("/api/services", func(sub chilib.Router) {
		sub.Use(middleware.RequireAuth(cfg))
		sub.Get("/", handlers.HandleListServices(k8s))
	})

	// Serve frontend static files in production
	if cfg.IsProduction {
		frontendFS := http.Dir(cfg.FrontendPath)
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
				respond.Error(w, errs.NotFound)
				return
			}
			// Try to serve the static file; fall back to index.html for SPA routes
			if f, err := frontendFS.Open(r.URL.Path); err == nil {
				f.Close()
				http.FileServer(frontendFS).ServeHTTP(w, r)
				return
			}
			http.ServeFile(w, r, filepath.Join(cfg.FrontendPath, "index.html"))
		})
	} else {
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
				respond.Error(w, errs.NotFound)
				return
			}
			http.NotFound(w, r)
		})
	}

	// SIGHUP resets TOC cache
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP)
	go func() {
		for range sigCh {
			handlers.ResetTocCache()
			services.Logger.Info("TOC cache reset via SIGHUP")
		}
	}()

	// Metrics server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		services.Logger.Info(fmt.Sprintf("Metrics on :%d", cfg.MetricsPort))
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.MetricsPort), mux); err != nil {
			services.Logger.Error("Metrics server failed", "error", err)
		}
	}()

	// Main server
	services.Logger.Info(fmt.Sprintf("Server on :%d", cfg.Port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r); err != nil {
		services.Logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
		next.ServeHTTP(w, r)
	})
}
