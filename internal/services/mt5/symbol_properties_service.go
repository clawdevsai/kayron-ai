package mt5

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
)

// SymbolPropertiesService handles symbol properties queries
type SymbolPropertiesService struct {
	client *Client
	logger *logger.Logger
}

// NewSymbolPropertiesService creates a new symbol properties service
func NewSymbolPropertiesService(client *Client) *SymbolPropertiesService {
	return &SymbolPropertiesService{
		client: client,
		logger: logger.New("SymbolPropertiesService"),
	}
}

// GetSymbolProperties retrieves symbol properties from MT5
func (sps *SymbolPropertiesService) GetSymbolProperties(ctx context.Context, symbol string) (*models.SymbolProperties, error) {
	sps.logger.Info(fmt.Sprintf("Querying properties for symbol: %s", symbol))

	url := fmt.Sprintf("%s/symbols/%s/properties", sps.client.baseURL, symbol)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := sps.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MT5 API error: %d", resp.StatusCode)
	}

	var result struct {
		Symbol string `json:"symbol"`
		Digits int32  `json:"digits"`
		TickSize string `json:"tickSize"`
		LotMin string `json:"lotMin"`
		LotMax string `json:"lotMax"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	props := &models.SymbolProperties{
		Symbol:   result.Symbol,
		Digits:   result.Digits,
		TickSize: result.TickSize,
		LotMin:   result.LotMin,
		LotMax:   result.LotMax,
	}

	return props, nil
}
