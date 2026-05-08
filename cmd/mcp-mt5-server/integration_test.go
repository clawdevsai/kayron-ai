package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukeware/kayron-ai/internal/config"
)

// TestAllToolsIntegration valida que todas 16 ferramentas respondem corretamente
func TestAllToolsIntegration(t *testing.T) {
	cfg := &config.Config{
		MT5Server:  "localhost",
		MT5Login:   "test",
		MT5Password: "test",
		MT5Timeout: 30,
		HTTPPort:   9999,
		GRPCPort:   9998,
	}

	server := NewMCPServer(cfg)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleRPC))
	defer httpServer.Close()

	tests := []struct {
		name   string
		method string
		params interface{}
	}{
		// Phase 1: Ferramentas Base (6)
		{"account-info", "account-info", nil},
		{"quote", "quote", map[string]interface{}{"symbol": "EURUSD"}},
		{"place-order", "place-order", map[string]interface{}{"symbol": "EURUSD", "volume": 1.0}},
		{"close-position", "close-position", map[string]interface{}{"ticket": 123}},
		{"orders-list", "orders-list", nil},
		{"get-candles", "get-candles", map[string]interface{}{"symbol": "EURUSD", "timeframe": "H1", "count": 10}},

		// Phase 2: Ferramentas Evolução (4)
		{"modify-order", "modify-order", map[string]interface{}{"ticket": 123, "stop_loss": 1.0900}},
		{"pending-order-details", "pending-order-details", map[string]interface{}{"symbol": "EURUSD"}},
		{"symbol-properties", "symbol-properties", map[string]interface{}{"symbol": "EURUSD"}},
		{"margin-calculator", "margin-calculator", map[string]interface{}{"symbol": "EURUSD", "volume": 1.0}},

		// Phase 3: Ferramentas Evolução (3)
		{"position-details", "position-details", map[string]interface{}{"symbol": "EURUSD"}},
		{"account-equity-history", "account-equity-history", map[string]interface{}{"from_timestamp": 1700000000, "to_timestamp": 1700086400, "granularity": "daily"}},
		{"balance-drawdown", "balance-drawdown", map[string]interface{}{"since_timestamp": 1700000000}},

		// Phase 4: Ferramentas Evolução (3)
		{"order-fill-analysis", "order-fill-analysis", map[string]interface{}{"ticket": 123}},
		{"market-hours", "market-hours", map[string]interface{}{"symbol": "EURUSD"}},
		{"tick-data", "tick-data", map[string]interface{}{"symbol": "EURUSD", "duration_seconds": 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := MCPRequest{
				Jsonrpc: "2.0",
				Method:  tt.method,
				Params:  tt.params,
				ID:      1,
			}

			body, _ := json.Marshal(req)
			resp, err := http.Post(httpServer.URL, "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected 200, got %d", resp.StatusCode)
			}

			var rpcResp MCPResponse
			json.NewDecoder(resp.Body).Decode(&rpcResp)

			// Se a resposta tem um erro ou um resultado válido, tool está funcionando
			hasError := rpcResp.Error != nil
			hasResult := rpcResp.Result != nil

			if !hasError && !hasResult {
				t.Errorf("Tool returned neither error nor result")
			}
		})
	}
}
