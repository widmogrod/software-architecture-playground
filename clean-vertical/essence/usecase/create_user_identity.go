package usecase

type CreateUserIdentity struct {
	UUID         string
	EmailAddress EmailAddress
}

type ResultOfCreateUserIdentity struct {
	ValidationError  *ValidationError
	SuccessfulResult *struct {
		UUID string
	}
}

type ValidationError struct {
	EmailAddressAlreadyExists bool
}

func (r *ResultOfCreateUserIdentity) IsSuccess() bool {
	return r.SuccessfulResult != nil && r.ValidationError == nil
}

func (r *ResultOfCreateUserIdentity) ConflictEmailExists() {
	if r.ValidationError == nil {
		r.ValidationError = &ValidationError{true}
	}

	r.ValidationError.EmailAddressAlreadyExists = true
}

func (r *ResultOfCreateUserIdentity) SucceedWithUUID(uuid string) {
	r.SuccessfulResult = &struct {
		UUID string
	}{
		uuid,
	}
}
