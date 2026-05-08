package mt5

import (
	"context"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
)

// ModifyOrderService handles order modification
type ModifyOrderService struct {
	client *Client
	logger *logger.Logger
}

// NewModifyOrderService creates a new modify order service
func NewModifyOrderService(client *Client) *ModifyOrderService {
	return &ModifyOrderService{
		client: client,
		logger: logger.New("ModifyOrderService"),
	}
}

// ModifyOrder modifies an existing order
func (ms *ModifyOrderService) ModifyOrder(ctx context.Context, order *models.ModifyOrder) (*models.ModifyOrderResult, error) {
	ms.logger.Info(fmt.Sprintf("Modifying order %d", order.Ticket))
	url := fmt.Sprintf("%s/orders/%d/modify", ms.client.baseURL, order.Ticket)

	payload := map[string]interface{}{}
	if order.Price.IsPositive() {
		payload["price"] = order.Price.String()
	}
	if order.StopLoss.IsPositive() {
		payload["stopLoss"] = order.StopLoss.String()
	}
	if order.TakeProfit.IsPositive() {
		payload["takeProfit"] = order.TakeProfit.String()
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ms.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &models.ModifyOrderResult{
			Ticket:   order.Ticket,
			Status:   "error",
			ErrorMsg: "MT5 server error",
		}, nil
	}

	var result struct {
		Status   string `json:"status"`
		ErrorMsg string `json:"errorMsg,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &models.ModifyOrderResult{
		Ticket:   order.Ticket,
		Status:   result.Status,
		ErrorMsg: result.ErrorMsg,
	}, nil
}
