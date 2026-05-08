package mt5

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lukeware/kayron-ai/internal/errors"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/shopspring/decimal"
)

// Client wraps the MT5 WebAPI HTTP client
type Client struct {
	baseURL    string
	login      string
	password   string
	httpClient *http.Client
	logger     *logger.Logger
	timeout    time.Duration
}

// NewClient creates a new MT5 client
func NewClient(baseURL, login, password string, timeout time.Duration) *Client {
	return &Client{
		baseURL:  baseURL,
		login:    login,
		password: password,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger:  logger.New("MT5Client"),
		timeout: timeout,
	}
}

// AccountInfo represents MT5 account information
type AccountInfo struct {
	Login       int64           `json:"login"`
	Balance     decimal.Decimal `json:"balance"`
	Equity      decimal.Decimal `json:"equity"`
	Margin      decimal.Decimal `json:"margin"`
	FreeMargin  decimal.Decimal `json:"free_margin"`
	MarginLevel decimal.Decimal `json:"margin_level"`
	Currency    string          `json:"currency"`
}

// Quote represents a market quote
type Quote struct {
	Symbol string          `json:"symbol"`
	Bid    decimal.Decimal `json:"bid"`
	Ask    decimal.Decimal `json:"ask"`
	Time   int64           `json:"time"`
}

// Order represents an MT5 order
type Order struct {
	Ticket     uint64          `json:"ticket"`
	Symbol     string          `json:"symbol"`
	Side       string          `json:"side"`
	Volume     decimal.Decimal `json:"volume"`
	OpenPrice  decimal.Decimal `json:"open_price"`
	OpenTime   int64           `json:"open_time"`
	StopLoss   decimal.Decimal `json:"stop_loss"`
	TakeProfit decimal.Decimal `json:"take_profit"`
	Status     string          `json:"status"`
	Comment    string          `json:"comment"`
}

// GetAccount retrieves account information
func (c *Client) GetAccount() (*AccountInfo, error) {
	startTime := time.Now()

	url := fmt.Sprintf("%s/api/account", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.logger.Error("Failed to create request", err)
		return nil, errors.ConnectionFailed(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.login, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		latency := time.Since(startTime).Milliseconds()
		c.logger.ErrorWithLatency("Request failed", err, latency)
		return nil, errors.ConnectionFailed(err.Error())
	}
	defer resp.Body.Close()

	latency := time.Since(startTime).Milliseconds()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.AuthenticationFailed("Invalid credentials")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var accountInfo AccountInfo
	if err := json.NewDecoder(resp.Body).Decode(&accountInfo); err != nil {
		c.logger.ErrorWithLatency("Failed to decode response", err, latency)
		return nil, err
	}

	c.logger.InfoWithLatency(fmt.Sprintf("Retrieved account info: %d", accountInfo.Login), latency)
	return &accountInfo, nil
}

// GetQuote retrieves a market quote
func (c *Client) GetQuote(symbol string) (*Quote, error) {
	startTime := time.Now()

	url := fmt.Sprintf("%s/api/quote/%s", c.baseURL, symbol)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.ConnectionFailed(err.Error())
	}

	req.SetBasicAuth(c.login, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		latency := time.Since(startTime).Milliseconds()
		c.logger.ErrorWithLatency("Quote request failed", err, latency)
		return nil, errors.ConnectionFailed(err.Error())
	}
	defer resp.Body.Close()

	latency := time.Since(startTime).Milliseconds()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.InvalidSymbol(symbol)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var quote Quote
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		c.logger.ErrorWithLatency("Failed to decode quote", err, latency)
		return nil, err
	}

	c.logger.InfoWithLatency(fmt.Sprintf("Retrieved quote: %s", symbol), latency)
	return &quote, nil
}

// PlaceOrder places a new order
func (c *Client) PlaceOrder(symbol string, side string, volume decimal.Decimal, price decimal.Decimal, sl, tp decimal.Decimal, comment string) (*Order, error) {
	startTime := time.Now()

	orderReq := map[string]interface{}{
		"symbol":      symbol,
		"side":        side,
		"volume":      volume.String(),
		"price":       price.String(),
		"stop_loss":   sl.String(),
		"take_profit": tp.String(),
		"comment":     comment,
	}

	body, _ := json.Marshal(orderReq)
	url := fmt.Sprintf("%s/api/order", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, errors.ConnectionFailed(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.login, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		latency := time.Since(startTime).Milliseconds()
		c.logger.ErrorWithLatency("PlaceOrder request failed", err, latency)
		return nil, errors.ConnectionFailed(err.Error())
	}
	defer resp.Body.Close()

	latency := time.Since(startTime).Milliseconds()

	if resp.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.NewMT5Error(errors.ErrInvalidVolume, "Invalid order parameters", string(body))
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var order Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		c.logger.ErrorWithLatency("Failed to decode order response", err, latency)
		return nil, err
	}

	c.logger.InfoWithLatency(fmt.Sprintf("Placed order: %d (%s %s)", order.Ticket, side, symbol), latency)
	return &order, nil
}

// ClosePosition closes an open position
func (c *Client) ClosePosition(ticket uint64, volume decimal.Decimal) (*Order, error) {
	startTime := time.Now()

	closeReq := map[string]interface{}{
		"ticket": ticket,
		"volume": volume.String(),
	}

	body, _ := json.Marshal(closeReq)
	url := fmt.Sprintf("%s/api/order/%d/close", c.baseURL, ticket)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, errors.ConnectionFailed(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.login, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		latency := time.Since(startTime).Milliseconds()
		c.logger.ErrorWithLatency("ClosePosition request failed", err, latency)
		return nil, errors.ConnectionFailed(err.Error())
	}
	defer resp.Body.Close()

	latency := time.Since(startTime).Milliseconds()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NewMT5Error(errors.ErrPositionNotFound, "Position not found", fmt.Sprintf("ticket: %d", ticket))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var order Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		c.logger.ErrorWithLatency("Failed to decode close response", err, latency)
		return nil, err
	}

	c.logger.InfoWithLatency(fmt.Sprintf("Closed position: %d", ticket), latency)
	return &order, nil
}

// ListOrders lists all orders for the account
func (c *Client) ListOrders(filter string) ([]*Order, error) {
	startTime := time.Now()

	url := fmt.Sprintf("%s/api/orders?filter=%s", c.baseURL, filter)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.ConnectionFailed(err.Error())
	}

	req.SetBasicAuth(c.login, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		latency := time.Since(startTime).Milliseconds()
		c.logger.ErrorWithLatency("ListOrders request failed", err, latency)
		return nil, errors.ConnectionFailed(err.Error())
	}
	defer resp.Body.Close()

	latency := time.Since(startTime).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var orders []*Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		c.logger.ErrorWithLatency("Failed to decode orders", err, latency)
		return nil, err
	}

	c.logger.InfoWithLatency(fmt.Sprintf("Retrieved %d orders", len(orders)), latency)
	return orders, nil
}
