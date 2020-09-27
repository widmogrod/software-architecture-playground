package inmemory

import (
	"github.com/widmogrod/software-architecture-playground/clean-vertical/essence/usecase"
	"testing"
)

func Specification(t *testing.T) {
	usecase.SpecHelloWorld(t)
	usecase.SpecRegisterAccountWithEmail(t)
	SpecConfirmAccountActivation(t)
}
