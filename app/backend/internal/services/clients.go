package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/mephalrith/noodles/backend/internal/config"
	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/model"
)

const (
	// ApprovedConfigMap holds clients that staff have approved.
	ApprovedConfigMap = "noodles-clients"
	// PendingConfigMap holds sign-in requests awaiting approval.
	PendingConfigMap = "noodles-clients-pending"
	// clientsKey is the ConfigMap data key holding the YAML list of entries.
	clientsKey = "clients.yaml"
	// PendingCap is the maximum number of pending requests accepted at once.
	PendingCap = 20

	clientsRefreshInterval = 30 * time.Second
)

// ClientRegistry resolves and manages non-staff clients.
type ClientRegistry interface {
	Lookup(email string) (*model.ClientEntry, bool)
	IsPending(email string) bool
	ListApproved() []model.ClientEntry
	ListPending() []model.ClientEntry
	AddPending(ctx context.Context, entry model.ClientEntry) error
	Approve(ctx context.Context, email string, role model.Role) error
	Reject(ctx context.Context, email string) error
	Refresh(ctx context.Context) error
}

var _ ClientRegistry = (*K8sClientRegistry)(nil)

// K8sClientRegistry stores the client registry in two ConfigMaps and keeps an
// in-memory copy refreshed on a ticker and invalidated on every write.
type K8sClientRegistry struct {
	client    kubernetes.Interface
	namespace string

	mu       sync.RWMutex
	approved []model.ClientEntry
	pending  []model.ClientEntry
}

// NewClientRegistry builds a registry backed by the in-cluster API server. In
// development there is no cluster, so the registry degrades to an in-memory one.
func NewClientRegistry(cfg *config.Config) *K8sClientRegistry {
	reg := &K8sClientRegistry{namespace: cfg.ClientsNamespace}

	if !cfg.IsProduction {
		Logger.Info("Clients: dev mode, using in-memory client registry")
		return reg
	}

	restCfg, err := rest.InClusterConfig()
	if err != nil {
		Logger.Error("Clients: failed to load in-cluster config", "error", err)
		return reg
	}

	clientSet, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		Logger.Error("Clients: failed to create client", "error", err)
		return reg
	}
	reg.client = clientSet

	return reg
}

// NewClientRegistryWithClient is used by tests and by callers that already hold
// a clientset.
func NewClientRegistryWithClient(client kubernetes.Interface, namespace string) *K8sClientRegistry {
	return &K8sClientRegistry{client: client, namespace: namespace}
}

// StartRefresh keeps the cache warm until ctx is cancelled.
func (r *K8sClientRegistry) StartRefresh(ctx context.Context) {
	if err := r.Refresh(ctx); err != nil {
		Logger.Error("Clients: initial refresh failed", "error", err)
	}

	go func() {
		ticker := time.NewTicker(clientsRefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := r.Refresh(ctx); err != nil {
					Logger.Error("Clients: refresh failed", "error", err)
				}
			}
		}
	}()
}

// Refresh reloads both lists from the cluster.
func (r *K8sClientRegistry) Refresh(ctx context.Context) error {
	if r.client == nil {
		return nil
	}

	approved, approvedErr := r.load(ctx, ApprovedConfigMap)
	pending, pendingErr := r.load(ctx, PendingConfigMap)

	r.mu.Lock()
	if approvedErr == nil {
		r.approved = approved
	}
	if pendingErr == nil {
		r.pending = pending
	}
	r.mu.Unlock()

	return errors.Join(approvedErr, pendingErr)
}

// Lookup returns the approved entry for an email, matched case-insensitively.
func (r *K8sClientRegistry) Lookup(email string) (*model.ClientEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if i := indexOfEmail(r.approved, email); i >= 0 {
		entry := r.approved[i]
		return &entry, true
	}
	return nil, false
}

// IsPending reports whether an email already has a pending request.
func (r *K8sClientRegistry) IsPending(email string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return indexOfEmail(r.pending, email) >= 0
}

func (r *K8sClientRegistry) ListApproved() []model.ClientEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.ClientEntry(nil), r.approved...)
}

func (r *K8sClientRegistry) ListPending() []model.ClientEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.ClientEntry(nil), r.pending...)
}

