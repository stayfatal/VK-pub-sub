package pubsub

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type PubSub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)

	Publish(subject string, msg interface{}) error

	Close(ctx context.Context) error
}

type pubSub struct {
	// {key - subject value - map {key - sub_uuid value chan}}
	subjects map[string]map[string]chan interface{}
	// true if Close was called
	closed *atomic.Bool
	mu     *sync.RWMutex
	wg     *sync.WaitGroup
}

func NewPubSub() PubSub {
	return &pubSub{
		subjects: make(map[string]map[string]chan interface{}),
		closed:   &atomic.Bool{},
		mu:       &sync.RWMutex{},
		wg:       &sync.WaitGroup{},
	}
}

func (sp *pubSub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	sp.mu.Lock()
	if _, ok := sp.subjects[subject]; !ok {
		sp.subjects[subject] = make(map[string]chan interface{}, 0)
	}
	sp.mu.Unlock()

	subscription, ch := newSubscription(subject, sp)
	sp.wg.Add(1)
	go func() {
		defer sp.wg.Done()

		for {
			mes, ok := <-ch
			if !ok {
				return
			}
			go cb(mes) // for situations when mesHandler is slow
		}
	}()

	return subscription, nil
}

func (sp *pubSub) Publish(subject string, msg interface{}) error {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	channels, ok := sp.subjects[subject]
	if !ok {
		return ErrNonExistentSubject
	}

	// broadcast
	for _, ch := range channels {
		ch <- msg
	}

	return nil
}

func (sp *pubSub) Close(ctx context.Context) error {
	sp.closed.Swap(true)
	// close and delete cant be separated cuz then publish will write to close channel
	sp.mu.Lock()
	for _, subs := range sp.subjects {
		for _, ch := range subs {
			close(ch)
		}
	}
	sp.subjects = make(map[string]map[string]chan interface{})
	sp.mu.Unlock()

	waited := make(chan struct{})
	go func() {
		sp.wg.Wait()
		waited <- struct{}{}
	}()
	select {
	case <-waited:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type subscription struct {
	uuid    string
	subject string
	pubSub  *pubSub
}

func newSubscription(subject string, pubSub *pubSub) (Subscription, chan interface{}) {
	uuid := uuid.NewString()
	ch := make(chan interface{}, 1)
	pubSub.mu.Lock()
	pubSub.subjects[subject][uuid] = ch
	pubSub.mu.Unlock()

	return &subscription{
		uuid:    uuid,
		subject: subject,
		pubSub:  pubSub,
	}, ch
}

func (s *subscription) Unsubscribe() {
	if !s.pubSub.closed.Load() {
		// close and delete cant be separated cuz then publish will write to close channel
		s.pubSub.mu.Lock()
		close(s.pubSub.subjects[s.subject][s.uuid])
		delete(s.pubSub.subjects[s.subject], s.uuid)
		s.pubSub.mu.Unlock()
	}
}
