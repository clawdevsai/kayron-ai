package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// PositionDetailsServiceHandler wraps position details operations
type PositionDetailsServiceHandler struct {
	service *mt5.PositionDetailsService
	logger  *logger.Logger
}

// NewPositionDetailsServiceHandler creates a new position details handler
func NewPositionDetailsServiceHandler(service *mt5.PositionDetailsService) *PositionDetailsServiceHandler {
	return &PositionDetailsServiceHandler{
		service: service,
		logger:  logger.New("PositionDetailsHandler"),
	}
}

// GetPositionDetails handles position details requests for a symbol
func (h *PositionDetailsServiceHandler) GetPositionDetails(ctx context.Context, req *api.PositionDetailsRequest) (*api.PositionDetailsResponse, error) {
	h.logger.Info("GetPositionDetails request handling")

	positions, err := h.service.GetPositionDetails(ctx, req.Symbol)
	if err != nil {
		h.logger.Error("GetPositionDetails failed", err)
		return &api.PositionDetailsResponse{
			Positions: []*api.PositionItem{},
		}, nil
	}

	items := make([]*api.PositionItem, len(positions))
	for i, pos := range positions {
		items[i] = &api.PositionItem{
			Ticket:       int64(pos.Ticket),
			Symbol:       pos.Symbol,
			Type:         pos.Type,
			Volume:       pos.Volume,
			OpenPrice:    pos.OpenPrice,
			CurrentPrice: pos.CurrentPrice,
			Profit:       pos.ProfitLoss,
			Swap:         pos.Swap,
			OpenTime:     pos.OpenTime,
		}
	}

	return &api.PositionDetailsResponse{
		Positions: items,
	}, nil
}
