package usecase

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
	"testing/quick"
)

func SpecHelloWorld(t *testing.T) {
	t.Run("HelloWorld: On every string input result should contain it", func(t *testing.T) {
		err := quick.Check(func(name string) bool {
			ctx := dispatch.Background()
			cmd := HelloWorld{
				Name: name,
			}

			res := dispatch.Invoke(ctx, cmd)

			if assert.IsType(t, ResultOfHelloWorld{}, res) {
				out := res.(ResultOfHelloWorld)
				if !assert.NotNil(t, out.SuccessfulResult, "result MUST NOT be nil") {
					return false
				}
				if !assert.Contains(t, out.SuccessfulResult, name, "MUST contains pass name") {
					return false
				}
			}

			return true
		}, nil)

		if err != nil {
			t.Error(err)
		}
	})
}
