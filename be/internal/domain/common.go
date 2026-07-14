package domain

type TokenClaims struct {
	UserID    int
	SessionID string
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	SessionID    string
}

type MemberFilter struct {
	TreeID      int
	Name        *string
	ArabicName  *string
	EnglishName *string
	Gender      *string
	IsMarried   *bool
}
