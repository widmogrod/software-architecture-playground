package usecase

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
)

func SpecConfirmAccountActivation(t *testing.T) {
	email := EmailAddress("user-eamil-2@example.com")
	ctx := context.Background()

	t.Run("ConfirmAccountActivation: When register new account, you must activate it and get access token", func(t *testing.T) {
		result := dispatch.Invoke(ctx, RegisterAccountWithEmail{EmailAddress: email})
		rorwe := result.(ResultOfRegisteringWithEmail)
		if !rorwe.IsSuccessful() {
			t.Fatal("fresh registration didn't succeed")
		}

		res := dispatch.Invoke(ctx, ConfirmAccountActivation{
			ActivationToken: rorwe.SuccessfulResult.ActivationTokenTestOnly,
		})
		resca := res.(ResultOfConfirmationOfAccountActivation)
		assert.Nil(t, resca.ValidationError)
		assert.NotNil(t, resca.SuccessfulResult)
	})

}
