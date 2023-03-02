package schemaless

import (
	"sync"
)

func NewInMemoryInterpreter() *InMemoryInterpreter {
	return &InMemoryInterpreter{
		channels: make(map[DAG]chan Message),
		errors:   make(map[DAG]error),
		handlers: make(map[DAG]map[string]Handler),
	}
}

type InMemoryInterpreter struct {
	lock     sync.Mutex
	channels map[DAG]chan Message
	errors   map[DAG]error
	handlers map[DAG]map[string]Handler
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
				//fmt.Printf("Merge: gorutine starting %T\n", x)
				for {
					select {
					case msg := <-i.channelForNode(x.Input[0]):
						//fmt.Printf("Merge: recieved %T msg=%v\n", x, msg)
						if err := i.handerByTypeAndKey(x, msg).Process(msg, i.returning(x)); err != nil {
							i.recordError(x, err)
							return
						}
						//case <-time.After(1 * time.Second):
						//	fmt.Printf("Merge: timeout for %T\n", x)
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
	//fmt.Printf("element %v error %s", x, err)
	i.errors[x] = err
}

func (i *InMemoryInterpreter) handerByTypeAndKey(x DAG, msg Message) Handler {
	key := i.keyFromMessage(msg)
	if _, ok := i.handlers[x]; !ok {
		i.handlers[x] = make(map[string]Handler)
	}
	if _, ok := i.handlers[x][key]; !ok {
		i.handlers[x][key] = MustMatchDAG(
			x,
			func(x *Map) Handler {
				h := x.OnMap
				return h
			},
			func(x *Merge) Handler {
				// TODO: to figure out how to make merge handler don't inhetit previous state
				// ????? This is bug for failing test

				return x.OnMerge()
			},
			func(x *Load) Handler {
				h := x.OnLoad
				return h
			},
		)
	}

	return i.handlers[x][key]
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
