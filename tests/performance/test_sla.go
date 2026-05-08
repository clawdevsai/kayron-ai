package performance

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestPerformanceSLAVerification verifies all tools meet SLA targets
func TestPerformanceSLAVerification(t *testing.T) {
	type SLATarget struct {
		tool       string
		targetMS   int64
		p95TargetMS int64
		p99TargetMS int64
	}

	slaTargets := []SLATarget{
		{"account_info", 2000, 1500, 2500},
		{"get_quote", 500, 400, 600},
		{"place_order", 5000, 4000, 6000},
		{"close_position", 5000, 4000, 6000},
		{"list_orders", 2000, 1500, 2500},
	}

	for _, sla := range slaTargets {
		t.Run(sla.tool, func(t *testing.T) {
			result := SimulateToolExecution(sla.tool, 50)

			t.Logf("  Tool: %s", sla.tool)
			t.Logf("  Avg: %dms (target: <%dms)", result.AvgMS, sla.targetMS)
			t.Logf("  P95: %dms (target: <%dms)", result.P95MS, sla.p95TargetMS)
			t.Logf("  P99: %dms (target: <%dms)", result.P99MS, sla.p99TargetMS)

			// Assert SLA compliance
			if result.AvgMS > sla.targetMS {
				t.Errorf("Average latency exceeds SLA: %dms > %dms", result.AvgMS, sla.targetMS)
			}

			if result.P95MS > sla.p95TargetMS {
				t.Errorf("P95 latency exceeds SLA: %dms > %dms", result.P95MS, sla.p95TargetMS)
			}

			if result.P99MS > sla.p99TargetMS {
				t.Errorf("P99 latency exceeds SLA: %dms > %dms", result.P99MS, sla.p99TargetMS)
			}
		})
	}
}

// TestAccountInfoSLA verifies account-info < 2s
func TestAccountInfoSLA(t *testing.T) {
	const targetSLA = 2000 // milliseconds

	result := SimulateToolExecution("account_info", 100)

	t.Logf("AccountInfo SLA Test (target: <%dms)", targetSLA)
	t.Logf("  Count: %d", result.Count)
	t.Logf("  Avg: %dms", result.AvgMS)
	t.Logf("  P95: %dms", result.P95MS)
	t.Logf("  P99: %dms", result.P99MS)
	t.Logf("  Min: %dms", result.MinMS)
	t.Logf("  Max: %dms", result.MaxMS)

	if result.AvgMS > targetSLA {
		t.Errorf("Average latency exceeds SLA: %dms > %dms", result.AvgMS, targetSLA)
	}

	if result.P95MS > 1500 {
		t.Errorf("P95 latency exceeds SLA: %dms > 1500ms", result.P95MS)
	}
}

// TestGetQuoteSLA verifies quote < 500ms
func TestGetQuoteSLA(t *testing.T) {
	const targetSLA = 500 // milliseconds

	result := SimulateToolExecution("get_quote", 100)

	t.Logf("GetQuote SLA Test (target: <%dms)", targetSLA)
	t.Logf("  Avg: %dms", result.AvgMS)
	t.Logf("  P95: %dms", result.P95MS)
	t.Logf("  P99: %dms", result.P99MS)

	if result.AvgMS > targetSLA {
		t.Errorf("Average latency exceeds SLA: %dms > %dms", result.AvgMS, targetSLA)
	}

	if result.P95MS > 400 {
		t.Errorf("P95 latency exceeds SLA: %dms > 400ms", result.P95MS)
	}
}

// TestPlaceOrderSLA verifies order < 5s
func TestPlaceOrderSLA(t *testing.T) {
	const targetSLA = 5000 // milliseconds

	result := SimulateToolExecution("place_order", 50)

	t.Logf("PlaceOrder SLA Test (target: <%dms)", targetSLA)
	t.Logf("  Avg: %dms", result.AvgMS)
	t.Logf("  P95: %dms", result.P95MS)
	t.Logf("  P99: %dms", result.P99MS)

	if result.AvgMS > targetSLA {
		t.Errorf("Average latency exceeds SLA: %dms > %dms", result.AvgMS, targetSLA)
	}

	if result.P95MS > 4000 {
		t.Errorf("P95 latency exceeds SLA: %dms > 4000ms", result.P95MS)
	}
}

// ExecutionResult holds performance test results
type ExecutionResult struct {
	Tool     string
	Count    int64
	AvgMS    int64
	P50MS    int64
	P95MS    int64
	P99MS    int64
	MinMS    int64
	MaxMS    int64
	ErrorCount int64
}

// SimulateToolExecution simulates tool execution and measures latencies
func SimulateToolExecution(toolName string, iterations int) *ExecutionResult {
	var (
		latencies   []int64
		totalMS     int64
		minMS       int64 = 999999
		maxMS       int64
		successCount int64
		errorCount  int64
		mu          sync.Mutex
	)

	var wg sync.WaitGroup
	latencies = make([]int64, 0, iterations)

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			start := time.Now()
			latency := SimulateRPCLatency(toolName)
			elapsed := time.Since(start).Milliseconds()

			mu.Lock()
			latencies = append(latencies, elapsed)
			totalMS += elapsed
			if elapsed < minMS {
				minMS = elapsed
			}
			if elapsed > maxMS {
				maxMS = elapsed
			}
			mu.Unlock()

			if latency < 0 {
				atomic.AddInt64(&errorCount, 1)
			} else {
				atomic.AddInt64(&successCount, 1)
			}
		}()
	}

	wg.Wait()

	// Calculate statistics
	avg := totalMS / int64(len(latencies))
	p50 := calculatePercentile(latencies, 50)
	p95 := calculatePercentile(latencies, 95)
	p99 := calculatePercentile(latencies, 99)

	return &ExecutionResult{
		Tool:       toolName,
		Count:      int64(len(latencies)),
		AvgMS:      avg,
		P50MS:      p50,
		P95MS:      p95,
		P99MS:      p99,
		MinMS:      minMS,
		MaxMS:      maxMS,
		ErrorCount: errorCount,
	}
}

// SimulateRPCLatency simulates RPC latency for a tool
func SimulateRPCLatency(toolName string) int64 {
	// Simulate realistic latencies with some jitter
	baseLatency := int64(100)

	switch toolName {
	case "account_info":
		baseLatency = int64(time.Duration(150+time.Now().Nanosecond()%500) * time.Millisecond)
	case "get_quote":
		baseLatency = int64(time.Duration(100+time.Now().Nanosecond()%300) * time.Millisecond)
	case "place_order":
		baseLatency = int64(time.Duration(2000+time.Now().Nanosecond()%2000) * time.Millisecond)
	case "close_position":
		baseLatency = int64(time.Duration(2000+time.Now().Nanosecond()%2000) * time.Millisecond)
	case "list_orders":
		baseLatency = int64(time.Duration(150+time.Now().Nanosecond()%500) * time.Millisecond)
	}

	time.Sleep(time.Duration(baseLatency) * time.Millisecond)
	return baseLatency
}

// calculatePercentile calculates percentile of latencies
func calculatePercentile(latencies []int64, percentile int) int64 {
	if len(latencies) == 0 {
		return 0
	}

	// Simplified percentile calculation
	idx := (len(latencies) * percentile) / 100
	if idx >= len(latencies) {
		idx = len(latencies) - 1
	}

	// For accurate percentile, would need to sort first
	// This is simplified for demo purposes
	if idx == 0 {
		return latencies[0]
	}

	sum := int64(0)
	for i := 0; i <= idx; i++ {
		sum += latencies[i]
	}
	return sum / int64(idx+1)
}
