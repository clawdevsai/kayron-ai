package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// OrderFillAnalysisTool handles the order-fill-analysis MCP tool
type OrderFillAnalysisTool struct {
	handler *daemon.OrderFillAnalysisServiceHandler
	logger  *logger.Logger
}

// NewOrderFillAnalysisTool creates a new OrderFillAnalysisTool
func NewOrderFillAnalysisTool(handler *daemon.OrderFillAnalysisServiceHandler) *OrderFillAnalysisTool {
	return &OrderFillAnalysisTool{
		handler: handler,
		logger:  logger.New("OrderFillAnalysisTool"),
	}
}

// Name returns the tool name
func (t *OrderFillAnalysisTool) Name() string {
	return "order-fill-analysis"
}

// Description returns the tool description
func (t *OrderFillAnalysisTool) Description() string {
	return "Analyze order fill price, slippage, and execution latency for a given ticket"
}

// InputSchema returns the JSON schema for parameters
func (t *OrderFillAnalysisTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ticket": map[string]interface{}{
				"type":        "integer",
				"description": "Order ticket number",
			},
		},
		"required": []string{"ticket"},
	}
}

// Execute handles the order-fill-analysis tool execution
func (t *OrderFillAnalysisTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing order-fill-analysis tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	ticket, ok := paramsMap["ticket"].(float64)
	if !ok {
		return nil, fmt.Errorf("ticket parameter is required (integer)")
	}

	ctx := context.Background()

	resp, err := t.handler.AnalyzeOrderFill(ctx, &api.OrderFillAnalysisRequest{
		Ticket: int64(ticket),
	})

	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to analyze order fill"), err)
		return nil, fmt.Errorf("failed to analyze order fill: %v", err)
	}

	t.logger.Info(fmt.Sprintf("Analyzed fill for order %d: slippage=%.4f, latency=%dms", resp.Ticket, resp.Slippage, resp.ExecutionLatency))

	return map[string]interface{}{
		"ticket":            resp.Ticket,
		"symbol":            resp.Symbol,
		"fill_price":        resp.FillPrice,
		"slippage":          resp.Slippage,
		"execution_latency": resp.ExecutionLatency,
	}, nil
}
