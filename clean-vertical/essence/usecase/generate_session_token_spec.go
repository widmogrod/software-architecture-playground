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
		assert.NotEmpty(t, resgst.SuccessfulResult.AccessToken, "")
		assert.NotEmpty(t, resgst.SuccessfulResult.RefreshToken, "")
		assert.NotEqual(t, resgst.SuccessfulResult.AccessToken, resgst.SuccessfulResult.RefreshToken)
	})
}
