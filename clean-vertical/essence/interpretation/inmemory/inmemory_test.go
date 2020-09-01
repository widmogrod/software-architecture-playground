package inmemory

import (
	"../../algebra/dispatch"
	"../../interpretation"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	inMemory := New()
	dispatch.Interpret(inMemory)

	os.Exit(m.Run())
}

func TestInMemoryImplementationConformsToSpecification(t *testing.T) {
	interpretation.Specification(t)
}
