package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mephalrith/noodles/backend/internal/config"
	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/model"
	"github.com/mephalrith/noodles/backend/internal/services"
)

func init() {
	services.InitLogger(&config.Config{})
	services.InitMetrics()
}

// fakeRegistry is an in-memory ClientRegistry for the resolution tests.
type fakeRegistry struct {
	approved []model.ClientEntry
	pending  []model.ClientEntry
	full     bool
	addErr   error
	adds     int

	approveErr error
	rejectErr  error
	approvals  []model.ClientEntry
	rejected   []string
}

func (f *fakeRegistry) Lookup(email string) (*model.ClientEntry, bool) {
	for i, e := range f.approved {
		if strings.EqualFold(e.Email, email) {
			return &f.approved[i], true
		}
	}
	return nil, false
}

func (f *fakeRegistry) IsPending(email string) bool {
	for _, e := range f.pending {
		if strings.EqualFold(e.Email, email) {
			return true
		}
	}
	return false
}

func (f *fakeRegistry) ListApproved() []model.ClientEntry { return f.approved }

func (f *fakeRegistry) ListPending() []model.ClientEntry { return f.pending }

func (f *fakeRegistry) AddPending(_ context.Context, entry model.ClientEntry) error {
	if f.full {
		return errs.PendingFull
	}
	if f.addErr != nil {
		return f.addErr
	}
	f.adds++
	f.pending = append(f.pending, entry)
	return nil
}

func (f *fakeRegistry) Approve(_ context.Context, email string, role model.Role) error {
	if f.approveErr != nil {
		return f.approveErr
	}
	f.approvals = append(f.approvals, model.ClientEntry{Email: email, Role: role})
	return nil
}

func (f *fakeRegistry) Reject(_ context.Context, email string) error {
	if f.rejectErr != nil {
		return f.rejectErr
	}
	f.rejected = append(f.rejected, email)
	return nil
}

func (f *fakeRegistry) Refresh(context.Context) error { return nil }

func testConfig() *config.Config {
	return &config.Config{
		Auth: config.AuthConfig{
			AdminGroups:   []string{"noodles-org:admin"},
			AllowedGroups: []string{"noodles-org:admin", "noodles-org:developer"},
		},
	}
}

func googleClaims(email string, verified any) map[string]any {
	return map[string]any{
		"sub":            "google-" + email,
		"email":          email,
		"email_verified": verified,
		"name":           "Test User",
	}
}

func TestResolveIdentityStaff(t *testing.T) {
	cases := []struct {
		name   string
		groups []string
		want   model.Role
	}{
		{"admin group", []string{"noodles-org:admin"}, model.RoleAdmin},
		{"developer group", []string{"noodles-org:developer"}, model.RoleViewer},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reg := &fakeRegistry{}
			user, err := resolveIdentity(context.Background(), googleClaims("staff@example.com", true), tc.groups, testConfig(), reg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.Role != tc.want {
				t.Fatalf("expected role %q, got %q", tc.want, user.Role)
			}
			if reg.adds != 0 {
				t.Fatal("staff login must not create a pending entry")
			}
		})
	}
}

func TestResolveIdentityStaffWinsOverClient(t *testing.T) {
	reg := &fakeRegistry{approved: []model.ClientEntry{{Email: "staff@example.com", Role: model.RoleClientAdmin}}}

	user, err := resolveIdentity(context.Background(), googleClaims("staff@example.com", true), []string{"noodles-org:admin"}, testConfig(), reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Role != model.RoleAdmin {
		t.Fatalf("expected staff role to win, got %q", user.Role)
	}
}

func TestResolveIdentityApprovedClient(t *testing.T) {
	cases := []struct {
		name  string
		entry model.ClientEntry
		want  model.Role
	}{
		{"client", model.ClientEntry{Email: "bob@example.com", Role: model.RoleClient}, model.RoleClient},
		{"client admin", model.ClientEntry{Email: "bob@example.com", Role: model.RoleClientAdmin}, model.RoleClientAdmin},
		{"unexpected role downgrades", model.ClientEntry{Email: "bob@example.com", Role: model.RoleAdmin}, model.RoleClient},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reg := &fakeRegistry{approved: []model.ClientEntry{tc.entry}}
			user, err := resolveIdentity(context.Background(), googleClaims("Bob@Example.com", true), nil, testConfig(), reg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.Role != tc.want {
				t.Fatalf("expected role %q, got %q", tc.want, user.Role)
			}
			if reg.adds != 0 {
				t.Fatal("approved client must not create a pending entry")
			}
		})
	}
}

func TestResolveIdentityUnknownCreatesPending(t *testing.T) {
	reg := &fakeRegistry{}
	ctx := context.Background()

	user, err := resolveIdentity(ctx, googleClaims("new@example.com", true), nil, testConfig(), reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Role != model.RolePending {
		t.Fatalf("expected pending role, got %q", user.Role)
	}
	if reg.adds != 1 {
		t.Fatalf("expected 1 pending entry, got %d", reg.adds)
	}
	if reg.pending[0].CreatedAt == "" {
		t.Fatal("expected createdAt to be set")
	}

	// A second sign-in must not duplicate the request.
	if _, err = resolveIdentity(ctx, googleClaims("NEW@example.com", true), nil, testConfig(), reg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg.adds != 1 {
		t.Fatalf("expected no duplicate pending entry, got %d", reg.adds)
	}
}

func TestResolveIdentityUnverifiedEmail(t *testing.T) {
	cases := []struct {
		name   string
		claims map[string]any
	}{
		{"explicitly false", googleClaims("new@example.com", false)},
		{"missing claim", map[string]any{"sub": "x", "email": "new@example.com"}},
		{"missing email", googleClaims("", true)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reg := &fakeRegistry{}
			_, err := resolveIdentity(context.Background(), tc.claims, nil, testConfig(), reg)
			if !errors.Is(err, errs.EmailUnverified) {
				t.Fatalf("expected errs.EmailUnverified, got %v", err)
			}
			if reg.adds != 0 {
				t.Fatal("no pending entry may be written for an unverified email")
			}
		})
	}
}

func TestResolveIdentityRequestsClosed(t *testing.T) {
	reg := &fakeRegistry{full: true}

	_, err := resolveIdentity(context.Background(), googleClaims("new@example.com", true), nil, testConfig(), reg)
	if !errors.Is(err, errs.RequestsClosed) {
		t.Fatalf("expected errs.RequestsClosed, got %v", err)
	}
}

func TestResolveIdentityAlreadyPending(t *testing.T) {
	reg := &fakeRegistry{pending: []model.ClientEntry{{Email: "wait@example.com", Role: model.RolePending}}}

	user, err := resolveIdentity(context.Background(), googleClaims("wait@example.com", true), nil, testConfig(), reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Role != model.RolePending {
		t.Fatalf("expected pending role, got %q", user.Role)
	}
	if reg.adds != 0 {
		t.Fatal("existing pending user must not be re-added")
	}
}

func TestResolveIdentityNoRegistry(t *testing.T) {
	_, err := resolveIdentity(context.Background(), googleClaims("new@example.com", true), nil, testConfig(), nil)
	if !errors.Is(err, errs.NotAuthorized) {
		t.Fatalf("expected errs.NotAuthorized, got %v", err)
	}
}
