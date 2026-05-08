package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// MarginCalculatorServiceHandler wraps margin calculation operations
type MarginCalculatorServiceHandler struct {
	service *mt5.MarginCalculatorService
	logger  *logger.Logger
}

// NewMarginCalculatorServiceHandler creates a new margin calculator handler
func NewMarginCalculatorServiceHandler(service *mt5.MarginCalculatorService) *MarginCalculatorServiceHandler {
	return &MarginCalculatorServiceHandler{
		service: service,
		logger:  logger.New("MarginCalculatorHandler"),
	}
}

// CalculateMarginRequirement handles margin calculation requests
func (h *MarginCalculatorServiceHandler) CalculateMarginRequirement(ctx context.Context, req *api.MarginCalculatorRequest) (*api.MarginCalculatorResponse, error) {
	h.logger.Info("CalculateMarginRequirement request handling")

	result, err := h.service.CalculateMarginRequirement(ctx, req.Symbol, req.Volume)
	if err != nil {
		h.logger.Error("CalculateMarginRequirement failed", err)
		return &api.MarginCalculatorResponse{
			MarginRequired: "0",
			MarginPercent:  0,
		}, nil
	}

	// Convert MarginPercentage decimal to float64
	marginPercent, _ := result.MarginPercentageDecimal.Float64()

	return &api.MarginCalculatorResponse{
		MarginRequired: result.MarginRequired,
		MarginPercent:  marginPercent,
	}, nil
}
