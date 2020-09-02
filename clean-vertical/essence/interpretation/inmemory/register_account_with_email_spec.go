package inmemory

import (
	"../../algebra/dispatch"
	. "../../usecase"
	"context"
	"testing"
)

func SpecRegisterAccountWithEmail(t *testing.T) {
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
}
