package handlers

import (
	"net/http"
	"regexp"

	chilib "github.com/go-chi/chi/v5"

	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/middleware"
	"github.com/mephalrith/noodles/backend/internal/respond"
	"github.com/mephalrith/noodles/backend/internal/services"
)

var nameRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9\-.]*[a-z0-9])?$`)

func validateTarget(w http.ResponseWriter, r *http.Request, k8s *services.K8sService, namespace, name string) bool {
	if !nameRE.MatchString(namespace) || !nameRE.MatchString(name) {
		respond.Error(w, errs.InvalidTarget)
		return false
	}
	managed, err := k8s.IsManagedNamespace(r.Context(), namespace)
	if err != nil || !managed {
		respond.Error(w, errs.NamespaceNotManaged)
		return false
	}
	return true
}

func HandleListDeployments(k8s *services.K8sService, argo *services.ArgoCDService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deployments, err := k8s.ListDeployments(r.Context())
		if err != nil {
			services.Logger.Error("Failed listing deployments", "error", err)
			respond.Error(w, errs.Internal("Failed to list deployments"))
			return
		}

		argoMap := argo.GetHealthMap()
		for i := range deployments {
			dep := &deployments[i]
			if dep.ArgoApp != "" {
				if status, ok := argoMap[dep.ArgoApp]; ok {
					dep.HealthStatus = status.Health
					dep.SyncStatus = status.Sync
				}
			}
		}

		respond.OK(w, deployments)
	}
}

func HandleRestartDeployment(k8s *services.K8sService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := chilib.URLParam(r, "namespace")
		name := chilib.URLParam(r, "name")
		if !validateTarget(w, r, k8s, namespace, name) {
			return
		}

		if err := k8s.RestartDeployment(r.Context(), namespace, name); err != nil {
			services.Logger.Error("Restart failed", "namespace", namespace, "name", name, "error", err)
			respond.Error(w, errs.Internal("Restart failed"))
			return
		}

		user := middleware.UserFromContext(r.Context())
		services.DeploymentActions.With(map[string]string{"action": "restart", "namespace": namespace, "deployment": name}).Inc()
		services.Logger.Info("Restart requested", "namespace", namespace, "name", name, "user", user.Email)

		respond.OK(w, map[string]bool{"ok": true})
	}
}

func HandlePauseDeployment(k8s *services.K8sService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := chilib.URLParam(r, "namespace")
		name := chilib.URLParam(r, "name")
		if !validateTarget(w, r, k8s, namespace, name) {
			return
		}

		if err := k8s.PauseDeployment(r.Context(), namespace, name); err != nil {
			services.Logger.Error("Pause failed", "namespace", namespace, "name", name, "error", err)
			respond.Error(w, errs.Internal(err.Error()))
			return
		}

		user := middleware.UserFromContext(r.Context())
		services.DeploymentActions.With(map[string]string{"action": "pause", "namespace": namespace, "deployment": name}).Inc()
		services.Logger.Info("Pause requested", "namespace", namespace, "name", name, "user", user.Email)

		respond.OK(w, map[string]bool{"ok": true})
	}
}

func HandleResumeDeployment(k8s *services.K8sService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := chilib.URLParam(r, "namespace")
		name := chilib.URLParam(r, "name")
		if !validateTarget(w, r, k8s, namespace, name) {
			return
		}

		if err := k8s.ResumeDeployment(r.Context(), namespace, name); err != nil {
			services.Logger.Error("Resume failed", "namespace", namespace, "name", name, "error", err)
			respond.Error(w, errs.Internal(err.Error()))
			return
		}

		user := middleware.UserFromContext(r.Context())
		services.DeploymentActions.With(map[string]string{"action": "resume", "namespace": namespace, "deployment": name}).Inc()
		services.Logger.Info("Resume requested", "namespace", namespace, "name", name, "user", user.Email)

		respond.OK(w, map[string]bool{"ok": true})
	}
}
