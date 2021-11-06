package invoker

import "fmt"

var _ Function = &FunctionInMemory{}

type FunctionInMemory struct {
	F func(input FunctionInput) FunctionOutput
}

func (f *FunctionInMemory) Call(input FunctionInput) FunctionOutput {
	return f.F(input)
}

var _ FunctionRegistry = &InMemoryFunctionRegistry{}

type InMemoryFunctionRegistry struct {
	reg map[FunctionID]*FunctionInMemory
}

func (r *InMemoryFunctionRegistry) Get(name FunctionID) (error, Function) {
	if _, ok := r.reg[name]; ok {
		return nil, r.reg[name]
	}

	return fmt.Errorf("in-memory-function-registry: function '%s' does not exists", name), nil
}

func (r *InMemoryFunctionRegistry) Register(name FunctionID, fun *FunctionInMemory) error {
	if _, ok := r.reg[name]; ok {
		return fmt.Errorf("in-memory-function-registry: function '%s' already registred", name)

	}

	r.reg[name] = fun
	return nil
}

func NewInMemoryFunctionRegistry() *InMemoryFunctionRegistry {
	return &InMemoryFunctionRegistry{
		reg: make(map[FunctionID]*FunctionInMemory),
	}
}
