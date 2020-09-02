package usecase

import (
	"context"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
)

func init() {
	dispatch.Register(HandleConfirmAccountActivation)
}

type ConfirmAccountActivation struct {
	ActivationToken string
}

type ResultOfConfirmationOfAccountActivation struct {
	ValidationError *struct {
		InvalidActivationToken bool
	}
	SuccessfulResult *SessionToken
}

func (r *ResultOfConfirmationOfAccountActivation) SessionToken(token *SessionToken) {
	r.SuccessfulResult = token
}

func NewInvalidActivationTokenError() *struct{ InvalidActivationToken bool } {
	return &struct {
		InvalidActivationToken bool
	}{
		InvalidActivationToken: true,
	}
}

func HandleConfirmAccountActivation(ctx context.Context, input ConfirmAccountActivation) ResultOfConfirmationOfAccountActivation {
	output := ResultOfConfirmationOfAccountActivation{}

	res := dispatch.Invoke(ctx, MarkAccountActivationTokenAsUse{
		ActivationToken: input.ActivationToken,
	})
	atu := res.(ResultOfMarkingAccountActivationTokenAsUsed)
	if !atu.IsSuccessful() && atu.ValidationError.InvalidToken {
		output.ValidationError = NewInvalidActivationTokenError()
		return output
	}

	res = dispatch.Invoke(ctx, GenerateSessionToken{
		UserUUID: atu.SuccessfulResult.UserUUID,
	})
	st := res.(ResultOfGeneratingSessionToken)

	output.SessionToken(&st.SuccessfulResult)

	return output
}
