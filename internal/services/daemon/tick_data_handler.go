package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// TickDataServiceHandler wraps tick data operations
type TickDataServiceHandler struct {
	service *mt5.TickDataService
	logger  *logger.Logger
}

// NewTickDataServiceHandler creates a new tick data handler
func NewTickDataServiceHandler(service *mt5.TickDataService) *TickDataServiceHandler {
	return &TickDataServiceHandler{
		service: service,
		logger:  logger.New("TickDataHandler"),
	}
}

// GetTickData handles tick data requests
func (h *TickDataServiceHandler) GetTickData(ctx context.Context, req *api.TickDataRequest) (*api.TickDataResponse, error) {
	h.logger.Info("GetTickData request handling")

	ticks, err := h.service.GetTickData(ctx, req.Symbol, req.DurationSeconds)
	if err != nil {
		h.logger.Error("GetTickData failed", err)
		return &api.TickDataResponse{
			Symbol: req.Symbol,
			Ticks:  []*api.TickItem{},
		}, nil
	}

	items := make([]*api.TickItem, len(ticks))
	for i, tick := range ticks {
		items[i] = &api.TickItem{
			Timestamp: tick.Timestamp,
			Bid:       tick.Bid,
			Ask:       tick.Ask,
		}
	}

	return &api.TickDataResponse{
		Symbol: req.Symbol,
		Ticks:  items,
	}, nil
}
