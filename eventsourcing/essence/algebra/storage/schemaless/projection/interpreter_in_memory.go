package schemaless

import (
	"fmt"
	"sync"
)

func NewInMemoryInterpreter() *InMemoryInterpreter {
	return &InMemoryInterpreter{
		channels: make(map[DAG]chan Message),
		errors:   make(map[DAG]error),
		handlers: make(map[DAG]map[string]Handler),
		byKeys:   make(map[DAG]map[string]Message),
	}
}

type InMemoryInterpreter struct {
	lock     sync.Mutex
	channels map[DAG]chan Message
	errors   map[DAG]error
	handlers map[DAG]map[string]Handler
	byKeys   map[DAG]map[string]Message
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
						if err := x.OnMap.Process(msg, i.returning(x)); err != nil {
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
							//fmt.Printf("Merge: recieved prev=%s msg=%s \n", i.keyFromMessage(prev), i.keyFromMessage(msg))
							if err := x.OnMerge.Process2(prev, msg, i.returning(x)); err != nil {
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
				if err := x.OnLoad.Process(&Combine{}, i.returning(x)); err != nil {
					i.recordError(x, err)
					return
				}
			}()

			return nil
		},
	)
}

func (i *InMemoryInterpreter) returning(x DAG) func(Message) {
	return func(msg Message) {
		switch x.(type) {
		case *Merge:
			//switch z := msg.(type) {
			//case *Combine:
			//	b, _ := schema.ToJSON(z.Data)
			//	fmt.Printf("Merge: returning %s %s \n", i.keyFromMessage(msg), string(b))
			//
			//}
			//delete(i.byKeys[y], i.keyFromMessage(msg))
			i.byKeys[x][i.keyFromMessage(msg)] = msg
		}
		i.channelForNode(x) <- msg
	}
}

func (i *InMemoryInterpreter) channelForNode(x DAG) chan Message {
	i.lock.Lock()
	defer i.lock.Unlock()
	if _, ok := i.channels[x]; !ok {
		i.channels[x] = make(chan Message)
	}
	return i.channels[x]
}

func (i *InMemoryInterpreter) recordError(x DAG, err error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	fmt.Printf("element %v error %s", x, err)
	i.errors[x] = err
}

func (i *InMemoryInterpreter) keyFromMessage(msg Message) string {
	return MustMatchMessage(
		msg,
		func(x *Combine) string {
			return x.Key
		},
		func(x *Retract) string {
			return x.Key
		},
		func(x *Both) string {
			return x.Key
		},
	)
}

func (i *InMemoryInterpreter) byKey(x *Merge, key string) (Message, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()
	if _, ok := i.byKeys[x]; !ok {
		i.byKeys[x] = make(map[string]Message)
	}
	if _, ok := i.byKeys[x][key]; !ok {
		return nil, false
	}
	return i.byKeys[x][key], true
}
