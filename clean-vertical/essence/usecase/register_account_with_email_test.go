package usecase

import (
	"../algebra/dispatch"
	"testing"
)

func Test_RegisterAccountWithEmail_InvalidEmail(t *testing.T) {
	res := HandleRegisterAccountWithEmail(RegisterAccountWithEmail{"ç$€§invalid-email!"})
	if res.ValidationError.EmailAddress.InvalidPattern != true {
		t.Error("email should be invalid")
	}
}

func Test_RegisterAccountWithEmail_EverythingFine(t *testing.T) {
	dispatch.ShouldInvokeAndReturn(t, func(t *testing.T, identity CreateUserIdentity) ResultOfCreateUserIdentity {
		if identity.EmailAddress != "email-isvalid@example.com" {
			t.Fatal("error email miss match")
		}

		return ResultOfCreateUserIdentity{
			SuccessfulResult: &struct {
				UUID string
			}{
				UUID: "some uuid",
			},
		}
	})

	res := HandleRegisterAccountWithEmail(RegisterAccountWithEmail{"email-isvalid@example.com"})
	if res.ValidationError.EmailAddress.InvalidPattern != false {
		t.Error("email should be valid")
	}
}
