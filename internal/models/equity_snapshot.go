package models

import "github.com/shopspring/decimal"

// EquitySnapshot represents account equity at point in time
type EquitySnapshot struct {
	AccountID int64
	Timestamp int64           // Unix epoch
	Equity    decimal.Decimal
	Balance   decimal.Decimal
}
