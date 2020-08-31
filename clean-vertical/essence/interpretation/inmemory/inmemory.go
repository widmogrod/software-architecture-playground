package inmemory

import (
	. "../../usecase"
	"time"
)

func New() *InMemoryInterpretation {
	return &InMemoryInterpretation{
		identityStore: make(map[string]struct {
			UUID         string
			EmailAddress EmailAddress
		}),
	}
}

type InMemoryInterpretation struct {
	identityStore map[string]struct {
		UUID         string
		EmailAddress EmailAddress
	}
}

func (i *InMemoryInterpretation) HandleCreateUserIdentity(input CreateUserIdentity) ResultOfCreateUserIdentity {
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

	output.SuccedWithUUID(uuid)

	return *output
}
