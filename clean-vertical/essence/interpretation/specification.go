package interpretation

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
	"testing"
)

func Specification(t *testing.T) {
	usecase.SpecHelloWorld(t)
	usecase.SpecRegisterAccountWithEmail(t)
	usecase.SpecConfirmAccountActivation(t)
	usecase.SpecCreateUserIdentity(t)
	usecase.SpecGenerateSessionToken(t)
}
