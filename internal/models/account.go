package models

import (
	"github.com/shopspring/decimal"
)

// TradingAccount represents MT5 account information
type TradingAccount struct {
	Balance    decimal.Decimal `json:"balance"`
	Equity     decimal.Decimal `json:"equity"`
	Margin     decimal.Decimal `json:"margin"`
	FreeMargin decimal.Decimal `json:"free_margin"`
	Currency   string          `json:"currency"`
}

// NewTradingAccount creates a new TradingAccount
func NewTradingAccount(balance, equity, margin, freeMargin decimal.Decimal, currency string) *TradingAccount {
	return &TradingAccount{
		Balance:    balance,
		Equity:     equity,
		Margin:     margin,
		FreeMargin: freeMargin,
		Currency:   currency,
	}
}
