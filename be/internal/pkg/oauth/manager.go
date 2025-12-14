package oauth

import (
	"context"
	"fmt"
	"log"

	"github.com/escalopa/family-tree/internal/config"
	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
)

type OAuthManager struct {
	providers map[string]OAuthProvider
}

func NewOAuthManager(cfg *config.OAuthConfig) *OAuthManager {
	manager := &OAuthManager{
		providers: make(map[string]OAuthProvider),
	}

	if googleCfg, ok := cfg.Providers["google"]; ok {
		manager.providers["google"] = NewGoogleProvider(
			googleCfg.ClientID,
			googleCfg.ClientSecret,
			cfg.GetRedirectURL("google"),
			googleCfg.Scopes,
		)
		log.Printf("Initialized Google OAuth provider")
	}

	// Add more providers here as needed
	// if facebookCfg, ok := cfg.Providers["facebook"]; ok {
	//     manager.providers["facebook"] = NewFacebookProvider(...)
	// }

	return manager
}

func (m *OAuthManager) GetProvider(providerName string) (OAuthProvider, error) {
	provider, ok := m.providers[providerName]
	if !ok {
		log.Printf("OAuth provider not found: %s", providerName)
		return nil, domain.NewValidationError(fmt.Sprintf("OAuth provider '%s' not supported", providerName))
	}
	return provider, nil
}

func (m *OAuthManager) GetAuthURL(providerName, state string) (string, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	return provider.GetAuthURL(state), nil
}

func (m *OAuthManager) HandleCallback(ctx context.Context, providerName, code string) (*oauth2.Token, *UserInfo, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, nil, err
	}

	// Exchange code for token
	token, err := provider.Exchange(ctx, code)
	if err != nil {
		log.Printf("OAuth exchange failed for provider %s: %v", providerName, err)
		return nil, nil, err
	}

	// Get user info
	userInfo, err := provider.GetUserInfo(ctx, token)
	if err != nil {
		log.Printf("Failed to get user info from %s: %v", providerName, err)
		return nil, nil, err
	}

	return token, userInfo, nil
}

func (m *OAuthManager) GetSupportedProviders() []string {
	providers := make([]string, 0, len(m.providers))
	for name := range m.providers {
		providers = append(providers, name)
	}
	return providers
}
