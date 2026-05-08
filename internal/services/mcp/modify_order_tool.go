package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

type ModifyOrderTool struct {
	handler *daemon.ModifyOrderServiceHandler
	logger  *logger.Logger
}

func NewModifyOrderTool(handler *daemon.ModifyOrderServiceHandler) *ModifyOrderTool {
	return &ModifyOrderTool{
		handler: handler,
		logger:  logger.New("ModifyOrderTool"),
	}
}

func (t *ModifyOrderTool) Name() string {
	return "modify_order"
}

func (t *ModifyOrderTool) Description() string {
	return "Modify existing order parameters (price, stop loss, take profit)"
}

func (t *ModifyOrderTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ticket": map[string]interface{}{
				"type":        "integer",
				"description": "Order ticket number",
			},
			"price": map[string]interface{}{
				"type":        "string",
				"description": "New order price (decimal string, optional)",
			},
			"stopLoss": map[string]interface{}{
				"type":        "string",
				"description": "New stop loss price (decimal string, optional)",
			},
			"takeProfit": map[string]interface{}{
				"type":        "string",
				"description": "New take profit price (decimal string, optional)",
			},
		},
		"required": []string{"ticket"},
	}
}

func (t *ModifyOrderTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing modify-order tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	ticket, ok := paramsMap["ticket"].(float64)
	if !ok {
		return nil, fmt.Errorf("ticket parameter is required and must be a number")
	}

	price, _ := paramsMap["price"].(string)
	stopLoss, _ := paramsMap["stopLoss"].(string)
	takeProfit, _ := paramsMap["takeProfit"].(string)

	ctx := context.Background()

	resp, err := t.handler.ModifyOrder(ctx, &api.ModifyOrderRequest{
		Ticket:     int64(ticket),
		Price:      price,
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
	})

	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to modify order %d", int64(ticket)), err)
		return nil, fmt.Errorf("failed to modify order: %v", err)
	}

	t.logger.Info(fmt.Sprintf("Order %d modified successfully", int64(ticket)))

	return map[string]interface{}{
		"ticket":   resp.Ticket,
		"status":   resp.Status,
		"errorMsg": resp.ErrorMsg,
	}, nil
}
