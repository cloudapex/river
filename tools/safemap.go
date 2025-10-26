package tools

import (
	"sync"
)

// BeeMap is a map with lock
type BeeMap struct {
	lock *sync.RWMutex
	bm   map[any]any
}

// NewBeeMap return new safemap
func NewBeeMap() *BeeMap {
	return &BeeMap{
		lock: new(sync.RWMutex),
		bm:   make(map[any]any),
	}
}

// Get from maps return the k's value
func (m *BeeMap) Get(k any) any {
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
func (m *BeeMap) Set(k any, v any) bool {
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
func (m *BeeMap) Check(k any) bool {
	m.lock.RLock()
	if _, ok := m.bm[k]; !ok {
		m.lock.RUnlock()
		return false
	}
	m.lock.RUnlock()
	return true
}

// Delete the given key and value.
func (m *BeeMap) Delete(k any) {
	m.lock.Lock()
	delete(m.bm, k)
	m.lock.Unlock()
}

// DeleteAll DeleteAll
func (m *BeeMap) DeleteAll() {
	m.lock.Lock()
	for k := range m.bm {
		delete(m.bm, k)
	}

	m.lock.Unlock()
}

// Items returns all items in safemap.
func (m *BeeMap) Items() map[any]any {
	m.lock.RLock()
	r := make(map[any]any)
	for k, v := range m.bm {
		r[k] = v
	}
	m.lock.RUnlock()
	return r
}
