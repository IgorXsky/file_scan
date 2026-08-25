package handler

import (
	"encoding/json"
	"net/http"

	"github.com/IgorXsky/file-scan/internal/clamav"
)

type HealthHandler struct {
	client *clamav.Client
}

func NewHealthHandler(client *clamav.Client) *HealthHandler {
	return &HealthHandler{client: client}
}

type healthResponse struct {
	Status string `json:"status"`
	ClamAV string `json:"clamav"`
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{Status: "ok", ClamAV: "ok"}
	status := http.StatusOK

	if err := h.client.Ping(); err != nil {
		resp.Status = "degraded"
		resp.ClamAV = err.Error()
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
