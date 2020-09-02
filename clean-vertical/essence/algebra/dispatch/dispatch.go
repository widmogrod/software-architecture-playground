package dispatch

import (
	"context"
	"reflect"
	"testing"
)

var handlers map[string]interface{}

func init() {
	handlers = make(map[string]interface{})
}

func Invoke(ctx context.Context, cmd interface{}) interface{} {
	name := reflect.TypeOf(cmd).Name()
	if h, ok := handlers[name]; ok {
		return reflect.ValueOf(h).Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(cmd),
		})[0].Interface()
	}
	return nil
}

func Register(handler interface{}) {
	name := reflect.TypeOf(handler).In(1).Name()
	handlers[name] = handler
}

func ShouldInvokeAndReturn(t *testing.T, v interface{}) {
	name := reflect.TypeOf(v).In(2).Name()
	handlers[name] = func(ctx context.Context, input interface{}) interface{} {
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
}

func Interpret(class interface{}) {
	r := reflect.TypeOf(class)
	for i := 0; i < r.NumMethod(); i++ {
		met := r.Method(i)
		name := met.Type.In(2).Name()
		handlers[name] = func(ctx context.Context, input interface{}) interface{} {
			return met.Func.Call([]reflect.Value{
				reflect.ValueOf(class),
				reflect.ValueOf(ctx),
				reflect.ValueOf(input),
			})[0].Interface()
		}
	}
}
