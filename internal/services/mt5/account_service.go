package mt5

import (
	"context"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
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

	// Call MT5 WebAPI client to get real account info
	accountInfo, err := as.client.GetAccount()
	if err != nil {
		as.logger.Error("Failed to retrieve account info from MT5", err)
		return nil, err
	}

	account := models.NewTradingAccount(
		accountInfo.Balance,
		accountInfo.Equity,
		accountInfo.Margin,
		accountInfo.FreeMargin,
		accountInfo.Currency,
	)
	as.logger.Info("Account information retrieved successfully")

	return account, nil
}
