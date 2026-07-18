package respond

import (
	"encoding/json"
	"net/http"

	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/services"
)

// JSON writes a JSON response with the given status code.
// Logs encoding failures since the response is already partially written.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		services.Logger.Error("Failed to encode JSON response", "error", err)
	}
}

// Error writes an errs.Error as a JSON response.
func Error(w http.ResponseWriter, e *errs.Error) {
	JSON(w, e.Status, e)
}

// OK writes a 200 JSON response.
func OK(w http.ResponseWriter, v any) {
	JSON(w, http.StatusOK, v)
}
