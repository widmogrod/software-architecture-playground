package schemaless

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
)

func TestAvgHandler(t *testing.T) {
	h := &AvgHandler{}
	assert.Equal(t, float64(0), h.avg)

	l := ListAssert{t: t}

	err := h.Process(&Combine{
		Data: schema.MkInt(1),
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(0, &Both{
		Retract: Retract{
			Data: schema.MkFloat(0),
		},
		Combine: Combine{
			Data: schema.MkFloat(1),
		},
	})
	assert.Equal(t, float64(1), h.avg)
	assert.Equal(t, 1, h.count)

	err = h.Process(&Combine{
		Data: schema.MkInt(11),
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(1, &Both{
		Retract: Retract{
			Data: schema.MkFloat(1),
		},
		Combine: Combine{
			Data: schema.MkFloat(6),
		},
	})
	assert.Equal(t, float64(6), h.avg)
	assert.Equal(t, 2, h.count)

	err = h.Process(&Combine{
		Data: schema.MkInt(3),
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(2, &Both{
		Retract: Retract{
			Data: schema.MkFloat(6),
		},
		Combine: Combine{
			Data: schema.MkFloat(5),
		},
	})
	assert.Equal(t, float64(5), h.avg)
	assert.Equal(t, 3, h.count)

	err = h.Process(&Retract{
		Data: schema.MkInt(1),
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(3, &Both{
		Retract: Retract{
			Data: schema.MkFloat(5),
		},
		Combine: Combine{
			Data: schema.MkFloat(7),
		},
	})
	assert.Equal(t, float64(7), h.avg)
	assert.Equal(t, 2, h.count)

	err = h.Process(&Both{
		Retract: Retract{
			Data: schema.MkInt(3),
		},
		Combine: Combine{
			Data: schema.MkInt(10),
		},
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(4, &Both{
		Retract: Retract{
			Data: schema.MkFloat(7),
		},
		Combine: Combine{
			Data: schema.MkFloat(10.5),
		},
	})
	assert.Equal(t, float64(10.5), h.avg)
	assert.Equal(t, 2, h.count)
}
