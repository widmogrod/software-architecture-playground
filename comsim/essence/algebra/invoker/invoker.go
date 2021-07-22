package invoker

type (
	FunctionID     = string
	FunctionInput  = interface{}
	FunctionOutput = interface{}
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

func (i *Invoke) Invoke(name FunctionID, input FunctionInput) (error, FunctionOutput) {
	err, f := i.Get(name)
	if err != nil {
		return err, nil
	}

	return nil, f.Call(input)
}
