package invoker

import (
	"flag"
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
	res, err := i.Invoke("a", "World")
	assert.NoError(t, err)
	assert.Equal(t, "Hello World", res)
}

var worksInGitHubActions = flag.Bool("i-work-in-github-actions", false, "Integration that tests work locally with docker but not in GitHub Actions (yet)")

func TestInvokeInDocker(t *testing.T) {
	if !*worksInGitHubActions {
		t.Skip("Skipping tests because this don't work in GitHub actions")
	}

	fr := NewDockerFunctionRegistry()
	err := fr.Register("a", "./demo-func")
	assert.NoError(t, err)

	i := NewInvoker(fr)
	res, err := i.Invoke("a", "World")
	assert.NoError(t, err)
	assert.Equal(t, "Hello World, from Docker!", res)
}
