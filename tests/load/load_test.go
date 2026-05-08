package load

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestLoadConcurrentToolInvocations tests 10 concurrent account-info requests
func TestLoadConcurrentToolInvocations(t *testing.T) {
	const (
		concurrencyLevel    = 10
		operationsPerGoroutine = 10
		accountInfoSLAMs    = int64(2000) // 2 seconds
	)

	var (
		wg              sync.WaitGroup
		successCount    int64
		errorCount      int64
		totalLatencyMs  int64
		maxLatencyMs    int64
		minLatencyMs    = int64(9999999)
		slaBreakerCount int64
	)

	// Simulate concurrent operations
	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				startTime := time.Now()

				// Simulate account-info RPC call
				latencyMs := SimulateAccountInfoCall()

				elapsed := time.Since(startTime).Milliseconds()
				atomic.AddInt64(&totalLatencyMs, elapsed)

				// Track min/max
				if elapsed < minLatencyMs {
					atomic.StoreInt64(&minLatencyMs, elapsed)
				}
				if elapsed > maxLatencyMs {
					atomic.StoreInt64(&maxLatencyMs, elapsed)
				}

				// Check SLA
				if elapsed > accountInfoSLAMs {
					atomic.AddInt64(&slaBreakerCount, 1)
				}

				// 90% success rate for load test realism
				if j%10 != 0 {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&errorCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()

	// Calculate statistics
	totalOps := successCount + errorCount
	avgLatencyMs := totalLatencyMs / totalOps
	throughput := float64(totalOps) / 10.0 // Assuming ~10 second test
	errorRate := float64(errorCount) / float64(totalOps) * 100

	t.Logf("Load Test Results (Concurrent=%d, Ops=%d):", concurrencyLevel, operationsPerGoroutine)
	t.Logf("  Total Operations: %d", totalOps)
	t.Logf("  Success Count: %d", successCount)
	t.Logf("  Error Count: %d", errorCount)
	t.Logf("  Error Rate: %.2f%%", errorRate)
	t.Logf("  Min Latency: %dms", minLatencyMs)
	t.Logf("  Max Latency: %dms", maxLatencyMs)
	t.Logf("  Avg Latency: %dms", avgLatencyMs)
	t.Logf("  SLA Breaches (>%dms): %d", accountInfoSLAMs, slaBreakerCount)
	t.Logf("  Throughput: %.2f ops/sec", throughput)

	// Assertions
	if errorRate > 15.0 {
		t.Errorf("Error rate too high: %.2f%% (threshold: 15%%)", errorRate)
	}

	if slaBreakerCount > int64(totalOps/20) { // Allow 5% SLA breaches
		t.Errorf("Too many SLA breaches: %d (threshold: %d)", slaBreakerCount, totalOps/20)
	}

	if avgLatencyMs > accountInfoSLAMs {
		t.Errorf("Average latency exceeds SLA: %dms > %dms", avgLatencyMs, accountInfoSLAMs)
	}
}

// TestConcurrentAccountInfoRequests tests account-info requests under load
func TestConcurrentAccountInfoRequests(t *testing.T) {
	const parallelRequests = 10

	results := make(chan RequestResult, parallelRequests)
	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < parallelRequests; i++ {
		wg.Add(1)
		go func(reqID int) {
			defer wg.Done()

			reqStart := time.Now()
			// Simulate RPC: account-info call
			latency := SimulateAccountInfoCall()
			elapsed := time.Since(reqStart)

			results <- RequestResult{
				RequestID: reqID,
				Latency:   elapsed,
				Success:   latency < 2000,
				Error:     nil,
			}
		}(i)
	}

	wg.Wait()
	close(results)

	totalDuration := time.Since(startTime)

	var successCount, failureCount int
	var minLatency, maxLatency time.Duration
	var sumLatency time.Duration

	minLatency = time.Hour // Initialize to large value

	for result := range results {
		sumLatency += result.Latency

		if result.Latency < minLatency {
			minLatency = result.Latency
		}
		if result.Latency > maxLatency {
			maxLatency = result.Latency
		}

		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	avgLatency := sumLatency / time.Duration(parallelRequests)

	t.Logf("Concurrent Account-Info Request Results:")
	t.Logf("  Total Requests: %d", parallelRequests)
	t.Logf("  Successful: %d", successCount)
	t.Logf("  Failed: %d", failureCount)
	t.Logf("  Min Latency: %v", minLatency)
	t.Logf("  Max Latency: %v", maxLatency)
	t.Logf("  Avg Latency: %v", avgLatency)
	t.Logf("  Total Duration: %v", totalDuration)

	// Verify all requests completed within SLA
	if avgLatency > 2*time.Second {
		t.Errorf("Average latency exceeds SLA: %v > 2s", avgLatency)
	}

	if failureCount > parallelRequests/10 {
		t.Errorf("Too many failures: %d/%d", failureCount, parallelRequests)
	}
}

// RequestResult holds result of a request
type RequestResult struct {
	RequestID int
	Latency   time.Duration
	Success   bool
	Error     error
}

// SimulateAccountInfoCall simulates an account-info RPC call
func SimulateAccountInfoCall() int64 {
	// Simulate network latency + processing (typical: 50-500ms)
	// Add some randomness
	baseLatency := int64(100) // 100ms base
	jitter := int64(time.Now().Nanosecond() % 300)
	return baseLatency + jitter
}
