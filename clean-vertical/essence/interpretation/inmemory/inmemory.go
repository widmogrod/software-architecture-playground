package inmemory

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation"
	. "github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
	"math/rand"
	"strconv"
	"time"
)

var _ interpretation.Interpretation = &InMemory{}

func New() *InMemory {
	return &InMemory{
		identityStore:    make(map[string]*identity),
		activationTokens: make([]activationToken, 0),
	}
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

func (i *InMemory) HandleHelloWorld(ctx dispatch.Context, input HelloWorld) ResultOfHelloWorld {
	return ResultOfHelloWorld{
		SuccessfulResult: "Hello, " + input.Name,
	}
}

func (i *InMemory) HandleCreateUserIdentity(ctx dispatch.Context, input CreateUserIdentity) ResultOfCreateUserIdentity {
	output := &ResultOfCreateUserIdentity{}
	idx := string(input.EmailAddress)

	// is persisted
	if _, ok := i.identityStore[idx]; ok {
		output.ValidationError = NewConflictEmailExistsError()
		return *output
	}

	i.identityStore[idx] = &identity{
		UUID:         input.UUID,
		EmailAddress: input.EmailAddress,
	}

	output.SuccessfulResult = NewCreateUserIdentityWithUUID(input.UUID)

	return *output
}

func (i *InMemory) HandleGenerateSessionToken(ctx dispatch.Context, input GenerateSessionToken) ResultOfGeneratingSessionToken {
	return ResultOfGeneratingSessionToken{
		SuccessfulResult: SessionToken{
			AccessToken:  strconv.Itoa(rand.Int()),
			RefreshToken: strconv.Itoa(rand.Int()),
		},
	}
}

func (i *InMemory) HandleCreateAccountActivationToken(ctx dispatch.Context, input CreateAccountActivationToken) ResultOfCreateAccountActivationToken {
	token := time.Now().String()
	i.activationTokens = append(i.activationTokens, activationToken{
		UUID:                 input.UUID,
		EmailActivationToken: token,
	})

	return ResultOfCreateAccountActivationToken{
		SuccessfulResult: token,
	}
}

func (i *InMemory) HandleMarkAccountActivationTokenAsUse(ctx dispatch.Context, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed {
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
