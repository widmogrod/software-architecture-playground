package schemaless

import "sync"

func NewInMemoryInterpreter() *InMemoryInterpreter {
	return &InMemoryInterpreter{
		channels: make(map[DAG]chan Message),
		errors:   make(map[DAG]error),
	}
}

type InMemoryInterpreter struct {
	lock     sync.Mutex
	channels map[DAG]chan Message
	errors   map[DAG]error
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
						if err := x.OnMerge.Process(msg, i.returning(x)); err != nil {
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

func (i *InMemoryInterpreter) returning(x DAG) func(Message) error {
	return func(msg Message) error {
		i.channelForNode(x) <- msg
		return nil
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
