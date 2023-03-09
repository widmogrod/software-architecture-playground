package schemaless

import (
	"container/list"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"sync"
	"time"
)

func DefaultInMemoryInterpreter() *InMemoryInterpreter {
	return &InMemoryInterpreter{
		pubsub:  NewPubSub(),
		byKeys:  make(map[Node]map[string]Item),
		running: make(map[Node]struct{}),
	}
}

type InMemoryInterpreter struct {
	lock     sync.Mutex
	pubsub   *PubSub
	byKeys   map[Node]map[string]Item
	running  map[Node]struct{}
	finished sync.WaitGroup
	// what difference between process time and event time
	// should answers question
	// - are there any events in the system, that a process should wait?
	watermark time.Time
}

func (i *InMemoryInterpreter) Run(nodes []Node) error {
	for _, node := range nodes {
		if err := i.run(node); err != nil {
			return err
		}
	}

	return nil
}

func (i *InMemoryInterpreter) run(dag Node) error {
	if dag == nil {
		return nil
	}

	if _, ok := i.running[dag]; ok {
		return nil
	}
	i.running[dag] = struct{}{}

	return MustMatchNode(
		dag,
		func(x *Map) error {
			i.finished.Add(1)
			go func() {
				defer i.finished.Done()

				var lastOffset int = 0

				for {
					msg, err := i.pubsub.Subscribe(x.Input, lastOffset)
					if i.shouldClose(err) {
						log.Debugln("Map: close", i.str(x))
						i.pubsub.Finish(x)
						return
					} else if i.shouldProcess(err) {
						lastOffset = msg.Offset
						log.Debugln("Map: ", i.str(x), msg.Aggregate != nil, msg.Retract != nil)
						switch true {
						case msg.Aggregate != nil && msg.Retract == nil,
							msg.Aggregate != nil && msg.Retract != nil && !x.Ctx.ShouldRetract():

							err := x.OnMap.Process(*msg.Aggregate, func(item Item) {
								i.pubsub.Publish(x, Message{
									Key:       item.Key,
									Aggregate: &item,
								})
							})
							if err != nil {
								panic(err)
							}

						case msg.Aggregate != nil && msg.Retract != nil && x.Ctx.ShouldRetract():
							buff := NewDual()
							err := x.OnMap.Process(*msg.Aggregate, buff.ReturningAggregate)
							if err != nil {
								panic(err)
							}
							err = x.OnMap.Retract(*msg.Retract, buff.ReturningRetract)
							if err != nil {
								panic(err)
							}

							if !buff.IsValid() {
								panic("Map(1); asymmetry " + i.str(x))
							}

							for _, msg := range buff.List() {
								i.pubsub.Publish(x, *msg)
							}

						case msg.Aggregate == nil && msg.Retract != nil && x.Ctx.ShouldRetract():
							err := x.OnMap.Retract(*msg.Retract, func(item Item) {
								i.pubsub.Publish(x, Message{
									Key:     item.Key,
									Retract: &item,
								})
							})
							if err != nil {
								panic(err)
							}

						case msg.Aggregate == nil && msg.Retract != nil && !x.Ctx.ShouldRetract():
							log.Debugln("ignored retraction", i.str(x))

						default:
							panic("not implemented Map(3); " + i.str(x) + " " + ToStrMessage(msg))
						}

						log.Debugln("√", i.str(x))
					} else {
						<-time.After(10 * time.Millisecond)
					}
				}
			}()
			return i.run(x.Input)
		},
		func(x *Merge) error {
			i.finished.Add(1)
			go func() {
				defer i.finished.Done()

				var lastOffset int = 0
				prev := make(map[string]*Item)

				for {
					msg, err := i.pubsub.Subscribe(x.Input, lastOffset)
					if i.shouldClose(err) {
						log.Debugln("Merge: close", i.str(x))
						i.pubsub.Finish(x)
						return
					} else if i.shouldProcess(err) {
						lastOffset = msg.Offset

						if msg.Retract == nil && msg.Aggregate == nil {
							panic("message has not Aggretate nor Retract. not implemented (1)")
						}

						log.Debugln("Merge 👯: ", i.str(x), msg.Aggregate != nil, msg.Retract != nil)

						if _, ok := prev[msg.Key]; ok {
							base := prev[msg.Key]

							if msg.Retract != nil && x.Ctx.ShouldRetract() {
								log.Debugln("❌retracting in merge", i.str(x))
								retract := Item{
									Key:  msg.Key,
									Data: schema.MkList(base.Data, msg.Retract.Data),
								}

								if err := x.OnMerge.Retract(retract, func(item Item) {
									base = &item
									i.pubsub.Publish(x, Message{
										Key:     msg.Key,
										Retract: &item,
									})
								}); err != nil {
									panic(err)
								}
							}

							if msg.Aggregate != nil {
								log.Debugln("✅aggregate in merge", i.str(x))
								merge := Item{
									Key:  msg.Key,
									Data: schema.MkList(base.Data, msg.Aggregate.Data),
								}
								err := x.OnMerge.Process(merge, func(item Item) {
									p := base
									base = &item
									i.pubsub.Publish(x, Message{
										Key:       msg.Key,
										Aggregate: &item,
										Retract:   p,
									})
								})
								if err != nil {
									panic(err)
								}
							}

							prev[msg.Key] = base

						} else {
							if msg.Retract != nil {
								panic("no previous state, and requesing retracting. not implemented (2)" + ToStrMessage(msg))
							}

							prev[msg.Key] = msg.Aggregate
							i.pubsub.Publish(x, Message{
								Key:       msg.Key,
								Aggregate: msg.Aggregate,
							})
						}
					} else {
						// wait
						<-time.After(10 * time.Millisecond)
					}
				}
			}()
			return i.run(x.Input)
		},
		func(x *Load) error {
			i.finished.Add(1)
			go func() {
				defer i.finished.Done()

				if err := x.OnLoad.Process(Item{}, func(item Item) {
					i.pubsub.Publish(x, Message{
						Key:       item.Key,
						Aggregate: &item,
						Retract:   nil,
					})
				}); err != nil {
					panic(err)
				}
				log.Debugln("Load: finish", i.str(x))
				i.pubsub.Finish(x)
			}()

			return nil
		},
		func(x *Join) error {
			i.finished.Add(1)
			go func() {
				defer i.finished.Done()

				lastOffset := make([]int, len(x.Input))
				for idx, _ := range x.Input {
					lastOffset[idx] = 0
				}

				for {
					for idx, y := range x.Input {
						msg, err := i.pubsub.Subscribe(y, lastOffset[idx])
						if i.shouldClose(err) {
							log.Debugln("Joining close", i.str(x), err)
							i.pubsub.Finish(x)
							return
						} else if i.shouldProcess(err) {
							log.Debugln("Joining loop published", i.str(x), ToStrMessage(msg))
							lastOffset[idx] = msg.Offset
							// join streams and publish
							i.pubsub.Publish(x, Message{
								Key:       msg.Key,
								Aggregate: msg.Aggregate,
								Retract:   msg.Retract,
							})
						} else {
							// wait for next message
							time.Sleep(100 * time.Millisecond)
						}
					}
				}
			}()

			return nil
		},
	)
}

