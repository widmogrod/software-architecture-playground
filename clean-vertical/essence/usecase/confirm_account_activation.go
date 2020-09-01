package usecase

import (
	"../algebra/dispatch"
)

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

func HandleConfirmAccountActivation(input ConfirmAccountActivation) ResultOfConfirmationOfAccountActivation {
	output := ResultOfConfirmationOfAccountActivation{}

	res := dispatch.Invoke(MarkAccountActivationTokenAsUse{
		ActivationToken: input.ActivationToken,
	})
	atu := res.(ResultOfMarkingAccountActivationTokenAsUsed)
	if !atu.IsSuccessful() && atu.ValidationError.InvalidToken {
		output.ValidationError = NewInvalidActivationTokenError()
		return output
	}

	res = dispatch.Invoke(GenerateSessionToken{
		UserUUID: atu.SuccessfulResult.UserUUID,
	})
	st := res.(ResultOfGeneratingSessionToken)

	output.SessionToken(&st.SuccessfulResult)

	return output
}
