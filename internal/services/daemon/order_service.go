package daemon

import (
	"context"
	"fmt"
	"sync"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
	"github.com/shopspring/decimal"
)

// OrderServiceHandler handles gRPC order requests
type OrderServiceHandler struct {
	mt5Service *mt5.OrderService
	queue      *models.Queue
	logger     *logger.Logger
	mu         sync.Mutex
}

// NewOrderServiceHandler creates a new OrderServiceHandler
func NewOrderServiceHandler(mt5Service *mt5.OrderService, queue *models.Queue) *OrderServiceHandler {
	return &OrderServiceHandler{
		mt5Service: mt5Service,
		queue:      queue,
		logger:     logger.New("OrderServiceHandler"),
	}
}

// PlaceOrder handles the PlaceOrder gRPC call with FIFO sequencing and idempotency
func (h *OrderServiceHandler) PlaceOrder(ctx context.Context, req *api.PlaceOrderRequest) (*api.OrderResponse, error) {
	h.logger.Info(fmt.Sprintf("PlaceOrder request: symbol=%s type=%s volume=%f", req.Symbol, req.Side, req.Volume))

	// Convert float64 to decimal
	volume := decimal.NewFromFloat(req.Volume)
	price := decimal.NewFromFloat(req.Price)

	orderType := models.OrderType(req.Type)
	order := models.NewOrder(req.Symbol, orderType, volume, price, req.IdempotencyKey)

	// Validate order
	if err := h.mt5Service.ValidateOrder(order); err != nil {
		h.logger.Error("Order validation failed", err)
		return nil, err
	}

	// FIFO sequencing - serialize order placement
	h.mu.Lock()
	defer h.mu.Unlock()

	// Place order through MT5 service (handles idempotency)
	ticket, err := h.mt5Service.PlaceOrder(ctx, order)
	if err != nil {
		h.logger.Error("Failed to place order", err)
		return nil, err
	}

	// Enqueue order for persistence
	// 	if err := h.queue.Enqueue(order); err != nil {
	// 		h.logger.Error("Failed to enqueue order", err)
	// 		// Don't fail the response, but log the issue
	// 	}

	resp := &api.OrderResponse{
		Ticket:    ticket,
		FillPrice: price.String(),
		Status:    "filled",
	}

	h.logger.Info(fmt.Sprintf("Order placed successfully: ticket=%d", ticket))
	return resp, nil
}
