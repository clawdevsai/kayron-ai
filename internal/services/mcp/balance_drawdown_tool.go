package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// BalanceDrawdownTool handles the balance-drawdown MCP tool
type BalanceDrawdownTool struct {
	handler *daemon.BalanceDrawdownServiceHandler
	logger  *logger.Logger
}

// NewBalanceDrawdownTool creates a new BalanceDrawdownTool
func NewBalanceDrawdownTool(handler *daemon.BalanceDrawdownServiceHandler) *BalanceDrawdownTool {
	return &BalanceDrawdownTool{
		handler: handler,
		logger:  logger.New("BalanceDrawdownTool"),
	}
}

// Name returns the tool name
func (t *BalanceDrawdownTool) Name() string {
	return "balance-drawdown"
}

// Description returns the tool description
func (t *BalanceDrawdownTool) Description() string {
	return "Calculate maximum drawdown percentage since a given timestamp"
}

// InputSchema returns the JSON schema for parameters
func (t *BalanceDrawdownTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"since_timestamp": map[string]interface{}{
				"type":        "integer",
				"description": "Start timestamp (Unix epoch in seconds, e.g., account creation time)",
			},
		},
		"required": []string{"since_timestamp"},
	}
}

// Execute handles the balance-drawdown tool execution
func (t *BalanceDrawdownTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing balance-drawdown tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	sinceTS, ok := paramsMap["since_timestamp"].(float64)
	if !ok {
		return nil, fmt.Errorf("since_timestamp parameter is required (integer)")
	}

	ctx := context.Background()

	resp, err := t.handler.CalculateDrawdown(ctx, &api.BalanceDrawdownRequest{
		SinceTimestamp: int64(sinceTS),
	})

	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to calculate drawdown"), err)
		return nil, fmt.Errorf("failed to calculate drawdown: %v", err)
	}

	t.logger.Info(fmt.Sprintf("Calculated drawdown: %.2f%%", resp.DrawdownPercent))

	return map[string]interface{}{
		"max_equity":        resp.MaxEquity,
		"current_equity":    resp.CurrentEquity,
		"drawdown_percent":  resp.DrawdownPercent,
		"since_timestamp":   int64(sinceTS),
		"since_date":        time.Unix(int64(sinceTS), 0).Format("2006-01-02"),
	}, nil
}
