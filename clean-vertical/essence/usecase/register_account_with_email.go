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
	// TODO private property?
	SuccessfulResult *RegisterAccountWithEmailSuccessfulResult
}

type RegisterAccountWithEmailSuccessfulResult struct {
	PleaseConfirmEmailLink bool
	// *TestOnly - is for convenience of spec tests, this value MUST never be expose
	ActivationTokenTestOnly string
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

func NewConfirmEmailLinkSuccess(token string) *RegisterAccountWithEmailSuccessfulResult {
	return &RegisterAccountWithEmailSuccessfulResult{
		PleaseConfirmEmailLink:  true,
		ActivationTokenTestOnly: token,
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

	tok := dispatch.Invoke(ctx, CreateAccountActivationToken{
		UUID: rocui.SuccessfulResult.UUID,
	})
	tokres := tok.(ResultOfCreateAccountActivationToken)
	output.SuccessfulResult = NewConfirmEmailLinkSuccess(tokres.SuccessfulResult)
	return output
}
