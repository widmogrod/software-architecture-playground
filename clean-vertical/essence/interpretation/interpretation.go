package interpretation

import (
	. "../usecase"
	"context"
)

type Interpretation interface {
	HandleCreateUserIdentity(ctx context.Context, input CreateUserIdentity) ResultOfCreateUserIdentity
	HandleGenerateSessionToken(ctx context.Context, input GenerateSessionToken) ResultOfGeneratingSessionToken
	HandleMarkAccountActivationTokenAsUse(ctx context.Context, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed
}
