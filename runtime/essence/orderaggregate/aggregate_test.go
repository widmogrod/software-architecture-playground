package orderaggregate

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/aggssert"
	"strings"
	"testing"
	"time"
)

var (
	okUserID    = "666"
	okProductID = "p7"
	okQuantity  = "3"
	now         = time.Now()
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
	assert.NotEmpty(t, a.Ref().ID)

	aggssert.Reproducible(t, a, NewOrderAggregate())

	// Assert specific state
	if assert.NotEmpty(t, a.state) {
		assert.Equal(t, okUserID, a.state.UserID)
		assert.Equal(t, okProductID, a.state.ProductID)
		assert.Equal(t, okQuantity, a.state.ProductQuantity)
	}

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

func TestOrderAggregateState_a_scenario(t *testing.T) {
	a := NewOrderAggregate()
	err := a.Handle(&OrderCreateCMD{
		UserID:    okUserID,
		ProductID: okProductID,
		Quantity:  okQuantity,
	})
	assert.NoError(t, err)

	err = a.Handle(&OrderCollectPaymentsCMD{
		Method: "apple",
		Amount: 100,
		Date:   &now,
	})
	assert.NoError(t, err)

	aggssert.Reproducible(t, a, NewOrderAggregate())

	// Assert specific state
	if assert.NotEmpty(t, a.state) {
		assert.Equal(t, okUserID, a.state.UserID)
		assert.Equal(t, okProductID, a.state.ProductID)
		assert.Equal(t, okQuantity, a.state.ProductQuantity)

		assert.True(t, a.state.PaymentCollected)
	}

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
		&OrderPaymentsCollected{
			Method: "apple",
			Amount: 100,
			Date:   &now,
		},
	)
}

func TestOrderAggregateState_cmd_permutation(t *testing.T) {
	commands := []interface{}{
		&OrderCreateCMD{
			UserID:    okUserID,
			ProductID: okProductID,
			Quantity:  okQuantity,
		}, &OrderCollectPaymentsCMD{
			Method: "apple",
			Amount: 100,
			Date:   &now,
		},
		&OrderCreateCMD{
			UserID:    okUserID,
			ProductID: okProductID,
			Quantity:  okQuantity,
		},
	}

	Perm(commands, func(perm []interface{}) {
		a := NewOrderAggregate()
		in := make([]string, len(perm)+1)
		var err error
		for i, v := range perm {
			err = a.Handle(v)
			if err != nil {
				in[i] = fmt.Sprintf("%T [x]", v)
			} else {
				in[i] = fmt.Sprintf("%T [√]", v)
			}
		}

		if err != nil {
			in[len(perm)] = fmt.Sprintf("  err= %s", err.Error())
		} else {
			in[len(perm)] = fmt.Sprintf("state= %#v", a.State())
		}

		aggssert.Reproducible(t, a, NewOrderAggregate())

		t.Logf("%s\n", strings.Join(
			in,
			" ",
		))
	})
}

func Perm(p []interface{}, f func(perm []interface{})) {
	perm(p, f, 0)
}

func perm(a []interface{}, f func(perm []interface{}), i int) {
	if i > len(a) {
		f(a)
		return
	}
	perm(a, f, i+1)
	for j := i + 1; j < len(a); j++ {
		a[i], a[j] = a[j], a[i]
		perm(a, f, i+1)
		a[i], a[j] = a[j], a[i]
	}
}
