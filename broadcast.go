package warpdrive

import (
	"context"
	"sync"
)

type Broadcast[T any] struct {
	mu   sync.Mutex
	subs map[chan<- T]any
}

func NewBroadcast[T any]() *Broadcast[T] {
	return &Broadcast[T]{subs: make(map[chan<- T]any)}
}

func (s *Broadcast[T]) Emit(v T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for sub := range s.subs {
		sub <- v
	}
}

func (s *Broadcast[T]) Listen(ctx context.Context, c chan<- T) {
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		delete(s.subs, c)
		s.mu.Unlock()
	}()
	s.mu.Lock()
	s.subs[c] = nil
	s.mu.Unlock()
	return
}
