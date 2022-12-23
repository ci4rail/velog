// Package processdatastore is a package that provides a process data store.
// The process data store is used to store process data objects associated with an address. For each address, only the most recent object is stored.
// The number of updates for each address is also stored.
// The typical use case is to call Write() from one goroutine and List()/Read() from another goroutine, which periodically outputs the process data store.
// The process data store is thread safe.
package processdatastore

import (
	"fmt"
	"sort"
	"sync"
)

// StoreEntry is the process data store for a single address
type StoreEntry struct {
	RecentObject Object
	numUpdates   int
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
	e.numUpdates++
	e.RecentObject = o
}

// Read reads the entry for the specified address from the process data store.
// In addition, it returns the number of updates for the address since the last call to Read().
// If the address has never got an update, an error is returned.
func (s *Store) Read(address uint32) (Object, int, error) {
	s.Lock()
	defer s.Unlock()

	e, ok := s.entry[address]
	if !ok {
		return nil, 0, fmt.Errorf("no entries for address %d", address)
	}
	numUpdates := e.numUpdates
	e.numUpdates = 0
	return e.RecentObject, numUpdates, nil
}

// List returns a list of all addresses in the process data store which have received any updates since the store creation.
// The list is sorted in ascending order.
func (s *Store) List() []int {
	s.Lock()
	defer s.Unlock()

	keys := make([]int, 0, len(s.entry))
	for k := range s.entry {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	return keys
}
