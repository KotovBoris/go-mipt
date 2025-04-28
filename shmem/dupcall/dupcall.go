//go:build !solution

package dupcall

import (
	"context"
	"sync"
)

type Call struct {
	mu          sync.Mutex
	running     int
	innerCtx    context.Context
	innerCancel context.CancelFunc
	ready       chan struct{}
	result      interface{}
	err         error
}

func (o *Call) Do(ctx context.Context, cb func(context.Context) (interface{}, error)) (interface{}, error) {
	o.mu.Lock()
	o.running++
	start := false
	if o.running == 1 {
		o.innerCtx, o.innerCancel = context.WithCancel(context.Background())
		o.ready = make(chan struct{})
		start = true
	}
	o.mu.Unlock()

	if start {
		go func() {
			r, err := cb(o.innerCtx)
			o.mu.Lock()
			o.result = r
			o.err = err
			close(o.ready)
			o.mu.Unlock()
		}()
	}

	select {
	case <-ctx.Done():
		o.decrement()
		return nil, ctx.Err()
	case <-o.ready:
		res, err := o.getResult()
		o.decrement()
		return res, err
	}
}

func (o *Call) decrement() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.running--
	if o.running == 0 && o.innerCancel != nil {
		o.innerCancel()
	}
}

func (o *Call) getResult() (interface{}, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.result, o.err
}
