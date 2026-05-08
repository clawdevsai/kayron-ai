package logger

import (
	"sync"
)

// LatencyTracker tracks latency metrics (p50, p95, p99) per tool
type LatencyTracker struct {
	mu      sync.RWMutex
	metrics map[string]*ToolLatencyMetrics
}

// ToolLatencyMetrics holds latency data for a single tool
type ToolLatencyMetrics struct {
	mu          sync.RWMutex
	name        string
	latencies   []int64 // in milliseconds
	count       int64
	totalMS     int64
	minMS       int64
	maxMS       int64
}

// NewLatencyTracker creates a new latency tracker
func NewLatencyTracker() *LatencyTracker {
	return &LatencyTracker{
		metrics: make(map[string]*ToolLatencyMetrics),
	}
}

// RecordLatency records a latency measurement for a tool
func (lt *LatencyTracker) RecordLatency(toolName string, latencyMS int64) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if _, exists := lt.metrics[toolName]; !exists {
		lt.metrics[toolName] = &ToolLatencyMetrics{
			name:      toolName,
			latencies: make([]int64, 0, 1000),
			minMS:     latencyMS,
			maxMS:     latencyMS,
		}
	}

	metric := lt.metrics[toolName]
	metric.mu.Lock()
	defer metric.mu.Unlock()

	metric.latencies = append(metric.latencies, latencyMS)
	metric.count++
	metric.totalMS += latencyMS

	if latencyMS < metric.minMS {
		metric.minMS = latencyMS
	}
	if latencyMS > metric.maxMS {
		metric.maxMS = latencyMS
	}
}

// GetMetrics returns latency metrics for a tool
func (lt *LatencyTracker) GetMetrics(toolName string) *ToolLatencyMetrics {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	if metric, exists := lt.metrics[toolName]; exists {
		return metric
	}
	return nil
}

// GetPercentile calculates percentile latency
func (tlm *ToolLatencyMetrics) GetPercentile(percentile int) int64 {
	tlm.mu.RLock()
	defer tlm.mu.RUnlock()

	if len(tlm.latencies) == 0 {
		return 0
	}

	// Simple percentile calculation (placeholder)
	// In production, use proper quantile algorithm
	idx := (len(tlm.latencies) * percentile) / 100
	if idx >= len(tlm.latencies) {
		idx = len(tlm.latencies) - 1
	}

	return tlm.latencies[idx]
}

// GetP50 returns 50th percentile (median)
func (tlm *ToolLatencyMetrics) GetP50() int64 {
	return tlm.GetPercentile(50)
}

// GetP95 returns 95th percentile
func (tlm *ToolLatencyMetrics) GetP95() int64 {
	return tlm.GetPercentile(95)
}

// GetP99 returns 99th percentile
func (tlm *ToolLatencyMetrics) GetP99() int64 {
	return tlm.GetPercentile(99)
}

// GetAverage returns average latency
func (tlm *ToolLatencyMetrics) GetAverage() int64 {
	tlm.mu.RLock()
	defer tlm.mu.RUnlock()

	if tlm.count == 0 {
		return 0
	}
	return tlm.totalMS / tlm.count
}

// GetStats returns all statistics as map
func (tlm *ToolLatencyMetrics) GetStats() map[string]int64 {
	tlm.mu.RLock()
	defer tlm.mu.RUnlock()

	return map[string]int64{
		"count":   tlm.count,
		"min_ms":  tlm.minMS,
		"max_ms":  tlm.maxMS,
		"avg_ms":  tlm.GetAverage(),
		"p50_ms":  tlm.GetP50(),
		"p95_ms":  tlm.GetP95(),
		"p99_ms":  tlm.GetP99(),
	}
}

// GetAllMetrics returns all tool metrics
func (lt *LatencyTracker) GetAllMetrics() map[string]map[string]int64 {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	result := make(map[string]map[string]int64)
	for name, metric := range lt.metrics {
		result[name] = metric.GetStats()
	}
	return result
}

// Reset clears all metrics
func (lt *LatencyTracker) Reset() {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	lt.metrics = make(map[string]*ToolLatencyMetrics)
}
