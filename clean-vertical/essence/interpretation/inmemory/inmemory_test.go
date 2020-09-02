package inmemory

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
)

func TestInMemoryImplementationConformsToSpecification(t *testing.T) {
	inMemory := New()
	dispatch.Interpret(inMemory)
	Specification(t)
}
