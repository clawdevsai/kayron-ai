package mt5

import (
	"context"
	"fmt"
	"time"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
)

// QuoteService handles MT5 quote queries
type QuoteService struct {
	client *Client
	logger *logger.Logger
}

// NewQuoteService creates a new QuoteService
func NewQuoteService(client *Client) *QuoteService {
	return &QuoteService{
		client: client,
		logger: logger.New("QuoteService"),
	}
}

// GetQuote retrieves the current quote for a symbol from MT5
func (qs *QuoteService) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	qs.logger.Info(fmt.Sprintf("Querying quote for %s", symbol))

	// Call MT5 WebAPI client to get real quote
	quoteData, err := qs.client.GetQuote(symbol)
	if err != nil {
		qs.logger.Error(fmt.Sprintf("Failed to retrieve quote for %s", symbol), err)
		return nil, err
	}

	quote := models.NewQuote(symbol, quoteData.Bid, quoteData.Ask, time.Now())
	qs.logger.Info(fmt.Sprintf("Quote retrieved for %s: bid=%v, ask=%v", symbol, quoteData.Bid, quoteData.Ask))

	return quote, nil
}
