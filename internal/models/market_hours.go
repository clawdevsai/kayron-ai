package models

// MarketHours represents trading hours for symbol
type MarketHours struct {
	Symbol    string
	OpenTime  int    // HHMM format (0900 = 9:00 AM)
	CloseTime int    // HHMM format (1700 = 5:00 PM)
	Timezone  string // "GMT", "EST", etc.
	IsClosed  bool   // true if market closed today
}
