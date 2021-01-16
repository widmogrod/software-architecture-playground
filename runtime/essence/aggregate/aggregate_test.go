package aggregate

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/aggssert"
	"testing"
	"time"
)

var (
	okUserID    = "666"
	okProductID = "p7"
	okQuantity  = "3"
)

func TestOrderAggregateState_new_aggregate_has_empty_state(t *testing.T) {
	a := NewOrderAggregate()
	aggssert.Empty(t, a)
}

func TestOrderAggregateState_aggregate_state_equal_to_new_replay_state(t *testing.T) {
	a := NewOrderAggregate()
	err := a.Handle(&OrderCreateCMD{
		UserID:    okUserID,
		ProductID: okProductID,
		Quantity:  okQuantity,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, a.ref.ID)

	aggssert.Reproducible(t, a, NewOrderAggregate())

	// Assert specific state
	if assert.NotEmpty(t, a.state) {
		assert.Equal(t, okUserID, a.state.UserID)
		assert.Equal(t, okProductID, a.state.ProductID)
		assert.Equal(t, okQuantity, a.state.ProductQuantity)
	}

	now := time.Now()
	aggssert.ChangesSequence(t, a.Changes(),
		&OrderCreated{
			OrderID:   a.Ref().ID,
			UserID:    okUserID,
			CreatedAt: &now,
		},
		&ProductAdded{
			ProductID: okProductID,
			Quantity:  okQuantity,
		},
	)
}

func (a *OrderCreated) AssertEqual(t *testing.T, value interface{}) bool {
	if b, ok := value.(*OrderCreated); ok {
		return assert.Equal(t, a.UserID, b.UserID) &&
			assert.Equal(t, a.OrderID, b.OrderID) &&
			assert.WithinDuration(t, *a.CreatedAt, *b.CreatedAt, time.Second)
	}

	return false
}
