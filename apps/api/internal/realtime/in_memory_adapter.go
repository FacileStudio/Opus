package realtime

import (
	"context"
	"errors"
	"sync"
)

// InMemoryAdapter delivers broadcast envelopes within a single process.
type InMemoryAdapter struct {
	mu       sync.RWMutex
	handlers []Handler
	closed   bool
}

// NewInMemoryAdapter creates a new in-memory broadcast adapter.
func NewInMemoryAdapter() *InMemoryAdapter {
	return &InMemoryAdapter{
		handlers: make([]Handler, 0),
	}
}

// Publish delivers an envelope to all registered subscribers.
func (a *InMemoryAdapter) Publish(_ context.Context, env Envelope) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.closed {
		return errors.New("adapter is closed")
	}
	for _, handler := range a.handlers {
		handler(env)
	}
	return nil
}

// Subscribe registers a handler for incoming envelopes.
func (a *InMemoryAdapter) Subscribe(_ context.Context, handler Handler) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return errors.New("adapter is closed")
	}
	a.handlers = append(a.handlers, handler)
	return nil
}

// Close terminates the in-memory adapter and clears subscribers.
func (a *InMemoryAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.closed = true
	a.handlers = nil
	return nil
}
