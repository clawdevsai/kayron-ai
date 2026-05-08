package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// PendingOrderServiceHandler wraps pending order operations
type PendingOrderServiceHandler struct {
	service *mt5.PendingOrderService
	logger  *logger.Logger
}

// NewPendingOrderServiceHandler creates a new pending order handler
func NewPendingOrderServiceHandler(service *mt5.PendingOrderService) *PendingOrderServiceHandler {
	return &PendingOrderServiceHandler{
		service: service,
		logger:  logger.New("PendingOrderHandler"),
	}
}

// GetPendingOrderDetails handles pending order details requests
func (h *PendingOrderServiceHandler) GetPendingOrderDetails(ctx context.Context, req *api.PendingOrderDetailsRequest) (*api.PendingOrderDetailsResponse, error) {
	h.logger.Info("GetPendingOrderDetails request handling")

	orders, err := h.service.GetPendingOrders(ctx, req.Symbol, req.Status, req.CreatedAfter)
	if err != nil {
		h.logger.Error("GetPendingOrderDetails failed", err)
		return &api.PendingOrderDetailsResponse{
			Orders: []*api.OrderItem{},
		}, nil
	}

	apiOrders := make([]*api.OrderItem, len(orders))
	for i, o := range orders {
		apiOrders[i] = &api.OrderItem{
			Ticket:     o.Ticket,
			Symbol:     o.Symbol,
			Type:       o.Type,
			Volume:     o.Volume,
			Price:      o.Price,
			Status:     o.Status,
			OpenTime:   o.OpenTime,
			FillPrice:  o.FillPrice,
			ProfitLoss: o.ProfitLoss,
		}
	}

	return &api.PendingOrderDetailsResponse{
		Orders: apiOrders,
	}, nil
}
