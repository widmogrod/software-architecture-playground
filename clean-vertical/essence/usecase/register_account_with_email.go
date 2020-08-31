package usecase

import (
	"../algebra/dispatch"
	"github.com/badoux/checkmail"
)

func init() {
	dispatch.When(RegisterAccountWithEmail{}, HandleRegisterAccountWithEmail)
}

type RegisterAccountWithEmail struct {
	EmailAddress EmailAddress
}

type ResultOfRegisteringWithEmail struct {
	ValidationError struct {
		EmailAddress struct {
			InvalidPattern bool
		}
	}
	SuccessfulResult struct {
		PleaseConfirmEmailLink bool
	}
}

type EmailAddress string

func (e EmailAddress) IsValid() bool {
	return checkmail.ValidateFormat(string(e)) == nil
}

func HandleRegisterAccountWithEmail(input RegisterAccountWithEmail) ResultOfRegisteringWithEmail {
	output := ResultOfRegisteringWithEmail{}

	if !input.EmailAddress.IsValid() {
		output.ValidationError.EmailAddress.InvalidPattern = true
		return output
	}

	res := dispatch.Dispatch(CreateUserIdentity{
		UUID:         "todo-generate-uuid",
		EmailAddress: input.EmailAddress,
	})

	if rocui, ok := res.(ResultOfCreateUserIdentity); ok && rocui.IsSuccess() {
		output.SuccessfulResult.PleaseConfirmEmailLink = true
		return output
	}

	panic("Never reach this path")
}
