package contract_test

import (
	"encoding/json"
	"testing"

	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestClosePositionSchema(t *testing.T) {
	// Test valid position JSON schema
	volume, _ := decimal.NewFromString("1.0")
	entryPrice, _ := decimal.NewFromString("1.0950")

	position := models.NewPosition(100001, "EURUSD", models.PositionTypeLong, volume, entryPrice)

	// Marshal to JSON
	data, err := json.Marshal(position)
	assert.NoError(t, err)

	// Unmarshal and verify schema
	var posData map[string]interface{}
	err = json.Unmarshal(data, &posData)
	assert.NoError(t, err)

	// Verify required fields
	assert.Contains(t, posData, "ticket")
	assert.Contains(t, posData, "symbol")
	assert.Contains(t, posData, "type")
	assert.Contains(t, posData, "volume")
	assert.Contains(t, posData, "entry_price")
	assert.Contains(t, posData, "profit")
}

func TestPositionProfitCalculation(t *testing.T) {
	tests := []struct {
		name        string
		posType     models.PositionType
		volume      string
		entryPrice  string
		currentPrice string
		expectedProfit string
	}{
		{
			name:         "Long position profit",
			posType:      models.PositionTypeLong,
			volume:       "1.0",
			entryPrice:   "1.0950",
			currentPrice: "1.0960",
			expectedProfit: "0.0010",
		},
		{
			name:         "Long position loss",
			posType:      models.PositionTypeLong,
			volume:       "1.0",
			entryPrice:   "1.0950",
			currentPrice: "1.0940",
			expectedProfit: "-0.0010",
		},
		{
			name:         "Short position profit",
			posType:      models.PositionTypeShort,
			volume:       "1.0",
			entryPrice:   "1.0950",
			currentPrice: "1.0940",
			expectedProfit: "0.0010",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume, _ := decimal.NewFromString(tt.volume)
			entryPrice, _ := decimal.NewFromString(tt.entryPrice)
			currentPrice, _ := decimal.NewFromString(tt.currentPrice)
			expectedProfit, _ := decimal.NewFromString(tt.expectedProfit)

			position := models.NewPosition(100001, "EURUSD", tt.posType, volume, entryPrice)
			position.UpdateProfit(currentPrice)

			assert.Equal(t, expectedProfit, position.Profit)
		})
	}
}
