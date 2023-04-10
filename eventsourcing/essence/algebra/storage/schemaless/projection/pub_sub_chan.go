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
	once sync.Once
}

func (s *subscriber[T]) Close() {
	s.CloseWithErr(nil)
}

func (s *subscriber[T]) CloseWithErr(err error) {
	s.once.Do(func() {
		s.done <- err
		close(s.done)
	})
}

func NewPubSubChan[T any]() *PubSubChan[T] {
	return &PubSubChan[T]{
		lock:        &sync.RWMutex{},
		channel:     make(chan T, 100),
		subscribers: nil,
	}
}

type PubSubChan[T any] struct {
	lock        *sync.RWMutex
	channel     chan T
	subscribers []*subscriber[T]
	isClosed    atomic.Bool
	once        sync.Once
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
			go func(sub *subscriber[T]) {
				defer wg.Done()

				if msg, ok := any(result).(Message); ok {
					if msg.finished {
						//sub.Close()
						return
					}
				}
				err := sub.f(result)
				if err != nil {
					log.Errorf("PubSubChan.Process: %s", err)
					sub.CloseWithErr(err)
				}
			}(sub)
		}
		s.lock.RUnlock()
		wg.Wait()

		if msg, ok := any(result).(Message); ok {
			if msg.finished {
				s.Close()
				break
			}
		}
	}

	s.lock.RLock()
	for _, sub := range s.subscribers {
		sub.Close()
	}
	s.lock.RUnlock()
}

func (s *PubSubChan[T]) Subscribe(f func(T) error) error {
	if s.isClosed.Load() {
		return fmt.Errorf("PubSubChan.Subscribe: channel is closed %w", ErrFinished)
	}

	sub := &subscriber[T]{
		f:    f,
		done: make(chan error),
	}

	s.lock.Lock()
	s.subscribers = append(s.subscribers, sub)
	s.lock.Unlock()

	err := <-sub.done

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
	s.once.Do(func() {
		s.isClosed.Store(true)
		close(s.channel)
	})
}
