package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// PendingOrderTool handles the pending-order-details MCP tool
type PendingOrderTool struct {
	handler *daemon.PendingOrderServiceHandler
	logger  *logger.Logger
}

// NewPendingOrderTool creates a new PendingOrderTool
func NewPendingOrderTool(handler *daemon.PendingOrderServiceHandler) *PendingOrderTool {
	return &PendingOrderTool{
		handler: handler,
		logger:  logger.New("PendingOrderTool"),
	}
}

// Name returns the tool name
func (t *PendingOrderTool) Name() string {
	return "pending-order-details"
}

// Description returns the tool description
func (t *PendingOrderTool) Description() string {
	return "Query pending orders with filters"
}

// InputSchema returns the JSON schema for parameters
func (t *PendingOrderTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"symbol": map[string]interface{}{
				"type":        "string",
				"description": "Symbol filter (optional)",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Order status filter (optional)",
			},
			"createdAfter": map[string]interface{}{
				"type":        "integer",
				"description": "Unix timestamp filter - orders created after this time (optional)",
			},
		},
	}
}

// Execute handles the pending-order-details tool execution
func (t *PendingOrderTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing pending-order-details tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	symbol, _ := paramsMap["symbol"].(string)
	status, _ := paramsMap["status"].(string)

	var createdAfter int64
	if ca, ok := paramsMap["createdAfter"].(float64); ok {
		createdAfter = int64(ca)
	}

	ctx := context.Background()

	resp, err := t.handler.GetPendingOrderDetails(ctx, &api.PendingOrderDetailsRequest{
		Symbol:       symbol,
		Status:       status,
		CreatedAfter: createdAfter,
	})

	if err != nil {
		t.logger.Error("Failed to get pending orders", err)
		return nil, fmt.Errorf("failed to retrieve pending orders: %v", err)
	}

	orders := make([]map[string]interface{}, len(resp.Orders))
	for i, o := range resp.Orders {
		orders[i] = map[string]interface{}{
			"ticket":     o.Ticket,
			"symbol":     o.Symbol,
			"type":       o.Type,
			"volume":     o.Volume,
			"price":      o.Price,
			"status":     o.Status,
			"openTime":   o.OpenTime,
			"fillPrice":  o.FillPrice,
			"profitLoss": o.ProfitLoss,
		}
	}

	t.logger.Info(fmt.Sprintf("Retrieved %d pending orders", len(orders)))

	return map[string]interface{}{
		"orders": orders,
		"count":  len(orders),
	}, nil
}
