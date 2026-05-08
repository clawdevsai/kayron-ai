package mt5

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
)

// PendingOrderService handles pending order queries
type PendingOrderService struct {
	client *Client
	logger *logger.Logger
}

// NewPendingOrderService creates a new pending order service
func NewPendingOrderService(client *Client) *PendingOrderService {
	return &PendingOrderService{
		client: client,
		logger: logger.New("PendingOrderService"),
	}
}

// GetPendingOrders retrieves pending orders with filters
func (ps *PendingOrderService) GetPendingOrders(ctx context.Context, symbol, status string, createdAfter int64) ([]*models.OrderItem, error) {
	ps.logger.Info(fmt.Sprintf("Querying pending orders: symbol=%s, status=%s", symbol, status))

	url := fmt.Sprintf("%s/orders?status=%s", ps.client.baseURL, status)
	if symbol != "" {
		url += fmt.Sprintf("&symbol=%s", symbol)
	}
	if createdAfter > 0 {
		url += fmt.Sprintf("&createdAfter=%d", createdAfter)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ps.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []*models.OrderItem{}, nil
	}

	var result struct {
		Orders []struct {
			Ticket    int64  `json:"ticket"`
			Symbol    string `json:"symbol"`
			Type      string `json:"type"`
			Volume    string `json:"volume"`
			Price     string `json:"price"`
			Status    string `json:"status"`
			OpenTime  int64  `json:"openTime"`
			FillPrice string `json:"fillPrice"`
			ProfitLoss string `json:"profitLoss"`
		} `json:"orders"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	orders := make([]*models.OrderItem, len(result.Orders))
	for i, o := range result.Orders {
		orders[i] = &models.OrderItem{
			Ticket:     o.Ticket,
			Symbol:     o.Symbol,
			Type:       o.Type,
			Volume:     o.Volume,
			Price:      o.Price,
			Status:     o.Status,
			OpenTime:   o.OpenTime,
			FillPrice:  o.FillPrice,
			ProfitLoss: o.ProfitLoss,
		}
	}

	return orders, nil
}
