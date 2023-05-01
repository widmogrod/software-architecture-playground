package projection

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
	"time"
)

func TestTriggering(t *testing.T) {
	ctx := context.Background()
	handler := &TriggerHandler{
		td: &AtPeriod{
			Duration: 100 * time.Millisecond,
		},
		wd: &FixedWindow{
			Width: 100 * time.Millisecond,
		},
	}

	returning := &ListAssert{t: t}

	go Trigger(ctx, func() {
		err := handler.Triggered(returning.Returning)
		assert.NoError(t, err)
	}, handler.td)

	for item := range GenerateItemsEvery(withTime(10, 0), 20, 10*time.Millisecond) {
		err := handler.Process(item, returning.Returning)
		assert.NoError(t, err)
	}

	time.Sleep(100 * time.Millisecond)

	assert.Empty(t, handler.buffer)
	t.Log(len(returning.Items))
	returning.AssertAt(0, Item{
		Key: "key",
		Data: schema.MkList(
			schema.MkInt(0), schema.MkInt(1), schema.MkInt(2), schema.MkInt(3), schema.MkInt(4),
			schema.MkInt(5), schema.MkInt(6), schema.MkInt(7), schema.MkInt(8), schema.MkInt(9),
		),
		EventTime: withTime(10, 0) + (100 * int64(time.Millisecond)),
		Window: &Window{
			Start: withTime(10, 0),
			End:   withTime(10, 0) + (100 * int64(time.Millisecond)),
		},
	})

}
