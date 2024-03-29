package projection

import (
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
)

var _ Handler = &FilterHandler{}

type FilterHandler struct {
	Where *predicate.WherePredicates
}

func (f *FilterHandler) Process(x Item, returning func(Item)) error {
	if f.Where.Evaluate(x.Data) {
		returning(x)
	}

	return nil
}

func (f *FilterHandler) Retract(x Item, returning func(Item)) error {
	return f.Process(x, returning)
}