// AddPending records a new sign-in request. It is a no-op if the email is
// already pending, and refuses new entries once PendingCap is reached.
func (r *K8sClientRegistry) AddPending(ctx context.Context, entry model.ClientEntry) error {
	entry.Email = normalizeEmail(entry.Email)
	entry.Role = model.RolePending
	if entry.CreatedAt == "" {
		entry.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	pending, err := r.load(ctx, PendingConfigMap)
	if err != nil {
		return err
	}

	if indexOfEmail(pending, entry.Email) >= 0 {
		r.setPending(pending)
		return nil
	}
	if len(pending) >= PendingCap {
		r.setPending(pending)
		return errs.PendingFull
	}

	pending = append(pending, entry)
	if err := r.save(ctx, PendingConfigMap, pending); err != nil {
		return err
	}

	r.setPending(pending)
	Logger.Info("Clients: pending request recorded", "email", entry.Email, "pendingCount", len(pending))
	return nil
}

// Approve moves an email from the pending list into the approved registry.
func (r *K8sClientRegistry) Approve(ctx context.Context, email string, role model.Role) error {
	if role != model.RoleClient && role != model.RoleClientAdmin {
		return errs.InvalidClientRole
	}
	email = normalizeEmail(email)

	pending, err := r.load(ctx, PendingConfigMap)
	if err != nil {
		return err
	}
	i := indexOfEmail(pending, email)
	if i < 0 {
		return errs.NotPending
	}

	entry := pending[i]
	entry.Role = role
	pending = append(pending[:i:i], pending[i+1:]...)

	approved, err := r.load(ctx, ApprovedConfigMap)
	if err != nil {
		return err
	}
	if j := indexOfEmail(approved, email); j >= 0 {
		approved[j] = entry
	} else {
		approved = append(approved, entry)
	}

	if err := r.save(ctx, ApprovedConfigMap, approved); err != nil {
		return err
	}
	if err := r.save(ctx, PendingConfigMap, pending); err != nil {
		return err
	}

	r.mu.Lock()
	r.approved = approved
	r.pending = pending
	r.mu.Unlock()

	Logger.Info("Clients: request approved", "email", email, "role", role)
	return nil
}

// Reject drops an email from the pending list without approving it.
func (r *K8sClientRegistry) Reject(ctx context.Context, email string) error {
	email = normalizeEmail(email)

	pending, err := r.load(ctx, PendingConfigMap)
	if err != nil {
		return err
	}
	i := indexOfEmail(pending, email)
	if i < 0 {
		return errs.NotPending
	}

	pending = append(pending[:i:i], pending[i+1:]...)
	if err := r.save(ctx, PendingConfigMap, pending); err != nil {
		return err
	}

	r.setPending(pending)
	Logger.Info("Clients: request rejected", "email", email)
	return nil
}

func (r *K8sClientRegistry) setPending(pending []model.ClientEntry) {
	r.mu.Lock()
	r.pending = pending
	r.mu.Unlock()
}

// load reads a ConfigMap and parses its clients.yaml key. A missing or
// malformed key degrades to an empty list rather than failing the caller.
func (r *K8sClientRegistry) load(ctx context.Context, name string) ([]model.ClientEntry, error) {
	if r.client == nil {
		r.mu.RLock()
		defer r.mu.RUnlock()
		if name == ApprovedConfigMap {
			return append([]model.ClientEntry(nil), r.approved...), nil
		}
		return append([]model.ClientEntry(nil), r.pending...), nil
	}

	cm, err := r.client.CoreV1().ConfigMaps(r.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("reading configmap %s/%s: %w", r.namespace, name, err)
	}

	raw, ok := cm.Data[clientsKey]
	if !ok || strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	var entries []model.ClientEntry
	if err := yaml.Unmarshal([]byte(raw), &entries); err != nil {
		Logger.Error("Clients: malformed registry, treating as empty", "configMap", name, "error", err)
		return nil, nil
	}

	for i := range entries {
		entries[i].Email = normalizeEmail(entries[i].Email)
	}
	return entries, nil
}

func (r *K8sClientRegistry) save(ctx context.Context, name string, entries []model.ClientEntry) error {
	if r.client == nil {
		return nil
	}

	if entries == nil {
		entries = []model.ClientEntry{}
	}
	raw, err := yaml.Marshal(entries)
	if err != nil {
		return fmt.Errorf("serializing clients for %s: %w", name, err)
	}

	cm, err := r.client.CoreV1().ConfigMaps(r.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("reading configmap %s/%s: %w", r.namespace, name, err)
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data[clientsKey] = string(raw)

	if _, err := r.client.CoreV1().ConfigMaps(r.namespace).Update(ctx, cm, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("updating configmap %s/%s: %w", r.namespace, name, err)
	}
	return nil
}

func indexOfEmail(entries []model.ClientEntry, email string) int {
	target := normalizeEmail(email)
	if target == "" {
		return -1
	}
	for i := range entries {
		if normalizeEmail(entries[i].Email) == target {
			return i
		}
	}
	return -1
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
