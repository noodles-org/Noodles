package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/middleware"
	"github.com/mephalrith/noodles/backend/internal/model"
	"github.com/mephalrith/noodles/backend/internal/respond"
	"github.com/mephalrith/noodles/backend/internal/services"
)

type approveRequest struct {
	Email string     `json:"email"`
	Role  model.Role `json:"role"`
}

type rejectRequest struct {
	Email string `json:"email"`
}

func HandleListPendingClients(registry services.ClientRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respond.OK(w, registry.ListPending())
	}
}

func HandleListApprovedClients(registry services.ClientRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respond.OK(w, registry.ListApproved())
	}
}

func HandleApproveClient(registry services.ClientRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req approveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
			respond.Error(w, errs.BadRequest)
			return
		}
		if req.Role != model.RoleClient && req.Role != model.RoleClientAdmin {
			respond.Error(w, errs.InvalidClientRole)
			return
		}

		if err := registry.Approve(r.Context(), req.Email, req.Role); err != nil {
			respondRegistryError(w, err, "Clients: approve failed", req.Email)
			return
		}

		services.Logger.Info("Clients: access approved", "email", req.Email, "role", req.Role, "by", actorEmail(r.Context()))

		respond.OK(w, map[string]bool{"ok": true})
	}
}

func HandleRejectClient(registry services.ClientRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req rejectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
			respond.Error(w, errs.BadRequest)
			return
		}

		if err := registry.Reject(r.Context(), req.Email); err != nil {
			respondRegistryError(w, err, "Clients: reject failed", req.Email)
			return
		}

		services.Logger.Info("Clients: access rejected", "email", req.Email, "by", actorEmail(r.Context()))

		respond.OK(w, map[string]bool{"ok": true})
	}
}

// actorEmail is the email of the staff member performing the action.
func actorEmail(ctx context.Context) string {
	if user := middleware.UserFromContext(ctx); user != nil {
		return user.Email
	}
	return ""
}

// respondRegistryError maps registry failures onto their HTTP responses.
func respondRegistryError(w http.ResponseWriter, err error, msg, email string) {
	var known *errs.Error
	if errors.As(err, &known) {
		respond.Error(w, known)
		return
	}
	services.Logger.Error(msg, "email", email, "error", err)
	respond.Error(w, errs.Internal("Failed to update client registry"))
}
