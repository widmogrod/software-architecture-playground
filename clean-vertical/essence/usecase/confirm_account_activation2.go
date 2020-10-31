package usecase

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
)

func AsyncHandleConfirmAccountActivation(ctx dispatch.Context, input ConfirmAccountActivation) AsyncResult {
	wf := dispatch.NewFlowAround(ResultOfConfirmationOfAccountActivation{})
	wf.OnEffect(func(atu ResultOfMarkingAccountActivationTokenAsUsed) *dispatch.ActivityResult {
		return wf.
			If(func() bool {
				return !atu.IsSuccessful() && atu.ValidationError.InvalidToken
			}).
			Then(wf.End.With(func(aggregate ResultOfConfirmationOfAccountActivation) ResultOfConfirmationOfAccountActivation {
				aggregate.ValidationError = NewInvalidActivationTokenError()
				return aggregate
			})).
			Else(wf.Invoke.With(func() GenerateSessionToken {
				return GenerateSessionToken{
					UserUUID: atu.SuccessfulResult.UserUUID,
				}
			}))
	})
	wf.OnEffect(func(st ResultOfGeneratingSessionToken) *dispatch.ActivityResult {
		return wf.End.With(func(aggregate ResultOfConfirmationOfAccountActivation) ResultOfConfirmationOfAccountActivation {
			aggregate.SuccessfulResult = &st.SuccessfulResult
			return aggregate
		})
	})
	wf.OnFailure(func() {
		// all failures
	})

	result := wf.Run(func() MarkAccountActivationTokenAsUse {
		return MarkAccountActivationTokenAsUse{
			ActivationToken: input.ActivationToken,
		}
	})

	wf.Log()

	return AsyncResult{
		InvocationID: result.InvocationID(),
	}
}
