package integration_test

import (
	"context"
	"testing"

	"github.com/lukeware/kayron-ai/internal/services/mt5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestQuoteIntegration(t *testing.T) {
	// Setup
	mockClient := setupMockMT5Client()
	quoteService := mt5.NewQuoteService(mockClient)

	// Test
	quote, err := quoteService.GetQuote(context.Background(), "EURUSD")

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, quote)
	assert.Equal(t, "EURUSD", quote.Symbol)
	assert.True(t, quote.Bid.GreaterThan(decimal.Zero))
	assert.True(t, quote.Ask.GreaterThan(decimal.Zero))
	assert.True(t, quote.Ask.GreaterThanOrEqual(quote.Bid))
}

func TestQuoteStaleDataHandling(t *testing.T) {
	t.Run("quote caching", func(t *testing.T) {
		// Test quote caching behavior
		// Should return cached quote if available and not stale
	})

	t.Run("stale quote detection", func(t *testing.T) {
		// Test detection of stale quotes
		// Should mark quote as stale based on timestamp
	})
}

func TestQuoteMultipleSymbols(t *testing.T) {
	mockClient := setupMockMT5Client()
	quoteService := mt5.NewQuoteService(mockClient)

	symbols := []string{"EURUSD", "GBPUSD", "USDJPY"}
	for _, symbol := range symbols {
		quote, err := quoteService.GetQuote(context.Background(), symbol)
		assert.NoError(t, err)
		assert.Equal(t, symbol, quote.Symbol)
	}
}
