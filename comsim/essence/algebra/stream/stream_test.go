package stream

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestChannelStream(t *testing.T) {
	s := NewChannelStream()
	go s.Work()

	size := 10
	for range rand.Perm(size) {
		s.Push(*GenerateMessage())
	}

	Spec(t, s, &specC{
		batchSize: size,
	})
}

type specC struct {
	batchSize int
}

func Spec(t *testing.T, s Streamer, c *specC) {
	t.Run("Push", func(t *testing.T) {
		for range rand.Perm(c.batchSize) {
			s.Push(*GenerateMessage())
		}
	})
	t.Run("Fetch cannot return more than batch size", func(t *testing.T) {
		result := s.Fetch(c.batchSize)
		assert.Truef(t, c.batchSize >= len(result), "c.batchSize = %d && len(result) = %d", c.batchSize, len(result))
		assert.True(t, c.batchSize >= 1)
	})
}