func (i *InMemoryInterpreter) WaitForDone() {
	i.finished.Wait()
}

func ToStrMessage(msg Message) string {
	return fmt.Sprintf("Message{Key: %s, Retract: %s, Aggregate: %s}",
		msg.Key,
		ToStrItem(msg.Retract),
		ToStrItem(msg.Aggregate))
}

func ToStrItem(item *Item) string {
	if item == nil {
		return "nil"
	}
	bytes, err := schema.ToJSON(item.Data)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("Item{Key: %s, Data: %s}",
		item.Key, string(bytes))
}

func (i *InMemoryInterpreter) str(x Node) string {
	return ToStr(x)
}

func ToStr(x Node) string {
	return MustMatchNode(
		x,
		func(x *Map) string {
			return fmt.Sprintf("map(%s, r=%v)", x.Ctx.Name(), x.Ctx.ShouldRetract())
		},
		func(x *Merge) string {
			return fmt.Sprintf("merge(%s, r=%v)", x.Ctx.Name(), x.Ctx.ShouldRetract())
		},
		func(x *Load) string {
			return fmt.Sprintf("load(%s, r=%v)", x.Ctx.Name(), x.Ctx.ShouldRetract())
		},
		func(x *Join) string {
			return fmt.Sprintf("join(%s, r=%v)", x.Ctx.Name(), x.Ctx.ShouldRetract())
		},
	)
}

func (i *InMemoryInterpreter) keyFromMessage(msg Item) string {
	return msg.Key
}

func (i *InMemoryInterpreter) byKey(x *Merge, key string) (Item, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if _, ok := i.byKeys[x]; !ok {
		i.byKeys[x] = make(map[string]Item)
	}
	if _, ok := i.byKeys[x][key]; !ok {
		return Item{}, false
	}
	return i.byKeys[x][key], true
}

func (i *InMemoryInterpreter) shouldClose(err error) bool {
	return errors.Is(err, ErrFinished)
}

func (i *InMemoryInterpreter) shouldProcess(err error) bool {
	return err == nil
}

func NewPubSub() *PubSub {
	return &PubSub{
		publisher: make(map[Node]*list.List),
		finished:  make(map[Node]bool),
	}
}

type PubSub struct {
	lock      sync.Mutex
	publisher map[Node]*list.List
	finished  map[Node]bool
}

var (
	ErrNoPublisher      = errors.New("cannot subscribe, no publisher")
	ErrFinished         = errors.New("cannot subscribe, to finished publisher")
	ErrAboveKnownOffset = errors.New("cannot subscribe, above known offset")
)

func (p *PubSub) Subscribe(to Node, fromOffset int) (Message, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.publisher[to]; !ok {
		return Message{}, ErrNoPublisher
	}

	if p.publisher[to].Len() <= fromOffset {
		if _, ok := p.finished[to]; ok {
			return Message{}, ErrFinished
		}
		return Message{}, ErrAboveKnownOffset
	}

	var i int
	var msg Message
	var found bool
	for e := p.publisher[to].Front(); e != nil; e = e.Next() {
		if i == fromOffset {
			found = true
			msg = e.Value.(Message)
			break
		}
		i++
	}

	if !found {
		panic("offset not found")
	}

	return msg, nil
}

func (p *PubSub) Publish(key Node, msg Message) {
	if msg.Offset != 0 {
		panic("cannot publish message with offset")
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.finished[key]; ok {
		panic("cannot publish to finished node")
	}

	if _, ok := p.publisher[key]; !ok {
		p.publisher[key] = list.New()
	}

	msg.Offset = p.publisher[key].Len() + 1
	p.publisher[key].PushBack(msg)
}

// Finish is called when a node won't publish any more messages
func (p *PubSub) Finish(key Node) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.finished[key] = true
}
