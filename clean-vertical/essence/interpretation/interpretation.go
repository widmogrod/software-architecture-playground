package interpretation

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	. "github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
)

type Interpretation interface {
	HandleCreateUserIdentity(ctx dispatch.Context, input CreateUserIdentity) ResultOfCreateUserIdentity
	HandleGenerateSessionToken(ctx dispatch.Context, input GenerateSessionToken) ResultOfGeneratingSessionToken
	HandleMarkAccountActivationTokenAsUse(ctx dispatch.Context, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed
	HandleCreateAccountActivationToken(ctx dispatch.Context, input CreateAccountActivationToken) ResultOfCreateAccountActivationToken
	HandleHelloWorld(ctx dispatch.Context, input HelloWorld) ResultOfHelloWorld
}
