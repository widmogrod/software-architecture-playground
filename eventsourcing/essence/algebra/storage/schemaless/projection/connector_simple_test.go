package schemaless

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
)

func TestGenerateHandler(t *testing.T) {
	generate := []Message{
		&Combine{
			Data: schema.FromGo(Game{
				Players: []string{"a", "b"},
				Winner:  "a",
			}),
		},
		&Combine{
			Data: schema.FromGo(Game{
				Players: []string{"a", "b"},
				Winner:  "b",
			}),
		},
		&Combine{
			Data: schema.FromGo(Game{
				Players: []string{"a", "b"},
				IsDraw:  true,
			}),
		},
	}

	h := &GenerateHandler{
		load: func(returning func(message Message) error) error {
			for idx, msg := range generate {
				err := returning(msg)
				assert.NoError(t, err, "failed to returning message at index=%d", idx)
			}
			return nil
		},
	}

	l := &ListAssert{
		t: t,
	}
	err := h.Process(&Combine{}, l.Returning)
	assert.NoError(t, err)

	l.AssertLen(3)

	for idx, msg := range generate {
		l.AssertAt(idx, msg)
	}
}
