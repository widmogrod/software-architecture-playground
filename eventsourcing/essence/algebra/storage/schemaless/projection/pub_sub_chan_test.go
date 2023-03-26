package projection

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPubSubChan(t *testing.T) {
	psc := NewPubSubChan[string]()
	go psc.Process()

	var err = errors.New("foo")

	done := make(chan struct{})

	go func() {
		defer close(done)
		err2 := psc.Subscribe(func(msg string) error {
			assert.Equal(t, "foo", msg)
			return err
		})
		assert.Error(t, err2, err)
	}()

	err3 := psc.Publish("foo")
	assert.NoError(t, err3)

	<-done
}
