package daemon

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// OrdersServiceHandler handles gRPC orders listing requests
type OrdersServiceHandler struct {
	mt5Service *mt5.OrdersService
	logger     *logger.Logger
}

// NewOrdersServiceHandler creates a new OrdersServiceHandler
func NewOrdersServiceHandler(mt5Service *mt5.OrdersService) *OrdersServiceHandler {
	return &OrdersServiceHandler{
		mt5Service: mt5Service,
		logger:     logger.New("OrdersServiceHandler"),
	}
}

// GetOrders handles the GetOrders gRPC call
func (h *OrdersServiceHandler) GetOrders(ctx context.Context, req *api.GetOrdersRequest) (*api.OrdersList, error) {
	h.logger.Info("GetOrders request")

	var orders []*api.Order
	var err error

	// If symbol is specified, get orders for that symbol only
	if req.Symbol != "" {
		h.logger.Info(fmt.Sprintf("Filtering orders for symbol: %s", req.Symbol))
		pendingOrders, err := h.mt5Service.GetPendingOrdersBySymbol(ctx, req.Symbol)
		if err != nil {
			h.logger.Error("Failed to get orders", err)
			return nil, err
		}

		for _, o := range pendingOrders {
			orders = append(orders, &api.Order{
				Ticket:   o.Ticket,
				Symbol:   o.Symbol,
				Type:     string(o.Type),
				Volume:   o.Volume.String(),
				Price:    o.Price.String(),
				Status:   string(o.Status),
				FillPrice: o.FillPrice.String(),
				ProfitLoss: o.ProfitLoss.String(),
			})
		}
	} else {
		// Get all pending orders
		pendingOrders, err := h.mt5Service.GetPendingOrders(ctx)
		if err != nil {
			h.logger.Error("Failed to get orders", err)
			return nil, err
		}

		for _, o := range pendingOrders {
			orders = append(orders, &api.Order{
				Ticket:   o.Ticket,
				Symbol:   o.Symbol,
				Type:     string(o.Type),
				Volume:   o.Volume.String(),
				Price:    o.Price.String(),
				Status:   string(o.Status),
				FillPrice: o.FillPrice.String(),
				ProfitLoss: o.ProfitLoss.String(),
			})
		}
	}

	resp := &api.OrdersList{
		Orders: orders,
		Count:  int32(len(orders)),
	}

	h.logger.Info(fmt.Sprintf("Orders retrieved: count=%d", len(orders)))
	return resp, nil
}
