package integration_test

import (
	"context"
	"testing"

	"github.com/lukeware/kayron-ai/internal/services/mt5"
	"github.com/stretchr/testify/assert"
)

func TestGetPendingOrdersIntegration(t *testing.T) {
	// Setup
	mockClient := setupMockMT5Client()
	ordersService := mt5.NewOrdersService(mockClient)

	// Test
	orders, err := ordersService.GetPendingOrders(context.Background())

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, orders)
	assert.IsType(t, []*struct{}{}, orders)
}

func TestGetPendingOrdersEmpty(t *testing.T) {
	mockClient := setupMockMT5Client()
	ordersService := mt5.NewOrdersService(mockClient)

	orders, err := ordersService.GetPendingOrders(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, orders)
	assert.Equal(t, 0, len(orders))
}

func TestGetPendingOrdersBySymbol(t *testing.T) {
	mockClient := setupMockMT5Client()
	ordersService := mt5.NewOrdersService(mockClient)

	orders, err := ordersService.GetPendingOrdersBySymbol(context.Background(), "EURUSD")

	assert.NoError(t, err)
	assert.NotNil(t, orders)
}
