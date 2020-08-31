package inmemory

import (
	"../../algebra/dispatch"
	"../../usecase"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	interpretation := New()
	dispatch.Interpret(interpretation)

	os.Exit(m.Run())
}

func Test_UsingExistingEmailDuringRegistration_Must_Fail(t *testing.T) {
	email := usecase.EmailAddress("user-eamil@example.com")

	result := dispatch.Invoke(usecase.RegisterAccountWithEmail{EmailAddress: email})
	rorwe := result.(usecase.ResultOfRegisteringWithEmail)
	if !rorwe.IsSuccessful() {
		t.Fatal("fresh registration didn't succeed")
	}

	result = dispatch.Invoke(usecase.RegisterAccountWithEmail{EmailAddress: email})
	rorwe = result.(usecase.ResultOfRegisteringWithEmail)
	if rorwe.IsSuccessful() {
		t.Fatal("reuse of an email must not be allowed!")
	}
}
