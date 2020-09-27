package usecase

type MarkAccountActivationTokenAsUse struct {
	ActivationToken string
}

type ResultOfMarkingAccountActivationTokenAsUsed struct {
	ValidationError  *MarkAccountActivationTokenAsUseValidationError
	SuccessfulResult *MarkAccountActivationTokenAsUseSuccessfulResult
}

type MarkAccountActivationTokenAsUseValidationError struct {
	InvalidToken bool
}

type MarkAccountActivationTokenAsUseSuccessfulResult struct {
	UserUUID string
}

func (r ResultOfMarkingAccountActivationTokenAsUsed) IsSuccessful() bool {
	return r.SuccessfulResult != nil
}

func NewAccountActivationInvalidTokenError() *MarkAccountActivationTokenAsUseValidationError {
	return &MarkAccountActivationTokenAsUseValidationError{
		InvalidToken: true,
	}
}

func NewAccountActivatedViaTokenSuccess(uuid string) *MarkAccountActivationTokenAsUseSuccessfulResult {
	return &MarkAccountActivationTokenAsUseSuccessfulResult{
		UserUUID: uuid,
	}
}
