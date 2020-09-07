package usecase

import (
	"context"
	"github.com/badoux/checkmail"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
)

func init() {
	dispatch.Register(HandleRegisterAccountWithEmail)
}

type RegisterAccountWithEmail struct {
	EmailAddress EmailAddress
}

type ResultOfRegisteringWithEmail struct {
	ValidationError *struct {
		EmailAddress struct {
			InvalidPattern bool
			InUse          bool
		}
	}
	SuccessfulResult *struct {
		PleaseConfirmEmailLink bool
	}
}

func (r *ResultOfRegisteringWithEmail) IsSuccessful() bool {
	return r.SuccessfulResult != nil && r.SuccessfulResult.PleaseConfirmEmailLink == true
}

type EmailAddress string

func (e EmailAddress) IsValid() bool {
	return checkmail.ValidateFormat(string(e)) == nil
}

// TODO refactor this abomination
func NewInvalidEmailPatternError() *struct {
	EmailAddress struct {
		InvalidPattern bool
		InUse          bool
	}
} {
	return &struct {
		EmailAddress struct {
			InvalidPattern bool
			InUse          bool
		}
	}{
		EmailAddress: struct {
			InvalidPattern bool
			InUse          bool
		}{
			InvalidPattern: true,
		},
	}
}

func NewEmailInUserError() *struct {
	EmailAddress struct {
		InvalidPattern bool
		InUse          bool
	}
} {
	return &struct {
		EmailAddress struct {
			InvalidPattern bool
			InUse          bool
		}
	}{
		EmailAddress: struct {
			InvalidPattern bool
			InUse          bool
		}{
			InUse: true,
		},
	}
}

func HandleRegisterAccountWithEmail(ctx context.Context, input RegisterAccountWithEmail) ResultOfRegisteringWithEmail {
	output := ResultOfRegisteringWithEmail{}

	if !input.EmailAddress.IsValid() {
		output.ValidationError = NewInvalidEmailPatternError()
		return output
	}

	res := dispatch.Invoke(ctx, CreateUserIdentity{
		UUID:         "todo-generate-uuid",
		EmailAddress: input.EmailAddress,
	})
	rocui := res.(ResultOfCreateUserIdentity)
	if !rocui.IsSuccess() && rocui.ValidationError.EmailAddressAlreadyExists {
		output.ValidationError = NewEmailInUserError()
		return output
	}

	output.SuccessfulResult = &struct{ PleaseConfirmEmailLink bool }{PleaseConfirmEmailLink: true}
	return output
}
