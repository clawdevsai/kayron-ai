package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Instrument represents a trading instrument
type Instrument struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

// Quote represents market quote for an instrument
type Quote struct {
	Symbol    string          `json:"symbol"`
	Bid       decimal.Decimal `json:"bid"`
	Ask       decimal.Decimal `json:"ask"`
	Spread    decimal.Decimal `json:"spread"`
	Timestamp time.Time       `json:"timestamp"`
}

// NewQuote creates a new Quote
func NewQuote(symbol string, bid, ask decimal.Decimal, timestamp time.Time) *Quote {
	spread := ask.Sub(bid)
	return &Quote{
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Spread:    spread,
		Timestamp: timestamp,
	}
}
