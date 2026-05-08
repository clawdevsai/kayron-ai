package daemon

import (
	"context"
	"testing"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

func TestGetPendingOrderDetails(t *testing.T) {
	client := mt5.NewClient("http://localhost:8228", "", "", 30)
	service := mt5.NewPendingOrderService(client)
	handler := NewPendingOrderServiceHandler(service)

	tests := []struct {
		name    string
		req     *api.PendingOrderDetailsRequest
		wantErr bool
	}{
		{
			name: "all pending orders",
			req: &api.PendingOrderDetailsRequest{
				Symbol:       "",
				Status:       "PENDING",
				CreatedAfter: 0,
			},
			wantErr: false,
		},
		{
			name: "pending orders by symbol",
			req: &api.PendingOrderDetailsRequest{
				Symbol:       "EURUSD",
				Status:       "PENDING",
				CreatedAfter: 0,
			},
			wantErr: false,
		},
		{
			name: "orders since timestamp",
			req: &api.PendingOrderDetailsRequest{
				Symbol:       "",
				Status:       "",
				CreatedAfter: 1700000000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.GetPendingOrderDetails(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPendingOrderDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if resp == nil {
				t.Fatal("GetPendingOrderDetails() returned nil response")
			}

			if resp.Orders == nil {
				t.Error("GetPendingOrderDetails() orders is nil")
			}
		})
	}
}
