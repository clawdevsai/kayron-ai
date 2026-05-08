package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// MarketHoursTool handles the market-hours MCP tool
type MarketHoursTool struct {
	handler *daemon.MarketHoursServiceHandler
	logger  *logger.Logger
}

// NewMarketHoursTool creates a new MarketHoursTool
func NewMarketHoursTool(handler *daemon.MarketHoursServiceHandler) *MarketHoursTool {
	return &MarketHoursTool{
		handler: handler,
		logger:  logger.New("MarketHoursTool"),
	}
}

// Name returns the tool name
func (t *MarketHoursTool) Name() string {
	return "market-hours"
}

// Description returns the tool description
func (t *MarketHoursTool) Description() string {
	return "Get market trading hours for a symbol (opening/closing times and timezone)"
}

// InputSchema returns the JSON schema for parameters
func (t *MarketHoursTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"symbol": map[string]interface{}{
				"type":        "string",
				"description": "Symbol name (e.g., EURUSD)",
			},
		},
		"required": []string{"symbol"},
	}
}

// Execute handles the market-hours tool execution
func (t *MarketHoursTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing market-hours tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	symbol, ok := paramsMap["symbol"].(string)
	if !ok || symbol == "" {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	ctx := context.Background()

	resp, err := t.handler.GetMarketHours(ctx, &api.MarketHoursRequest{
		Symbol: symbol,
	})

	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to get market hours for %s", symbol), err)
		return nil, fmt.Errorf("failed to get market hours: %v", err)
	}

	t.logger.Info(fmt.Sprintf("Retrieved market hours for %s: %04d-%04d", resp.Symbol, resp.OpenTime, resp.CloseTime))

	// Format times for readability
	openTimeStr := fmt.Sprintf("%02d:%02d", resp.OpenTime/100, resp.OpenTime%100)
	closeTimeStr := fmt.Sprintf("%02d:%02d", resp.CloseTime/100, resp.CloseTime%100)

	return map[string]interface{}{
		"symbol":     resp.Symbol,
		"open_time":  resp.OpenTime,
		"close_time": resp.CloseTime,
		"open_time_formatted":  openTimeStr,
		"close_time_formatted": closeTimeStr,
		"timezone":   resp.Timezone,
		"is_closed":  resp.IsClosed,
	}, nil
}
