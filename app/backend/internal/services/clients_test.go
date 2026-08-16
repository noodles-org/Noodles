package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/mephalrith/noodles/backend/internal/config"
	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/model"
)

const testNamespace = "dashboard"

func init() {
	InitLogger(&config.Config{})
}

func newTestRegistry(t *testing.T, approved, pending string) *K8sClientRegistry {
	t.Helper()

	client := fake.NewSimpleClientset(
		configMap(ApprovedConfigMap, approved),
		configMap(PendingConfigMap, pending),
	)

	reg := NewClientRegistryWithClient(client, testNamespace)
	if err := reg.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	return reg
}

func configMap(name, data string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace},
		Data:       map[string]string{clientsKey: data},
	}
}

func storedPending(t *testing.T, reg *K8sClientRegistry) []model.ClientEntry {
	t.Helper()
	entries, err := reg.load(context.Background(), PendingConfigMap)
	if err != nil {
		t.Fatalf("load pending: %v", err)
	}
	return entries
}

func TestLookupParsesApprovedYAML(t *testing.T) {
	reg := newTestRegistry(t, `
- email: bob@example.com
  name: Bob
  sub: google|1
  role: client_admin
  createdAt: "2026-01-01T00:00:00Z"
`, "[]")

	entry, ok := reg.Lookup("BOB@Example.com ")
	if !ok {
		t.Fatal("expected approved entry for bob@example.com")
	}
	if entry.Role != model.RoleClientAdmin {
		t.Fatalf("expected client_admin, got %q", entry.Role)
	}

	if _, ok := reg.Lookup("nobody@example.com"); ok {
		t.Fatal("expected no entry for unknown email")
	}
}

func TestMalformedYAMLDegradesToEmpty(t *testing.T) {
	reg := newTestRegistry(t, "not: [a valid, list", "")

	if got := len(reg.ListApproved()); got != 0 {
		t.Fatalf("expected empty registry, got %d entries", got)
	}
	if got := len(reg.ListPending()); got != 0 {
		t.Fatalf("expected empty pending list, got %d entries", got)
	}
}

func TestAddPendingIsIdempotent(t *testing.T) {
	reg := newTestRegistry(t, "[]", "[]")
	ctx := context.Background()

	entry := model.ClientEntry{Email: "New@example.com", Name: "New", Sub: "google|9"}
	if err := reg.AddPending(ctx, entry); err != nil {
		t.Fatalf("first AddPending: %v", err)
	}
	if err := reg.AddPending(ctx, entry); err != nil {
		t.Fatalf("second AddPending: %v", err)
	}

	stored := storedPending(t, reg)
	if len(stored) != 1 {
		t.Fatalf("expected 1 pending entry, got %d", len(stored))
	}
	if stored[0].Email != "new@example.com" {
		t.Fatalf("expected normalized email, got %q", stored[0].Email)
	}
	if stored[0].Role != model.RolePending {
		t.Fatalf("expected pending role, got %q", stored[0].Role)
	}
	if stored[0].CreatedAt == "" {
		t.Fatal("expected createdAt to be set")
	}
	if !reg.IsPending("NEW@example.com") {
		t.Fatal("expected cache to report the email as pending")
	}
}

func TestAddPendingEnforcesCap(t *testing.T) {
	reg := newTestRegistry(t, "[]", "[]")
	ctx := context.Background()

	for i := 0; i < PendingCap; i++ {
		email := fmt.Sprintf("user%d@example.com", i)
		if err := reg.AddPending(ctx, model.ClientEntry{Email: email}); err != nil {
			t.Fatalf("AddPending %d: %v", i, err)
		}
	}

	err := reg.AddPending(ctx, model.ClientEntry{Email: "overflow@example.com"})
	if !errors.Is(err, errs.PendingFull) {
		t.Fatalf("expected errs.PendingFull, got %v", err)
	}

	stored := storedPending(t, reg)
	if len(stored) != PendingCap {
		t.Fatalf("expected %d pending entries, got %d", PendingCap, len(stored))
	}
}

func TestApproveMovesEntry(t *testing.T) {
	reg := newTestRegistry(t, "[]", `
- email: bob@example.com
  name: Bob
  sub: google|1
  role: pending
`)
	ctx := context.Background()

	if err := reg.Approve(ctx, "Bob@Example.com", model.RoleClient); err != nil {
		t.Fatalf("approve: %v", err)
	}

	if len(storedPending(t, reg)) != 0 {
		t.Fatal("expected pending list to be empty after approval")
	}

	entry, ok := reg.Lookup("bob@example.com")
	if !ok {
		t.Fatal("expected approved entry after approval")
	}
	if entry.Role != model.RoleClient {
		t.Fatalf("expected client role, got %q", entry.Role)
	}
	if entry.Sub != "google|1" {
		t.Fatalf("expected sub to be carried over, got %q", entry.Sub)
	}
}

func TestApproveRejectsUnknownEmailAndBadRole(t *testing.T) {
	reg := newTestRegistry(t, "[]", `
- email: bob@example.com
  role: pending
`)
	ctx := context.Background()

	if err := reg.Approve(ctx, "ghost@example.com", model.RoleClient); !errors.Is(err, errs.NotPending) {
		t.Fatalf("expected errs.NotPending, got %v", err)
	}
	if err := reg.Approve(ctx, "bob@example.com", model.RoleAdmin); !errors.Is(err, errs.InvalidClientRole) {
		t.Fatalf("expected errs.InvalidClientRole, got %v", err)
	}
	if len(storedPending(t, reg)) != 1 {
		t.Fatal("expected pending list to be untouched")
	}
}

func TestRejectRemovesEntry(t *testing.T) {
	reg := newTestRegistry(t, "[]", `
- email: bob@example.com
  role: pending
- email: ann@example.com
  role: pending
`)
	ctx := context.Background()

	if err := reg.Reject(ctx, "bob@example.com"); err != nil {
		t.Fatalf("reject: %v", err)
	}

	stored := storedPending(t, reg)
	if len(stored) != 1 || stored[0].Email != "ann@example.com" {
		t.Fatalf("unexpected pending list after reject: %+v", stored)
	}
	if err := reg.Reject(ctx, "bob@example.com"); !errors.Is(err, errs.NotPending) {
		t.Fatalf("expected errs.NotPending, got %v", err)
	}
}
