package usecase

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
	"time"
)

func SpecGenerateSessionToken(t *testing.T) {
	uuid := time.Now().String()
	t.Run("GenerateSessionToken: should generate session token", func(t *testing.T) {
		ctx := dispatch.Background()
		cmd := GenerateSessionToken{
			UserUUID: uuid,
		}

		res := dispatch.Invoke(ctx, cmd)
		resgst := res.(ResultOfGeneratingSessionToken)
		assert.NotEmpty(t, resgst.SuccessfulResult.AccessToken, "AccessToken is empty")
		assert.NotEmpty(t, resgst.SuccessfulResult.RefreshToken, "RefreshToken is empty")
		assert.NotEqual(t, resgst.SuccessfulResult.AccessToken, resgst.SuccessfulResult.RefreshToken, "AccessToken must be different to RefreshToken")
	})

	t.Run("GenerateSessionToken: consecutive token generation should create unique tokens each time", func(t *testing.T) {
		ctx := dispatch.Background()
		cmd := GenerateSessionToken{
			UserUUID: uuid,
		}

		token1 := dispatch.Invoke(ctx, cmd).(ResultOfGeneratingSessionToken).SuccessfulResult
		token2 := dispatch.Invoke(ctx, cmd).(ResultOfGeneratingSessionToken).SuccessfulResult

		assert.NotEqual(t, token1.AccessToken, token2.AccessToken, "newly created AccessToken is identical to previous and MUST NOT be")
		assert.NotEqual(t, token1.RefreshToken, token2.RefreshToken, "newly created RefreshToken is identical to previous and MUST NOT be")
	})
}
