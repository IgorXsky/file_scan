package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/IgorXsky/file-scan/internal/clamav"
)

type ScanHandler struct {
	client       *clamav.Client
	maxFileSize  int64
}

func NewScanHandler(client *clamav.Client, maxFileSize int64) *ScanHandler {
	return &ScanHandler{client: client, maxFileSize: maxFileSize}
}

type scanResponse struct {
	Clean     bool   `json:"clean"`
	Signature string `json:"signature,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (h *ScanHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, scanResponse{Error: "method not allowed"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxFileSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Error("read form file", "error", err)
		writeJSON(w, http.StatusBadRequest, scanResponse{Error: "invalid file: " + err.Error()})
		return
	}
	defer file.Close()

	slog.Info("scanning file", "filename", header.Filename, "size", header.Size)

	result, err := h.client.ScanStream(file)
	if err != nil {
		slog.Error("clamav scan failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, scanResponse{Error: "scan unavailable"})
		return
	}

	status := http.StatusOK
	resp := scanResponse{
		Clean:     result.Clean,
		Signature: result.Signature,
		Error:     result.Error,
	}

	if !result.Clean && result.Error == "" {
		slog.Warn("infected file detected", "filename", header.Filename, "signature", result.Signature)
	}

	writeJSON(w, status, resp)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
