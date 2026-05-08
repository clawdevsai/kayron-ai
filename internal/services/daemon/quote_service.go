package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// QuoteServiceHandler handles gRPC quote requests
type QuoteServiceHandler struct {
	mt5Service *mt5.QuoteService
	logger     *logger.Logger
}

// NewQuoteServiceHandler creates a new QuoteServiceHandler
func NewQuoteServiceHandler(mt5Service *mt5.QuoteService) *QuoteServiceHandler {
	return &QuoteServiceHandler{
		mt5Service: mt5Service,
		logger:     logger.New("QuoteServiceHandler"),
	}
}

// GetQuote handles the GetQuote gRPC call
func (h *QuoteServiceHandler) GetQuote(ctx context.Context, req *api.GetQuoteRequest) (*api.Quote, error) {
	h.logger.Info("GetQuote request for " + req.Symbol)

	quote, err := h.mt5Service.GetQuote(ctx, req.Symbol)
	if err != nil {
		h.logger.Error("Failed to get quote", err)
		return nil, err
	}

	resp := &api.Quote{
		Symbol:    quote.Symbol,
		Bid:       quote.Bid.String(),
		Ask:       quote.Ask.String(),
		Spread:    quote.Spread.String(),
		Timestamp: quote.Timestamp.Unix(),
	}

	h.logger.Info("Quote response sent for " + quote.Symbol)
	return resp, nil
}
