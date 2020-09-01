package usecase

type GenerateSessionToken struct {
	UserUUID string
}

type ResultOfGeneratingSessionToken struct {
	SuccessfulResult SessionToken
}

type SessionToken struct {
	AccessToken  string
	RefreshToken string
}
