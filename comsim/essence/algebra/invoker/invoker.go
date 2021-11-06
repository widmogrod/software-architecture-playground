package invoker

type (
	FunctionID = string
	// FunctionInput and FunctionOutput represent a data.
	// In a way we can think about it as parametric polymorphism of f: a -> b
	// To enable decoupling and scalability of functions so that they can leave in process or on other machine,
	// input and output need to be though as data that can we can transmit over network, in other words message passing.
	// For prototype purpose is represented as string, but for production we should thing about it as bytes.
	FunctionInput  = string
	FunctionOutput = string
)

type Func = func(input FunctionInput) FunctionOutput

type Function interface {
	Call(input FunctionInput) FunctionOutput
}

type FunctionRegistry interface {
	Get(name FunctionID) (error, Function)
}

func NewInvoker(fr FunctionRegistry) *Invoke {
	return &Invoke{fr: fr}
}

type Invoke struct {
	fr FunctionRegistry
}

func (i *Invoke) Get(name FunctionID) (error, Function) {
	return i.fr.Get(name)
}

func (i *Invoke) Invoke(name FunctionID, input FunctionInput) (FunctionOutput, error) {
	err, f := i.Get(name)
	if err != nil {
		return "", err
	}

	return f.Call(input), nil
}
