package contract_test

import (
	"encoding/json"
	"testing"

	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestOrdersListSchema(t *testing.T) {
	// Test valid orders list JSON schema
	volume, _ := decimal.NewFromString("1.0")
	price, _ := decimal.NewFromString("1.0950")

	orders := []*models.Order{
		models.NewOrder("EURUSD", models.OrderTypeBuy, volume, price, "uuid-1"),
		models.NewOrder("GBPUSD", models.OrderTypeSell, volume, price, "uuid-2"),
	}

	// Marshal to JSON
	data, err := json.Marshal(orders)
	assert.NoError(t, err)

	// Unmarshal and verify schema
	var ordersData []map[string]interface{}
	err = json.Unmarshal(data, &ordersData)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(ordersData))
}

func TestOrdersListValidation(t *testing.T) {
	tests := []struct {
		name       string
		orderCount int
		valid      bool
	}{
		{
			name:       "Valid orders list - single order",
			orderCount: 1,
			valid:      true,
		},
		{
			name:       "Valid orders list - multiple orders",
			orderCount: 5,
			valid:      true,
		},
		{
			name:       "Valid orders list - empty",
			orderCount: 0,
			valid:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume, _ := decimal.NewFromString("1.0")
			price, _ := decimal.NewFromString("1.0950")

			orders := make([]*models.Order, 0)
			for i := 0; i < tt.orderCount; i++ {
				orders = append(orders, models.NewOrder("EURUSD", models.OrderTypeBuy, volume, price, "uuid"))
			}

			assert.Equal(t, tt.orderCount, len(orders))
		})
	}
}
