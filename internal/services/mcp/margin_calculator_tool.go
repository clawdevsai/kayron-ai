package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// MarginCalculatorTool handles the margin-calculator MCP tool
type MarginCalculatorTool struct {
	handler *daemon.MarginCalculatorServiceHandler
	logger  *logger.Logger
}

// NewMarginCalculatorTool creates a new MarginCalculatorTool
func NewMarginCalculatorTool(handler *daemon.MarginCalculatorServiceHandler) *MarginCalculatorTool {
	return &MarginCalculatorTool{
		handler: handler,
		logger:  logger.New("MarginCalculatorTool"),
	}
}

// Name returns the tool name
func (t *MarginCalculatorTool) Name() string {
	return "margin-calculator"
}

// Description returns the tool description
func (t *MarginCalculatorTool) Description() string {
	return "Calculate margin requirement for a given volume on a trading symbol"
}

// InputSchema returns the JSON schema for parameters
func (t *MarginCalculatorTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"symbol": map[string]interface{}{
				"type":        "string",
				"description": "Symbol name (e.g. EURUSD)",
			},
			"volume": map[string]interface{}{
				"type":        "string",
				"description": "Trade volume as decimal (e.g. 0.1)",
			},
		},
		"required": []string{"symbol", "volume"},
	}
}

// Execute handles the margin-calculator tool execution
func (t *MarginCalculatorTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing margin-calculator tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	symbol, ok := paramsMap["symbol"].(string)
	if !ok || symbol == "" {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	volume, ok := paramsMap["volume"].(string)
	if !ok || volume == "" {
		return nil, fmt.Errorf("volume parameter is required")
	}

	ctx := context.Background()

	resp, err := t.handler.CalculateMarginRequirement(ctx, &api.MarginCalculatorRequest{
		Symbol: symbol,
		Volume: volume,
	})

	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to calculate margin for %s", symbol), err)
		return nil, fmt.Errorf("failed to calculate margin requirement: %v", err)
	}

	t.logger.Info(fmt.Sprintf("Calculated margin for %s", symbol))

	return map[string]interface{}{
		"marginRequired": resp.MarginRequired,
		"marginPercent":  resp.MarginPercent,
	}, nil
}
