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
	r := reflect.TypeOf(cmd)
	if h, ok := handlers[r.Name()]; ok {
		return reflect.ValueOf(h).Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(cmd),
		})[0].Interface()
	}
	return nil
}

func When(cmd interface{}, fn interface{}) {
	r := reflect.TypeOf(cmd)
	handlers[r.Name()] = fn
}

func ShouldInvokeAndReturn(t *testing.T, v interface{}) {
	cmd := reflect.TypeOf(v).In(2).Name()
	handlers[cmd] = func(ctx context.Context, input interface{}) interface{} {
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
		cmd := met.Type.In(2)
		handlers[cmd.Name()] = func(ctx context.Context, input interface{}) interface{} {
			return met.Func.Call([]reflect.Value{
				reflect.ValueOf(class),
				reflect.ValueOf(ctx),
				reflect.ValueOf(input),
			})[0].Interface()
		}
	}
}
