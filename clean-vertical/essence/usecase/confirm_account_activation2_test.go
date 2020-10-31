package usecase

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
)

func TestAsyncHandleConfirmAccountActivation(t *testing.T) {
	token := "actication token"
	ctx := dispatch.Background()
	res := AsyncHandleConfirmAccountActivation(ctx, ConfirmAccountActivation{
		ActivationToken: token,
	})

	assert.NotEmpty(t, res.InvocationID)

	t.Log("TODO: pooling for result")
}
