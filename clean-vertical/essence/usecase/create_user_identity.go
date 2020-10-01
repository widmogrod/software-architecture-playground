package usecase

type CreateUserIdentity struct {
	UUID         string
	EmailAddress EmailAddress
}

type ResultOfCreateUserIdentity struct {
	ValidationError  *CreateUserIdentityValidationError
	SuccessfulResult *CreateUserIdentitySuccessfulResult
}

type CreateUserIdentityValidationError struct {
	EmailAddressAlreadyExists bool
}

type CreateUserIdentitySuccessfulResult struct {
	UUID string
}

func NewConflictEmailExistsError() *CreateUserIdentityValidationError {
	return &CreateUserIdentityValidationError{
		EmailAddressAlreadyExists: true,
	}
}

func NewCreateUserIdentityWithUUID(uuid string) *CreateUserIdentitySuccessfulResult {
	return &CreateUserIdentitySuccessfulResult{
		UUID: uuid,
	}
}

func (r *ResultOfCreateUserIdentity) IsSuccess() bool {
	return r.SuccessfulResult != nil && r.ValidationError == nil
}
