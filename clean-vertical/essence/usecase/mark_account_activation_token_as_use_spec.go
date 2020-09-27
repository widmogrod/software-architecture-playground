package usecase

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
	"time"
)

func SpecMarkAccountActivationTokenAsUse(t *testing.T) {
	ctx := context.Background()
	token := time.Now().String()

	t.Run("MarkAccountActivationTokenAsUse: For invalid token it should return appropriate error", func(t *testing.T) {
		cmd := MarkAccountActivationTokenAsUse{
			ActivationToken: token,
		}

		res := dispatch.Invoke(ctx, cmd)
		resact := res.(ResultOfMarkingAccountActivationTokenAsUsed)

		if assert.NotNil(t, resact.ValidationError) {
			assert.True(t, resact.ValidationError.InvalidToken)
		}

		assert.Nil(t, resact.SuccessfulResult)
		assert.False(t, resact.IsSuccessful())
	})

	t.Run("MarkAccountActivationTokenAsUse: prepare token", func(t *testing.T) {
		uuid := time.Now().String()
		res := dispatch.Invoke(ctx, CreateAccountActivationToken{
			UUID: uuid,
		})
		restok := res.(ResultOfCreateAccountActivationToken)

		t.Run("MarkAccountActivationTokenAsUse: activate existing token", func(t *testing.T) {
			cmd := MarkAccountActivationTokenAsUse{
				ActivationToken: restok.SuccessfulResult,
			}

			res = dispatch.Invoke(ctx, cmd)
			resact := res.(ResultOfMarkingAccountActivationTokenAsUsed)

			assert.Nil(t, resact.ValidationError)
			if assert.NotNil(t, resact.SuccessfulResult) {
				assert.Equal(t, resact.SuccessfulResult.UserUUID, uuid)
			}
		})
	})
}
