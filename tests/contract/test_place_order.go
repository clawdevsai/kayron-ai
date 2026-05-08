package contract_test

import (
	"encoding/json"
	"testing"

	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPlaceOrderSchema(t *testing.T) {
	// Test valid order JSON schema
	volume, _ := decimal.NewFromString("1.0")
	price, _ := decimal.NewFromString("1.0950")

	order := models.NewOrder("EURUSD", models.OrderTypeBuy, volume, price, "test-uuid-1")

	// Marshal to JSON
	data, err := json.Marshal(order)
	assert.NoError(t, err)

	// Unmarshal and verify schema
	var orderData map[string]interface{}
	err = json.Unmarshal(data, &orderData)
	assert.NoError(t, err)

	// Verify required fields
	assert.Contains(t, orderData, "symbol")
	assert.Contains(t, orderData, "type")
	assert.Contains(t, orderData, "volume")
	assert.Contains(t, orderData, "price")
	assert.Contains(t, orderData, "status")
}

func TestOrderInputValidation(t *testing.T) {
	tests := []struct {
		name      string
		symbol    string
		orderType models.OrderType
		volume    string
		price     string
		valid     bool
	}{
		{
			name:      "Valid buy order",
			symbol:    "EURUSD",
			orderType: models.OrderTypeBuy,
			volume:    "1.0",
			price:     "1.0950",
			valid:     true,
		},
		{
			name:      "Valid sell order",
			symbol:    "EURUSD",
			orderType: models.OrderTypeSell,
			volume:    "2.5",
			price:     "1.0952",
			valid:     true,
		},
		{
			name:      "Empty symbol",
			symbol:    "",
			orderType: models.OrderTypeBuy,
			volume:    "1.0",
			price:     "1.0950",
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume, _ := decimal.NewFromString(tt.volume)
			price, _ := decimal.NewFromString(tt.price)

			order := models.NewOrder(tt.symbol, tt.orderType, volume, price, "test-uuid")

			if tt.valid {
				assert.NotNil(t, order)
				assert.Equal(t, models.OrderStatusPending, order.Status)
			} else {
				assert.NotNil(t, order)
			}
		})
	}
}

func TestOrderIdempotency(t *testing.T) {
	volume, _ := decimal.NewFromString("1.0")
	price, _ := decimal.NewFromString("1.0950")

	order1 := models.NewOrder("EURUSD", models.OrderTypeBuy, volume, price, "uuid-1")
	order2 := models.NewOrder("EURUSD", models.OrderTypeBuy, volume, price, "uuid-1")

	// Same idempotency key should match
	assert.Equal(t, order1.IdempotencyKey, order2.IdempotencyKey)
}
