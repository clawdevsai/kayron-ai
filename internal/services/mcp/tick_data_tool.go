package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// TickDataTool handles the tick-data MCP tool
type TickDataTool struct {
	handler *daemon.TickDataServiceHandler
	logger  *logger.Logger
}

// NewTickDataTool creates a new TickDataTool
func NewTickDataTool(handler *daemon.TickDataServiceHandler) *TickDataTool {
	return &TickDataTool{
		handler: handler,
		logger:  logger.New("TickDataTool"),
	}
}

// Name returns the tool name
func (t *TickDataTool) Name() string {
	return "tick-data"
}

// Description returns the tool description
func (t *TickDataTool) Description() string {
	return "Get tick data (bid/ask quotes) for a symbol over a specified duration"
}

// InputSchema returns the JSON schema for parameters
func (t *TickDataTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"symbol": map[string]interface{}{
				"type":        "string",
				"description": "Symbol name (e.g., EURUSD)",
			},
			"duration_seconds": map[string]interface{}{
				"type":        "integer",
				"description": "Duration in seconds (1-300, default: 10)",
			},
		},
		"required": []string{"symbol"},
	}
}

// Execute handles the tick-data tool execution
func (t *TickDataTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing tick-data tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	symbol, ok := paramsMap["symbol"].(string)
	if !ok || symbol == "" {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	duration := int32(10) // Default 10 seconds
	if durFloat, ok := paramsMap["duration_seconds"].(float64); ok {
		duration = int32(durFloat)
	}

	// Validate duration
	if duration < 1 {
		duration = 1
	}
	if duration > 300 {
		duration = 300
	}

	ctx := context.Background()

	resp, err := t.handler.GetTickData(ctx, &api.TickDataRequest{
		Symbol:          symbol,
		DurationSeconds: duration,
	})

	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to get tick data for %s", symbol), err)
		return nil, fmt.Errorf("failed to get tick data: %v", err)
	}

	t.logger.Info(fmt.Sprintf("Retrieved %d ticks for %s", len(resp.Ticks), resp.Symbol))

	// Format ticks for output
	ticks := make([]map[string]interface{}, len(resp.Ticks))
	for i, tick := range resp.Ticks {
		ticks[i] = map[string]interface{}{
			"timestamp": tick.Timestamp,
			"time":      time.Unix(tick.Timestamp, 0).Format("2006-01-02 15:04:05"),
			"bid":       tick.Bid,
			"ask":       tick.Ask,
		}
	}

	return map[string]interface{}{
		"symbol":            resp.Symbol,
		"ticks":             ticks,
		"tick_count":        len(ticks),
		"duration_seconds":  duration,
	}, nil
}
