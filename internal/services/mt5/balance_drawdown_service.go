package mt5

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/shopspring/decimal"
)

// BalanceDrawdownResult represents drawdown calculation result
type BalanceDrawdownResult struct {
	MaxEquity       string
	CurrentEquity   string
	DrawdownPercent float64
}

// BalanceDrawdownService handles balance drawdown calculations
type BalanceDrawdownService struct {
	client              *Client
	equityHistoryServ   *AccountEquityHistoryService
	logger              *logger.Logger
}

// NewBalanceDrawdownService creates a new balance drawdown service
func NewBalanceDrawdownService(client *Client, equityHistoryService *AccountEquityHistoryService) *BalanceDrawdownService {
	return &BalanceDrawdownService{
		client:            client,
		equityHistoryServ: equityHistoryService,
		logger:            logger.New("BalanceDrawdownService"),
	}
}

// CalculateDrawdown calculates max drawdown since a given timestamp
func (s *BalanceDrawdownService) CalculateDrawdown(ctx context.Context, sinceTimestamp int64) (*BalanceDrawdownResult, error) {
	s.logger.Info(fmt.Sprintf("Calculating drawdown since %d", sinceTimestamp))

	// Get current account info
	accountInfo, err := s.client.GetAccount()
	if err != nil {
		s.logger.Error("Failed to get account info", err)
		return nil, fmt.Errorf("failed to get account info: %v", err)
	}

	currentEquity := accountInfo.Equity

	// Get equity history since timestamp
	// Use a time range: from sinceTimestamp to now
	now := int64(0) // 0 means current time in the service
	snapshots, err := s.equityHistoryServ.GetEquityHistory(ctx, sinceTimestamp, now, "daily")
	if err != nil {
		s.logger.Error("Failed to get equity history", err)
		return nil, fmt.Errorf("failed to get equity history: %v", err)
	}

	// If no history, use current equity as max
	maxEquity := currentEquity

	// Find maximum equity in history
	for _, snap := range snapshots {
		equity, _ := decimal.NewFromString(snap.Equity)
		if equity.GreaterThan(maxEquity) {
			maxEquity = equity
		}
	}

	// Calculate drawdown percentage
	// drawdown = ((max - current) / max) * 100
	var drawdownPercent float64 = 0.0
	if maxEquity.GreaterThan(decimal.Zero) {
		drawdown := maxEquity.Sub(currentEquity).Div(maxEquity).Mul(decimal.NewFromInt(100))
		drawdownPercent, _ = drawdown.Float64()
	}

	return &BalanceDrawdownResult{
		MaxEquity:       maxEquity.String(),
		CurrentEquity:   currentEquity.String(),
		DrawdownPercent: drawdownPercent,
	}, nil
}
