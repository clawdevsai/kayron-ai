package integration_test

import (
	"context"
	"testing"

	"github.com/lukeware/kayron-ai/internal/services/mt5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAccountInfoIntegration(t *testing.T) {
	// Setup
	mockClient := setupMockMT5Client()
	accountService := mt5.NewAccountService(mockClient)

	// Test
	account, err := accountService.GetAccount(context.Background())

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, "USD", account.Currency)
	assert.True(t, account.Balance.GreaterThan(decimal.Zero))
	assert.True(t, account.Equity.GreaterThan(decimal.Zero))
	assert.True(t, account.FreeMargin.GreaterThan(decimal.Zero))
}

func TestAccountInfoErrorHandling(t *testing.T) {
	t.Run("disconnected terminal", func(t *testing.T) {
		// Test error handling for disconnected MT5 terminal
		// This would typically fail with connection error
	})

	t.Run("invalid account", func(t *testing.T) {
		// Test error handling for invalid account
	})
}
