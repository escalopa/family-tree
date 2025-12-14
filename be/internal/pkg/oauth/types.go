package oauth

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
)

// OAuth Provider Constants
const (
	ProviderGoogle = "google"
	// Add more providers here
	// ProviderFacebook = "facebook"
)

// OAuthProvider interface for OAuth providers
type OAuthProvider interface {
	GetAuthURL(state string) string
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error)
	GetProviderName() string
}

// ProviderFactory is a function that creates a new OAuth provider instance
type ProviderFactory func(clientID, clientSecret, redirectURL string, scopes []string) OAuthProvider

// ProviderFactories maps provider names to their factory functions
var ProviderFactories = map[string]ProviderFactory{
	ProviderGoogle: NewGoogleProvider,
	// Add more providers here as needed
	// ProviderFacebook: NewFacebookProvider,
}
