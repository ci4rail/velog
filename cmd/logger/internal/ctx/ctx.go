package ctx

import (
	"context"
	"fmt"
	"sync"
)

type key int

const (
	keyWg key = iota
)

// NewWgContext creates a new context with a WaitGroup
func NewWgContext() (context.Context, context.CancelFunc, *sync.WaitGroup) {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, keyWg, wg)
	return ctx, cancel, wg
}

// WgFromContext returns the WaitGroup from the context and increments the WaitGroup counter
func WgFromContext(ctx context.Context) (*sync.WaitGroup, error) {
	wg, ok := ctx.Value(keyWg).(*sync.WaitGroup)
	if !ok {
		return nil, fmt.Errorf("context does not contain WaitGroup")
	}
	wg.Add(1)
	return wg, nil
}
