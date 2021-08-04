package stream

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStream(t *testing.T) {
	s := NewRandomStream()
	Spec(t, s, &specC{
		batchSize: 100,
	})
}

type specC struct {
	batchSize int
}

func Spec(t *testing.T, s Streamer, c *specC) {
	t.Run("Fetch cannot return more than batch size", func(t *testing.T) {
		result := s.Fetch(c.batchSize)
		assert.Truef(t, c.batchSize >= len(result), "c.batchSize = %d && len(result) = %d", c.batchSize, len(result))
		assert.True(t, c.batchSize >= 1)
	})
}
