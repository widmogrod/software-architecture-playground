package usecase

type MarkAccountActivationTokenAsUse struct {
	ActivationToken string
}

type ResultOfMarkingAccountActivationTokenAsUsed struct {
	ValidationError *struct {
		InvalidToken bool
	}
	SuccessfulResult *struct {
		UserUUID string
	}
}

func (r ResultOfMarkingAccountActivationTokenAsUsed) IsSuccessful() bool {
	return r.SuccessfulResult != nil
}
