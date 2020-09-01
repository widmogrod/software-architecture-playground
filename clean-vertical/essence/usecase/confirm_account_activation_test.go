package usecase

import (
	"../algebra/dispatch"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_HandleConfirmAccountActivation_token_is_invalid(t *testing.T) {
	token := "invalid-token"
	dispatch.ShouldInvokeAndReturn(t, func(t *testing.T, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed {
		assert.Equal(t, token, input.ActivationToken)
		return ResultOfMarkingAccountActivationTokenAsUsed{
			ValidationError: &struct {
				InvalidToken bool
			}{
				InvalidToken: true,
			},
		}
	})
	res := HandleConfirmAccountActivation(ConfirmAccountActivation{
		ActivationToken: token,
	})
	if !res.ValidationError.InvalidActivationToken {
		t.Fatal("invalid activation token must return ValidationError")
	}
}

func Test_HandleConfirmAccountActivation_successful(t *testing.T) {
	userUUID := "asd-21d-d21-d21"
	token := "valid-token"
	accessToken := "accessToken"
	refreshToken := "refreshToken"
	dispatch.ShouldInvokeAndReturn(t, func(t *testing.T, input MarkAccountActivationTokenAsUse) ResultOfMarkingAccountActivationTokenAsUsed {
		assert.Equal(t, token, input.ActivationToken)
		return ResultOfMarkingAccountActivationTokenAsUsed{
			SuccessfulResult: &struct {
				UserUUID string
			}{
				UserUUID: userUUID,
			},
		}
	})
	dispatch.ShouldInvokeAndReturn(t, func(t *testing.T, input GenerateSessionToken) ResultOfGeneratingSessionToken {
		assert.Equal(t, userUUID, input.UserUUID)
		return ResultOfGeneratingSessionToken{
			SuccessfulResult: SessionToken{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		}
	})
	res := HandleConfirmAccountActivation(ConfirmAccountActivation{
		ActivationToken: token,
	})

	assert.Equal(t, accessToken, res.SuccessfulResult.AccessToken)
	assert.Equal(t, refreshToken, res.SuccessfulResult.RefreshToken)
}
