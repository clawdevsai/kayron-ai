package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// MarketHoursServiceHandler wraps market hours operations
type MarketHoursServiceHandler struct {
	service *mt5.MarketHoursService
	logger  *logger.Logger
}

// NewMarketHoursServiceHandler creates a new market hours handler
func NewMarketHoursServiceHandler(service *mt5.MarketHoursService) *MarketHoursServiceHandler {
	return &MarketHoursServiceHandler{
		service: service,
		logger:  logger.New("MarketHoursHandler"),
	}
}

// GetMarketHours handles market hours requests
func (h *MarketHoursServiceHandler) GetMarketHours(ctx context.Context, req *api.MarketHoursRequest) (*api.MarketHoursResponse, error) {
	h.logger.Info("GetMarketHours request handling")

	hours, err := h.service.GetMarketHours(ctx, req.Symbol)
	if err != nil {
		h.logger.Error("GetMarketHours failed", err)
		return &api.MarketHoursResponse{
			Symbol:    req.Symbol,
			OpenTime:  0,
			CloseTime: 0,
			Timezone:  "",
			IsClosed:  true,
		}, nil
	}

	return &api.MarketHoursResponse{
		Symbol:    hours.Symbol,
		OpenTime:  hours.OpenTime,
		CloseTime: hours.CloseTime,
		Timezone:  hours.Timezone,
		IsClosed:  hours.IsClosed,
	}, nil
}
