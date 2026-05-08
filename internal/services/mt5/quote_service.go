package mt5

import (
	"context"
	"fmt"
	"time"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/shopspring/decimal"
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

	// Call MT5 client to get quote
	// This is a placeholder - actual implementation depends on MT5 API
	bid, _ := decimal.NewFromString("1.0950")
	ask, _ := decimal.NewFromString("1.0952")

	quote := models.NewQuote(symbol, bid, ask, time.Now())
	qs.logger.Info(fmt.Sprintf("Quote retrieved for %s: bid=%v, ask=%v", symbol, bid, ask))

	return quote, nil
}
