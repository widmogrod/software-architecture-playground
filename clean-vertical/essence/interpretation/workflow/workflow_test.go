package workflow

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation/inmemory"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
	"testing"
	"time"
)

var program *dispatch.Workflow

func init() {
	program = dispatch.NewWorkflow()

	dispatch.SetDefault(program)
}

func TestWorkflowImplementationConformsToSpecification(t *testing.T) {
	inMemory := inmemory.New()
	//program := dispatch.NewWorkflow()
	program.Interpretation(inMemory)

	go program.Log()

	usecase.SpecHelloWorld(t)
	//usecase.SpecRegisterAccountWithEmail(t)

	//interpretation.Specification(t)
	//usecase.SpecConfirmAccountActivation(t)

	time.Sleep(time.Second * 3)
}
