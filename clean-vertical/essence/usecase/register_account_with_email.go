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
			InUse          bool
		}
	}
	SuccessfulResult struct {
		PleaseConfirmEmailLink bool
	}
}

func (r *ResultOfRegisteringWithEmail) IsSuccessful() bool {
	return r.SuccessfulResult.PleaseConfirmEmailLink == true
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

	res := dispatch.Invoke(CreateUserIdentity{
		UUID:         "todo-generate-uuid",
		EmailAddress: input.EmailAddress,
	})

	rocui := res.(ResultOfCreateUserIdentity)
	if !rocui.IsSuccess() && rocui.ValidationError.EmailAddressAlreadyExists {
		output.ValidationError.EmailAddress.InUse = true
		return output
	}

	output.SuccessfulResult.PleaseConfirmEmailLink = true
	return output
}
