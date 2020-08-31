package usecase

import "../algebra/dispatch"

func init() {
	dispatch.When(CreateUserIdentity{}, HandleCreateUserIdentity)
}

type CreateUserIdentity struct {
	UUID         string
	EmailAddress EmailAddress
}

type ResultOfCreateUserIdentity struct {
	ValidationError struct {
		EmailAddressAlreadyExists bool
	}
	SuccessfulResult *struct {
		UUID string
	}
}

func (r ResultOfCreateUserIdentity) IsSuccess() bool {
	return r.SuccessfulResult != nil
}

func HandleCreateUserIdentity(c CreateUserIdentity) ResultOfCreateUserIdentity {
	// TODO implement
	return ResultOfCreateUserIdentity{}
}
