package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// OrdersListTool handles the orders-list MCP tool
type OrdersListTool struct {
	handler *daemon.OrdersServiceHandler
	logger  *logger.Logger
}

// NewOrdersListTool creates a new OrdersListTool
func NewOrdersListTool(handler *daemon.OrdersServiceHandler) *OrdersListTool {
	return &OrdersListTool{
		handler: handler,
		logger:  logger.New("OrdersListTool"),
	}
}

// Execute handles the orders-list tool execution
func (t *OrdersListTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing orders-list tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		paramsMap = make(map[string]interface{})
	}

	symbol := ""
	if s, ok := paramsMap["symbol"].(string); ok {
		symbol = s
	}

	ctx := context.Background()

	// Call gRPC handler
	ordersList, err := t.handler.GetOrders(ctx, &api.GetOrdersRequest{Symbol: symbol})
	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to get orders: %v", err))
		return nil, fmt.Errorf("failed to retrieve orders: %v", err)
	}

	orders := make([]map[string]interface{}, 0)
	for _, o := range ordersList.Orders {
		orders = append(orders, map[string]interface{}{
			"ticket":      o.Ticket,
			"symbol":      o.Symbol,
			"type":        o.Type,
			"volume":      o.Volume,
			"price":       o.Price,
			"status":      o.Status,
			"fill_price":  o.FillPrice,
			"profit_loss": o.ProfitLoss,
		})
	}

	result := map[string]interface{}{
		"orders": orders,
		"count":  ordersList.Count,
	}

	t.logger.Info(fmt.Sprintf("Orders list retrieved: count=%d", ordersList.Count))
	return result, nil
}
