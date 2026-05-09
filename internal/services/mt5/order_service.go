package mt5

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
)

// OrderService handles MT5 order placement and management
type OrderService struct {
	client            *Client
	logger            *logger.Logger
	idempotencyCache  *models.IdempotencyCache
}

// NewOrderService creates a new OrderService
func NewOrderService(client *Client) *OrderService {
	return &OrderService{
		client:           client,
		logger:           logger.New("OrderService"),
		idempotencyCache: models.NewIdempotencyCache(),
	}
}

// PlaceOrder places an order on MT5
func (os *OrderService) PlaceOrder(ctx context.Context, order *models.Order) (int64, error) {
	os.logger.Info(fmt.Sprintf("Placing %s order for %s volume=%.2f price=%.2f", order.Type, order.Symbol, order.Volume, order.Price))

	// Check idempotency cache
	if ticket, exists := os.idempotencyCache.Get(order.IdempotencyKey); exists {
		os.logger.Info(fmt.Sprintf("Order already placed with idempotency key %s, ticket=%d", order.IdempotencyKey, ticket))
		return ticket, nil
	}

	// Validate order parameters
	if order.Volume.IsNegative() || order.Volume.IsZero() {
		return 0, fmt.Errorf("invalid order volume: %v", order.Volume)
	}

	if order.Price.IsNegative() || order.Price.IsZero() {
		return 0, fmt.Errorf("invalid order price: %v", order.Price)
	}

	// Call MT5 WebAPI client to place real order
	placedOrder, err := os.client.PlaceOrder(
		order.Symbol,
		string(order.Type),
		order.Volume,
		order.Price,
		*order.StopLoss,
		*order.TakeProfit,
		order.Comment,
	)
	if err != nil {
		os.logger.Error("Failed to place order on MT5", err)
		return 0, err
	}

	ticket := int64(placedOrder.Ticket)

	// Cache the order
	os.idempotencyCache.Set(order.IdempotencyKey, ticket)

	os.logger.Info(fmt.Sprintf("Order placed successfully, ticket=%d", ticket))
	return ticket, nil
}

// ValidateOrder validates order parameters
func (os *OrderService) ValidateOrder(order *models.Order) error {
	if order.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	if order.Type != models.OrderTypeBuy && order.Type != models.OrderTypeSell {
		return fmt.Errorf("invalid order type: %s", order.Type)
	}

	if order.Volume.IsNegative() || order.Volume.IsZero() {
		return fmt.Errorf("invalid order volume: %v", order.Volume)
	}

	if order.Price.IsNegative() || order.Price.IsZero() {
		return fmt.Errorf("invalid order price: %v", order.Price)
	}

	if order.StopLoss != nil && order.Price.Equal(*order.StopLoss) {
		return fmt.Errorf("stop loss cannot equal entry price")
	}

	if order.TakeProfit != nil && order.Price.Equal(*order.TakeProfit) {
		return fmt.Errorf("take profit cannot equal entry price")
	}

	return nil
}

// GetIdempotencyCache returns the idempotency cache
func (os *OrderService) GetIdempotencyCache() *models.IdempotencyCache {
	return os.idempotencyCache
}
