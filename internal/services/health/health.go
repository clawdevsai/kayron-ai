package health

import (
	"encoding/json"
	"net/http"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
)

// Handler handles health check requests
type Handler struct {
	queue  *models.Queue
	logger *logger.Logger
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status            string `json:"status"`
	TerminalConnected bool   `json:"terminal_connected"`
	QueueLength       int    `json:"queue_length"`
}

// NewHandler creates a new health check handler
func NewHandler(queue *models.Queue) *Handler {
	return &Handler{
		queue:  queue,
		logger: logger.New("HealthHandler"),
	}
}

// ServeHTTP handles HTTP requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queueLength, err := h.queue.GetQueueLength()
	if err != nil {
		h.logger.Error("Failed to get queue length", err)
		queueLength = -1
	}

	response := HealthResponse{
		Status:            "ok",
		TerminalConnected: true, // TODO: check actual connection status
		QueueLength:       queueLength,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
