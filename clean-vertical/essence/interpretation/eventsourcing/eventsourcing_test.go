package eventsourcing

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation"
	"testing"
)

func TestEventSourcingImplementationConformsToSpecification(t *testing.T) {
	implementation := New()
	dispatch.Interpret(implementation)
	interpretation.Specification(t)
}
