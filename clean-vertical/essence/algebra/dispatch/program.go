package dispatch

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func NewProgram() *program {
	return &program{
		handlers: &sync.Map{},
	}
}

type program struct {
	handlers *sync.Map
}

func (p *program) Invoke(ctx Context, cmd interface{}) interface{} {
	name := reflect.TypeOf(cmd).Name()
	if h, ok := p.handlers.Load(name); ok {
		return reflect.ValueOf(h).Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(cmd),
		})[0].Interface()
	}

	return errors.New("dispatch: No handler for a cmd of a type = " + name)
}

func (p *program) Interpretation(class interface{}) {
	r := reflect.TypeOf(class)
	for i := 0; i < r.NumMethod(); i++ {
		met := r.Method(i)
		name := met.Type.In(2).Name()
		handler := func(ctx Context, input interface{}) interface{} {
			return met.Func.Call([]reflect.Value{
				reflect.ValueOf(class),
				reflect.ValueOf(ctx),
				reflect.ValueOf(input),
			})[0].Interface()
		}

		p.handlers.Store(name, handler)
	}
}

func (p *program) RegisterGlobalHandler(handler interface{}) {
	name := reflect.TypeOf(handler).In(1).Name()
	p.handlers.Store(name, handler)
}

func (p *program) ShouldInvokeAndReturn(t *testing.T, v interface{}) {
	name := reflect.TypeOf(v).In(2).Name()
	handler := func(ctx Context, input interface{}) interface{} {
		res := reflect.ValueOf(v).Call([]reflect.Value{
			reflect.ValueOf(t),
			reflect.ValueOf(ctx),
			reflect.ValueOf(input),
		})

		if len(res) > 0 {
			return res[0].Interface()
		}

		return nil
	}

	p.handlers.Store(name, handler)
}
