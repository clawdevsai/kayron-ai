package models

// OrderItem represents an order item
type OrderItem struct {
	Ticket     int64
	Symbol     string
	Type       string
	Volume     string
	Price      string
	Status     string
	OpenTime   int64
	FillPrice  string
	ProfitLoss string
}
