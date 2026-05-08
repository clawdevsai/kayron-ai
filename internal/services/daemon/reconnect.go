package daemon

import (
	"context"
	"math"
	"time"

	"github.com/lukeware/kayron-ai/internal/logger"
	mt5client "github.com/lukeware/kayron-ai/internal/services/mt5"
)

// Reconnector handles auto-reconnect logic with exponential backoff
type Reconnector struct {
	client         *mt5client.Client
	maxRetries     int
	baseBackoff    time.Duration
	maxBackoff     time.Duration
	heartbeatTick  time.Duration
	logger         *logger.Logger
	isConnected    bool
	lastHeartbeat  time.Time
	consecutiveFails int
}

// NewReconnector creates a new reconnector
func NewReconnector(client *mt5client.Client) *Reconnector {
	return &Reconnector{
		client:        client,
		maxRetries:    5,
		baseBackoff:   time.Second,
		maxBackoff:    time.Minute,
		heartbeatTick: time.Second * 5,
		logger:        logger.New("Reconnector"),
		isConnected:   false,
	}
}

// Start begins the auto-reconnect loop
func (r *Reconnector) Start(ctx context.Context) {
	ticker := time.NewTicker(r.heartbeatTick)
	defer ticker.Stop()

	r.logger.Info("Reconnector started with 5s heartbeat interval")

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Reconnector stopped")
			return
		case <-ticker.C:
			r.performHealthCheck(ctx)
		}
	}
}

// performHealthCheck performs a health check and reconnects if needed
func (r *Reconnector) performHealthCheck(ctx context.Context) {
	// Try to get account info as a health check
	_, err := r.client.GetAccount()

	if err == nil {
		if !r.isConnected {
			r.logger.Info("Successfully reconnected to MT5")
			r.isConnected = true
			r.consecutiveFails = 0
		}
		r.lastHeartbeat = time.Now()
		return
	}

	r.consecutiveFails++
	r.isConnected = false

	if r.consecutiveFails > r.maxRetries {
		r.logger.Warn("Max reconnection retries exceeded")
		return
	}

	backoff := r.calculateBackoff()
	r.logger.WarnWithError(
		"Health check failed, retrying in "+backoff.String(),
		err,
	)

	time.Sleep(backoff)
}

// calculateBackoff calculates exponential backoff with max cap
func (r *Reconnector) calculateBackoff() time.Duration {
	backoff := time.Duration(math.Pow(2, float64(r.consecutiveFails))) * r.baseBackoff
	if backoff > r.maxBackoff {
		backoff = r.maxBackoff
	}
	return backoff
}

// IsConnected returns the current connection status
func (r *Reconnector) IsConnected() bool {
	return r.isConnected
}

// LastHeartbeat returns the timestamp of the last successful health check
func (r *Reconnector) LastHeartbeat() time.Time {
	return r.lastHeartbeat
}

// Reset resets the reconnection state
func (r *Reconnector) Reset() {
	r.consecutiveFails = 0
	r.isConnected = false
	r.lastHeartbeat = time.Time{}
}
