package usecase

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
)

func init() {
	dispatch.RegisterGlobalHandler(HandleConfirmAccountActivation)
}

type ConfirmAccountActivation struct {
	ActivationToken string
}

type ResultOfConfirmationOfAccountActivation struct {
	ValidationError  *ConfirmAccountActivationValidationError
	SuccessfulResult *SessionToken
}

type AsyncResult struct {
	InvocationID string
}

type AsyncResultOfConfirmationOfAccountActivation struct {
	InvocationID string
	Status       string
	Result       *ResultOfConfirmationOfAccountActivation
}

type ConfirmAccountActivationValidationError struct {
	InvalidActivationToken bool
}

func NewInvalidActivationTokenError() *ConfirmAccountActivationValidationError {
	return &ConfirmAccountActivationValidationError{
		InvalidActivationToken: true,
	}
}

func HandleConfirmAccountActivation(ctx dispatch.Context, input ConfirmAccountActivation) ResultOfConfirmationOfAccountActivation {
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

	output.SuccessfulResult = &st.SuccessfulResult
	return output
}
