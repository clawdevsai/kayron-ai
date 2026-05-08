package health

import (
	"sync"
	"time"
)

// HealthStatus represents terminal health status
type HealthStatus struct {
	TerminalConnected    bool      `json:"terminal_connected"`
	LastHeartbeat        time.Time `json:"last_heartbeat"`
	LastHeartbeatString  string    `json:"last_heartbeat_string"`
	SecondsSinceHeartbeat int64    `json:"seconds_since_heartbeat"`
	UpSinceMilliseconds  int64     `json:"up_since_milliseconds"`
	IsHealthy            bool      `json:"is_healthy"`
	Message              string    `json:"message"`
}

// HealthMonitor monitors terminal connection health
type HealthMonitor struct {
	mu                    sync.RWMutex
	connected             bool
	lastHeartbeat         time.Time
	startTime             time.Time
	heartbeatTimeoutSecs  int64
	maxSecondsSinceHealth int64
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(heartbeatTimeoutSecs int64) *HealthMonitor {
	return &HealthMonitor{
		connected:            false,
		lastHeartbeat:        time.Now(),
		startTime:            time.Now(),
		heartbeatTimeoutSecs: heartbeatTimeoutSecs, // Default 10 seconds
		maxSecondsSinceHealth: heartbeatTimeoutSecs,
	}
}

// RecordHeartbeat records a successful heartbeat
func (hm *HealthMonitor) RecordHeartbeat() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.lastHeartbeat = time.Now()
	hm.connected = true
}

// RecordDisconnect records a disconnection
func (hm *HealthMonitor) RecordDisconnect() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.connected = false
}

// SetConnected sets the connected status
func (hm *HealthMonitor) SetConnected(connected bool) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.connected = connected
	if connected {
		hm.lastHeartbeat = time.Now()
	}
}

// GetStatus returns current health status
func (hm *HealthMonitor) GetStatus() *HealthStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	now := time.Now()
	lastHeartbeat := hm.lastHeartbeat
	secondsSince := int64(now.Sub(lastHeartbeat).Seconds())
	upSince := int64(now.Sub(hm.startTime).Milliseconds())

	// Health is ok if heartbeat was received within timeout
	isHealthy := hm.connected && (secondsSince < hm.heartbeatTimeoutSecs)

	message := "OK"
	if !hm.connected {
		message = "Terminal desconectado"
	} else if secondsSince >= hm.heartbeatTimeoutSecs {
		message = "Heartbeat expirado - verificando reconexão"
	}

	return &HealthStatus{
		TerminalConnected:     hm.connected,
		LastHeartbeat:         lastHeartbeat,
		LastHeartbeatString:   lastHeartbeat.Format(time.RFC3339),
		SecondsSinceHeartbeat: secondsSince,
		UpSinceMilliseconds:   upSince,
		IsHealthy:             isHealthy,
		Message:               message,
	}
}

// IsConnected returns whether terminal is connected
func (hm *HealthMonitor) IsConnected() bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	return hm.connected
}

// IsHealthy returns whether terminal is in good health
func (hm *HealthMonitor) IsHealthy() bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if !hm.connected {
		return false
	}

	secondsSince := int64(time.Since(hm.lastHeartbeat).Seconds())
	return secondsSince < hm.heartbeatTimeoutSecs
}

// GetLastHeartbeatAge returns seconds since last heartbeat
func (hm *HealthMonitor) GetLastHeartbeatAge() int64 {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	return int64(time.Since(hm.lastHeartbeat).Seconds())
}

// Reset resets the health monitor
func (hm *HealthMonitor) Reset() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.connected = false
	hm.lastHeartbeat = time.Now()
	hm.startTime = time.Now()
}
