package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/model"
)

func postJSON(t *testing.T, h http.HandlerFunc, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h(w, httptest.NewRequest(http.MethodPost, "/api/clients/approve", strings.NewReader(body)))
	return w
}

func TestHandleListClients(t *testing.T) {
	reg := &fakeRegistry{
		approved: []model.ClientEntry{{Email: "a@example.com", Role: model.RoleClient}},
		pending:  []model.ClientEntry{{Email: "b@example.com"}},
	}

	w := httptest.NewRecorder()
	HandleListApprovedClients(reg)(w, httptest.NewRequest(http.MethodGet, "/api/clients", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "a@example.com") {
		t.Fatalf("approved list: got %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	HandleListPendingClients(reg)(w, httptest.NewRequest(http.MethodGet, "/api/clients/pending", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "b@example.com") {
		t.Fatalf("pending list: got %d %s", w.Code, w.Body.String())
	}
}

func TestHandleApproveClient(t *testing.T) {
	reg := &fakeRegistry{}
	w := postJSON(t, HandleApproveClient(reg), `{"email":"b@example.com","role":"client_admin"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(reg.approvals) != 1 || reg.approvals[0].Role != model.RoleClientAdmin {
		t.Fatalf("unexpected approvals: %+v", reg.approvals)
	}
}

func TestHandleApproveClientValidation(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{"malformed body", `{`, http.StatusBadRequest},
		{"missing email", `{"role":"client"}`, http.StatusBadRequest},
		{"staff role", `{"email":"b@example.com","role":"admin"}`, http.StatusBadRequest},
		{"unknown role", `{"email":"b@example.com","role":"wizard"}`, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reg := &fakeRegistry{}
			w := postJSON(t, HandleApproveClient(reg), tc.body)
			if w.Code != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, w.Code)
			}
			if len(reg.approvals) != 0 {
				t.Fatal("invalid request must not approve anyone")
			}
		})
	}
}

func TestHandleApproveClientNotPending(t *testing.T) {
	reg := &fakeRegistry{approveErr: errs.NotPending}
	w := postJSON(t, HandleApproveClient(reg), `{"email":"ghost@example.com","role":"client"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleRejectClient(t *testing.T) {
	reg := &fakeRegistry{}
	w := postJSON(t, HandleRejectClient(reg), `{"email":"b@example.com"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(reg.rejected) != 1 || reg.rejected[0] != "b@example.com" {
		t.Fatalf("unexpected rejections: %+v", reg.rejected)
	}
}

func TestHandleRejectClientNotPending(t *testing.T) {
	reg := &fakeRegistry{rejectErr: errs.NotPending}
	w := postJSON(t, HandleRejectClient(reg), `{"email":"ghost@example.com"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
