package domain

// TokenClaims represents the token claims used across the application
type TokenClaims struct {
	UserID    int
	Email     string
	RoleID    int
	SessionID string
}

// AuthTokens represents authentication tokens
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	SessionID    string
}

// MemberFilter is used for searching members
type MemberFilter struct {
	ArabicName  *string
	EnglishName *string
	Gender      *string
	IsMarried   *bool
}
