package store

import (
	"sync"

	"project/parser"
)

// MemoryStore keeps all entries in memory
type MemoryStore struct {
	mu      sync.RWMutex
	entries []parser.Entry
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		entries: make([]parser.Entry, 0, 10_000),
	}
}

// Save appends one entry (Safe for concurrent calls)
func (m *MemoryStore) Save(e parser.Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, e)
	return nil
}

// Snapshot returns a copy of stored entries for reading without holding the lock
func (m *MemoryStore) Snapshot() []parser.Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]parser.Entry, len(m.entries))
	copy(out, m.entries)
	return out
}
