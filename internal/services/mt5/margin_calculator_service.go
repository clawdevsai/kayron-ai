package mt5

import (
	"context"
	"fmt"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/shopspring/decimal"
)

// MarginCalculatorService handles margin calculations
type MarginCalculatorService struct {
	client *Client
	logger *logger.Logger
}

// NewMarginCalculatorService creates a new margin calculator service
func NewMarginCalculatorService(client *Client) *MarginCalculatorService {
	return &MarginCalculatorService{
		client: client,
		logger: logger.New("MarginCalculatorService"),
	}
}

// MarginRequirementResult represents margin calculation result
type MarginRequirementResult struct {
	MarginRequired        string          // Margin required as decimal string
	MarginPercentageDecimal decimal.Decimal // Margin % as decimal for float64 conversion
}

// CalculateMarginRequirement calculates required margin for a given volume
// Returns margin required in account currency
func (s *MarginCalculatorService) CalculateMarginRequirement(ctx context.Context, symbol string, volume string) (*MarginRequirementResult, error) {
	s.logger.Info(fmt.Sprintf("Calculating margin for %s volume=%s", symbol, volume))

	// Parse volume as decimal
	volumeDec, err := decimal.NewFromString(volume)
	if err != nil {
		return nil, fmt.Errorf("invalid volume: %v", err)
	}

	// Get current quote for symbol
	quote, err := s.client.GetQuote(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote for %s: %v", symbol, err)
	}

	currentPrice := quote.Ask

	// Get account info for balance
	accountInfo, err := s.client.GetAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %v", err)
	}

	accountBalance := accountInfo.Balance

	// Calculate margin requirement
	// Margin = Volume * ContractSize * Price / Leverage
	// For forex: Volume * 100000 * Price / Leverage
	// Assuming 1:100 leverage by default
	leverage := decimal.NewFromInt(100)
	contractSize := decimal.NewFromInt(100000)

	marginRequired := volumeDec.Mul(contractSize).Mul(currentPrice).Div(leverage)

	// Calculate margin percentage
	marginPercentage := marginRequired.Div(accountBalance).Mul(decimal.NewFromInt(100))

	return &MarginRequirementResult{
		MarginRequired:        marginRequired.String(),
		MarginPercentageDecimal: marginPercentage,
	}, nil
}
