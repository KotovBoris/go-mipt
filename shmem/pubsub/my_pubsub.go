//go:build !solution

package pubsub

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrPubSubClosed = errors.New("pubsub is closed")

type publishTask struct {
	subj     string
	msg      interface{}
	snapshot []*MySubscription
}

var _ Subscription = (*MySubscription)(nil)

type MySubscription struct {
	pub   *MyPubSub
	subj  string
	cb    MsgHandler
	msgCh chan interface{}
	done  chan struct{}
	once  sync.Once
}

func (s *MySubscription) run() {
	for msg := range s.msgCh {
		s.cb(msg)
	}
	close(s.done)
}

func (s *MySubscription) Unsubscribe() {
	s.once.Do(func() {
		s.pub.mu.Lock()
		if subs, ok := s.pub.subs[s.subj]; ok {
			delete(subs, s)
			if len(subs) == 0 {
				delete(s.pub.subs, s.subj)
			}
		}
		s.pub.mu.Unlock()
		close(s.msgCh)
		<-s.done
	})
}

var _ PubSub = (*MyPubSub)(nil)

type MyPubSub struct {
	mu        sync.RWMutex
	subs      map[string]map[*MySubscription]struct{}
	publishCh chan publishTask
	closed    bool
	processWG sync.WaitGroup
	subsWG    sync.WaitGroup
}

func NewPubSub() PubSub {
	pub := &MyPubSub{
		subs:      make(map[string]map[*MySubscription]struct{}),
		publishCh: make(chan publishTask, 1000),
	}
	pub.processWG.Add(1)
	go pub.processLoop()
	return pub
}

func (p *MyPubSub) processLoop() {
	defer p.processWG.Done()
	for task := range p.publishCh {
		for _, sub := range task.snapshot {
			sub.msgCh <- task.msg
		}
	}
}

func (p *MyPubSub) Subscribe(subj string, cb MsgHandler) (Subscription, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, ErrPubSubClosed
	}
	if _, ok := p.subs[subj]; !ok {
		p.subs[subj] = make(map[*MySubscription]struct{})
	}
	sub := &MySubscription{
		pub:   p,
		subj:  subj,
		cb:    cb,
		msgCh: make(chan interface{}, 1000),
		done:  make(chan struct{}),
	}
	p.subs[subj][sub] = struct{}{}
	p.subsWG.Add(1)
	go func() {
		defer p.subsWG.Done()
		sub.run()
	}()
	return sub, nil
}

func (p *MyPubSub) Publish(subj string, msg interface{}) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return ErrPubSubClosed
	}
	var snapshot []*MySubscription
	if subs, ok := p.subs[subj]; ok {
		snapshot = make([]*MySubscription, 0, len(subs))
		for sub := range subs {
			snapshot = append(snapshot, sub)
		}
	}
	p.mu.RUnlock()
	task := publishTask{
		subj:     subj,
		msg:      msg,
		snapshot: snapshot,
	}
	select {
	case p.publishCh <- task:
		return nil
	case <-time.After(100 * time.Millisecond):
		return errors.New("publish queue is full")
	}
}

func (p *MyPubSub) Close(ctx context.Context) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	close(p.publishCh)
	p.mu.Unlock()
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.processWG.Wait()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	p.mu.Lock()
	for _, subs := range p.subs {
		for sub := range subs {
			close(sub.msgCh)
		}
	}
	p.mu.Unlock()
	doneSubs := make(chan struct{})
	go func() {
		defer close(doneSubs)
		p.subsWG.Wait()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-doneSubs:
		return nil
	}
}
