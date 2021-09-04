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

// HandlerCreateUserIdentity is an conceptualisation exercise
// where handlers have implementation detail, and how they could look like.
// In this example you may notice that uniqueness of email is business
// requirement that can be enforce by storing user identity under Key
// which value is an email address, but this also implicitly limit
// how user can be retrieved form storage - and that is - user identity can be retrieved only by email address.
// Such thing can limit other business needs, and therefore is not recommended
// This should be important and visible implementation detail, but not express in business logic.
// Good place is to have it in essence/implementation/<your-impl>
//func HandlerCreateUserIdentity(ctx dispatch.Context, input CreateUserIdentity) ResultOfCreateUserIdentity {
//	output := ResultOfCreateUserIdentity{}
//
//	res := dispatch.Invoke(ctx, KeyExists{
//		Key: input.EmailAddress,
//	})
//	reskv := res.(KexExistsResult)
//
//	if reskv.Exists {
//		output.ValidationError = NewConflictEmailExistsError()
//		return output
//	}
//
//	res = dispatch.Invoke(ctx, KVSet{
//		Key: input.EmailAddress,
//		Right: UserIdentityAggregate{
//			UUID:         input.UUID,
//			EmailAddress: input.EmailAddress,
//		},
//	})
//
//	return output
//}
