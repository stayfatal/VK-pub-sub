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
	// флаг поднят когда был вызван Close()
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

// MessageHandler будет запущен асинхронно
// Чтобы избежать утечки горутин контролируйте выполнение MessageHandler
func (sp *pubSub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	if sp.closed.Load() {
		return nil, ErrClosed
	}

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
			go cb(mes) // запускаем асинхронно чтобы не ждать выполнения долгой обработки
		}
	}()

	return subscription, nil
}

func (sp *pubSub) Publish(subject string, msg interface{}) error {
	if sp.closed.Load() {
		return ErrClosed
	}

	// мьютекс повешен на всю функцию для того чтобы не дать другим частям кода закрыть канал во время записи туда
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	channels, ok := sp.subjects[subject]
	if !ok || len(channels) == 0 {
		return nil
	}

	// рассылка всем сабам
	for _, ch := range channels {
		ch <- msg
	}

	return nil
}

func (sp *pubSub) Close(ctx context.Context) error {
	sp.closed.Swap(true)
	// лочим целый блок потому что нельзя разделить закрытие канала и его удаление из мапы
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
	uuid     string
	subject  string
	pubSub   *pubSub
	unsubbed *atomic.Bool
}

func newSubscription(subject string, pubSub *pubSub) (Subscription, chan interface{}) {
	uuid := uuid.NewString()
	ch := make(chan interface{}, 1)
	pubSub.mu.Lock()
	pubSub.subjects[subject][uuid] = ch
	pubSub.mu.Unlock()

	return &subscription{
		uuid:     uuid,
		subject:  subject,
		pubSub:   pubSub,
		unsubbed: &atomic.Bool{},
	}, ch
}

func (s *subscription) Unsubscribe() {
	if s.pubSub.closed.Load() || s.unsubbed.Load() {
		return
	}
	s.unsubbed.Swap(true)
	// лочим целый блок потому что нельзя разделить закрытие канала и его удаление из мапы
	s.pubSub.mu.Lock()
	close(s.pubSub.subjects[s.subject][s.uuid])
	delete(s.pubSub.subjects[s.subject], s.uuid)
	s.pubSub.mu.Unlock()
}
