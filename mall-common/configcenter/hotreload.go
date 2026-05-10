package configcenter

import "sync"

// HotConfig is a concurrency-safe holder for a hot-reloadable config value.
// Read via Get() from request handlers; update via Set() from the Watcher callback.
type HotConfig[T any] struct {
	mu  sync.RWMutex
	val T
}

// NewHotConfig initialises a HotConfig with an initial value.
func NewHotConfig[T any](initial T) *HotConfig[T] {
	return &HotConfig[T]{val: initial}
}

// Get returns the current value (read-locked).
func (h *HotConfig[T]) Get() T {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.val
}

// Set replaces the value atomically (write-locked). Called from Watcher callbacks.
func (h *HotConfig[T]) Set(v T) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.val = v
}
