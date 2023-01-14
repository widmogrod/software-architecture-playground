package machine

func NewSimpleMachine[C, S any](f func(C, S) (S, error)) *Machine[C, S] {
	return &Machine[C, S]{
		handle: f,
	}
}

type Machine[C, S any] struct {
	state  S
	handle func(C, S) (S, error)
}

func (o *Machine[C, S]) Handle(cmd C) error {
	state, err := o.handle(cmd, o.state)
	if err != nil {
		return err
	}

	o.state = state
	return nil
}

func (o *Machine[C, S]) State() S {
	return o.state
}
