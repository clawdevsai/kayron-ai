package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Candle represents OHLC price bar
type Candle struct {
	Symbol    string
	Timeframe string // M1, M5, M15, H1, D, W
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    int64
	Timestamp time.Time
}

// TimeframeToMinutes converts timeframe string to minutes
func TimeframeToMinutes(tf string) int {
	switch tf {
	case "M1":
		return 1
	case "M5":
		return 5
	case "M15":
		return 15
	case "H1":
		return 60
	case "D":
		return 1440
	case "W":
		return 10080
	default:
		return 60 // default H1
	}
}
