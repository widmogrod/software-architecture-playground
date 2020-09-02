package inmemory

import (
	"../../algebra/dispatch"
	. "../../usecase"
	"context"
	"testing"
)

func SpecConfirmAccountActivation(t *testing.T) {
	email := EmailAddress("user-eamil@example.com")

	ctx := context.Background()
	result := dispatch.Invoke(ctx, RegisterAccountWithEmail{EmailAddress: email})
	rorwe := result.(ResultOfRegisteringWithEmail)
	if !rorwe.IsSuccessful() {
		//t.Fatal("fresh registration didn't succeed")
	}

	//res := dispatch.Invoke(ConfirmAccountActivation{
	//	ActivationToken: "",
	//})
	//
	//_ = res.(ResultOfConfirmationOfAccountActivation)
	//caa.SuccessfulResult.RefreshToken

}
