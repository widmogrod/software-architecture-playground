package projection

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
)

func TestDefaultInMemoryInterpreter(t *testing.T) {
	dag := NewDAGBuilder()
	_ = dag.
		Load(&GenerateHandler{Load: func(push func(message Item)) error {
			push(Item{Key: "1", Data: schema.FromGo(1)})
			return nil
		}}).
		Map(&MapHandler[int, int]{
			F: func(x int, returning func(key string, value int)) error {
				returning("x", x+1)
				return nil
			},
		}, WithName("DoSomething"))

	t.Run("normal run, finishes", func(t *testing.T) {
		interpreter := DefaultInMemoryInterpreter()
		assert.NotNil(t, interpreter)

		err := interpreter.Run(context.Background(), dag.Build())
		assert.NoError(t, err)

		stats := interpreter.StatsSnapshotAndReset()
		assert.Equal(t, 1, stats["load[root.Load0].returning"])
		assert.Equal(t, 1, stats["map[DoSomething].returning.aggregate"])

		// second time should return zero
		stats = interpreter.StatsSnapshotAndReset()
		assert.Equal(t, 0, stats["load[root.Load0].returning"])
		assert.Equal(t, 0, stats["map[DoSomething].returning.aggregate"])
	})

	// should be able to run again the same DAG

	t.Run("run on closed context should not execute", func(t *testing.T) {
		// should not execute when context is cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		interpreter := DefaultInMemoryInterpreter()
		err := interpreter.Run(ctx, dag.Build())
		assert.NoError(t, err)

		stats := interpreter.StatsSnapshotAndReset()
		assert.Equal(t, 0, stats["load[root.Load0].returning"])
		assert.Equal(t, 0, stats["map[DoSomething].returning.aggregate"])
	})

	t.Run("executing the same DAG twice should not execute twice", func(t *testing.T) {
		interpreter := DefaultInMemoryInterpreter()
		assert.NotNil(t, interpreter)

		err := interpreter.Run(context.Background(), dag.Build())
		assert.NoError(t, err)

		err = interpreter.Run(context.Background(), dag.Build())
		assert.ErrorIs(t, err, ErrInterpreterNotInNewState)
	})
}
