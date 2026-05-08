package mt5

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/shopspring/decimal"
)

// CandleService handles MT5 candle data requests
type CandleService struct {
	client *Client
	logger *logger.Logger
}

// NewCandleService creates a new CandleService
func NewCandleService(client *Client) *CandleService {
	return &CandleService{
		client: client,
		logger: logger.New("CandleService"),
	}
}

// GetCandles fetches OHLC candles from MT5
func (cs *CandleService) GetCandles(ctx context.Context, symbol string, timeframe string, count int32) ([]models.Candle, error) {
	cs.logger.Info(fmt.Sprintf("Fetching %d %s candles for %s", count, timeframe, symbol))

	// Build request URL
	url := fmt.Sprintf("%s/symbols/%s/candles?tf=%s&count=%d", cs.client.baseURL, symbol, timeframe, count)

	// Make HTTP request
	resp, err := cs.client.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch candles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("MT5 API error: status %d", resp.StatusCode)
	}

	// Parse response
	var rawCandles []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawCandles); err != nil {
		return nil, fmt.Errorf("failed to parse candles response: %w", err)
	}

	// Convert to Candle models
	candles := make([]models.Candle, 0, len(rawCandles))
	for _, raw := range rawCandles {
		open, _ := decimal.NewFromString(fmt.Sprintf("%v", raw["o"]))
		high, _ := decimal.NewFromString(fmt.Sprintf("%v", raw["h"]))
		low, _ := decimal.NewFromString(fmt.Sprintf("%v", raw["l"]))
		close, _ := decimal.NewFromString(fmt.Sprintf("%v", raw["c"]))

		volume := int64(0)
		if v, ok := raw["v"].(float64); ok {
			volume = int64(v)
		}

		timestamp := int64(0)
		if t, ok := raw["t"].(float64); ok {
			timestamp = int64(t)
		}

		candles = append(candles, models.Candle{
			Symbol:    symbol,
			Timeframe: timeframe,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			Timestamp: time.Unix(timestamp, 0),
		})
	}

	cs.logger.Info(fmt.Sprintf("Retrieved %d candles for %s", len(candles), symbol))
	return candles, nil
}

// ValidateTimeframe checks if timeframe is supported
func (cs *CandleService) ValidateTimeframe(tf string) error {
	validFrames := map[string]bool{
		"M1": true, "M5": true, "M15": true, "H1": true, "D": true, "W": true,
	}
	if !validFrames[tf] {
		return fmt.Errorf("unsupported timeframe: %s", tf)
	}
	return nil
}
