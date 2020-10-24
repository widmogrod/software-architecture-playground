package usecase

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
)

func Test_HandleConfirmAccountActivation_token_is_invalid(t *testing.T) {
	token := "invalid-token"
	dispatch.ShouldInvokeAndReturn(t, func(t *testing.T, _ dispatch.Context, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed {
		assert.Equal(t, token, input.ActivationToken)
		return ResultOfMarkingAccountActivationTokenAsUsed{
			ValidationError: NewAccountActivationInvalidTokenError(),
		}
	})

	ctx := dispatch.Background()
	res := HandleConfirmAccountActivation(ctx, ConfirmAccountActivation{
		ActivationToken: token,
	})
	if !res.ValidationError.InvalidActivationToken {
		t.Fatal("invalid activation token must return CreateUserIdentityValidationError")
	}
}

func Test_HandleConfirmAccountActivation_successful(t *testing.T) {
	userUUID := "asd-21d-d21-d21"
	token := "valid-token"
	accessToken := "accessToken"
	refreshToken := "refreshToken"
	dispatch.ShouldInvokeAndReturn(t, func(t *testing.T, _ dispatch.Context, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed {
		assert.Equal(t, token, input.ActivationToken)
		return ResultOfMarkingAccountActivationTokenAsUsed{
			SuccessfulResult: NewAccountActivatedViaTokenSuccess(userUUID),
		}
	})
	dispatch.ShouldInvokeAndReturn(t, func(t *testing.T, _ dispatch.Context, input GenerateSessionToken) ResultOfGeneratingSessionToken {
		assert.Equal(t, userUUID, input.UserUUID)
		return ResultOfGeneratingSessionToken{
			SuccessfulResult: SessionToken{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		}
	})

	ctx := dispatch.Background()
	res := HandleConfirmAccountActivation(ctx, ConfirmAccountActivation{
		ActivationToken: token,
	})

	assert.Equal(t, accessToken, res.SuccessfulResult.AccessToken)
	assert.Equal(t, refreshToken, res.SuccessfulResult.RefreshToken)
}
