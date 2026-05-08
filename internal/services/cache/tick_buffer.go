package cache

import (
	"sync"

	"github.com/lukeware/kayron-ai/internal/models"
)

// TickBuffer circular buffer for market ticks per symbol
type TickBuffer struct {
	symbol string
	ticks  []*models.Tick
	head   int // next write position
	size   int // current number of ticks
	max    int // buffer capacity
	mu     sync.RWMutex
}

// NewTickBuffer creates circular buffer, capacity 1000 ticks per symbol
func NewTickBuffer(symbol string) *TickBuffer {
	return &TickBuffer{
		symbol: symbol,
		ticks:  make([]*models.Tick, 1000),
		head:   0,
		size:   0,
		max:    1000,
	}
}

// Write adds tick to buffer, overwrites oldest if full
func (b *TickBuffer) Write(tick *models.Tick) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.ticks[b.head] = tick
	b.head = (b.head + 1) % b.max

	if b.size < b.max {
		b.size++
	}
}

// Read returns copy of all ticks in chronological order
func (b *TickBuffer) Read() []*models.Tick {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.size == 0 {
		return []*models.Tick{}
	}

	result := make([]*models.Tick, b.size)
	for i := 0; i < b.size; i++ {
		idx := (b.head - b.size + i + b.max) % b.max
		result[i] = b.ticks[idx]
	}
	return result
}

// Clear empties buffer
func (b *TickBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.head = 0
	b.size = 0
	b.ticks = make([]*models.Tick, b.max)
}

// TickBufferRegistry manages buffers per symbol
type TickBufferRegistry struct {
	buffers map[string]*TickBuffer
	mu      sync.RWMutex
}

// NewTickBufferRegistry creates registry
func NewTickBufferRegistry() *TickBufferRegistry {
	return &TickBufferRegistry{
		buffers: make(map[string]*TickBuffer),
	}
}

// Get returns or creates buffer for symbol
func (r *TickBufferRegistry) Get(symbol string) *TickBuffer {
	r.mu.RLock()
	if buf, exists := r.buffers[symbol]; exists {
		r.mu.RUnlock()
		return buf
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after lock
	if buf, exists := r.buffers[symbol]; exists {
		return buf
	}

	buf := NewTickBuffer(symbol)
	r.buffers[symbol] = buf
	return buf
}

// Write adds tick to symbol buffer
func (r *TickBufferRegistry) Write(symbol string, tick *models.Tick) {
	buf := r.Get(symbol)
	buf.Write(tick)
}

// Read returns ticks for symbol
func (r *TickBufferRegistry) Read(symbol string) []*models.Tick {
	buf := r.Get(symbol)
	return buf.Read()
}

// Clear empties all buffers
func (r *TickBufferRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, buf := range r.buffers {
		buf.Clear()
	}
}
