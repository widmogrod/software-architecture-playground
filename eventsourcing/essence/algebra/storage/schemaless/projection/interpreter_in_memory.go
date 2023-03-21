package projection

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"sync"
	"time"
)

func DefaultInMemoryInterpreter() *InMemoryInterpreter {
	return &InMemoryInterpreter{
		pubsub:  NewPubSub[Node](),
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

type InMemoryInterpreter struct {
	lock     sync.Mutex
	pubsub   *PubSub[Node]
	byKeys   map[Node]map[string]Item
	running  map[Node]struct{}
	finished sync.WaitGroup
	status   ExecutionStatus
	// what difference between process time and event time
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

		if err := i.run(ctx, node); err != nil {
			i.lock.Lock()
			i.status = ExecutionStatusError
			i.lock.Unlock()

			return fmt.Errorf("interpreter.Run(2) %w", err)
		}
	}

	i.waitForDone()
	i.lock.Lock()
	i.status = ExecutionStatusFinished
	i.lock.Unlock()

	return nil
}

func (i *InMemoryInterpreter) run(ctx context.Context, dag Node) error {
	// this is because a dag builder when use WithName creates empty node
	// fix this!
	if dag == nil {
		//panic("fix nodes that are nil! fix dag builder!")
		return nil
	}

	//if _, ok := i.running[dag]; ok {
	//	return nil
	//}
	//i.running[dag] = struct{}{}

	err := i.pubsub.Register(dag)
	if err != nil {
		panic(err)
	}

	return MustMatchNode(
		dag,
		func(x *Map) error {
			i.finished.Add(1)
			go func() {
				// continue
				defer i.finished.Done()

				select {
				case <-ctx.Done():
					return
				default:
				}

				var lastOffset int = 0

				err := i.pubsub.Subscribe(
					ctx,
					x.Input,
					lastOffset,
					func(msg Message) {
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
					},
				)
				if err != nil {
					panic(err)
				}

				log.Debugln("Map: Finish", i.str(x))
				i.pubsub.Finish(x)
			}()

			//return i.run(x.Input)
			return nil
		},
		func(x *Merge) error {
			i.finished.Add(1)
			go func() {
				// continue
				defer i.finished.Done()

				select {
				case <-ctx.Done():
					return
				default:
				}

				var lastOffset int = 0
				prev := make(map[string]*Item)

				err := i.pubsub.Subscribe(
					ctx,
					x.Input,
					lastOffset,
					func(msg Message) {
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
								panic(err)
							}
						}
					},
				)
				if err != nil {
					panic(err)
				}

				log.Debugln("Merge: Finish", i.str(x))
				i.pubsub.Finish(x)
			}()

			//return i.run(x.Input)
			return nil
		},
		func(x *Load) error {
			i.finished.Add(1)
			go func() {
				defer i.finished.Done()

				select {
				case <-ctx.Done():
					return
				default:
				}

				err := x.OnLoad.Process(Item{}, func(item Item) {
					if item.EventTime == 0 {
						item.EventTime = time.Now().UnixNano()
					}

					// calculate watermark
					if item.EventTime > i.watermark {
						i.watermark = item.EventTime
					}

					i.stats.Incr(fmt.Sprintf("load[%s].returning", x.Ctx.Name()), 1)

					err := i.pubsub.Publish(ctx, x, Message{
						Key:       item.Key,
						Aggregate: &item,
						Retract:   nil,
					})
					if err != nil {
						panic(err)
					}
				})

				if err != nil {
					panic(err)
				}

				log.Debugln("Load: Finish", i.str(x))
				i.pubsub.Finish(x)
			}()

			return nil
		},
		func(x *Join) error {
			i.finished.Add(1)
			go func() {
				defer i.finished.Done()

				select {
				case <-ctx.Done():
					return
				default:
					// continue
				}

				lastOffset := make([]int, len(x.Input))
				for idx, _ := range x.Input {
					lastOffset[idx] = 0
				}

				wg := sync.WaitGroup{}
				for idx, y := range x.Input {
					wg.Add(1)

					go func(idx int, y Node) {
						defer wg.Done()

						err := i.pubsub.Subscribe(
							ctx,
							y,
							lastOffset[idx],
							func(msg Message) {
								lastOffset[idx] = msg.Offset

								i.stats.Incr(fmt.Sprintf("join[%s].returning", x.Ctx.Name()), 1)

								// join streams and publish
								err := i.pubsub.Publish(ctx, x, Message{
									Key:       msg.Key,
									Aggregate: msg.Aggregate,
									Retract:   msg.Retract,
								})
								if err != nil {
									panic(err)
								}
							},
						)

						if err != nil {
							panic(err)
						}
					}(idx, y)
				}
				wg.Wait()

				log.Debugln("Join: Finish", i.str(x))
				i.pubsub.Finish(x)
			}()

			return nil
		},
	)
}

func (i *InMemoryInterpreter) waitForDone() {
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
