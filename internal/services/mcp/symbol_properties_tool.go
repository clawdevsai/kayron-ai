package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// SymbolPropertiesTool handles the symbol-properties MCP tool
type SymbolPropertiesTool struct {
	handler *daemon.SymbolPropertiesServiceHandler
	logger  *logger.Logger
}

// NewSymbolPropertiesTool creates a new SymbolPropertiesTool
func NewSymbolPropertiesTool(handler *daemon.SymbolPropertiesServiceHandler) *SymbolPropertiesTool {
	return &SymbolPropertiesTool{
		handler: handler,
		logger:  logger.New("SymbolPropertiesTool"),
	}
}

// Name returns the tool name
func (t *SymbolPropertiesTool) Name() string {
	return "symbol-properties"
}

// Description returns the tool description
func (t *SymbolPropertiesTool) Description() string {
	return "Get symbol trading properties (digits, tick size, lot min/max)"
}

// InputSchema returns the JSON schema for parameters
func (t *SymbolPropertiesTool) InputSchema() map[string]interface{} {
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

// Execute handles the symbol-properties tool execution
func (t *SymbolPropertiesTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing symbol-properties tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	symbol, ok := paramsMap["symbol"].(string)
	if !ok || symbol == "" {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	ctx := context.Background()

	resp, err := t.handler.GetSymbolProperties(ctx, &api.SymbolPropertiesRequest{
		Symbol: symbol,
	})

	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to get properties for %s", symbol), err)
		return nil, fmt.Errorf("failed to retrieve symbol properties: %v", err)
	}

	t.logger.Info(fmt.Sprintf("Retrieved properties for %s", symbol))

	return map[string]interface{}{
		"symbol":   resp.Symbol,
		"digits":   resp.Digits,
		"tickSize": resp.TickSize,
		"lotMin":   resp.LotMin,
		"lotMax":   resp.LotMax,
	}, nil
}
