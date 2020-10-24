package awsstepfunction

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation"
	. "github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
)

var _ interpretation.Interpretation = &AWSStepFunction{}

type AWSStepFunction struct {
}

func (a AWSStepFunction) HandleCreateUserIdentity(ctx dispatch.Context, input CreateUserIdentity) ResultOfCreateUserIdentity {
	panic("implement me")
}

func (a AWSStepFunction) HandleGenerateSessionToken(ctx dispatch.Context, input GenerateSessionToken) ResultOfGeneratingSessionToken {
	panic("implement me")
}

func (a AWSStepFunction) HandleMarkAccountActivationTokenAsUse(ctx dispatch.Context, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed {
	panic("implement me")
}

func (a AWSStepFunction) HandleHelloWorld(ctx dispatch.Context, input HelloWorld) ResultOfHelloWorld {
	//ctx.PauseActivity().ResumeActivityWithResultURL()
	//activity.CallbackFromURL().Put(ResultOfHelloWorld{})

	panic("implement me")
}

func (a AWSStepFunction) HandleCreateAccountActivationToken(ctx dispatch.Context, input CreateAccountActivationToken) ResultOfCreateAccountActivationToken {
	panic("implement me")
}
