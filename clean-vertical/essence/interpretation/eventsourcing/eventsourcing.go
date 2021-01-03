package eventsourcing

import (
	"container/list"
	"fmt"
	"github.com/segmentio/ksuid"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/interpretation"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
	"math/rand"
	"strconv"
	"sync"
)

var _ interpretation.Interpretation = &EventSourcing{}

func New() *EventSourcing {
	return &EventSourcing{
		IdentityStore:        NewEventStore(),
		ActivationTokenStore: NewEventStore(),
	}
}

type EventSourcing struct {
	IdentityStore        *EventStore
	ActivationTokenStore *EventStore
}

func NewEventStore() *EventStore {
	return &EventStore{
		log:  list.New(),
		lock: &sync.Mutex{},
		err:  nil,
	}
}

type EventStore struct {
	lock sync.Locker
	log  *list.List
	err  error
}

func (a *EventStore) Append(input interface{}) *AggregateAppendResult {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.err != nil {
		return &AggregateAppendResult{
			Ok:  a,
			Err: a.err,
		}
	}

	a.log.PushBack(input)

	return &AggregateAppendResult{
		Ok: a,
	}
}

type Reduced struct {
	StopReduction bool
	Value         interface{}
}

func (a *EventStore) Reduce(f func(cmd interface{}, result *Reduced) *Reduced, init interface{}) *AggregateResultResult {
	a.lock.Lock()
	defer a.lock.Unlock()

	result := &Reduced{
		StopReduction: false,
		Value:         init,
	}

	if a.err != nil {
		return &AggregateResultResult{
			Ok:  result,
			Err: a.err,
		}
	}

	for e := a.log.Front(); e != nil; e = e.Next() {
		result = f(e.Value, result)
		if result.StopReduction {
			break
		}
	}

	return &AggregateResultResult{
		Ok: result,
	}
}

type AggregateAppendResult struct {
	Ok  *EventStore
	Err error
}

type AggregateResultResult struct {
	Ok  *Reduced
	Err error
}

func (e *EventSourcing) HandleCreateUserIdentity(ctx dispatch.Context, input usecase.CreateUserIdentity) usecase.ResultOfCreateUserIdentity {
	type aggregationContainer = struct {
		Duplicated bool
	}

	result := &aggregationContainer{
		Duplicated: false,
	}
	err := e.IdentityStore.
		Append(input).Ok.
		Reduce(func(value interface{}, result *Reduced) *Reduced {
			agg := result.Value.(*aggregationContainer)
			if input2, ok := value.(usecase.CreateUserIdentity); ok {
				// we have duplicate
				if input2.EmailAddress == input.EmailAddress {
					// let's check if that's not our entry
					if input.UUID != input2.UUID {
						// and when not, then we have duplicate
						agg.Duplicated = true
					}

					// either way, we're should stop reduction
					result.StopReduction = true
				}
			}

			return result
		}, result).Err

	if err != nil {
		// check your assumptions here!
		panic("this should never happen! but here we go: " + err.Error())
	}

	output := usecase.ResultOfCreateUserIdentity{}

	if result.Duplicated {
		output.ValidationError = usecase.NewConflictEmailExistsError()
	} else {
		output.SuccessfulResult = usecase.NewCreateUserIdentityWithUUID(input.UUID)
	}

	return output
}

func (e *EventSourcing) HandleGenerateSessionToken(ctx dispatch.Context, input usecase.GenerateSessionToken) usecase.ResultOfGeneratingSessionToken {
	output := usecase.ResultOfGeneratingSessionToken{
		SuccessfulResult: usecase.SessionToken{
			AccessToken:  strconv.Itoa(rand.Int()),
			RefreshToken: strconv.Itoa(rand.Int()),
		},
	}
	return output
}

type ActivationTokenEntity struct {
	Token string
	UUID  string
}

func (e *EventSourcing) HandleMarkAccountActivationTokenAsUse(ctx dispatch.Context, input usecase.MarkAccountActivationTokenAsUse) usecase.ResultOfMarkingAccountActivationTokenAsUsed {
	type aggregationContainer = struct {
		InvalidToken  bool
		ActivatedUUID string
	}

	result := &aggregationContainer{
		InvalidToken: true,
	}
	err := e.ActivationTokenStore.
		Reduce(func(cmd interface{}, result *Reduced) *Reduced {
			agg := result.Value.(*aggregationContainer)

			activationToken := cmd.(ActivationTokenEntity)
			if activationToken.Token == input.ActivationToken {
				agg.InvalidToken = false
				agg.ActivatedUUID = activationToken.UUID

				result.StopReduction = true
			}

			return result
		}, result).Err

	if err != nil {
		// check your assumptions here!
		panic("this should never happen! but here we go: " + err.Error())
	}

	output := usecase.ResultOfMarkingAccountActivationTokenAsUsed{}
	if result.InvalidToken {
		output.ValidationError = usecase.NewAccountActivationInvalidTokenError()
	} else {
		output.SuccessfulResult = usecase.NewAccountActivatedViaTokenSuccess(result.ActivatedUUID)
	}
	return output
}

func (e *EventSourcing) HandleCreateAccountActivationToken(ctx dispatch.Context, input usecase.CreateAccountActivationToken) usecase.ResultOfCreateAccountActivationToken {
	token := ActivationTokenEntity{
		Token: ksuid.New().String(),
		UUID:  input.UUID,
	}
	err := e.ActivationTokenStore.
		Append(token).Err

	if err != nil {
		// check your assumptions here!
		panic("this should never happen! but here we go: " + err.Error())
	}

	output := usecase.ResultOfCreateAccountActivationToken{}
	output.SuccessfulResult = token.Token
	return output
}

func (e *EventSourcing) HandleHelloWorld(ctx dispatch.Context, input usecase.HelloWorld) usecase.ResultOfHelloWorld {
	return usecase.ResultOfHelloWorld{
		SuccessfulResult: fmt.Sprintf("Ola! %s", input),
	}
}
