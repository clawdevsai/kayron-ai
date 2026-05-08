package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// CandleTool handles the get-candles MCP tool
type CandleTool struct {
	handler *daemon.CandleServiceHandler
	logger  *logger.Logger
}

// NewCandleTool creates a new CandleTool
func NewCandleTool(handler *daemon.CandleServiceHandler) *CandleTool {
	return &CandleTool{
		handler: handler,
		logger:  logger.New("CandleTool"),
	}
}

// Execute handles the get-candles tool execution
func (t *CandleTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing get-candles tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters format")
	}

	symbol, ok := paramsMap["symbol"].(string)
	if !ok || symbol == "" {
		return nil, fmt.Errorf("symbol parameter required")
	}

	timeframe, ok := paramsMap["timeframe"].(string)
	if !ok || timeframe == "" {
		return nil, fmt.Errorf("timeframe parameter required")
	}

	countVal, ok := paramsMap["count"].(float64)
	if !ok {
		return nil, fmt.Errorf("count parameter required")
	}
	count := int32(countVal)

	ctx := context.Background()

	// Call gRPC handler
	response, err := t.handler.GetCandles(ctx, &api.GetCandlesRequest{
		Symbol:    symbol,
		Timeframe: timeframe,
		Count:     count,
	})
	if err != nil {
		t.logger.Error("Failed to get candles", err)
		return nil, fmt.Errorf("failed to retrieve candles: %v", err)
	}

	candles := make([]map[string]interface{}, 0, len(response.Candles))
	for _, c := range response.Candles {
		candles = append(candles, map[string]interface{}{
			"open":      c.Open,
			"high":      c.High,
			"low":       c.Low,
			"close":     c.Close,
			"volume":    c.Volume,
			"timestamp": c.Timestamp,
		})
	}

	result := map[string]interface{}{
		"symbol":    response.Symbol,
		"timeframe": response.Timeframe,
		"candles":   candles,
	}

	t.logger.Info(fmt.Sprintf("Candles retrieved: count=%d", len(candles)))
	return result, nil
}
