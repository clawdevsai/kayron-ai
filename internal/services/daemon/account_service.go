package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// AccountServiceHandler handles gRPC account info requests
type AccountServiceHandler struct {
	mt5Service *mt5.AccountService
	logger     *logger.Logger
}

// NewAccountServiceHandler creates a new AccountServiceHandler
func NewAccountServiceHandler(mt5Service *mt5.AccountService) *AccountServiceHandler {
	return &AccountServiceHandler{
		mt5Service: mt5Service,
		logger:     logger.New("AccountServiceHandler"),
	}
}

// GetAccountInfo handles the GetAccountInfo gRPC call
func (h *AccountServiceHandler) GetAccountInfo(ctx context.Context, req *api.GetAccountInfoRequest) (*api.AccountInfo, error) {
	h.logger.Info("GetAccountInfo request received")

	account, err := h.mt5Service.GetAccount(ctx)
	if err != nil {
		h.logger.Error("Failed to get account info", err)
		return nil, err
	}

	resp := &api.AccountInfo{
		Balance:    account.Balance.String(),
		Equity:     account.Equity.String(),
		MarginUsed: account.Margin.String(),
		MarginFree: account.FreeMargin.String(),
		Currency:   account.Currency,
	}

	h.logger.Info("AccountInfo response sent successfully")
	return resp, nil
}
