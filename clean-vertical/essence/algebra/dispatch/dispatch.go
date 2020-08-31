package dispatch

import (
	"reflect"
	"testing"
)

var handlers map[string]interface{}

func Invoke(cmd interface{}) interface{} {
	r := reflect.TypeOf(cmd)
	if h, ok := handlers[r.Name()]; ok {
		return reflect.ValueOf(h).Call([]reflect.Value{reflect.ValueOf(cmd)})[0].Interface()
	}
	return nil
}

func When(cmd interface{}, fn interface{}) {
	handlers = make(map[string]interface{})
	r := reflect.TypeOf(cmd)
	handlers[r.Name()] = fn
}

func ShouldInvokeAndReturn(t *testing.T, v interface{}) {
	cmd := reflect.TypeOf(v).In(1).Name()
	handlers[cmd] = func(in interface{}) interface{} {
		res := reflect.ValueOf(v).Call([]reflect.Value{
			reflect.ValueOf(t),
			reflect.ValueOf(in),
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
		cmd := met.Type.In(1)
		handlers[cmd.Name()] = func(input interface{}) interface{} {
			return met.Func.Call([]reflect.Value{
				reflect.ValueOf(class),
				reflect.ValueOf(input),
			})[0].Interface()
		}
	}
}
