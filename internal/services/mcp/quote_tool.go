package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// QuoteTool handles the quote MCP tool
type QuoteTool struct {
	handler *daemon.QuoteServiceHandler
	logger  *logger.Logger
}

// NewQuoteTool creates a new QuoteTool
func NewQuoteTool(handler *daemon.QuoteServiceHandler) *QuoteTool {
	return &QuoteTool{
		handler: handler,
		logger:  logger.New("QuoteTool"),
	}
}

// Execute handles the quote tool execution
func (t *QuoteTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing quote tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	symbol, ok := paramsMap["symbol"].(string)
	if !ok || symbol == "" {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	ctx := context.Background()

	// Call gRPC handler
	quote, err := t.handler.GetQuote(ctx, &api.GetQuoteRequest{Symbol: symbol})
	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to get quote for %s", symbol), err)
		return nil, fmt.Errorf("failed to retrieve quote for %s: %v", symbol, err)
	}

	result := map[string]interface{}{
		"symbol":    quote.Symbol,
		"bid":       quote.Bid,
		"ask":       quote.Ask,
		"spread":    quote.Spread,
		"timestamp": quote.Timestamp,
	}

	t.logger.Info(fmt.Sprintf("Quote retrieved for %s", symbol))
	return result, nil
}
