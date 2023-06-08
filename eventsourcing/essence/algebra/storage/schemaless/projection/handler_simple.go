package projection

var _ Handler = (*SimpleHandler)(nil)

type SimpleHandler struct {
	P func(x Item, returning func(Item)) error
	R func(x Item, returning func(Item)) error
}

func (s SimpleHandler) Process(x Item, returning func(Item)) error {
	return s.P(x, returning)
}

func (s SimpleHandler) Retract(x Item, returning func(Item)) error {
	return s.R(x, returning)
}
