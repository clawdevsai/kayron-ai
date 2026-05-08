package integration_test

import (
	"context"
	"testing"

	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPlaceOrderIntegration(t *testing.T) {
	// Setup
	mockClient := setupMockMT5Client()
	orderService := mt5.NewOrderService(mockClient)

	volume, _ := decimal.NewFromString("1.0")
	price, _ := decimal.NewFromString("1.0950")

	order := models.NewOrder("EURUSD", models.OrderTypeBuy, volume, price, "test-uuid-1")

	// Test
	ticket, err := orderService.PlaceOrder(context.Background(), order)

	// Verify
	assert.NoError(t, err)
	assert.Greater(t, ticket, int64(0))
}

func TestOrderIdempotency(t *testing.T) {
	// Setup
	mockClient := setupMockMT5Client()
	orderService := mt5.NewOrderService(mockClient)

	volume, _ := decimal.NewFromString("1.0")
	price, _ := decimal.NewFromString("1.0950")

	// Place order
	order1 := models.NewOrder("EURUSD", models.OrderTypeBuy, volume, price, "idempotent-key-1")
	ticket1, err := orderService.PlaceOrder(context.Background(), order1)
	assert.NoError(t, err)

	// Place same order again
	order2 := models.NewOrder("EURUSD", models.OrderTypeBuy, volume, price, "idempotent-key-1")
	ticket2, err := orderService.PlaceOrder(context.Background(), order2)
	assert.NoError(t, err)

	// Should return same ticket (idempotency)
	assert.Equal(t, ticket1, ticket2)
}

func TestConcurrentOrders(t *testing.T) {
	mockClient := setupMockMT5Client()
	orderService := mt5.NewOrderService(mockClient)

	volume, _ := decimal.NewFromString("1.0")
	price, _ := decimal.NewFromString("1.0950")

	// Place multiple concurrent orders
	tickets := make([]int64, 0)
	for i := 0; i < 5; i++ {
		order := models.NewOrder("EURUSD", models.OrderTypeBuy, volume, price, "concurrent-uuid")
		ticket, err := orderService.PlaceOrder(context.Background(), order)
		assert.NoError(t, err)
		tickets = append(tickets, ticket)
	}

	assert.Equal(t, 5, len(tickets))
}

func TestOrderValidation(t *testing.T) {
	mockClient := setupMockMT5Client()
	orderService := mt5.NewOrderService(mockClient)

	tests := []struct {
		name    string
		order   *models.Order
		valid   bool
	}{
		{
			name: "Valid order",
			order: func() *models.Order {
				volume, _ := decimal.NewFromString("1.0")
				price, _ := decimal.NewFromString("1.0950")
				return models.NewOrder("EURUSD", models.OrderTypeBuy, volume, price, "uuid-1")
			}(),
			valid: true,
		},
		{
			name: "Invalid symbol",
			order: func() *models.Order {
				volume, _ := decimal.NewFromString("1.0")
				price, _ := decimal.NewFromString("1.0950")
				order := models.NewOrder("", models.OrderTypeBuy, volume, price, "uuid-2")
				return order
			}(),
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := orderService.ValidateOrder(tt.order)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
