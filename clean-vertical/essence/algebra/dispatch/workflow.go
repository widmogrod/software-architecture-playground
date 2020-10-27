package dispatch

import (
	"container/list"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// DISCLAIMER! This code is an abomination, don't look at it.
// I wrote it as a way to let me understand what I search for in respect to workflow engine, what are challenges far ahead

var _ Program = &Workflow{}

func NewWorkflow() *Workflow {
	return &Workflow{
		program:   NewProgram(),
		partition: &sync.Map{},
		typesreg:  &sync.Map{},
	}
}

type Workflow struct {
	*program
	partition *sync.Map
	typesreg  *sync.Map
}

type state uint8

const (
	Pending = iota
	Processing
	Ok
	Err
)

type work struct {
	activityID    string
	name          string
	inputPayload  []byte
	state         state
	outputType    string
	outputPayload []byte
}

func (w *Workflow) Invoke(ctx Context, cmd interface{}) interface{} {
	activity := &list.List{}
	if store, loaded := w.partition.LoadOrStore(ctx.ActivityID(), activity); loaded {
		activity = store.(*list.List)
	}

	name := reflect.TypeOf(cmd).Name()
	body, err := json.Marshal(cmd)
	if err != nil {
		panic("dispatch: could not marshal cmd body. reason: " + err.Error())
	}

	w.typesreg.Store(name, reflect.TypeOf(cmd))

	activity.PushBack(&work{
		activityID:   ctx.ActivityID(),
		inputPayload: body,
		name:         name,
		state:        Pending,
	})

	return w.Result(ctx.ActivityID())
}

func (w *Workflow) Log() {
	jobQueue := make(chan *work)
	defer close(jobQueue)

	go func() {
		for {
			select {
			case work := <-jobQueue:
				work.state = Processing

				name := work.name
				// execute
				if h, ok := w.handlers.Load(name); ok {
					cmd, _ := w.typesreg.Load(name)
					cmdt := cmd.(reflect.Type)

					newcmd := reflect.New(cmdt)
					newcmd2 := newcmd.Interface()
					err := json.Unmarshal(work.inputPayload, newcmd2)

					if err != nil {
						work.state = Err
						work.outputType = "error"
						work.outputPayload = []byte(err.Error())
						return
					}

					newctx := FromActivityID(work.activityID)

					result := reflect.ValueOf(h).Call([]reflect.Value{
						reflect.ValueOf(newctx),
						newcmd.Elem(),
					})[0].Interface()

					//name := reflect.TypeOf(result).Name()
					body, err := json.Marshal(result)
					if err != nil {
						work.state = Err
						work.outputType = "error"
						work.outputPayload = []byte(err.Error())
						return
					}

					restyp := reflect.TypeOf(result)
					name := restyp.Name()

					w.typesreg.Store(name, restyp)

					work.state = Ok
					work.outputType = name
					work.outputPayload = body
					return
				}

				work.state = Err
				work.outputType = "error"
				work.outputPayload = []byte("dispatch: No handler for a inputPayload of a type = " + name)
			}
		}
	}()

	for {
		w.partition.Range(func(activityID, value interface{}) bool {
			activity := value.(*list.List)
			for e := activity.Front(); e != nil; e = e.Next() {
				// do something with e.Value
				work := e.Value.(*work)
				w.log(activityID.(string), work)

				switch work.state {
				case Pending:
					jobQueue <- work
				}
			}

			return true
		})
		time.Sleep(2 * time.Second)
	}
}

func (w *Workflow) log(activityID string, work *work) {
	switch work.state {
	case Pending:
		fmt.Printf("[pending]    %s: %s(%s) -> ??? \n", activityID, work.name, work.inputPayload)
	case Processing:
		fmt.Printf("[processing] %s: %s(%s) -> ...) \n", activityID, work.name, work.inputPayload)
	case Ok:
		fmt.Printf("[ok]         %s: %s(%s) -> Ok(%s(%s)) \n", activityID, work.name, work.inputPayload, work.outputType, work.outputPayload)
	case Err:
		fmt.Printf("[err]        %s: %s(%s) -> Err(%s(%s)) \n", activityID, work.name, work.inputPayload, work.outputType, work.outputPayload)
	}
}

func (w *Workflow) Result(activityID string) interface{} {
	value, ok := w.partition.Load(activityID)
	if !ok {
		panic("dispatch: Result could not find activity" + activityID)
	}
	activity := value.(*list.List)

	res := make(chan interface{}, 1)
	defer close(res)

	go func() {
		for {
			for e := activity.Front(); e != nil; e = e.Next() {
				// do something with e.Value
				work := e.Value.(*work)
				//w.log(activityID, work)

				switch work.state {
				case Ok:
					output, _ := w.typesreg.Load(work.outputType)
					outputt := output.(reflect.Type)

					newcmd := reflect.New(outputt)
					newcmd2 := newcmd.Interface()
					err := json.Unmarshal(work.outputPayload, newcmd2)
					if err != nil {
						panic("dispatch: Result could not be unmarshal. reason: " + err.Error())
					}

					res <- newcmd.Elem().Interface()
					return
				case Err:
					panic("dispatch: Result could failed. reason: " + string(work.outputPayload))
				}
			}
		}
	}()

	return <-res
}
