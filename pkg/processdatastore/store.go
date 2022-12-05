package processdatastore

import (
	"fmt"
	"sync"
)

// StoreEntry is the process data store for a single address
type StoreEntry struct {
	RecentObject Object
	NumUpdates   int
}

// Store is the process data store
type Store struct {
	sync.RWMutex
	entry map[uint32]*StoreEntry
}

// NewStore creates a new process data store
func NewStore() *Store {
	return &Store{
		entry: make(map[uint32]*StoreEntry),
	}
}

// Write writes an object to the process data store
func (s *Store) Write(o Object) {
	s.Lock()
	defer s.Unlock()

	e, ok := s.entry[o.Address()]
	if !ok {
		e = &StoreEntry{}
		s.entry[o.Address()] = e
	}
	e.NumUpdates++
	e.RecentObject = o
}

// ReadAndClearEntry reads the entry for the specified address from the process data store
// After reading the entry, the entry is cleared
func (s *Store) ReadAndClearEntry(address uint32) (*StoreEntry, error) {
	s.Lock()
	defer s.Unlock()

	e, ok := s.entry[address]
	if !ok {
		return nil, fmt.Errorf("no entries for address %d", address)
	}
	delete(s.entry, address)
	return e, nil
}
