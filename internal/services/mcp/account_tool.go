package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// AccountInfoTool handles the account-info MCP tool
type AccountInfoTool struct {
	handler *daemon.AccountServiceHandler
	logger  *logger.Logger
}

// NewAccountInfoTool creates a new AccountInfoTool
func NewAccountInfoTool(handler *daemon.AccountServiceHandler) *AccountInfoTool {
	return &AccountInfoTool{
		handler: handler,
		logger:  logger.New("AccountInfoTool"),
	}
}

// Execute handles the account-info tool execution
func (t *AccountInfoTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing account-info tool")

	ctx := context.Background()

	// Call gRPC handler
	accountInfo, err := t.handler.GetAccountInfo(ctx, &api.AccountInfoRequest{})
	if err != nil {
		t.logger.Error("Failed to get account info", err)
		return nil, fmt.Errorf("failed to retrieve account information: %v", err)
	}

	result := map[string]interface{}{
		"balance":      accountInfo.Balance,
		"equity":       accountInfo.Equity,
		"margin_used":  accountInfo.MarginUsed,
		"margin_free":  accountInfo.MarginFree,
		"currency":     accountInfo.Currency,
	}

	t.logger.Info("Account info retrieved successfully")
	return result, nil
}
