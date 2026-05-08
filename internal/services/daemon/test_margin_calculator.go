package daemon

import (
	"context"
	"testing"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

func TestCalculateMarginRequirement(t *testing.T) {
	client := mt5.NewClient("http://localhost:8228", "", "", 30)
	service := mt5.NewMarginCalculatorService(client)
	handler := NewMarginCalculatorServiceHandler(service)

	tests := []struct {
		name    string
		req     *api.MarginCalculatorRequest
		wantErr bool
	}{
		{
			name: "EURUSD 0.1 volume margin",
			req: &api.MarginCalculatorRequest{
				Symbol: "EURUSD",
				Volume: "0.1",
			},
			wantErr: false,
		},
		{
			name: "GBPUSD 0.5 volume margin",
			req: &api.MarginCalculatorRequest{
				Symbol: "GBPUSD",
				Volume: "0.5",
			},
			wantErr: false,
		},
		{
			name: "USDJPY 1.0 volume margin",
			req: &api.MarginCalculatorRequest{
				Symbol: "USDJPY",
				Volume: "1.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.CalculateMarginRequirement(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateMarginRequirement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if resp == nil {
				t.Fatal("CalculateMarginRequirement() returned nil response")
			}

			if resp.MarginRequired == "" {
				t.Error("CalculateMarginRequirement() margin required should not be empty")
			}
		})
	}
}
