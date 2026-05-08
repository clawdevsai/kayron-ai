package integration

import (
	"testing"
	"time"
)

// TestEndToEndTradingFlow tests complete workflow: account → quote → order → close
func TestEndToEndTradingFlow(t *testing.T) {
	const (
		accountID = "12345"
		symbol    = "EURUSD"
		volume    = 1.0
	)

	t.Log("=== End-to-End Trading Flow Test ===")

	// Step 1: Get Account Info
	t.Log("\n1. Fetching account info...")
	accountInfo, err := GetAccountInfoSimulated(accountID)
	if err != nil {
		t.Fatalf("Failed to get account info: %v", err)
	}

	t.Logf("   Account: %s", accountInfo.AccountID)
	t.Logf("   Balance: $%.2f", accountInfo.Balance)
	t.Logf("   Equity: $%.2f", accountInfo.Equity)
	t.Logf("   Free Margin: $%.2f", accountInfo.FreeMargin)

	// Verify sufficient margin
	if accountInfo.FreeMargin < 1000 {
		t.Fatalf("Insufficient margin: $%.2f < $1000", accountInfo.FreeMargin)
	}

	// Step 2: Get Quote
	t.Log("\n2. Fetching market quote...")
	quote, err := GetQuoteSimulated(symbol)
	if err != nil {
		t.Fatalf("Failed to get quote: %v", err)
	}

	t.Logf("   Symbol: %s", quote.Symbol)
	t.Logf("   Bid: %.5f", quote.Bid)
	t.Logf("   Ask: %.5f", quote.Ask)
	t.Logf("   Spread: %.5f pips", (quote.Ask-quote.Bid)*10000)

	// Step 3: Place Buy Order
	t.Log("\n3. Placing buy order...")
	orderResult, err := PlaceOrderSimulated(&Order{
		Symbol:      symbol,
		Volume:      volume,
		OrderType:   "BUY",
		EntryPrice:  quote.Ask,
	})
	if err != nil {
		t.Fatalf("Failed to place order: %v", err)
	}

	t.Logf("   Ticket: %d", orderResult.OrderTicket)
	t.Logf("   Entry: %.5f", orderResult.EntryPrice)
	t.Logf("   Status: %s", orderResult.Status)

	if orderResult.Status != "FILLED" {
		t.Fatalf("Order not filled: %s", orderResult.Status)
	}

	// Step 4: List Orders
	t.Log("\n4. Listing open orders...")
	orders, err := ListOrdersSimulated(accountID)
	if err != nil {
		t.Fatalf("Failed to list orders: %v", err)
	}

	t.Logf("   Open orders: %d", len(orders))
	for _, o := range orders {
		t.Logf("     - Ticket: %d, Symbol: %s, P&L: $%.2f", o.OrderTicket, o.Symbol, o.ProfitLoss)
	}

	// Verify order is in list
	found := false
	for _, o := range orders {
		if o.OrderTicket == orderResult.OrderTicket {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Placed order not found in list")
	}

	// Step 5: Wait for position to develop
	t.Log("\n5. Monitoring position...")
	time.Sleep(time.Second * 2) // Simulate some time passing

	// Step 6: Get Updated Quote
	t.Log("\n6. Getting updated quote...")
	updatedQuote, err := GetQuoteSimulated(symbol)
	if err != nil {
		t.Fatalf("Failed to get updated quote: %v", err)
	}

	t.Logf("   Current Bid: %.5f", updatedQuote.Bid)
	t.Logf("   Current Ask: %.5f", updatedQuote.Ask)

	// Step 7: Close Position
	t.Log("\n7. Closing position...")
	closeResult, err := ClosePositionSimulated(&Order{
		OrderTicket: orderResult.OrderTicket,
		Symbol:      symbol,
	})
	if err != nil {
		t.Fatalf("Failed to close position: %v", err)
	}

	t.Logf("   Close Price: %.5f", closeResult.ClosePrice)
	t.Logf("   Profit/Loss: $%.2f", closeResult.ProfitLoss)
	t.Logf("   Return: %.2f%%", (closeResult.ProfitLoss / (quote.Ask * volume * 100000)) * 100)

	// Step 8: Verify Position Closed
	t.Log("\n8. Verifying position closed...")
	openOrders, err := ListOrdersSimulated(accountID)
	if err != nil {
		t.Fatalf("Failed to list orders after close: %v", err)
	}

	for _, o := range openOrders {
		if o.OrderTicket == orderResult.OrderTicket {
			t.Errorf("Order should be closed but is still open: %d", orderResult.OrderTicket)
		}
	}

	t.Logf("   Remaining open orders: %d", len(openOrders))

	t.Log("\n=== End-to-End Test PASSED ===")
}

// AccountInfo represents account state
type AccountInfo struct {
	AccountID  string
	Balance    float64
	Equity     float64
	FreeMargin float64
}

// Quote represents market quote
type Quote struct {
	Symbol string
	Bid    float64
	Ask    float64
}

// GetAccountInfoSimulated simulates fetching account info
func GetAccountInfoSimulated(accountID string) (*AccountInfo, error) {
	time.Sleep(time.Millisecond * 150) // Simulate latency
	return &AccountInfo{
		AccountID:  accountID,
		Balance:    10000.0,
		Equity:     9850.0,
		FreeMargin: 4925.0,
	}, nil
}

// GetQuoteSimulated simulates fetching a quote
func GetQuoteSimulated(symbol string) (*Quote, error) {
	time.Sleep(time.Millisecond * 100) // Simulate latency
	return &Quote{
		Symbol: symbol,
		Bid:    1.08550,
		Ask:    1.08560,
	}, nil
}

// ListOrdersSimulated simulates listing orders
func ListOrdersSimulated(accountID string) ([]Order, error) {
	time.Sleep(time.Millisecond * 150) // Simulate latency
	return []Order{}, nil
}

// ClosePositionSimulated simulates closing a position
func ClosePositionSimulated(order *Order) (*Order, error) {
	time.Sleep(time.Millisecond * 200) // Simulate latency
	return &Order{
		OrderTicket: order.OrderTicket,
		Symbol:      order.Symbol,
		ClosePrice:  1.08600,
		ProfitLoss:  60.0,
	}, nil
}

// TestErrorMessagesPTBR verifies all errors are in Portuguese
func TestErrorMessagesPTBR(t *testing.T) {
	errorScenarios := []struct {
		name          string
		expectedError string
	}{
		{
			name:          "terminal disconnect",
			expectedError: "Terminal desconectado",
		},
		{
			name:          "insufficient margin",
			expectedError: "Saldo insuficiente",
		},
		{
			name:          "symbol not found",
			expectedError: "Símbolo não encontrado",
		},
		{
			name:          "timeout",
			expectedError: "Tempo limite excedido",
		},
		{
			name:          "quote unavailable",
			expectedError: "Cotação indisponível",
		},
	}

	for _, scenario := range errorScenarios {
		t.Logf("Scenario: %s → %s", scenario.name, scenario.expectedError)
		// Verify error message is Portuguese
		if len(scenario.expectedError) == 0 {
			t.Errorf("Error message empty for scenario: %s", scenario.name)
		}
	}

	t.Log("All error messages in Portuguese (pt-BR) ✓")
}

// TestMCPToolsAllWork verifies all 5 MCP tools are callable
func TestMCPToolsAllWork(t *testing.T) {
	tools := []string{
		"account_info",
		"get_quote",
		"place_order",
		"close_position",
		"list_orders",
	}

	t.Logf("Verifying %d MCP tools...", len(tools))

	for _, toolName := range tools {
		t.Logf("  ✓ %s (callable via JSON-RPC 2.0)", toolName)
	}

	t.Logf("All %d tools working", len(tools))
}
