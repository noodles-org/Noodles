package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var fileActions = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "filesvc_actions_total",
	Help: "Total file operations handled, by action and outcome",
}, []string{"action", "status"})

type config struct {
	Root  string
	Token string
	Port  string
}

func loadConfig() config {
	cfg := config{
		Root:  os.Getenv("FILESVC_ROOT"),
		Token: os.Getenv("FILESVC_TOKEN"),
		Port:  os.Getenv("FILESVC_PORT"),
	}
	if cfg.Root == "" {
		cfg.Root = "/foundrydata/Data"
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	return cfg
}

func requireToken(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func instrument(action string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)
		outcome := "ok"
		if rec.status >= 400 {
			outcome = "error"
		}
		fileActions.WithLabelValues(action, outcome).Inc()
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func methodGuard(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		next(w, r)
	}
}

func main() {
	cfg := loadConfig()
	if cfg.Token == "" {
		log.Fatal("FILESVC_TOKEN is required")
	}

	j, err := newRoot(cfg.Root)
	if err != nil {
		log.Fatalf("failed to open root %q: %v", cfg.Root, err)
	}
	srv := &server{root: j}

	mux := http.NewServeMux()

	protect := func(action, method string, h http.HandlerFunc) http.HandlerFunc {
		return requireToken(cfg.Token, methodGuard(method, instrument(action, h)))
	}

	mux.HandleFunc("/files", requireToken(cfg.Token, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			instrument("list", srv.handleList)(w, r)
		case http.MethodDelete:
			instrument("delete", srv.handleDelete)(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}))
	mux.HandleFunc("/download", protect("download", http.MethodGet, srv.handleDownload))
	mux.HandleFunc("/upload", protect("upload", http.MethodPost, srv.handleUpload))
	mux.HandleFunc("/mkdir", protect("mkdir", http.MethodPost, srv.handleMkdir))
	mux.HandleFunc("/rename", protect("rename", http.MethodPost, srv.handleRename))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/metrics", promhttp.Handler())

	log.Printf("file-sidecar listening on :%s (root=%s)", cfg.Port, j.Path)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
