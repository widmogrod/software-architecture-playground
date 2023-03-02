package schemaless

import "fmt"

type GenerateHandler struct {
	load func(push func(message Message) error) error
}

func (h *GenerateHandler) Process(msg Message, returning func(Message) error) error {
	return MustMatchMessage(
		msg,
		func(x *Combine) error {
			return h.load(returning)
		},
		func(x *Retract) error {
			return fmt.Errorf("generator cannot retract")
		},
		func(x *Both) error {
			return fmt.Errorf("generator cannot bot retract and combine")
		},
	)
}
