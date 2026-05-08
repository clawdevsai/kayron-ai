package mt5

import (
	"context"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/shopspring/decimal"
)

// AccountService handles MT5 account queries
type AccountService struct {
	client *Client
	logger *logger.Logger
}

// NewAccountService creates a new AccountService
func NewAccountService(client *Client) *AccountService {
	return &AccountService{
		client: client,
		logger: logger.New("AccountService"),
	}
}

// GetAccount retrieves the current account information from MT5
func (as *AccountService) GetAccount(ctx context.Context) (*models.TradingAccount, error) {
	as.logger.Info("Querying account information")

	// Call MT5 client to get account info
	// This is a placeholder - actual implementation depends on MT5 API
	balance, _ := decimal.NewFromString("10000.00")
	equity, _ := decimal.NewFromString("10500.00")
	margin, _ := decimal.NewFromString("2000.00")
	freeMargin, _ := decimal.NewFromString("8500.00")

	account := models.NewTradingAccount(balance, equity, margin, freeMargin, "USD")
	as.logger.Info("Account information retrieved successfully")

	return account, nil
}
