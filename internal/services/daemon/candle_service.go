package daemon

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// CandleServiceHandler handles gRPC candle requests
type CandleServiceHandler struct {
	mt5Service *mt5.CandleService
	logger     *logger.Logger
}

// NewCandleServiceHandler creates a new CandleServiceHandler
func NewCandleServiceHandler(mt5Service *mt5.CandleService) *CandleServiceHandler {
	return &CandleServiceHandler{
		mt5Service: mt5Service,
		logger:     logger.New("CandleServiceHandler"),
	}
}

// GetCandles handles the GetCandles gRPC call
func (h *CandleServiceHandler) GetCandles(ctx context.Context, req *api.GetCandlesRequest) (*api.GetCandlesResponse, error) {
	h.logger.Info(fmt.Sprintf("GetCandles request: symbol=%s timeframe=%s count=%d", req.Symbol, req.Timeframe, req.Count))

	// Validate timeframe
	if err := h.mt5Service.ValidateTimeframe(req.Timeframe); err != nil {
		h.logger.Error("Invalid timeframe", err)
		return nil, err
	}

	// Fetch candles
	candles, err := h.mt5Service.GetCandles(ctx, req.Symbol, req.Timeframe, req.Count)
	if err != nil {
		h.logger.Error("Failed to get candles", err)
		return nil, err
	}

	// Convert to API response
	apiCandles := make([]*api.CandleItem, 0, len(candles))
	for _, c := range candles {
		apiCandles = append(apiCandles, &api.CandleItem{
			Open:      c.Open.String(),
			High:      c.High.String(),
			Low:       c.Low.String(),
			Close:     c.Close.String(),
			Volume:    c.Volume,
			Timestamp: c.Timestamp.Unix(),
		})
	}

	resp := &api.GetCandlesResponse{
		Symbol:    req.Symbol,
		Timeframe: req.Timeframe,
		Candles:   apiCandles,
	}

	h.logger.Info(fmt.Sprintf("Candles retrieved: count=%d", len(apiCandles)))
	return resp, nil
}
