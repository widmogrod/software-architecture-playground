package schemaless

var _ Handler = &GenerateHandler{}

type GenerateHandler struct {
	load func(push func(message Item)) error
}

func (h *GenerateHandler) Process(_ Item, returning func(Item)) error {
	return h.load(returning)
}

func (h *GenerateHandler) Retract(x Item, returning func(Item)) error {
	//TODO implement me
	panic("implement me")
}
