package inmemory

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation"
	"testing"
)

func TestInMemoryImplementationConformsToSpecification(t *testing.T) {
	implementation := New()
	dispatch.Interpret(implementation)
	interpretation.Specification(t)
}
