package usecase

import (
	"../algebra/dispatch"
	"testing"
)

func SpecCreateUserIdentity(t *testing.T) {
	email := EmailAddress("user-eamil@example.com")

	result := dispatch.Invoke(RegisterAccountWithEmail{EmailAddress: email})
	rorwe := result.(ResultOfRegisteringWithEmail)
	if !rorwe.IsSuccessful() {
		t.Fatal("fresh registration didn't succeed")
	}

	result = dispatch.Invoke(RegisterAccountWithEmail{EmailAddress: email})
	rorwe = result.(ResultOfRegisteringWithEmail)
	if rorwe.IsSuccessful() {
		t.Fatal("reuse of an email must not be allowed!")
	}
}
