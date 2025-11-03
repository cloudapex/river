package tools

import (
	"sync"
)

// SafeMap is a map with lock
type SafeMap struct {
	lock *sync.RWMutex
	bm   map[any]any
}

// NewSafeMap return new safemap
func NewSafeMap() *SafeMap {
	return &SafeMap{
		lock: new(sync.RWMutex),
		bm:   make(map[any]any),
	}
}

// Get from maps return the k's value
func (m *SafeMap) Get(k any) any {
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
func (m *SafeMap) Set(k any, v any) bool {
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
func (m *SafeMap) Check(k any) bool {
	m.lock.RLock()
	if _, ok := m.bm[k]; !ok {
		m.lock.RUnlock()
		return false
	}
	m.lock.RUnlock()
	return true
}

// Delete the given key and value.
func (m *SafeMap) Delete(k any) {
	m.lock.Lock()
	delete(m.bm, k)
	m.lock.Unlock()
}

// DeleteAll DeleteAll
func (m *SafeMap) DeleteAll() {
	m.lock.Lock()
	for k := range m.bm {
		delete(m.bm, k)
	}

	m.lock.Unlock()
}

// Items returns all items in safemap.
func (m *SafeMap) Items() map[any]any {
	m.lock.RLock()
	r := make(map[any]any)
	for k, v := range m.bm {
		r[k] = v
	}
	m.lock.RUnlock()
	return r
}
