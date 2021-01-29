package usecase

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
	"time"
)

func SpecCreateUserIdentity(t *testing.T) {
	uuid := time.Now().String()
	uuid2 := time.Now().Add(1).String()
	email := EmailAddress("some-email-address@email.com")

	ctx := dispatch.Background()
	cmd := CreateUserIdentity{
		UUID:         uuid,
		EmailAddress: email,
	}

	t.Run("CreateUserIdentity: should create user identity when do not exists", func(t *testing.T) {
		res := dispatch.Invoke(ctx, cmd)
		resui := res.(ResultOfCreateUserIdentity)

		assert.NotNil(t, resui.SuccessfulResult)
		assert.Equal(t, resui.SuccessfulResult.UUID, uuid)

		t.Run("CreateUserIdentity: should prevent from creation of duplicates of emails even with different UUID", func(t *testing.T) {
			cmd = CreateUserIdentity{
				UUID:         uuid2,
				EmailAddress: email,
			}
			res := dispatch.Invoke(ctx, cmd)
			resui := res.(ResultOfCreateUserIdentity)
			assert.Nil(t, resui.SuccessfulResult)
			assert.NotNil(t, resui.ValidationError)
			assert.True(t, resui.ValidationError.EmailAddressAlreadyExists)
		})

	})
}
