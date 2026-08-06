package gateway

import (
	"encoding/json"
	"net/http"

	"github.com/vucongthanh92/courier/agent-gateway/helper/constants"
	"github.com/vucongthanh92/courier/agent-gateway/internal/domain/models"
)

func NewHTTPHandler(service *Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(constants.HealthzPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := service.Health(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"service": service.cfg.ServiceName,
		})
	})
	mux.HandleFunc(constants.AssistantInstructionsPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"instructions": service.SystemInstructions(),
		})
	})
	mux.HandleFunc(constants.SafetyEvaluatePath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req models.SafetyEvaluationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error": "invalid request body",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"result": service.EvaluateSafety(req.Text),
		})
	})
	return mux
}

func writeJSON(w http.ResponseWriter, status int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
