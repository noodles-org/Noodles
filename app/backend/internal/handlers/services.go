package handlers

import (
	"net/http"

	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/respond"
	"github.com/mephalrith/noodles/backend/internal/services"
)

func HandleListServices(k8s *services.K8sService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svcList, err := k8s.DiscoverServices(r.Context())
		if err != nil {
			services.Logger.Error("Service discovery failed", "error", err)
			respond.Error(w, errs.Internal("Failed to discover services"))
			return
		}

		respond.OK(w, svcList)
	}
}
