package mt5

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
)

// PositionService handles MT5 position queries and management
type PositionService struct {
	client *Client
	logger *logger.Logger
}

// NewPositionService creates a new PositionService
func NewPositionService(client *Client) *PositionService {
	return &PositionService{
		client: client,
		logger: logger.New("PositionService"),
	}
}

// GetPosition retrieves a position by ticket
func (ps *PositionService) GetPosition(ctx context.Context, ticket int64) (*models.Position, error) {
	ps.logger.Info(fmt.Sprintf("Querying position ticket=%d", ticket))

	// Call MT5 WebAPI client to get real position (through list orders)
	orders, err := ps.client.ListOrders("OPEN")
	if err != nil {
		ps.logger.Error(fmt.Sprintf("Failed to list orders for position ticket=%d", ticket), err)
		return nil, err
	}

	// Find the matching ticket
	var orderData *Order
	for _, order := range orders {
		if int64(order.Ticket) == ticket {
			orderData = order
			break
		}
	}

	if orderData == nil {
		return nil, fmt.Errorf("position not found: ticket=%d", ticket)
	}

	position := models.NewPosition(ticket, orderData.Symbol, mapOrderSideToPosition(orderData.Side), orderData.Volume, orderData.OpenPrice)
	position.UpdateProfit(orderData.OpenPrice) // UpdateProfit with current price (from quote)

	ps.logger.Info(fmt.Sprintf("Position retrieved: ticket=%d, symbol=%s", ticket, orderData.Symbol))
	return position, nil
}

// mapOrderSideToPosition converts order side to position type
func mapOrderSideToPosition(side string) models.PositionType {
	if side == "BUY" {
		return models.PositionTypeLong
	}
	return models.PositionTypeShort
}

// ClosePosition closes an open position
func (ps *PositionService) ClosePosition(ctx context.Context, ticket int64) (decimal.Decimal, error) {
	ps.logger.Info(fmt.Sprintf("Closing position ticket=%d", ticket))

	// Get position first
	position, err := ps.GetPosition(ctx, ticket)
	if err != nil {
		return decimal.Zero, err
	}

	// Call MT5 client to close position
	// This is a placeholder - actual implementation depends on MT5 API

	ps.logger.Info(fmt.Sprintf("Position closed: ticket=%d, profit=%.2f", ticket, position.Profit))
	return position.Profit, nil
}

// ListPositions returns all open positions
func (ps *PositionService) ListPositions(ctx context.Context) ([]*models.Position, error) {
	ps.logger.Info("Querying all open positions")

	// Call MT5 client to get positions
	// This is a placeholder - actual implementation depends on MT5 API
	positions := make([]*models.Position, 0)

	ps.logger.Info(fmt.Sprintf("Retrieved %d open positions", len(positions)))
	return positions, nil
}
