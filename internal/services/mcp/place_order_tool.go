package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// PlaceOrderTool handles the place-order MCP tool
type PlaceOrderTool struct {
	handler *daemon.OrderServiceHandler
	logger  *logger.Logger
}

// NewPlaceOrderTool creates a new PlaceOrderTool
func NewPlaceOrderTool(handler *daemon.OrderServiceHandler) *PlaceOrderTool {
	return &PlaceOrderTool{
		handler: handler,
		logger:  logger.New("PlaceOrderTool"),
	}
}

// Execute handles the place-order tool execution
func (t *PlaceOrderTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing place-order tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	// Extract required parameters
	symbol, ok := paramsMap["symbol"].(string)
	if !ok || symbol == "" {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	orderType, ok := paramsMap["type"].(string)
	if !ok || orderType == "" {
		return nil, fmt.Errorf("type parameter is required")
	}

	volume, ok := paramsMap["volume"].(string)
	if !ok || volume == "" {
		return nil, fmt.Errorf("volume parameter is required")
	}

	price, ok := paramsMap["price"].(string)
	if !ok || price == "" {
		return nil, fmt.Errorf("price parameter is required")
	}

	idempotencyKey, ok := paramsMap["idempotency_key"].(string)
	if !ok || idempotencyKey == "" {
		return nil, fmt.Errorf("idempotency_key parameter is required")
	}

	ctx := context.Background()

	// Call gRPC handler
	orderResp, err := t.handler.PlaceOrder(ctx, &api.PlaceOrderRequest{
		Symbol:         symbol,
		Type:           orderType,
		Volume:         volume,
		Price:          price,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to place order: %v", err))
		return nil, fmt.Errorf("failed to place order: %v", err)
	}

	result := map[string]interface{}{
		"ticket":     orderResp.Ticket,
		"fill_price": orderResp.FillPrice,
		"status":     orderResp.Status,
	}

	t.logger.Info(fmt.Sprintf("Order placed successfully: ticket=%d", orderResp.Ticket))
	return result, nil
}
