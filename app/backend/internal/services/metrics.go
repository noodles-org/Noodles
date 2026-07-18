package services

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	AuthEvents = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dashboard_auth_events_total",
		Help: "Total authentication events",
	}, []string{"status", "reason"})

	UniqueUsersGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dashboard_unique_authenticated_users",
		Help: "Total unique users authenticated since last restart",
	})

	DeploymentActions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dashboard_deployment_actions_total",
		Help: "Deployment management actions",
	}, []string{"action", "namespace", "deployment"})

	UnauthorizedAttempts = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dashboard_unauthorized_access_attempts_total",
		Help: "Attempts to perform admin actions without admin role",
	}, []string{"path"})

	HTTPRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dashboard_http_requests_total",
		Help: "Total HTTP requests",
	}, []string{"method", "route", "status_code"})

	HTTPDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "dashboard_http_request_duration_seconds",
		Help:    "HTTP request duration",
		Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"method", "route"})

	uniqueUsers   = make(map[string]struct{})
	uniqueUsersMu sync.Mutex
)

func TrackUniqueUser(userID string) {
	uniqueUsersMu.Lock()
	defer uniqueUsersMu.Unlock()
	uniqueUsers[userID] = struct{}{}
	UniqueUsersGauge.Set(float64(len(uniqueUsers)))
}

func InitMetrics() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			Namespace: "dashboard",
		}),
		AuthEvents,
		UniqueUsersGauge,
		DeploymentActions,
		UnauthorizedAttempts,
		HTTPRequests,
		HTTPDuration,
	)
	return reg
}
