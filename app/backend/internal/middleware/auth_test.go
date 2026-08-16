package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mephalrith/noodles/backend/internal/config"
	"github.com/mephalrith/noodles/backend/internal/model"
	"github.com/mephalrith/noodles/backend/internal/services"
)

func init() {
	services.InitLogger(&config.Config{})
	services.InitMetrics()
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func requestAs(role model.Role) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/api/deployments", nil)
	user := &model.User{Email: "u@example.com", Role: role}
	return r.WithContext(context.WithValue(r.Context(), userContextKey, user))
}

func TestRequireApproved(t *testing.T) {
	tests := []struct {
		role model.Role
		want int
	}{
		{model.RoleAdmin, http.StatusOK},
		{model.RoleViewer, http.StatusOK},
		{model.RoleClientAdmin, http.StatusOK},
		{model.RoleClient, http.StatusOK},
		{model.RolePending, http.StatusForbidden},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		RequireApproved(okHandler()).ServeHTTP(w, requestAs(tt.role))
		if w.Code != tt.want {
			t.Errorf("role %s: got %d, want %d", tt.role, w.Code, tt.want)
		}
	}
}

func TestRequireApprovedWithoutUser(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/deployments", nil)
	RequireApproved(okHandler()).ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Errorf("no user: got %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRequireRoleMutating(t *testing.T) {
	tests := []struct {
		role model.Role
		want int
	}{
		{model.RoleAdmin, http.StatusOK},
		{model.RoleClientAdmin, http.StatusOK},
		{model.RoleClient, http.StatusForbidden},
		{model.RoleViewer, http.StatusForbidden},
		{model.RolePending, http.StatusForbidden},
	}

	guard := RequireRole(model.RoleAdmin, model.RoleClientAdmin)
	for _, tt := range tests {
		w := httptest.NewRecorder()
		guard(okHandler()).ServeHTTP(w, requestAs(tt.role))
		if w.Code != tt.want {
			t.Errorf("role %s: got %d, want %d", tt.role, w.Code, tt.want)
		}
	}
}
