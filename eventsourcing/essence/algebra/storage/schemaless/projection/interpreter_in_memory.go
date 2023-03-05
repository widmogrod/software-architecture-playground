package schemaless

import (
	"container/list"
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"sync"
	"time"
)

func NewInMemoryInterpreter() *InMemoryInterpreter {
	return &InMemoryInterpreter{
		pubsub:  NewPubSub2(),
		errors:  make(map[DAG]error),
		byKeys:  make(map[DAG]map[string]Item),
		running: make(map[DAG]struct{}),
	}
}

type InMemoryInterpreter struct {
	lock    sync.Mutex
	pubsub  *PubSub2
	errors  map[DAG]error
	byKeys  map[DAG]map[string]Item
	running map[DAG]struct{}
}

func (i *InMemoryInterpreter) Run2(nodes []DAG) error {
	for _, node := range nodes {
		if err := i.Run(node); err != nil {
			return err
		}
	}

	return nil
}

func (i *InMemoryInterpreter) Run(dag DAG) error {
	if dag == nil {
		return nil
	}

	if _, ok := i.running[dag]; ok {
		//fmt.Printf("RUNNNIN! %s \n", i.str(dag))
		return nil
	}
	i.running[dag] = struct{}{}

	return MustMatchDAG(
		dag,
		func(x *Map) error {
			//fmt.Println("traverse", i.str(x))
			//ch := i.pubsub.Subscribe(x.Input, x)
			go func() {
				//fmt.Println("ok", i.str(x))
				//defer i.pubsub.Unsubscribe(x)

				var lastOffset int = 0

				for {
					msg, ok := i.pubsub.Subscribe(x.Input, lastOffset)
					if !ok {
						<-time.After(10 * time.Millisecond)
						continue
					}
					lastOffset = msg.Offset
					//select {
					//case msg := <-i.pubsub.Subscribe(x.Input, x)
					fmt.Println("Map: ", i.str(x), msg.Aggregate != nil, msg.Retract != nil)
					switch true {
					case msg.Aggregate != nil && msg.Retract == nil,
						msg.Aggregate != nil && msg.Retract != nil && !x.Name.ShouldRetract():

						err := x.OnMap.Process(*msg.Aggregate, func(item Item) {
							i.pubsub.Publish(x, Message{
								Key:       item.Key,
								Aggregate: &item,
							})
						})
						if err != nil {
							panic(err)
						}

					case msg.Aggregate != nil && msg.Retract != nil && x.Name.ShouldRetract():
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

					case msg.Aggregate == nil && msg.Retract != nil && x.Name.ShouldRetract():
						err := x.OnMap.Retract(*msg.Retract, func(item Item) {
							i.pubsub.Publish(x, Message{
								Key:     item.Key,
								Retract: &item,
							})
						})
						if err != nil {
							panic(err)
						}

					case msg.Aggregate == nil && msg.Retract != nil && !x.Name.ShouldRetract():
						fmt.Println("ignored retraction", i.str(x))

					default:
						panic("not implemented Map(3); " + i.str(x) + " " + ToStrMessage(msg))
					}

					fmt.Println("√", i.str(x))
				}
				//}
			}()
			return i.Run(x.Input)
		},
		func(x *Merge) error {
			//fmt.Println("traverse", i.str(x))
			//ch := i.pubsub.Subscribe(x.Input, x)
			go func() {
				//fmt.Println("ok", i.str(x))
				//defer i.pubsub.Unsubscribe(x)

				var lastOffset int = 0
				prev := make(map[string]*Item)

				for {
					msg, ok := i.pubsub.Subscribe(x.Input, lastOffset)
					if !ok {
						<-time.After(10 * time.Millisecond)
						continue
					}
					lastOffset = msg.Offset

					if msg.Retract == nil && msg.Aggregate == nil {
						panic("message has not Aggretate nor Retract. not implemented (1)")
					}

					fmt.Println("Merge 👯: ", i.str(x), msg.Aggregate != nil, msg.Retract != nil)
					if _, ok := prev[msg.Key]; ok {
						base := prev[msg.Key]

						if msg.Retract != nil && x.Name.ShouldRetract() {
							fmt.Println("❌retracting in merge", i.str(x))
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
							fmt.Println("✅aggregate in merge", i.str(x))
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
				}
			}()
			return i.Run(x.Input)
		},
		func(x *Load) error {
			//fmt.Println("traverse", i.str(x))
			go func() {
				//fmt.Println("ok", i.str(x))
				if err := x.OnLoad.Process(Item{}, func(item Item) {
					i.pubsub.Publish(x, Message{
						Key:       item.Key,
						Aggregate: &item,
						Retract:   nil,
					})
				}); err != nil {
					panic(err)
				}
			}()

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

//func (i *InMemoryInterpreter) WaitUntilFinished() {
//	//<-i.pubsub.Drained()
//	//<-time.After(500 * time.Millisecond)
//}

func (i *InMemoryInterpreter) str(x DAG) string {
	return ToStr(x)
}

func ToStr(x DAG) string {
	return MustMatchDAG(
		x,
		func(x *Map) string {
			return fmt.Sprintf("map(%s, r=%v)", x.Name.Name(), x.Name.ShouldRetract())
		},
		func(x *Merge) string {
			return fmt.Sprintf("merge(%s, r=%v)", x.Name.Name(), x.Name.ShouldRetract())
		},
		func(x *Load) string {
			return fmt.Sprintf("load(%s, r=%v)", x.Name.Name(), x.Name.ShouldRetract())
		},
	)
}

func (i *InMemoryInterpreter) recordError(x DAG, err error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	fmt.Printf("ERR: element %v error %s", x, err.Error())
	i.errors[x] = err
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

func NewPubSub2() *PubSub2 {
	return &PubSub2{
		publisher: make(map[DAG]*list.List),
	}
}

type PubSub2 struct {
	lock      sync.Mutex
	publisher map[DAG]*list.List
}

func (p *PubSub2) Subscribe(to DAG, fromOffset int) (Message, bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.publisher[to]; !ok {
		return Message{}, false
	}

	if p.publisher[to].Len() <= fromOffset {
		return Message{}, false
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

	return msg, true
}

func (p *PubSub2) Publish(key DAG, msg Message) {
	if msg.Offset != 0 {
		panic("cannot publish message with offset")
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.publisher[key]; !ok {
		p.publisher[key] = list.New()
	}

	msg.Offset = p.publisher[key].Len() + 1
	p.publisher[key].PushBack(msg)
}
