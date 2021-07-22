package invoker

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInvokeInMemory(t *testing.T) {
	fr := NewInMemoryFunctionRegistry()
	err := fr.Register("a", &FunctionInMemory{func(input FunctionInput) FunctionOutput {
		return fmt.Sprintf("Hello %s", input)
	}})
	assert.NoError(t, err)

	i := NewInvoker(fr)
	err, res := i.Invoke("a", "World")
	assert.NoError(t, err)
	assert.Equal(t, "Hello World", res)
}

func TestInvokeInDocker(t *testing.T) {
	fr := NewDockerFunctionRegistry()
	err := fr.Register("a", "./demo-func")
	assert.NoError(t, err)

	i := NewInvoker(fr)
	err, res := i.Invoke("a", "World")
	assert.NoError(t, err)
	assert.Equal(t, "Hello World, from Docker!", res)
}
