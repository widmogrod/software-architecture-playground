package usecase

import (
	"context"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
	"testing"
)

func SpecRegisterAccountWithEmail(t *testing.T) {
	t.Run("RegisterAccountWithEmail: Registering same email two times in a row should not succeed", func(t *testing.T) {
		email := EmailAddress("user-eamil@example.com")

		ctx := context.Background()
		result := dispatch.Invoke(ctx, RegisterAccountWithEmail{EmailAddress: email})
		rorwe := result.(ResultOfRegisteringWithEmail)
		if !rorwe.IsSuccessful() {
			t.Fatal("fresh registration didn't succeed")
		}

		result = dispatch.Invoke(ctx, RegisterAccountWithEmail{EmailAddress: email})
		rorwe = result.(ResultOfRegisteringWithEmail)
		if rorwe.IsSuccessful() {
			t.Fatal("reuse of an email must not be allowed!")
		}
	})
}
