package middleware

import (
	"fmt"
	"net/http"
	"time"

	chilib "github.com/go-chi/chi/v5"

	"github.com/mephalrith/noodles/backend/internal/services"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		duration := time.Since(start).Seconds()
		route := chilib.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = r.URL.Path
		}

		services.HTTPRequests.With(map[string]string{
			"method":      r.Method,
			"route":       route,
			"status_code": fmt.Sprintf("%d", sw.status),
		}).Inc()

		services.HTTPDuration.With(map[string]string{
			"method": r.Method,
			"route":  route,
		}).Observe(duration)
	})
}
