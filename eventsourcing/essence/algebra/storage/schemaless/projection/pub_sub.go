package projection

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
)

func NewPubSub[T comparable]() *PubSub[T] {
	lock := sync.RWMutex{}
	return &PubSub[T]{
		lock:      &lock,
		cond:      sync.NewCond(lock.RLocker()),
		publisher: make(map[T]*list.List),
		finished:  make(map[T]bool),
	}
}

type PubSub[T comparable] struct {
	lock      *sync.RWMutex
	cond      *sync.Cond
	publisher map[T]*list.List
	finished  map[T]bool
}

var (
	ErrNoPublisher       = errors.New("no publisher")
	ErrFinished          = errors.New("publisher is finished")
	ErrContextDone       = errors.New("context is done")
	ErrHandlerReturnErr  = errors.New("handler returned error")
	ErrPublishWithOffset = errors.New("cannot publish message with offset")
)

func (p *PubSub[T]) Register(key T) error {
	//log.Errorf("pubsub.Register(%s)\n", GetCtx(any(key).(Node)).name)
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.finished[key]; ok {
		return fmt.Errorf("pubsub.Register: key=%#v %w", key, ErrFinished)
	}

	if _, ok := p.publisher[key]; !ok {
		p.publisher[key] = list.New()
	} else {
		//log.Errorf("pubsub.Register(%s) ALREADY\n", GetCtx(any(key).(Node)).name)
	}
	p.cond.Broadcast()

	return nil
}

// Publish should return error, and not throw panic
// this is a temporary solution, for prototyping
func (p *PubSub[T]) Publish(ctx context.Context, key T, msg Message) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("pubsub.Publish: key=%#v ctx=%s %w", key, ctx.Err(), ErrContextDone)
	default:
		// continue
	}

	//log.Errorf("pubsub.Publish(%s)\n", GetCtx(any(key).(Node)).name)
	if msg.Offset != 0 {
		return fmt.Errorf("pubsub.Publish: key=%#v %w", key, ErrPublishWithOffset)
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.finished[key]; ok {
		return fmt.Errorf("pubsub.Publish: key=%#v %w", key, ErrFinished)
	}

	if _, ok := p.publisher[key]; !ok {
		p.publisher[key] = list.New()
	}

	msg.Offset = p.publisher[key].Len()
	p.publisher[key].PushBack(msg)
	p.cond.Broadcast()
	return nil
}

// Finish is called when a node won't publish any more messages
func (p *PubSub[T]) Finish(key T) {
	//log.Errorf("pubsub.Finish(%s)\n", GetCtx(any(key).(Node)).name)
	p.lock.Lock()
	p.finished[key] = true
	p.lock.Unlock()

	p.cond.Broadcast()
}

//TODO: refactor PubSub and Kinesis to share as much as they can!

func (p *PubSub[T]) Subscribe(ctx context.Context, node T, fromOffset int, f func(Message) error) error {
	//log.Errorf("pubsub.Subscribe(%s, %d)\n", GetCtx(any(node).(Node)).name, fromOffset)
	//p.cond.L.Lock()
	//for _, ok := p.publisher[node]; !ok; {
	//	p.cond.Wait()
	//}
	//appendLog := p.publisher[node]
	//p.cond.L.Unlock()

	p.lock.RLock()
	appendLog, ok := p.publisher[node]
	if !ok {
		p.lock.RUnlock()
		return ErrNoPublisher
	}
	p.lock.RUnlock()

	var prev *list.Element = nil

	// Until, there is no messages, wait
	p.cond.L.Lock()
	for appendLog.Len() == 0 && !p.finished[node] {
		p.cond.Wait()
	}
	if appendLog.Len() == 0 && p.finished[node] {
		p.cond.L.Unlock()
		return nil
	}

	// Select the offset to start reading messages from
	switch fromOffset {
	case 0:
		prev = appendLog.Front()
	case -1:
		prev = appendLog.Back()
	default:
		for e := appendLog.Front(); e != nil; e = e.Next() {
			prev = e
			if e.Value.(Message).Offset == fromOffset {
				break
			}
		}

		if prev == appendLog.Back() {
			p.cond.L.Unlock()
			return errors.New("offset not found")
		}
	}
	p.cond.L.Unlock()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("pubsub.Subscribe %s %w", ctx.Err(), ErrContextDone)

		default:
			msg := prev.Value.(Message)

			//log.Errorf("pubsub.Subscribe CALL (%s)\n", GetCtx(any(node).(Node)).name)
			err := f(msg)
			if err != nil {
				return fmt.Errorf("pubsub.Subscribe %s %w", err, ErrHandlerReturnErr)
			}

			// Wait for new changes to be available
			p.cond.L.Lock()
			for prev.Next() == nil && !p.finished[node] {
				p.cond.Wait()
			}
			if prev.Next() == nil && p.finished[node] {
				p.cond.L.Unlock()
				//log.Errorf("pubsub.Subscribe END(%s)\n", GetCtx(any(node).(Node)).name)
				return nil
			}

			prev = prev.Next()
			p.cond.L.Unlock()
		}
	}
}
