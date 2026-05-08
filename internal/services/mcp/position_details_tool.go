package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// PositionDetailsTool handles the position-details MCP tool
type PositionDetailsTool struct {
	handler *daemon.PositionDetailsServiceHandler
	logger  *logger.Logger
}

// NewPositionDetailsTool creates a new PositionDetailsTool
func NewPositionDetailsTool(handler *daemon.PositionDetailsServiceHandler) *PositionDetailsTool {
	return &PositionDetailsTool{
		handler: handler,
		logger:  logger.New("PositionDetailsTool"),
	}
}

// Name returns the tool name
func (t *PositionDetailsTool) Name() string {
	return "position-details"
}

// Description returns the tool description
func (t *PositionDetailsTool) Description() string {
	return "Get detailed information about open positions for a symbol (P&L, swap, open time)"
}

// InputSchema returns the JSON schema for parameters
func (t *PositionDetailsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"symbol": map[string]interface{}{
				"type":        "string",
				"description": "Symbol name (e.g. EURUSD)",
			},
		},
		"required": []string{"symbol"},
	}
}

// Execute handles the position-details tool execution
func (t *PositionDetailsTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing position-details tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	symbol, ok := paramsMap["symbol"].(string)
	if !ok || symbol == "" {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	ctx := context.Background()

	resp, err := t.handler.GetPositionDetails(ctx, &api.PositionDetailsRequest{
		Symbol: symbol,
	})

	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to get position details for %s", symbol), err)
		return nil, fmt.Errorf("failed to get position details: %v", err)
	}

	t.logger.Info(fmt.Sprintf("Retrieved position details for %s", symbol))

	positions := make([]map[string]interface{}, len(resp.Positions))
	for i, pos := range resp.Positions {
		positions[i] = map[string]interface{}{
			"ticket":       pos.Ticket,
			"symbol":       pos.Symbol,
			"type":         pos.Type,
			"volume":       pos.Volume,
			"openPrice":    pos.OpenPrice,
			"currentPrice": pos.CurrentPrice,
			"profit":       pos.Profit,
			"swap":         pos.Swap,
			"openTime":     pos.OpenTime,
		}
	}

	return map[string]interface{}{
		"symbol":    symbol,
		"positions": positions,
		"count":     len(positions),
	}, nil
}
