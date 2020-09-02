package inmemory

import (
	"testing"
)

func Specification(t *testing.T) {
	SpecRegisterAccountWithEmail(t)
	SpecConfirmAccountActivation(t)
}
