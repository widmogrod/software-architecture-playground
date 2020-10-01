package awsstepfunction

import (
	"context"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation"
	. "github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
)

var _ interpretation.Interpretation = &AWSStepFunction{}

type AWSStepFunction struct {
}

func (a AWSStepFunction) HandleCreateUserIdentity(ctx context.Context, input CreateUserIdentity) ResultOfCreateUserIdentity {
	panic("implement me")
}

func (a AWSStepFunction) HandleGenerateSessionToken(ctx context.Context, input GenerateSessionToken) ResultOfGeneratingSessionToken {
	panic("implement me")
}

func (a AWSStepFunction) HandleMarkAccountActivationTokenAsUse(ctx context.Context, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed {
	panic("implement me")
}

func (a AWSStepFunction) HandleHelloWorld(ctx context.Context, input HelloWorld) ResultOfHelloWorld {
	//ctx.PauseActivity().ResumeActivityWithResultURL()
	//activity.CallbackFromURL().Put(ResultOfHelloWorld{})

	panic("implement me")
}

func (a AWSStepFunction) HandleCreateAccountActivationToken(ctx context.Context, input CreateAccountActivationToken) ResultOfCreateAccountActivationToken {
	panic("implement me")
}
