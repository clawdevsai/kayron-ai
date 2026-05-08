package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// BalanceDrawdownServiceHandler wraps balance drawdown operations
type BalanceDrawdownServiceHandler struct {
	service *mt5.BalanceDrawdownService
	logger  *logger.Logger
}

// NewBalanceDrawdownServiceHandler creates a new balance drawdown handler
func NewBalanceDrawdownServiceHandler(service *mt5.BalanceDrawdownService) *BalanceDrawdownServiceHandler {
	return &BalanceDrawdownServiceHandler{
		service: service,
		logger:  logger.New("BalanceDrawdownHandler"),
	}
}

// CalculateDrawdown handles balance drawdown requests
func (h *BalanceDrawdownServiceHandler) CalculateDrawdown(ctx context.Context, req *api.BalanceDrawdownRequest) (*api.BalanceDrawdownResponse, error) {
	h.logger.Info("CalculateDrawdown request handling")

	result, err := h.service.CalculateDrawdown(ctx, req.SinceTimestamp)
	if err != nil {
		h.logger.Error("CalculateDrawdown failed", err)
		return &api.BalanceDrawdownResponse{
			MaxEquity:       "0",
			CurrentEquity:   "0",
			DrawdownPercent: 0,
		}, nil
	}

	return &api.BalanceDrawdownResponse{
		MaxEquity:       result.MaxEquity,
		CurrentEquity:   result.CurrentEquity,
		DrawdownPercent: result.DrawdownPercent,
	}, nil
}
