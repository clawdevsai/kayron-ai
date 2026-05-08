package models

import "github.com/shopspring/decimal"

// Tick represents market tick (bid/ask snapshot)
type Tick struct {
	Symbol    string
	Timestamp int64           // Unix epoch
	Bid       decimal.Decimal
	Ask       decimal.Decimal
}
