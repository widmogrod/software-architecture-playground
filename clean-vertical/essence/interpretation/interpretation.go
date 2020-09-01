package interpretation

import (
	. "../usecase"
)

type Interpretation interface {
	HandleCreateUserIdentity(input CreateUserIdentity) ResultOfCreateUserIdentity
	HandleGenerateSessionToken(input GenerateSessionToken) ResultOfGeneratingSessionToken
	HandleMarkAccountActivationTokenAsUse(input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed
}
