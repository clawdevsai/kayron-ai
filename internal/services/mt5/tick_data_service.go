package mt5

import (
	"context"
	"fmt"
	"time"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/shopspring/decimal"
)

// TickItem represents a single tick (bid/ask quote)
type TickItem struct {
	Timestamp int64
	Bid       string
	Ask       string
}

// TickDataService handles tick data streaming/history
type TickDataService struct {
	client *Client
	logger *logger.Logger
}

// NewTickDataService creates a new tick data service
func NewTickDataService(client *Client) *TickDataService {
	return &TickDataService{
		client: client,
		logger: logger.New("TickDataService"),
	}
}

// GetTickData retrieves tick data for a symbol over a time period
func (s *TickDataService) GetTickData(ctx context.Context, symbol string, durationSeconds int32) ([]*TickItem, error) {
	s.logger.Info(fmt.Sprintf("Getting tick data for %s over %d seconds", symbol, durationSeconds))

	// Placeholder implementation - in production, would stream real ticks from MT5
	// Returns mock tick data at 100ms intervals

	if durationSeconds <= 0 {
		durationSeconds = 10 // Default to 10 seconds
	}
	if durationSeconds > 300 {
		durationSeconds = 300 // Max 5 minutes
	}

	ticks := []*TickItem{}
	now := time.Now().Unix()

	// Generate ticks at 100ms intervals
	intervalMs := int64(100)
	tickIntervalSec := intervalMs / 1000
	if tickIntervalSec < 1 {
		tickIntervalSec = 1
	}

	baseBid, _ := decimal.NewFromString("1.0950")
	baseAsk, _ := decimal.NewFromString("1.0955")

	for i := int32(0); i < durationSeconds; i += int32(tickIntervalSec) {
		// Mock: bid/ask fluctuate slightly
		variationFactor, _ := decimal.NewFromString("0.0001")
		variation := decimal.NewFromInt(int64(i%5) - 2).Mul(variationFactor)
		bid := baseBid.Add(variation)
		ask := baseAsk.Add(variation)

		ticks = append(ticks, &TickItem{
			Timestamp: now - int64(durationSeconds) + int64(i),
			Bid:       bid.String(),
			Ask:       ask.String(),
		})
	}

	s.logger.Info(fmt.Sprintf("Generated %d ticks for %s", len(ticks), symbol))
	return ticks, nil
}
