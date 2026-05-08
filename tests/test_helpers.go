package tests

import (
	"fmt"
	"sync"
	"time"

	"github.com/lukeware/kayron-ai/internal/services/mt5"
	"github.com/shopspring/decimal"
)

// MockMT5Client provides a mock implementation of the MT5 client
type MockMT5Client struct {
	Accounts map[string]*mt5.AccountInfo
	Quotes   map[string]*mt5.Quote
	Orders   map[uint64]*mt5.Order
	mu       sync.RWMutex
	CallLog  []string
}

// NewMockMT5Client creates a new mock client
func NewMockMT5Client() *MockMT5Client {
	return &MockMT5Client{
		Accounts: make(map[string]*mt5.AccountInfo),
		Quotes:   make(map[string]*mt5.Quote),
		Orders:   make(map[uint64]*mt5.Order),
		CallLog:  []string{},
	}
}

// GetAccount returns mock account info
func (m *MockMT5Client) GetAccount() (*mt5.AccountInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallLog = append(m.CallLog, fmt.Sprintf("[%s] GetAccount", time.Now().Format(time.RFC3339)))

	if len(m.Accounts) == 0 {
		return nil, fmt.Errorf("account not found")
	}

	for _, account := range m.Accounts {
		return account, nil
	}
	return nil, nil
}

// GetQuote returns a mock quote
func (m *MockMT5Client) GetQuote(symbol string) (*mt5.Quote, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.CallLog = append(m.CallLog, fmt.Sprintf("[%s] GetQuote(%s)", time.Now().Format(time.RFC3339), symbol))

	if quote, exists := m.Quotes[symbol]; exists {
		return quote, nil
	}

	return nil, fmt.Errorf("symbol not found: %s", symbol)
}

// PlaceOrder creates a mock order
func (m *MockMT5Client) PlaceOrder(symbol, side string, volume, price, sl, tp decimal.Decimal, comment string) (*mt5.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallLog = append(m.CallLog, fmt.Sprintf("[%s] PlaceOrder(%s %s %s @ %s)", time.Now().Format(time.RFC3339), side, volume, symbol, price))

	ticket := uint64(len(m.Orders) + 1)
	order := &mt5.Order{
		Ticket:     ticket,
		Symbol:     symbol,
		Side:       side,
		Volume:     volume,
		OpenPrice:  price,
		OpenTime:   time.Now().Unix(),
		StopLoss:   sl,
		TakeProfit: tp,
		Status:     "OPEN",
		Comment:    comment,
	}

	m.Orders[ticket] = order
	return order, nil
}

// ClosePosition closes a mock position
func (m *MockMT5Client) ClosePosition(ticket uint64, volume decimal.Decimal) (*mt5.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallLog = append(m.CallLog, fmt.Sprintf("[%s] ClosePosition(ticket=%d, volume=%s)", time.Now().Format(time.RFC3339), ticket, volume))

	order, exists := m.Orders[ticket]
	if !exists {
		return nil, fmt.Errorf("position not found: %d", ticket)
	}

	order.Status = "CLOSED"
	return order, nil
}

// ListOrders returns mock orders
func (m *MockMT5Client) ListOrders(filter string) ([]*mt5.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.CallLog = append(m.CallLog, fmt.Sprintf("[%s] ListOrders(filter=%s)", time.Now().Format(time.RFC3339), filter))

	var result []*mt5.Order
	for _, order := range m.Orders {
		if filter == "ALL" || order.Status == filter {
			result = append(result, order)
		}
	}

	return result, nil
}

// Test fixtures
var (
	// DummyAccount is a test account fixture
	DummyAccount = &mt5.AccountInfo{
		Login:       123456,
		Balance:     decimal.NewFromInt(10000),
		Equity:      decimal.NewFromInt(9500),
		Margin:      decimal.NewFromInt(500),
		FreeMargin:  decimal.NewFromInt(9500),
		MarginLevel: decimal.NewFromInt(1900),
		Currency:    "USD",
	}

	// DummySymbols is a list of test symbols
	DummySymbols = []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD"}

	// DummyQuotes is test quote data
	DummyQuotes = map[string]*mt5.Quote{
		"EURUSD": {
			Symbol: "EURUSD",
			Bid:    decimal.NewFromString("1.0950"),
			Ask:    decimal.NewFromString("1.0952"),
			Time:   time.Now().Unix(),
		},
		"GBPUSD": {
			Symbol: "GBPUSD",
			Bid:    decimal.NewFromString("1.2750"),
			Ask:    decimal.NewFromString("1.2752"),
			Time:   time.Now().Unix(),
		},
	}
)

// SetupMockClient initializes a mock client with test data
func SetupMockClient() *MockMT5Client {
	client := NewMockMT5Client()
	client.Accounts["123456"] = DummyAccount

	for symbol, quote := range DummyQuotes {
		client.Quotes[symbol] = quote
	}

	return client
}

// Package-level helper for integration tests
var mockMT5ClientInstance *mt5.Client

// setupMockMT5Client returns a configured mock MT5 client for testing
func setupMockMT5Client() *mt5.Client {
	// For testing purposes, return a client pointing to a mock endpoint
	// In real tests, this would be mocked or use dependency injection
	return mt5.NewClient("http://localhost:8228", "test_login", "test_password", 30)
}
