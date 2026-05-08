package models

import "github.com/shopspring/decimal"

// ModifyOrder represents modification to pending order
type ModifyOrder struct {
	Ticket     int64
	Price      decimal.Decimal
	StopLoss   decimal.Decimal
	TakeProfit decimal.Decimal
}
