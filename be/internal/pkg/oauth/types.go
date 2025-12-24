package oauth

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
)

const (
	ProviderGoogle    = "google"
	ProviderYandex    = "yandex"
	ProviderVK        = "vk"
	ProviderFacebook  = "facebook"
	ProviderInstagram = "instagram"
	ProviderGitHub    = "github"
	ProviderGitLab    = "gitlab"
	ProviderLinkedIn  = "linkedin"
)

type OAuthProvider interface {
	GetAuthURL(state string) string
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error)
	GetProviderName() string
}

type ProviderFactory func(clientID, clientSecret, redirectURL, userInfoURL string, scopes []string) OAuthProvider

var ProviderFactories = map[string]ProviderFactory{
	ProviderGoogle:    NewGoogleProvider,
	ProviderYandex:    NewYandexProvider,
	ProviderVK:        NewVKProvider,
	ProviderFacebook:  NewFacebookProvider,
	ProviderInstagram: NewInstagramProvider,
	ProviderGitHub:    NewGitHubProvider,
	ProviderGitLab:    NewGitLabProvider,
	ProviderLinkedIn:  NewLinkedInProvider,
}
