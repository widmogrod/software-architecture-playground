package usecase

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
	"time"
)

func SpecCreateAccountActivationToken(t *testing.T) {
	ctx := dispatch.Background()
	uuid := time.Now().String()

	t.Run("CreateAccountActivationToken: should always create activation token", func(t *testing.T) {
		cmd := CreateAccountActivationToken{
			UUID: uuid,
		}

		res := dispatch.Invoke(ctx, cmd)
		resat := res.(ResultOfCreateAccountActivationToken)

		assert.NotEmpty(t, resat.SuccessfulResult)
	})
}
