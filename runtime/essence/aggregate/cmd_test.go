package aggregate

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func (a *OrderCreated) AssertEqual(t *testing.T, value interface{}) bool {
	if b, ok := value.(*OrderCreated); ok {
		return assert.Equal(t, a.UserID, b.UserID) &&
			assert.Equal(t, a.OrderID, b.OrderID) &&
			assert.WithinDuration(t, *a.CreatedAt, *b.CreatedAt, time.Second)
	}

	return false
}
