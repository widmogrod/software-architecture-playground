package projection

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"sync"
)

func DefaultInMemoryInterpreter() *InMemoryInterpreter {
	return &InMemoryInterpreter{
		pubsub: NewPubSubMultiChan[Node](),
		//pubsub:  NewPubSub[Node](),
		byKeys:  make(map[Node]map[string]Item),
		running: make(map[Node]struct{}),
		stats:   NewStatsCollector(),
	}
}

type ExecutionStatus int

const (
	ExecutionStatusNew ExecutionStatus = iota
	ExecutionStatusRunning
	ExecutionStatusError
	ExecutionStatusFinished
)

var (
	ErrInterpreterNotInNewState = fmt.Errorf("interpreter is not in new state")
)

type PubSubForInterpreter[T comparable] interface {
	Register(key T) error
	Publish(ctx context.Context, key T, msg Message) error
	Finish(ctx context.Context, key T)
	Subscribe(ctx context.Context, node T, fromOffset int, f func(Message) error) error
}

type InMemoryInterpreter struct {
	lock    sync.Mutex
	pubsub  PubSubForInterpreter[Node]
	byKeys  map[Node]map[string]Item
	running map[Node]struct{}
	status  ExecutionStatus
	// what differences between process time and event time
	// should answers question
	// - are there any events in the system, that a process should wait?
	watermark int64
	stats     StatsCollector
}

func (i *InMemoryInterpreter) Run(ctx context.Context, nodes []Node) error {
	i.lock.Lock()
	if i.status != ExecutionStatusNew {
		i.lock.Unlock()
		return fmt.Errorf("interpreter.Run state %d %w", i.status, ErrInterpreterNotInNewState)
	}
	i.status = ExecutionStatusRunning
	i.lock.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	group := &ExecutionGroup{
		ctx:    ctx,
		cancel: cancel,
	}

	// Registering new nodes makes sure that, in case of non-deterministic concurrency
	// when goroutine want to subscribe to a node, it will be registered, even if it's not publishing yet
	for _, node := range nodes {
		if node == nil {
			continue
		}

		err := i.pubsub.Register(node)
		if err != nil {
			i.lock.Lock()
			i.status = ExecutionStatusError
			i.lock.Unlock()

			return fmt.Errorf("interpreter.Run(1) %w", err)
		}
	}

	for _, node := range nodes {
		if node == nil {
			continue
		}

		func(node Node) {
			group.Go(func() (err error) {
				return i.run(ctx, node)
			})
		}(node)
	}

	if err := group.Wait(); err != nil {
		i.lock.Lock()
		i.status = ExecutionStatusError
		i.lock.Unlock()

		return fmt.Errorf("interpreter.Run(2) %w", err)
	}

	i.lock.Lock()
	i.status = ExecutionStatusFinished
	i.lock.Unlock()

	return nil
}

