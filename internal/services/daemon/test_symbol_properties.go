package daemon

import (
	"context"
	"testing"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

func TestGetSymbolProperties(t *testing.T) {
	client := mt5.NewClient("http://localhost:8228", "", "", 30)
	service := mt5.NewSymbolPropertiesService(client)
	handler := NewSymbolPropertiesServiceHandler(service)

	tests := []struct {
		name    string
		req     *api.SymbolPropertiesRequest
		wantErr bool
	}{
		{
			name: "EURUSD properties",
			req: &api.SymbolPropertiesRequest{
				Symbol: "EURUSD",
			},
			wantErr: false,
		},
		{
			name: "GBPUSD properties",
			req: &api.SymbolPropertiesRequest{
				Symbol: "GBPUSD",
			},
			wantErr: false,
		},
		{
			name: "USDJPY properties",
			req: &api.SymbolPropertiesRequest{
				Symbol: "USDJPY",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.GetSymbolProperties(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSymbolProperties() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if resp == nil {
				t.Fatal("GetSymbolProperties() returned nil response")
			}

			if resp.Symbol != tt.req.Symbol {
				t.Errorf("GetSymbolProperties() symbol = %s, want %s", resp.Symbol, tt.req.Symbol)
			}

			if resp.Digits == 0 && tt.req.Symbol != "" {
				t.Error("GetSymbolProperties() digits should not be 0")
			}
		})
	}
}
