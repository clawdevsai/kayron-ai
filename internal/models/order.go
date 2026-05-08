package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeBuy  OrderType = "buy"
	OrderTypeSell OrderType = "sell"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusFilled    OrderStatus = "filled"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRejected  OrderStatus = "rejected"
)

// Order represents a trading order
type Order struct {
	Ticket        int64            `json:"ticket"`
	Symbol        string           `json:"symbol"`
	Type          OrderType        `json:"type"`
	Volume        decimal.Decimal  `json:"volume"`
	Price         decimal.Decimal  `json:"price"`
	StopLoss      *decimal.Decimal `json:"stop_loss,omitempty"`
	TakeProfit    *decimal.Decimal `json:"take_profit,omitempty"`
	Status        OrderStatus      `json:"status"`
	FillPrice     decimal.Decimal  `json:"fill_price"`
	ProfitLoss    decimal.Decimal  `json:"profit_loss"`
	CreatedAt     time.Time        `json:"created_at"`
	FilledAt      *time.Time       `json:"filled_at,omitempty"`
	IdempotencyKey string          `json:"idempotency_key"`
}

// NewOrder creates a new Order
func NewOrder(symbol string, orderType OrderType, volume, price decimal.Decimal, idempotencyKey string) *Order {
	return &Order{
		Symbol:         symbol,
		Type:           orderType,
		Volume:         volume,
		Price:          price,
		Status:         OrderStatusPending,
		CreatedAt:      time.Now(),
		IdempotencyKey: idempotencyKey,
	}
}

// SetStopLoss sets the stop loss for the order
func (o *Order) SetStopLoss(sl decimal.Decimal) {
	o.StopLoss = &sl
}

// SetTakeProfit sets the take profit for the order
func (o *Order) SetTakeProfit(tp decimal.Decimal) {
	o.TakeProfit = &tp
}

// MarkFilled marks the order as filled
func (o *Order) MarkFilled(fillPrice decimal.Decimal) {
	o.Status = OrderStatusFilled
	o.FillPrice = fillPrice
	now := time.Now()
	o.FilledAt = &now
}
