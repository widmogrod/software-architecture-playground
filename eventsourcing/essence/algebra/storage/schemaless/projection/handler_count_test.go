package schemaless

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
)

func TestCountHandler(t *testing.T) {
	h := &CountHandler{}
	assert.Equal(t, 0, h.value)

	l := &ListAssert{t: t}
	err := h.Process(&Combine{
		Data: schema.MkInt(1),
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(0, &Both{
		Retract: Retract{
			Data: schema.MkInt(0),
		},
		Combine: Combine{
			Data: schema.MkInt(1),
		},
	})
	assert.Equal(t, 1, h.value)

	err = h.Process(&Combine{
		Data: schema.MkInt(2),
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(1, &Both{
		Retract: Retract{
			Data: schema.MkInt(1),
		},
		Combine: Combine{
			Data: schema.MkInt(3),
		},
	})
	assert.Equal(t, 3, h.value)

	err = h.Process(&Retract{
		Data: schema.MkInt(1),
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(2, &Both{
		Retract: Retract{
			Data: schema.MkInt(3),
		},
		Combine: Combine{
			Data: schema.MkInt(2),
		},
	})
	assert.Equal(t, 2, h.value)

	err = h.Process(&Both{
		Retract: Retract{
			Data: schema.MkInt(2),
		},
		Combine: Combine{
			Data: schema.MkInt(1),
		},
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(3, &Both{
		Retract: Retract{
			Data: schema.MkInt(2),
		},
		Combine: Combine{
			Data: schema.MkInt(1),
		},
	})
	assert.Equal(t, 1, h.value)
}
