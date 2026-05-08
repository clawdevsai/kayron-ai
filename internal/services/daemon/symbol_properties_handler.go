package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// SymbolPropertiesServiceHandler wraps symbol properties operations
type SymbolPropertiesServiceHandler struct {
	service *mt5.SymbolPropertiesService
	logger  *logger.Logger
}

// NewSymbolPropertiesServiceHandler creates a new symbol properties handler
func NewSymbolPropertiesServiceHandler(service *mt5.SymbolPropertiesService) *SymbolPropertiesServiceHandler {
	return &SymbolPropertiesServiceHandler{
		service: service,
		logger:  logger.New("SymbolPropertiesHandler"),
	}
}

// GetSymbolProperties handles symbol properties requests
func (h *SymbolPropertiesServiceHandler) GetSymbolProperties(ctx context.Context, req *api.SymbolPropertiesRequest) (*api.SymbolPropertiesResponse, error) {
	h.logger.Info("GetSymbolProperties request handling")

	props, err := h.service.GetSymbolProperties(ctx, req.Symbol)
	if err != nil {
		h.logger.Error("GetSymbolProperties failed", err)
		return &api.SymbolPropertiesResponse{
			Symbol: req.Symbol,
		}, nil
	}

	return &api.SymbolPropertiesResponse{
		Symbol:   props.Symbol,
		Digits:   props.Digits,
		TickSize: props.TickSize,
		LotMin:   props.LotMin,
		LotMax:   props.LotMax,
	}, nil
}
