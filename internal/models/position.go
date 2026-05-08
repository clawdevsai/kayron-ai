package models

import (
	"github.com/shopspring/decimal"
)

// PositionType represents the type of position
type PositionType string

const (
	PositionTypeLong  PositionType = "long"
	PositionTypeShort PositionType = "short"
)

// Position represents an open trading position
type Position struct {
	Ticket       int64           `json:"ticket"`
	Symbol       string          `json:"symbol"`
	Type         PositionType    `json:"type"`
	Volume       decimal.Decimal `json:"volume"`
	EntryPrice   decimal.Decimal `json:"entry_price"`
	CurrentPrice decimal.Decimal `json:"current_price"`
	Profit       decimal.Decimal `json:"profit"`
}

// NewPosition creates a new Position
func NewPosition(ticket int64, symbol string, posType PositionType, volume, entryPrice decimal.Decimal) *Position {
	return &Position{
		Ticket:     ticket,
		Symbol:     symbol,
		Type:       posType,
		Volume:     volume,
		EntryPrice: entryPrice,
		Profit:     decimal.Zero,
	}
}

// UpdateProfit recalculates profit based on current price
func (p *Position) UpdateProfit(currentPrice decimal.Decimal) {
	p.CurrentPrice = currentPrice
	priceDiff := currentPrice.Sub(p.EntryPrice)

	if p.Type == PositionTypeLong {
		p.Profit = priceDiff.Mul(p.Volume)
	} else {
		p.Profit = priceDiff.Mul(p.Volume).Neg()
	}
}
