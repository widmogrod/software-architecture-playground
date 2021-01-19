package aggssert

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/runtime"
	"testing"
)

type Aggregate interface {
	State() interface{}
	Changes() *runtime.EventStore
	Ref() *runtime.AggregateRef
	Apply(change interface{}) error
	Handle(cmd interface{}) error
}

func Reproducible(t *testing.T, reply, fresh Aggregate) bool {
	reply.Changes().Reduce(func(change interface{}, result *runtime.Reduced) *runtime.Reduced {
		result.StopReduction = !assert.NoError(t, fresh.Apply(change))
		return result
	}, nil)

	return assert.Equal(t, reply.State(), fresh.State())
}

func Empty(t *testing.T, a Aggregate) {
	assert.Empty(t, a.State())
	assert.Empty(t, a.Ref().ID)
	assert.NotEmpty(t, a.Ref().Type)
	ChangesLen(t, a, 0)
}

func ChangesLen(t *testing.T, a Aggregate, expected uint) bool {
	length := a.Changes().Reduce(func(change interface{}, result *runtime.Reduced) *runtime.Reduced {
		result.Value = result.Value.(int) + 1
		return result
	}, uint(0)).Ok

	return assert.Equal(t, length, expected)
}

type DelegatedEquality interface {
	AssertEqual(t *testing.T, value interface{}) bool
}

func CustomEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if compareTo, ok := expected.(DelegatedEquality); ok {
		return compareTo.AssertEqual(t, actual)
	}

	return assert.Equal(t, expected, actual, msgAndArgs)
}

func ChangesSequence(t *testing.T, store *runtime.EventStore, seq ...interface{}) bool {
	result := store.Reduce(func(change interface{}, result *runtime.Reduced) *runtime.Reduced {
		changes, ok := result.Value.([]interface{})
		if !assert.Truef(t, ok, "ChangesSequence is not of type `[]interface{}` but %T", result.Value) {
			result.StopReduction = true
			return result
		}

		if len(changes) == 0 {
			assert.Fail(t, "ChangesSequence has more events to compare", "Next object in sequence should have type %T", change)
			result.StopReduction = true
			return result
		}

		var expectedChange interface{}
		expectedChange, result.Value = changes[0], changes[1:]
		result.StopReduction = !CustomEqual(t, change, expectedChange)

		return result
	}, seq).Ok

	return assert.Len(t, result, 0)
}
