package schemaless

type GenerateHandler struct {
	load func(push func(message Item)) error
}

func (h *GenerateHandler) Process(_ Item, returning func(Item)) error {
	return h.load(returning)
}
