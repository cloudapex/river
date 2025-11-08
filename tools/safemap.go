package tools

import (
	"sync"
)

// SafeMap is a map with lock
type SafeMap[K comparable] struct {
	lock *sync.RWMutex
	bm   map[K]any
}

// NewSafeMap return new safemap
func NewSafeMap[K comparable]() *SafeMap[K] {
	return &SafeMap[K]{
		lock: new(sync.RWMutex),
		bm:   make(map[K]any),
	}
}

// Get from maps return the k's value
func (m *SafeMap[K]) Get(k K) any {
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
func (m *SafeMap[K]) Set(k K, v any) bool {
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
func (m *SafeMap[K]) Check(k K) bool {
	m.lock.RLock()
	if _, ok := m.bm[k]; !ok {
		m.lock.RUnlock()
		return false
	}
	m.lock.RUnlock()
	return true
}

// Delete the given key and value.
func (m *SafeMap[K]) Delete(k K) {
	m.lock.Lock()
	delete(m.bm, k)
	m.lock.Unlock()
}

// DeleteAll DeleteAll
func (m *SafeMap[K]) DeleteAll() {
	m.lock.Lock()
	for k := range m.bm {
		delete(m.bm, k)
	}

	m.lock.Unlock()
}

// Items returns all items in safemap.
func (m *SafeMap[K]) Items() map[K]any {
	m.lock.RLock()
	r := make(map[K]any)
	for k, v := range m.bm {
		r[k] = v
	}
	m.lock.RUnlock()
	return r
}
