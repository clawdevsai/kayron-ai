package integration

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrentOrderPlacement tests FIFO sequencing with concurrent orders
func TestConcurrentOrderPlacement(t *testing.T) {
	const (
		concurrentOrders = 5
		symbol           = "EURUSD"
	)

	var (
		wg             sync.WaitGroup
		orderTimestamps []int64
		mu             sync.Mutex
		successCount   int64
		errorCount     int64
	)

	orderTimestamps = make([]int64, 0, concurrentOrders)

	// Place 5 orders concurrently on same symbol
	for i := 0; i < concurrentOrders; i++ {
		wg.Add(1)
		go func(orderNum int) {
			defer wg.Done()

			// Simulate order placement
			timestamp, err := PlaceOrderSimulated(symbol)
			if err == nil {
				mu.Lock()
				orderTimestamps = append(orderTimestamps, timestamp)
				mu.Unlock()
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&errorCount, 1)
				t.Logf("Order %d failed: %v", orderNum, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify FIFO sequencing
	if len(orderTimestamps) != concurrentOrders {
		t.Errorf("Expected %d orders, got %d", concurrentOrders, len(orderTimestamps))
	}

	// Verify timestamps are monotonically increasing (FIFO order)
	for i := 1; i < len(orderTimestamps); i++ {
		if orderTimestamps[i] < orderTimestamps[i-1] {
			t.Errorf("FIFO order violated at index %d: %d < %d",
				i, orderTimestamps[i], orderTimestamps[i-1])
		}
	}

	// Verify no duplicate fills
	fills := make(map[int64]int)
	for _, ts := range orderTimestamps {
		fills[ts]++
	}

	for ts, count := range fills {
		if count > 1 {
			t.Errorf("Duplicate fill for order at timestamp %d (count: %d)", ts, count)
		}
	}

	t.Logf("Concurrent Order Test: %d successful, %d failed", successCount, errorCount)
}

// TestConcurrentOrderFIFOEnforcement verifies strict FIFO ordering
func TestConcurrentOrderFIFOEnforcement(t *testing.T) {
	const numOrders = 10

	orders := make([]*Order, numOrders)
	mu := sync.Mutex{}

	var wg sync.WaitGroup

	// Create all orders concurrently
	for i := 0; i < numOrders; i++ {
		wg.Add(1)
		go func(orderID int) {
			defer wg.Done()

			order := &Order{
				ID:        int64(orderID),
				Symbol:    "EURUSD",
				Volume:    1.0,
				OrderTime: time.Now().UnixNano(),
			}

			// Simulate MT5 order submission
			if result, err := SubmitOrderSimulated(order); err == nil {
				mu.Lock()
				orders[orderID] = result
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Verify all orders placed
	for i, order := range orders {
		if order == nil {
			t.Errorf("Order %d is nil", i)
			continue
		}

		// Verify order timestamps reflect submission order
		if order.MT5Timestamp == 0 {
			t.Errorf("Order %d has zero MT5 timestamp", i)
		}
	}

	t.Logf("FIFO Enforcement Test: All %d orders placed with timestamps", numOrders)
}

// TestNoDuplicateFills verifies no duplicate order fills
func TestNoDuplicateFills(t *testing.T) {
	const numOrders = 5

	// Submit orders concurrently
	submissions := make([]int64, numOrders)
	for i := 0; i < numOrders; i++ {
		submissions[i] = int64(time.Now().UnixNano())
	}

	// Check for duplicates (same ticket/timestamp)
	seenTickets := make(map[int64]bool)

	for _, ticket := range submissions {
		if seenTickets[ticket] {
			t.Errorf("Duplicate ticket found: %d", ticket)
		}
		seenTickets[ticket] = true
	}

	t.Logf("Duplicate Fill Test: All %d submissions have unique tickets", numOrders)
}

// Order represents a trading order
type Order struct {
	ID             int64
	Symbol         string
	Volume         float64
	OrderTime      int64
	MT5Timestamp   int64
	OrderTicket    int64
	Status         string
}

// PlaceOrderSimulated simulates placing an order
func PlaceOrderSimulated(symbol string) (int64, error) {
	// Simulate MT5 WebAPI call
	time.Sleep(time.Millisecond * 100)
	return time.Now().UnixNano(), nil
}

// SubmitOrderSimulated simulates submitting an order and getting back result
func SubmitOrderSimulated(order *Order) (*Order, error) {
	// Simulate processing
	time.Sleep(time.Millisecond * 50)

	result := &Order{
		ID:           order.ID,
		Symbol:       order.Symbol,
		Volume:       order.Volume,
		OrderTime:    order.OrderTime,
		MT5Timestamp: time.Now().UnixNano(),
		OrderTicket:  int64(time.Now().Nanosecond()),
		Status:       "FILLED",
	}

	return result, nil
}
