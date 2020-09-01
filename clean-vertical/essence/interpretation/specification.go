package interpretation

import (
	"../usecase"
	"testing"
)

func Specification(t *testing.T) {
	usecase.SpecCreateUserIdentity(t)
}
