package workflow

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation"
	. "github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
)

var _ interpretation.Interpretation = &App{}

type App struct {
}

func (a App) HandleCreateUserIdentity(ctx dispatch.Context, input CreateUserIdentity) ResultOfCreateUserIdentity {
	panic("implement me")
}

func (a App) HandleGenerateSessionToken(ctx dispatch.Context, input GenerateSessionToken) ResultOfGeneratingSessionToken {
	panic("implement me")
}

func (a App) HandleMarkAccountActivationTokenAsUse(ctx dispatch.Context, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed {
	panic("implement me")
}

func (a App) HandleHelloWorld(ctx dispatch.Context, input HelloWorld) ResultOfHelloWorld {
	//ctx.PauseActivity().ResumeActivityWithResultURL()
	//activity.CallbackFromURL().Put(ResultOfHelloWorld{})

	panic("implement me")
}

func (a App) HandleCreateAccountActivationToken(ctx dispatch.Context, input CreateAccountActivationToken) ResultOfCreateAccountActivationToken {
	panic("implement me")
}
