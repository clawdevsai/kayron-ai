package daemon

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// PositionServiceHandler handles gRPC position requests
type PositionServiceHandler struct {
	mt5Service *mt5.PositionService
	logger     *logger.Logger
}

// NewPositionServiceHandler creates a new PositionServiceHandler
func NewPositionServiceHandler(mt5Service *mt5.PositionService) *PositionServiceHandler {
	return &PositionServiceHandler{
		mt5Service: mt5Service,
		logger:     logger.New("PositionServiceHandler"),
	}
}

// ClosePosition handles the ClosePosition gRPC call
func (h *PositionServiceHandler) ClosePosition(ctx context.Context, req *api.ClosePositionRequest) (*api.PositionCloseResponse, error) {
	h.logger.Info(fmt.Sprintf("ClosePosition request: ticket=%d", req.Ticket))

	profitLoss, err := h.mt5Service.ClosePosition(ctx, req.Ticket)
	if err != nil {
		h.logger.Error("Failed to close position", err)
		return nil, err
	}

	resp := &api.PositionCloseResponse{
		Ticket:     req.Ticket,
		ProfitLoss: profitLoss.String(),
	}

	h.logger.Info(fmt.Sprintf("Position closed: ticket=%d, profitLoss=%s", req.Ticket, profitLoss.String()))
	return resp, nil
}
