package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// OrderFillAnalysisServiceHandler wraps order fill analysis operations
type OrderFillAnalysisServiceHandler struct {
	service *mt5.OrderFillAnalysisService
	logger  *logger.Logger
}

// NewOrderFillAnalysisServiceHandler creates a new order fill analysis handler
func NewOrderFillAnalysisServiceHandler(service *mt5.OrderFillAnalysisService) *OrderFillAnalysisServiceHandler {
	return &OrderFillAnalysisServiceHandler{
		service: service,
		logger:  logger.New("OrderFillAnalysisHandler"),
	}
}

// AnalyzeOrderFill handles order fill analysis requests
func (h *OrderFillAnalysisServiceHandler) AnalyzeOrderFill(ctx context.Context, req *api.OrderFillAnalysisRequest) (*api.OrderFillAnalysisResponse, error) {
	h.logger.Info("AnalyzeOrderFill request handling")

	result, err := h.service.AnalyzeOrderFill(ctx, req.Ticket)
	if err != nil {
		h.logger.Error("AnalyzeOrderFill failed", err)
		return &api.OrderFillAnalysisResponse{
			Ticket:           req.Ticket,
			Symbol:           "",
			FillPrice:        "0",
			Slippage:         "0",
			ExecutionLatency: 0,
		}, nil
	}

	return &api.OrderFillAnalysisResponse{
		Ticket:           result.Ticket,
		Symbol:           result.Symbol,
		FillPrice:        result.FillPrice,
		Slippage:         result.Slippage,
		ExecutionLatency: result.ExecutionLatency,
	}, nil
}