func (i *InMemoryInterpreter) run(ctx context.Context, dag Node) error {
	if dag == nil {
		//panic("fix nodes that are nil! fix dag builder!")
		return nil
	}

	return MustMatchNode(
		dag,
		func(x *Map) error {
			var lastOffset int = 0

			err := i.pubsub.Subscribe(
				ctx,
				x.Input,
				lastOffset,
				func(msg Message) error {
					lastOffset = msg.Offset
					log.Debugln("Map: ", i.str(x), msg.Aggregate != nil, msg.Retract != nil)
					log.Debugf("✉️: %+v %s\n", msg, i.str(x))
					switch true {
					case msg.Aggregate != nil && msg.Retract == nil,
						msg.Aggregate != nil && msg.Retract != nil && !x.Ctx.ShouldRetract():

						err := x.OnMap.Process(*msg.Aggregate, func(item Item) {
							i.stats.Incr(fmt.Sprintf("map[%s].returning.aggregate", x.Ctx.Name()), 1)

							err := i.pubsub.Publish(ctx, x, Message{
								Key:       item.Key,
								Aggregate: &item,
							})
							if err != nil {
								panic(err)
							}
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
							i.stats.Incr(fmt.Sprintf("map[%s].returning.aggregate", x.Ctx.Name()), 1)
							i.stats.Incr(fmt.Sprintf("map[%s].returning.retract", x.Ctx.Name()), 1)

							err := i.pubsub.Publish(ctx, x, *msg)
							if err != nil {
								panic(err)
							}
						}

					case msg.Aggregate == nil && msg.Retract != nil && x.Ctx.ShouldRetract():
						err := x.OnMap.Retract(*msg.Retract, func(item Item) {

							i.stats.Incr(fmt.Sprintf("map[%s].returning.aggregate", x.Ctx.Name()), 1)

							err := i.pubsub.Publish(ctx, x, Message{
								Key:     item.Key,
								Retract: &item,
							})
							if err != nil {
								panic(err)
							}
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

					return nil
				},
			)
			if err != nil {
				return fmt.Errorf("interpreter.Map(1) %w", err)
			}

			log.Debugln("Map: Finish", i.str(x))
			i.pubsub.Finish(ctx, x)

			return nil
		},
		func(x *Merge) error {
			var lastOffset int = 0
			prev := make(map[string]*Item)

			err := i.pubsub.Subscribe(
				ctx,
				x.Input,
				lastOffset,
				func(msg Message) error {
					lastOffset = msg.Offset

					if msg.Retract == nil && msg.Aggregate == nil {
						panic("message has not Aggretate nor Retract. not implemented (1)")
					}

					log.Debugln("Merge 👯: ", i.str(x), msg.Aggregate != nil, msg.Retract != nil)

					if _, ok := prev[msg.Key]; ok {
						base := prev[msg.Key]

						// TODO: retraction and aggregatoin don't happen in transactional way, even if message has both operations
						// this is a problem, because if retraction fails, then aggregation will be lost
						if msg.Retract != nil && x.Ctx.ShouldRetract() {
							log.Debugln("❌retracting in merge", i.str(x))
							retract := Item{
								Key:  msg.Key,
								Data: schema.MkList(base.Data, msg.Retract.Data),
							}

							if err := x.OnMerge.Retract(retract, func(item Item) {

								i.stats.Incr(fmt.Sprintf("merge[%s].returning.retract", x.Ctx.Name()), 1)

								base = &item
								err := i.pubsub.Publish(ctx, x, Message{
									Key:     msg.Key,
									Retract: &item,
								})
								if err != nil {
									panic(err)
								}
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
								i.stats.Incr(fmt.Sprintf("merge[%s].returning.aggregate", x.Ctx.Name()), 1)

								p := base
								base = &item
								// TODO: In feature, we should make better decision whenever send retractions or not.
								// For now, we always send retractions, they don't have to be treated as retraction by the receiver.
								// But, this has penalty related to throughput, and latency, and for some applications, it is not acceptable.
								err := i.pubsub.Publish(ctx, x, Message{
									Key:       msg.Key,
									Aggregate: &item,
									Retract:   p,
								})
								if err != nil {
									panic(err)
								}
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

						i.stats.Incr(fmt.Sprintf("merge[%s].returning.aggregate", x.Ctx.Name()), 1)

						prev[msg.Key] = msg.Aggregate
						err := i.pubsub.Publish(ctx, x, Message{
							Key:       msg.Key,
							Aggregate: msg.Aggregate,
						})
						if err != nil {
							return fmt.Errorf("interpreter.Merge(1) %w", err)
						}
					}

					return nil
				},
			)
			if err != nil {
				return fmt.Errorf("interpreter.Merge(1) %w", err)
			}

			log.Debugln("Merge: Finish", i.str(x))
			i.pubsub.Finish(ctx, x)

			return nil
		},
		func(x *Load) error {
			var err error
			err = x.OnLoad.Process(Item{}, func(item Item) {
				if err != nil {
					return
				}

				//if item.EventTime == 0 {
				//	item.EventTime = time.Now().UnixNano()
				//}
				//
				//// calculate watermark
				//if item.EventTime > i.watermark {
				//	i.watermark = item.EventTime
				//}

				i.stats.Incr(fmt.Sprintf("load[%s].returning", x.Ctx.Name()), 1)

				err = i.pubsub.Publish(ctx, x, Message{
					Key:       item.Key,
					Aggregate: &item,
					Retract:   nil,
				})
			})

			if err != nil {
				return fmt.Errorf("interpreter.Load(1) %w", err)
			}

			log.Debugln("Load: Finish", i.str(x))
			i.pubsub.Finish(ctx, x)

			return nil
		},
		func(x *Join) error {
			lastOffset := make([]int, len(x.Input))
			for idx, _ := range x.Input {
				lastOffset[idx] = 0
			}

			group := ExecutionGroup{ctx: ctx}

			for idx := range x.Input {
				func(idx int) {
					group.Go(func() error {
						return i.pubsub.Subscribe(
							ctx,
							x.Input[idx],
							lastOffset[idx],
							func(msg Message) error {
								lastOffset[idx] = msg.Offset

								i.stats.Incr(fmt.Sprintf("join[%s].returning", x.Ctx.Name()), 1)

								// join streams and publish
								err := i.pubsub.Publish(ctx, x, Message{
									Key:       msg.Key,
									Aggregate: msg.Aggregate,
									Retract:   msg.Retract,
								})

								if err != nil {
									return fmt.Errorf("interpreter.Join(1) %w", err)
								}

								return nil
							},
						)
					})
				}(idx)
			}

			if err := group.Wait(); err != nil {
				return fmt.Errorf("interpreter.Join(1) %w", err)
			}

			log.Debugln("Join: Finish", i.str(x))
			i.pubsub.Finish(ctx, x)

			return nil
		},
	)
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
			return fmt.Sprintf("Load(%s, r=%v)", x.Ctx.Name(), x.Ctx.ShouldRetract())
		},
		func(x *Join) string {
			return fmt.Sprintf("join(%s, r=%v)", x.Ctx.Name(), x.Ctx.ShouldRetract())
		},
	)
}

func (i *InMemoryInterpreter) StatsSnapshotAndReset() Stats {
	return i.stats.Snapshot()
}

type ExecutionGroup struct {
	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
	err    error
	once   sync.Once
}

func (g *ExecutionGroup) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		select {
		case <-g.ctx.Done():
			if err := g.ctx.Err(); err != nil {
				g.once.Do(func() {
					g.err = err
					if g.cancel != nil {
						g.cancel()
					}
				})
			}

		default:
			err := f()
			if err != nil {
				g.once.Do(func() {
					g.err = err
					if g.cancel != nil {
						g.cancel()
					}
				})
			}
		}
	}()
}

func (g *ExecutionGroup) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return nil
}
