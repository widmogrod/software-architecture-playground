package usecase

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
	"time"
)

func Test_RegisterAccountWithEmail_InvalidEmail(t *testing.T) {
	ctx := context.Background()
	res := HandleRegisterAccountWithEmail(ctx, RegisterAccountWithEmail{"ç$€§invalid-email!"})
	if res.ValidationError.EmailAddress.InvalidPattern != true {
		t.Error("email should be invalid")
	}
}

func Test_RegisterAccountWithEmail_EverythingFine(t *testing.T) {
	uuid := time.Now().String()
	email := EmailAddress("email-isvalid@example.com")
	dispatch.ShouldInvokeAndReturn(t, func(t *testing.T, _ context.Context, identity CreateUserIdentity) ResultOfCreateUserIdentity {
		if !assert.Equal(t, identity.EmailAddress, email) {
			t.Fatal("error email miss match")
		}

		res := ResultOfCreateUserIdentity{}
		res.SuccessfulResult = NewCreateUserIdentityWithUUID(uuid)
		return res
	})

	dispatch.ShouldInvokeAndReturn(t, func(t *testing.T, _ context.Context, input CreateAccountActivationToken) ResultOfCreateAccountActivationToken {
		if !assert.Equal(t, uuid, input.UUID) {
			t.Fatal("UUIDs must match")
		}

		return ResultOfCreateAccountActivationToken{
			SuccessfulResult: input.UUID,
		}
	})

	ctx := context.Background()
	res := HandleRegisterAccountWithEmail(ctx, RegisterAccountWithEmail{email})
	if res.ValidationError != nil && res.ValidationError.EmailAddress.InvalidPattern != false {
		t.Error("email should be valid")
	}
}
