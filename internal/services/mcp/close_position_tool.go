package mcp

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
)

// ClosePositionTool handles the close-position MCP tool
type ClosePositionTool struct {
	handler *daemon.PositionServiceHandler
	logger  *logger.Logger
}

// NewClosePositionTool creates a new ClosePositionTool
func NewClosePositionTool(handler *daemon.PositionServiceHandler) *ClosePositionTool {
	return &ClosePositionTool{
		handler: handler,
		logger:  logger.New("ClosePositionTool"),
	}
}

// Execute handles the close-position tool execution
func (t *ClosePositionTool) Execute(params interface{}) (interface{}, error) {
	t.logger.Info("Executing close-position tool")

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters")
	}

	ticketFloat, ok := paramsMap["ticket"].(float64)
	if !ok {
		return nil, fmt.Errorf("ticket parameter is required and must be a number")
	}

	ticket := int64(ticketFloat)

	ctx := context.Background()

	// Call gRPC handler
	closeResp, err := t.handler.ClosePosition(ctx, &api.ClosePositionRequest{Ticket: ticket})
	if err != nil {
		t.logger.Error("Failed to close position", err)
		return nil, fmt.Errorf("failed to close position: %v", err)
	}

	result := map[string]interface{}{
		"ticket":      closeResp.Ticket,
		"profit_loss": closeResp.ProfitLoss,
	}

	t.logger.Info(fmt.Sprintf("Position closed: ticket=%d", closeResp.Ticket))
	return result, nil
}
