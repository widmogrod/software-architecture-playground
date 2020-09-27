package interpretation

import (
	"context"
	. "github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
)

type Interpretation interface {
	HandleCreateUserIdentity(ctx context.Context, input CreateUserIdentity) ResultOfCreateUserIdentity
	HandleGenerateSessionToken(ctx context.Context, input GenerateSessionToken) ResultOfGeneratingSessionToken
	HandleMarkAccountActivationTokenAsUse(ctx context.Context, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed
	HandleCreateAccountActivationToken(ctx context.Context, input CreateAccountActivationToken) ResultOfCreateAccountActivationToken
	HandleHelloWorld(ctx context.Context, input HelloWorld) ResultOfHelloWorld
}
