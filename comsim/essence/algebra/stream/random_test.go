package stream

import (
	"testing"
)

func TestRandomStream(t *testing.T) {
	s := NewRandomStream()
	Spec(t, s, &specC{
		batchSize: 100,
	})
}
