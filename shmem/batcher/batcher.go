//go:build !solution

package batcher

import (
	"sync"

	"gitlab.com/slon/shad-go/batcher/slow"
)

type Batcher struct {
	slow *slow.Value

	mu      sync.Mutex
	value   interface{}
	ch      chan struct{}
	running bool
}

func NewBatcher(v *slow.Value) *Batcher {
	return &Batcher{
		slow: v,
		ch:   make(chan struct{}),
	}
}

func (b *Batcher) Load() interface{} {
	b.mu.Lock()
	started := false

	if !b.running {
		b.running = true
		started = true
		go b.doLoad()
	}
	ch := b.ch
	b.mu.Unlock()

	<-ch

	if started {
		b.mu.Lock()
		val := b.value
		b.mu.Unlock()
		return val
	}

	b.mu.Lock()
	if !b.running {
		b.running = true
		go b.doLoad()
	}
	ch = b.ch
	b.mu.Unlock()

	<-ch

	b.mu.Lock()
	val := b.value
	b.mu.Unlock()
	return val
}

func (b *Batcher) doLoad() {
	newVal := b.slow.Load()

	b.mu.Lock()
	b.value = newVal
	close(b.ch)
	b.ch = make(chan struct{})
	b.running = false
	b.mu.Unlock()
}
