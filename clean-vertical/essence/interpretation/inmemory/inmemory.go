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
		identityStore: make(map[string]struct {
			UUID         string
			EmailAddress EmailAddress
		}),
	}
}

type InMemory struct {
	identityStore map[string]struct {
		UUID         string
		EmailAddress EmailAddress
	}
}

func (i *InMemory) HandleCreateUserIdentity(ctx context.Context, input CreateUserIdentity) ResultOfCreateUserIdentity {
	output := &ResultOfCreateUserIdentity{}
	idx := string(input.EmailAddress)

	// is persisted
	if _, ok := i.identityStore[idx]; ok {
		output.ConflictEmailExists()
		return *output
	}

	uuid := time.Now().String()

	i.identityStore[idx] = struct {
		UUID         string
		EmailAddress EmailAddress
	}{
		UUID:         uuid,
		EmailAddress: input.EmailAddress,
	}

	output.SucceedWithUUID(uuid)

	return *output
}

func (i *InMemory) HandleGenerateSessionToken(ctx context.Context, input GenerateSessionToken) ResultOfGeneratingSessionToken {
	panic("implement me")
}

func (i *InMemory) HandleMarkAccountActivationTokenAsUse(ctx context.Context, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed {
	panic("implement me")
}
