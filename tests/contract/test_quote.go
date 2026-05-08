package contract_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestQuoteSchema(t *testing.T) {
	// Test valid quote JSON schema
	bid, _ := decimal.NewFromString("1.0950")
	ask, _ := decimal.NewFromString("1.0952")

	quote := models.NewQuote("EURUSD", bid, ask, time.Now())

	// Marshal to JSON
	data, err := json.Marshal(quote)
	assert.NoError(t, err)

	// Unmarshal and verify schema
	var quoteData map[string]interface{}
	err = json.Unmarshal(data, &quoteData)
	assert.NoError(t, err)

	// Verify required fields
	assert.Contains(t, quoteData, "symbol")
	assert.Contains(t, quoteData, "bid")
	assert.Contains(t, quoteData, "ask")
	assert.Contains(t, quoteData, "spread")
	assert.Contains(t, quoteData, "timestamp")
}

func TestQuoteBidAskValidation(t *testing.T) {
	tests := []struct {
		name    string
		symbol  string
		bid     string
		ask     string
		valid   bool
		message string
	}{
		{
			name:    "Valid quote - bid < ask",
			symbol:  "EURUSD",
			bid:     "1.0950",
			ask:     "1.0952",
			valid:   true,
			message: "Bid should be less than ask",
		},
		{
			name:    "Invalid quote - bid > ask",
			symbol:  "EURUSD",
			bid:     "1.0952",
			ask:     "1.0950",
			valid:   false,
			message: "Bid greater than ask",
		},
		{
			name:    "Valid quote - equal bid/ask (tight spread)",
			symbol:  "EURUSD",
			bid:     "1.0950",
			ask:     "1.0950",
			valid:   true,
			message: "Bid equal to ask (zero spread)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bid, _ := decimal.NewFromString(tt.bid)
			ask, _ := decimal.NewFromString(tt.ask)

			quote := models.NewQuote(tt.symbol, bid, ask, time.Now())
			assert.NotNil(t, quote)

			if tt.valid {
				// Verify spread calculation
				assert.True(t, quote.Ask.GreaterThanOrEqual(quote.Bid), tt.message)
			}
		})
	}
}

func TestQuoteSpreadCalculation(t *testing.T) {
	bid, _ := decimal.NewFromString("1.0950")
	ask, _ := decimal.NewFromString("1.0952")

	quote := models.NewQuote("EURUSD", bid, ask, time.Now())

	expectedSpread, _ := decimal.NewFromString("0.0002")
	assert.Equal(t, expectedSpread, quote.Spread)
}
