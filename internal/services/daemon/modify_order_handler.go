package daemon

import (
	"context"

	"github.com/lukeware/kayron-ai/api"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
	"github.com/shopspring/decimal"
)

// ModifyOrderServiceHandler wraps modify order operations
type ModifyOrderServiceHandler struct {
	service *mt5.ModifyOrderService
	logger  *logger.Logger
}

// NewModifyOrderServiceHandler creates a new modify order handler
func NewModifyOrderServiceHandler(service *mt5.ModifyOrderService) *ModifyOrderServiceHandler {
	return &ModifyOrderServiceHandler{
		service: service,
		logger:  logger.New("ModifyOrderHandler"),
	}
}

// ModifyOrder handles modify order requests
func (h *ModifyOrderServiceHandler) ModifyOrder(ctx context.Context, req *api.ModifyOrderRequest) (*api.ModifyOrderResponse, error) {
	h.logger.Info("ModifyOrder request handling")

	// Parse decimal values
	var price, stopLoss, takeProfit decimal.Decimal
	var err error

	if req.Price != "" {
		price, err = decimal.NewFromString(req.Price)
		if err != nil {
			h.logger.Error("Invalid price", err)
			return &api.ModifyOrderResponse{
				Ticket:   req.Ticket,
				Status:   "error",
				ErrorMsg: "Preço inválido",
			}, nil
		}
	}

	if req.StopLoss != "" {
		stopLoss, err = decimal.NewFromString(req.StopLoss)
		if err != nil {
			h.logger.Error("Invalid stop loss", err)
			return &api.ModifyOrderResponse{
				Ticket:   req.Ticket,
				Status:   "error",
				ErrorMsg: "StopLoss inválido",
			}, nil
		}
	}

	if req.TakeProfit != "" {
		takeProfit, err = decimal.NewFromString(req.TakeProfit)
		if err != nil {
			h.logger.Error("Invalid take profit", err)
			return &api.ModifyOrderResponse{
				Ticket:   req.Ticket,
				Status:   "error",
				ErrorMsg: "TakeProfit inválido",
			}, nil
		}
	}

	order := &models.ModifyOrder{
		Ticket:     req.Ticket,
		Price:      price,
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
	}

	result, err := h.service.ModifyOrder(ctx, order)
	if err != nil {
		h.logger.Error("ModifyOrder failed", err)
		return &api.ModifyOrderResponse{
			Ticket:   req.Ticket,
			Status:   "error",
			ErrorMsg: "Falha ao modificar ordem",
		}, nil
	}

	return &api.ModifyOrderResponse{
		Ticket:   result.Ticket,
		Status:   result.Status,
		ErrorMsg: result.ErrorMsg,
	}, nil
}
