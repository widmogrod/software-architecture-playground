package projection

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
)

type subscriber[T any] struct {
	f    func(T) error
	done chan error
}

func NewPubSubChan[T any]() *PubSubChan[T] {
	return &PubSubChan[T]{
		lock:        &sync.RWMutex{},
		channel:     make(chan T),
		subscribers: nil,
	}
}

type PubSubChan[T any] struct {
	lock        *sync.RWMutex
	channel     chan T
	subscribers []subscriber[T]
	isClosed    atomic.Bool
}

func (s *PubSubChan[T]) Publish(msg T) error {
	if msg2, ok := any(msg).(Message); ok {
		if msg2.finished {
			s.channel <- msg
			return nil
		}
	}

	if s.isClosed.Load() {
		return fmt.Errorf("PubSubChan.Publish: channel is closed %w", ErrFinished)
	}
	s.channel <- msg
	return nil
}

func (s *PubSubChan[T]) Process() {
	for result := range s.channel {
		wg := &sync.WaitGroup{}
		s.lock.RLock()
		for _, sub := range s.subscribers {
			wg.Add(1)
			go func(sub subscriber[T]) {
				defer wg.Done()

				if msg, ok := any(result).(Message); ok {
					if msg.finished {
						close(sub.done)
						return
					}
				}
				err := sub.f(result)
				if err != nil {
					log.Errorf("PubSubChan.Process: %s", err)
					sub.done <- err
				}
			}(sub)
		}
		s.lock.RUnlock()

		wg.Wait()

		if msg, ok := any(result).(Message); ok {
			if msg.finished {
				s.Close()
				return
			}
		}
	}

	s.lock.RLock()
	for _, sub := range s.subscribers {
		close(sub.done)
	}
	s.lock.RUnlock()
}

func (s *PubSubChan[T]) Subscribe(f func(T) error) error {
	if s.isClosed.Load() {
		return fmt.Errorf("PubSubChan.Subscribe: channel is closed %w", ErrFinished)
	}

	done := make(chan error)

	s.lock.Lock()
	s.subscribers = append(s.subscribers, subscriber[T]{
		f:    f,
		done: done,
	})
	s.lock.Unlock()

	err := <-done

	//s.lock.Lock()
	//for idx, sub := range s.subscribers {
	//	if sub.done == done {
	//		s.subscribers = append(s.subscribers[:idx], s.subscribers[idx+1:]...)
	//		break
	//	}
	//}
	//s.lock.Unlock()

	return err
}

func (s *PubSubChan[T]) Close() {
	if s.isClosed.Load() {
		return
	}

	s.isClosed.Store(true)
	close(s.channel)
}
