package tools

import (
	"sync"
)

// Map is a map with lock
type Map[K comparable] struct {
	lock *sync.RWMutex
	bm   map[K]any
}

// NewSafeMap return new safemap
func NewSafeMap[K comparable]() *Map[K] {
	return &Map[K]{
		lock: new(sync.RWMutex),
		bm:   make(map[K]any),
	}
}

// Get from maps return the k's value
func (m *Map[K]) Get(k K) any {
	m.lock.RLock()
	if val, ok := m.bm[k]; ok {
		m.lock.RUnlock()
		return val
	}
	m.lock.RUnlock()
	return nil
}

// Set Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *Map[K]) Set(k K, v any) bool {
	m.lock.Lock()
	if val, ok := m.bm[k]; !ok {
		m.bm[k] = v
		m.lock.Unlock()
	} else if val != v {
		m.bm[k] = v
		m.lock.Unlock()
	} else {
		m.lock.Unlock()
		return false
	}
	return true
}

// Check Returns true if k is exist in the map.
func (m *Map[K]) Check(k K) bool {
	m.lock.RLock()
	if _, ok := m.bm[k]; !ok {
		m.lock.RUnlock()
		return false
	}
	m.lock.RUnlock()
	return true
}

// Delete the given key and value.
func (m *Map[K]) Delete(k K) {
	m.lock.Lock()
	delete(m.bm, k)
	m.lock.Unlock()
}

// DeleteAll DeleteAll
func (m *Map[K]) DeleteAll() {
	m.lock.Lock()
	m.bm = make(map[K]any)
	m.lock.Unlock()
}

// Items returns all items in safemap.
func (m *Map[K]) Items() map[K]any {
	m.lock.RLock()
	r := make(map[K]any)
	for k, v := range m.bm {
		r[k] = v
	}
	m.lock.RUnlock()
	return r
}
