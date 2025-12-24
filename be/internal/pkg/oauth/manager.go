package oauth

import (
	"context"
	"log/slog"

	"github.com/escalopa/family-tree/internal/config"
	"github.com/escalopa/family-tree/internal/domain"
)

type OAuthManager struct {
	providers     map[string]OAuthProvider
	providerOrder []string
}

func NewOAuthManager(cfg *config.OAuthConfig) *OAuthManager {
	manager := &OAuthManager{
		providers:     make(map[string]OAuthProvider),
		providerOrder: cfg.GetProviderOrder(),
	}

	for _, providerName := range manager.providerOrder {
		providerCfg := cfg.Providers[providerName]

		factory, exists := ProviderFactories[providerName]
		if !exists {
			slog.Warn("OAuthManager.NewOAuthManager: provider factory not found", "provider", providerName)
			continue
		}

		redirectURL := cfg.GetRedirectURL(providerName)
		provider := factory(providerCfg.ClientID, providerCfg.ClientSecret, redirectURL, providerCfg.UserInfoURL, providerCfg.Scopes)
		manager.providers[providerName] = provider
		manager.providerOrder = append(manager.providerOrder, providerName)

		slog.Info("OAuthManager.NewOAuthManager: initialized provider", "provider", providerName, "order", providerCfg.Order)
	}

	return manager
}

func (m *OAuthManager) GetProvider(providerName string) (OAuthProvider, error) {
	provider, ok := m.providers[providerName]
	if !ok {
		slog.Warn("OAuthManager.GetProvider: OAuth provider not supported", "provider", providerName)
		return nil, domain.NewValidationError("error.oauth.provider_not_supported", map[string]string{"provider": providerName})
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

func (m *OAuthManager) GetUserInfo(ctx context.Context, providerName, code string) (*domain.OAuthUserInfo, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	token, err := provider.Exchange(ctx, code)
	if err != nil {
		slog.Error("OAuthManager.GetUserInfo: exchange code for token", "provider", providerName, "error", err)
		return nil, err
	}

	userInfo, err := provider.GetUserInfo(ctx, token)
	if err != nil {
		slog.Error("OAuthManager.GetUserInfo: get user info from provider", "provider", providerName, "error", err)
		return nil, err
	}

	return userInfo, nil
}

func (m *OAuthManager) GetSupportedProviders() []string {
	return m.providerOrder
}
