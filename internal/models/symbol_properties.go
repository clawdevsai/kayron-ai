package models

// SymbolProperties represents symbol trading properties
type SymbolProperties struct {
	Symbol   string
	Digits   int32  // Decimal places
	TickSize string // Minimum price movement
	LotMin   string // Minimum lot size
	LotMax   string // Maximum lot size
}
