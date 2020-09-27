package inmemory

import (
	"context"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation"
	. "github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
	"time"
)

var _ interpretation.Interpretation = &InMemory{}

func New() *InMemory {
	return &InMemory{
		identityStore:    make(map[string]*identity),
		activationTokens: make([]activationToken, 0),
	}
}

func String(s string) *string {
	return &s
}

type InMemory struct {
	identityStore    map[string]*identity
	activationTokens []activationToken
}

type activationToken struct {
	UUID                 string
	EmailActivationToken string
}

type identity struct {
	UUID         string
	EmailAddress EmailAddress
}

func (i *InMemory) HandleHelloWorld(ctx context.Context, input HelloWorld) ResultOfHelloWorld {
	return ResultOfHelloWorld{
		SuccessfulResult: "Hello, " + input.Name,
	}
}

func (i *InMemory) HandleCreateUserIdentity(ctx context.Context, input CreateUserIdentity) ResultOfCreateUserIdentity {
	output := &ResultOfCreateUserIdentity{}
	idx := string(input.EmailAddress)

	// is persisted
	if _, ok := i.identityStore[idx]; ok {
		output.ValidationError = NewConflictEmailExistsError()
		return *output
	}

	uuid := time.Now().String()

	i.identityStore[idx] = &identity{
		UUID:         uuid,
		EmailAddress: input.EmailAddress,
	}

	output.SuccessfulResult = NewCreateUserIdentityWithUUID(uuid)

	return *output
}

func (i *InMemory) HandleGenerateSessionToken(ctx context.Context, input GenerateSessionToken) ResultOfGeneratingSessionToken {
	// TODO implement
	return ResultOfGeneratingSessionToken{
		SuccessfulResult: SessionToken{
			AccessToken:  "",
			RefreshToken: "",
		},
	}
}

func (i *InMemory) HandleCreateAccountActivationToken(ctx context.Context, input CreateAccountActivationToken) ResultOfCreateAccountActivationToken {
	token := time.Now().String()
	i.activationTokens = append(i.activationTokens, activationToken{
		UUID:                 input.UUID,
		EmailActivationToken: token,
	})

	return ResultOfCreateAccountActivationToken{
		SuccessfulResult: token,
	}
}

func (i *InMemory) HandleMarkAccountActivationTokenAsUse(ctx context.Context, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed {
	// TODO: Naive implementation ahead! O(n) don't do it at home!
	for idx, token := range i.activationTokens {
		if token.EmailActivationToken == input.ActivationToken {
			// remove token from list
			i.activationTokens[idx] = i.activationTokens[len(i.activationTokens)-1]
			i.activationTokens = i.activationTokens[:len(i.activationTokens)-1]

			return ResultOfMarkingAccountActivationTokenAsUsed{
				SuccessfulResult: NewAccountActivatedViaTokenSuccess(token.UUID),
			}
		}
	}

	return ResultOfMarkingAccountActivationTokenAsUsed{
		ValidationError: NewAccountActivationInvalidTokenError(),
	}
}
