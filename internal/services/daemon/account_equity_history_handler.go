package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// AccountEquityHistoryServiceHandler wraps equity history operations
type AccountEquityHistoryServiceHandler struct {
	service *mt5.AccountEquityHistoryService
	logger  *logger.Logger
}

// NewAccountEquityHistoryServiceHandler creates a new equity history handler
func NewAccountEquityHistoryServiceHandler(service *mt5.AccountEquityHistoryService) *AccountEquityHistoryServiceHandler {
	return &AccountEquityHistoryServiceHandler{
		service: service,
		logger:  logger.New("AccountEquityHistoryHandler"),
	}
}

// GetEquityHistory handles equity history requests
func (h *AccountEquityHistoryServiceHandler) GetEquityHistory(ctx context.Context, req *api.EquityHistoryRequest) (*api.EquityHistoryResponse, error) {
	h.logger.Info("GetEquityHistory request handling")

	snapshots, err := h.service.GetEquityHistory(ctx, req.FromTimestamp, req.ToTimestamp, req.Granularity)
	if err != nil {
		h.logger.Error("GetEquityHistory failed", err)
		return &api.EquityHistoryResponse{
			Snapshots: []*api.EquitySnapshot{},
		}, nil
	}

	items := make([]*api.EquitySnapshot, len(snapshots))
	for i, snap := range snapshots {
		items[i] = &api.EquitySnapshot{
			Timestamp: snap.Timestamp,
			Equity:    snap.Equity,
			Balance:   snap.Balance,
		}
	}

	return &api.EquityHistoryResponse{
		Snapshots: items,
	}, nil
}
