package usecase

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/algebra/dispatch"
)

func AsyncHandleConfirmAccountActivation(ctx dispatch.Context, input ConfirmAccountActivation) AsyncResult {
	wf := dispatch.NewFlowAround(ResultOfConfirmationOfAccountActivation{})
	wf.
		OnEffect(ResultOfMarkingAccountActivationTokenAsUsed{}).
		Activity(wf.
			If(func(atu ResultOfMarkingAccountActivationTokenAsUsed) bool {
				return !atu.IsSuccessful() && atu.ValidationError.InvalidToken
			}).
			Then(wf.End.With(func(ctx, aggregate ResultOfConfirmationOfAccountActivation) ResultOfConfirmationOfAccountActivation {
				aggregate.ValidationError = NewInvalidActivationTokenError()
				return aggregate
			})).
			Else(wf.Invoke.With(func(ctx, atu ResultOfMarkingAccountActivationTokenAsUsed) (GenerateSessionToken, ResultOfGeneratingSessionToken) {
				return GenerateSessionToken{
					UserUUID: atu.SuccessfulResult.UserUUID,
				}, ResultOfGeneratingSessionToken{}
			})))

	wf.
		OnEffect(ResultOfGeneratingSessionToken{}).
		Activity(wf.End.With(func(ctx ResultOfGeneratingSessionToken, aggregate ResultOfConfirmationOfAccountActivation) ResultOfConfirmationOfAccountActivation {
			aggregate.SuccessfulResult = &ctx.SuccessfulResult
			return aggregate
		}))

	wf.OnFailure(func() {
		// all failures
	})

	result := wf.Run(func() (MarkAccountActivationTokenAsUse, ResultOfMarkingAccountActivationTokenAsUsed) {
		return MarkAccountActivationTokenAsUse{
			ActivationToken: input.ActivationToken,
		}, ResultOfMarkingAccountActivationTokenAsUsed{}
	})

	fmt.Println(dispatch.ToPlantText(wf))

	return AsyncResult{
		InvocationID: result.InvocationID(),
	}
}
