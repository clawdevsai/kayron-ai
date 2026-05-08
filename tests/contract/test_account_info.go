package contract_test

import (
	"encoding/json"
	"testing"

	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAccountInfoSchema(t *testing.T) {
	// Test valid account info JSON schema
	account := models.NewTradingAccount(
		decimal.NewFromInt(10000),
		decimal.NewFromInt(10500),
		decimal.NewFromInt(2000),
		decimal.NewFromInt(8500),
		"USD",
	)

	// Marshal to JSON
	data, err := json.Marshal(account)
	assert.NoError(t, err)

	// Unmarshal and verify schema
	var accountData map[string]interface{}
	err = json.Unmarshal(data, &accountData)
	assert.NoError(t, err)

	// Verify required fields
	assert.Contains(t, accountData, "balance")
	assert.Contains(t, accountData, "equity")
	assert.Contains(t, accountData, "margin")
	assert.Contains(t, accountData, "free_margin")
	assert.Contains(t, accountData, "currency")
}

func TestAccountInfoValidation(t *testing.T) {
	tests := []struct {
		name      string
		balance   decimal.Decimal
		equity    decimal.Decimal
		margin    decimal.Decimal
		freeMargin decimal.Decimal
		currency  string
		valid     bool
	}{
		{
			name:       "Valid account",
			balance:    decimal.NewFromInt(10000),
			equity:     decimal.NewFromInt(10500),
			margin:     decimal.NewFromInt(2000),
			freeMargin: decimal.NewFromInt(8500),
			currency:   "USD",
			valid:      true,
		},
		{
			name:       "Zero balance",
			balance:    decimal.Zero,
			equity:     decimal.NewFromInt(500),
			margin:     decimal.NewFromInt(2000),
			freeMargin: decimal.NewFromInt(8500),
			currency:   "USD",
			valid:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := models.NewTradingAccount(tt.balance, tt.equity, tt.margin, tt.freeMargin, tt.currency)
			assert.NotNil(t, account)
			assert.Equal(t, tt.currency, account.Currency)
		})
	}
}
