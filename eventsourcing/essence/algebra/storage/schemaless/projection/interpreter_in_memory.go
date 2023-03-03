package schemaless

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"sync"
)

func NewInMemoryInterpreter() *InMemoryInterpreter {
	return &InMemoryInterpreter{
		channels: make(map[DAG]chan Item),
		errors:   make(map[DAG]error),
		handlers: make(map[DAG]map[string]Handler),
		byKeys:   make(map[DAG]map[string]Item),
	}
}

type InMemoryInterpreter struct {
	lock     sync.Mutex
	channels map[DAG]chan Item
	errors   map[DAG]error
	handlers map[DAG]map[string]Handler
	byKeys   map[DAG]map[string]Item
}

func (i *InMemoryInterpreter) Run(dag DAG) error {
	if dag == nil {
		return nil
	}

	return MustMatchDAG(
		dag,
		func(x *Map) error {
			go func() {
				//fmt.Printf("Map: gorutine starting %T\n", x)
				for {
					select {
					case msg := <-i.channelForNode(x.Input):
						//fmt.Printf("Map: recieved %T msg=%v\n", x, msg)
						if err := i.callProcess(x.OnMap, msg, i.returning(x)); err != nil {
							i.recordError(x, err)
							return
						}
						//case <-time.After(1 * time.Second):
						//	fmt.Printf("Map: timeout for %T\n", x)
					}
				}
			}()
			return i.Run(x.Input)
		},
		func(x *Merge) error {
			go func() {
				for {
					select {
					case msg := <-i.channelForNode(x.Input[0]):
						// should a merge publish new message
						// what if merge from previous, should it recall previous mesages?
						// in a way it's liek a reduce, so past state should be kept somewhere
						prev, ok := i.byKey(x, i.keyFromMessage(msg))
						if ok {
							merge := Item{
								Key:  msg.Key,
								Data: schema.MkList(prev.Data, msg.Data),
							}
							//fmt.Printf("Merge: recieved prev=%s msg=%s \n", i.keyFromMessage(prev), i.keyFromMessage(msg))
							if err := i.callProcess(x.OnMerge, merge, i.returning(x)); err != nil {
								i.recordError(x, err)
								return
							}
						} else {
							i.returning(x)(msg)
						}
					}
				}
			}()
			return i.Run(x.Input[0])
		},
		func(x *Load) error {
			go func() {
				//fmt.Printf("Load: gorutine starting %T\n", x)
				if err := i.callProcess(x.OnLoad, Item{}, i.returning(x)); err != nil {
					i.recordError(x, err)
					return
				}
			}()

			return nil
		},
	)
}

func (i *InMemoryInterpreter) callProcess(handler Handler, x Item, returning func(Item)) error {
	return handler.Process(x, returning)
}

func (i *InMemoryInterpreter) returning(x DAG) func(Item) {
	return func(msg Item) {
		switch x.(type) {
		case *Merge:
			i.byKeys[x][i.keyFromMessage(msg)] = msg
		}
		i.channelForNode(x) <- msg
	}
}

func (i *InMemoryInterpreter) channelForNode(x DAG) chan Item {
	i.lock.Lock()
	defer i.lock.Unlock()
	if _, ok := i.channels[x]; !ok {
		i.channels[x] = make(chan Item)
	}
	return i.channels[x]
}

func (i *InMemoryInterpreter) recordError(x DAG, err error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	fmt.Printf("element %v error %s", x, err)
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
