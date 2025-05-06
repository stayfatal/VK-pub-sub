package pubsub

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublish(t *testing.T) {
	var (
		counter    int32
		subsAmount = 5
		subject    = "test subject"
		wg         = &sync.WaitGroup{}
		sp         = NewPubSub().(*pubSub)
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	sp.subjects[subject] = make(map[string]chan interface{}, 0)
	for i := 0; i < subsAmount; i++ {
		ch := make(chan interface{})
		uuid := uuid.NewString()
		sp.subjects[subject][uuid] = ch
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ch:
				atomic.AddInt32(&counter, 1)
			case <-ctx.Done():
				return
			}
		}()
	}
	err := sp.Publish(subject, "some message")
	wg.Wait()
	require.NoError(t, err)
	require.NoError(t, ctx.Err())
	assert.Equal(t, int32(subsAmount), counter)
}

func TestPublishWithZeroSubs(t *testing.T) {
	var (
		subject = "test subject"
		sp      = NewPubSub().(*pubSub)
	)

	err := sp.Publish(subject, "some message")
	require.NoError(t, err)
}

func TestSubscribe(t *testing.T) {
	var (
		counter    int32
		subsAmount = 5
		subject    = "test subject"
		testMsg    = "some message"
		wg         = &sync.WaitGroup{}
		sp         = NewPubSub().(*pubSub)
	)

	for i := 0; i < subsAmount; i++ {
		wg.Add(1)
		_, err := sp.Subscribe(
			subject,
			func(msg interface{}) {
				defer wg.Done()
				atomic.AddInt32(&counter, 1)
				assert.Equal(t, testMsg, msg)
			},
		)
		require.NoError(t, err)
	}

	for _, ch := range sp.subjects[subject] {
		ch <- testMsg
	}
	wg.Wait()
	assert.Equal(t, int32(subsAmount), counter)
}

func TestUnsubscribe(t *testing.T) {
	var (
		subsAmount = 5
		subject    = "test subject"
		sp         = NewPubSub().(*pubSub)
	)

	for i := 0; i < subsAmount; i++ {
		subscription, err := sp.Subscribe(
			subject,
			func(msg interface{}) {},
		)
		require.NoError(t, err)
		subscription.Unsubscribe()
	}

	for _, ch := range sp.subjects[subject] {
		_, ok := <-ch
		assert.False(t, ok)
	}
}

func TestCloseWithoutContext(t *testing.T) {
	var (
		subsAmount = 5
		subject    = "test subject"
		sp         = NewPubSub().(*pubSub)
	)

	for i := 0; i < subsAmount; i++ {
		sp.Subscribe(subject, func(msg interface{}) {})
	}

	err := sp.Close(context.Background())
	require.NoError(t, err)

	for _, ch := range sp.subjects[subject] {
		_, ok := <-ch
		assert.False(t, ok)
	}
}

func TestCloseWithContext(t *testing.T) {
	var (
		subsAmount = 5
		subject    = "test subject"
		sp         = NewPubSub().(*pubSub)
	)

	for i := 0; i < subsAmount; i++ {
		sp.Subscribe(subject, func(msg interface{}) {})
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := sp.Close(ctx)
	require.Error(t, err)
}

func TestUnsubscribeAfterClose(t *testing.T) {
	var (
		subsAmount = 5
		subject    = "test subject"
		sp         = NewPubSub().(*pubSub)
	)

	subs := []Subscription{}
	for i := 0; i < subsAmount; i++ {
		sub, err := sp.Subscribe(subject, func(msg interface{}) {})
		require.NoError(t, err)
		subs = append(subs, sub)
	}

	err := sp.Close(context.Background())
	require.NoError(t, err)

	for _, sub := range subs {
		sub.Unsubscribe()
	}
}

func TestMultipleUnsubscribe(t *testing.T) {
	var (
		subject = "test subject"
		sp      = NewPubSub().(*pubSub)
	)

	sub, err := sp.Subscribe(subject, func(msg interface{}) {})
	require.NoError(t, err)

	sub.Unsubscribe()
	sub.Unsubscribe()

	err = sp.Close(context.Background())
	require.NoError(t, err)
}

func TestSubscribeAfterClose(t *testing.T) {
	var (
		subject = "test subject"
		sp      = NewPubSub().(*pubSub)
	)

	err := sp.Close(context.Background())
	require.NoError(t, err)

	_, err = sp.Subscribe(subject, func(msg interface{}) {})
	require.Error(t, err)
}

func TestPublishAfterClose(t *testing.T) {
	var (
		subsAmount = 5
		subject    = "test subject"
		sp         = NewPubSub().(*pubSub)
	)

	for i := 0; i < subsAmount; i++ {
		_, err := sp.Subscribe(subject, func(msg interface{}) {})
		assert.Nil(t, err)
	}

	err := sp.Close(context.Background())
	require.NoError(t, err)

	err = sp.Publish(subject, "some mes")
	require.Error(t, err)
}

func TestConcurrencyPublishWhileClosing(t *testing.T) {
	var (
		subsAmount = 5
		subject    = "test subject"
		sp         = NewPubSub().(*pubSub)
	)

	for i := 0; i < subsAmount; i++ {
		_, err := sp.Subscribe(subject, func(msg interface{}) {})
		require.NoError(t, err)
	}

	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				sp.Publish(subject, "some mes")
			}
		}
	}()

	err := sp.Close(context.Background())
	require.NoError(t, err)
	stop <- struct{}{}
}

func TestRace(t *testing.T) {
	sp := NewPubSub()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sub, _ := sp.Subscribe("test", func(msg interface{}) {})
			_ = sp.Publish("test", "msg")
			sub.Unsubscribe()
		}()
	}
	wg.Wait()
	err := sp.Close(context.Background())
	require.NoError(t, err)
}
