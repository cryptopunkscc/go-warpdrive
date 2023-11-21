package warpdrive

import (
	"sync"
)

type Unsubscribe func()
type Listener chan<- interface{}

type Subscriptions struct {
	mu   sync.Mutex
	subs map[Listener]Unsubscribe
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{subs: map[Listener]Unsubscribe{}}
}

func (s *Subscriptions) Notify(data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for subscriber := range s.subs {
		subscriber <- data
	}
}

func (s *Subscriptions) Subscribe(c Listener) (unsub Unsubscribe) {
	unsub = func() {
		s.mu.Lock()
		delete(s.subs, c)
		s.mu.Unlock()
	}
	s.mu.Lock()
	s.subs[c] = unsub
	s.mu.Unlock()
	return
}
