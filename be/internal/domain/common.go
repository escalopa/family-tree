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
	Name      *string
	Gender    *string
	IsMarried *bool
}
