package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// AccountEquityHistoryTool handles the account-equity-history MCP tool
type AccountEquityHistoryTool struct {
	handler *daemon.AccountEquityHistoryServiceHandler
	logger  *logger.Logger
}

// NewAccountEquityHistoryTool creates a new AccountEquityHistoryTool
func NewAccountEquityHistoryTool(handler *daemon.AccountEquityHistoryServiceHandler) *AccountEquityHistoryTool {
	return &AccountEquityHistoryTool{
		handler: handler,
		logger:  logger.New("AccountEquityHistoryTool"),
	}
}

// Name returns the tool name
func (t *AccountEquityHistoryTool) Name() string {
	return "account-equity-history"
}

// Description returns the tool description
func (t *AccountEquityHistoryTool) Description() string {
	return "Get equity history snapshots for a date range (daily or hourly)"
}

// InputSchema returns the JSON schema for parameters
func (t *AccountEquityHistoryTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"from_timestamp": map[string]interface{}{
				"type":        "integer",
				"description": "Start timestamp (Unix epoch in seconds)",
			},
			"to_timestamp": map[string]interface{}{
				"type":        "integer",
				"description": "End timestamp (Unix epoch in seconds)",
			},
			"granularity": map[string]interface{}{
				"type":        "string",
				"description": "Snapshot granularity: 'daily' or 'hourly' (default: daily)",
				"enum":        []string{"daily", "hourly"},
			},
		},
		"required": []string{"from_timestamp", "to_timestamp"},
	}
}

// Execute handles the account-equity-history tool execution
func (t *AccountEquityHistoryTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing account-equity-history tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	fromTS, ok := paramsMap["from_timestamp"].(float64)
	if !ok {
		return nil, fmt.Errorf("from_timestamp parameter is required (integer)")
	}

	toTS, ok := paramsMap["to_timestamp"].(float64)
	if !ok {
		return nil, fmt.Errorf("to_timestamp parameter is required (integer)")
	}

	granularity, ok := paramsMap["granularity"].(string)
	if !ok {
		granularity = "daily"
	}

	// Validate timestamps
	if int64(fromTS) > int64(toTS) {
		return nil, fmt.Errorf("from_timestamp must be less than or equal to to_timestamp")
	}

	// Limit max range to 365 days for query efficiency
	maxRange := int64(365 * 24 * 3600) // 365 days in seconds
	if int64(toTS)-int64(fromTS) > maxRange {
		return nil, fmt.Errorf("date range cannot exceed 365 days")
	}

	ctx := context.Background()

	resp, err := t.handler.GetEquityHistory(ctx, &api.EquityHistoryRequest{
		FromTimestamp: int64(fromTS),
		ToTimestamp:   int64(toTS),
		Granularity:   granularity,
	})

	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to get equity history"), err)
		return nil, fmt.Errorf("failed to get equity history: %v", err)
	}

	t.logger.Info(fmt.Sprintf("Retrieved %d equity snapshots", len(resp.Snapshots)))

	snapshots := make([]map[string]interface{}, len(resp.Snapshots))
	for i, snap := range resp.Snapshots {
		snapshots[i] = map[string]interface{}{
			"timestamp": snap.Timestamp,
			"date":      time.Unix(snap.Timestamp, 0).Format("2006-01-02"),
			"equity":    snap.Equity,
			"balance":   snap.Balance,
		}
	}

	return map[string]interface{}{
		"from_timestamp": int64(fromTS),
		"to_timestamp":   int64(toTS),
		"granularity":    granularity,
		"snapshots":      snapshots,
		"count":          len(snapshots),
	}, nil
}
