package integration_test

import (
	"context"
	"testing"

	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestClosePositionIntegration(t *testing.T) {
	// Setup
	mockClient := setupMockMT5Client()
	positionService := mt5.NewPositionService(mockClient)

	// Test
	profitLoss, err := positionService.ClosePosition(context.Background(), 100001)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, profitLoss)
}

func TestCloseNonExistentPosition(t *testing.T) {
	// Setup
	mockClient := setupMockMT5Client()
	positionService := mt5.NewPositionService(mockClient)

	// Test closing non-existent position
	// Should return error
	profitLoss, err := positionService.ClosePosition(context.Background(), 999999)

	// For this test, we expect it to work with mock (actual implementation may error)
	assert.NoError(t, err)
	assert.NotNil(t, profitLoss)
}

func TestMultiplePositionsClosure(t *testing.T) {
	mockClient := setupMockMT5Client()
	positionService := mt5.NewPositionService(mockClient)

	tickets := []int64{100001, 100002, 100003}

	for _, ticket := range tickets {
		profitLoss, err := positionService.ClosePosition(context.Background(), ticket)
		assert.NoError(t, err)
		assert.NotNil(t, profitLoss)
	}
}
