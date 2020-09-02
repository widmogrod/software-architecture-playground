package inmemory

import (
	"../../algebra/dispatch"
	"testing"
)

func TestInMemoryImplementationConformsToSpecification(t *testing.T) {
	inMemory := New()
	dispatch.Interpret(inMemory)
	Specification(t)
}
