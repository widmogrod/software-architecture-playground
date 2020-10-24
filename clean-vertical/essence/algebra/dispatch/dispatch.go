package dispatch

import (
	"testing"
)

var defaultProgram Program

func init() {
	defaultProgram = NewProgram()
}

func SetDefault(p Program) {
	defaultProgram = p
}

func Invoke(ctx Context, cmd interface{}) interface{} {
	return defaultProgram.Invoke(ctx, cmd)
}

func RegisterGlobalHandler(handler interface{}) {
	defaultProgram.RegisterGlobalHandler(handler)
}

func ShouldInvokeAndReturn(t *testing.T, v interface{}) {
	defaultProgram.ShouldInvokeAndReturn(t, v)
}

func Interpret(class interface{}) {
	defaultProgram.Interpretation(class)
}
