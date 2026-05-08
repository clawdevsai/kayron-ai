package daemon

import (
	"context"
	"testing"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

func TestModifyOrder(t *testing.T) {
	// Create mock client and service for testing
	client := mt5.NewClient("http://localhost:8228", "", "", 30)
	service := mt5.NewModifyOrderService(client)
	handler := NewModifyOrderServiceHandler(service)

	tests := []struct {
		name    string
		req     *api.ModifyOrderRequest
		wantErr bool
	}{
		{
			name: "valid modify with price",
			req: &api.ModifyOrderRequest{
				Ticket: 12345,
				Price:  "1.1050",
			},
			wantErr: false,
		},
		{
			name: "valid modify with stop loss",
			req: &api.ModifyOrderRequest{
				Ticket:   12345,
				StopLoss: "1.1000",
			},
			wantErr: false,
		},
		{
			name: "valid modify with take profit",
			req: &api.ModifyOrderRequest{
				Ticket:     12345,
				TakeProfit: "1.1100",
			},
			wantErr: false,
		},
		{
			name: "invalid price",
			req: &api.ModifyOrderRequest{
				Ticket: 12345,
				Price:  "invalid",
			},
			wantErr: false, // returns error in response, not err return
		},
		{
			name: "invalid stop loss",
			req: &api.ModifyOrderRequest{
				Ticket:   12345,
				StopLoss: "not a number",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.ModifyOrder(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ModifyOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if resp == nil {
				t.Fatal("ModifyOrder() returned nil response")
			}

			if resp.Ticket != tt.req.Ticket {
				t.Errorf("ModifyOrder() ticket = %d, want %d", resp.Ticket, tt.req.Ticket)
			}

			if resp.Status == "" {
				t.Error("ModifyOrder() status is empty")
			}
		})
	}
}
