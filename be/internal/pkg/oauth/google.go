package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProvider struct {
	config      *oauth2.Config
	userInfoURL string
}

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func NewGoogleProvider(clientID, clientSecret, redirectURL, userInfoURL string, scopes []string) OAuthProvider {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	return &GoogleProvider{
		config:      oauthConfig,
		userInfoURL: userInfoURL,
	}
}

func (g *GoogleProvider) GetProviderName() string {
	return ProviderGoogle
}

func (g *GoogleProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (g *GoogleProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		slog.Error("GoogleProvider.Exchange: exchange code for token", "error", err)
		return nil, domain.NewExternalServiceError(err)
	}
	return token, nil
}

func (g *GoogleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error) {
	client := g.config.Client(ctx, token)
	resp, err := client.Get(g.userInfoURL)
	if err != nil {
		slog.Error("GoogleProvider.GetUserInfo: get user info from API", "error", err)
		return nil, domain.NewExternalServiceError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("GoogleProvider.GetUserInfo: non-200 status code", "status_code", resp.StatusCode)
		return nil, domain.NewExternalServiceError(fmt.Errorf("status code %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("GoogleProvider.GetUserInfo: read response body", "error", err)
		return nil, domain.NewInternalError(err)
	}

	var googleInfo googleUserInfo
	if err := json.Unmarshal(data, &googleInfo); err != nil {
		slog.Error("GoogleProvider.GetUserInfo: unmarshal response", "error", err)
		return nil, domain.NewInternalError(err)
	}

	return &domain.OAuthUserInfo{
		ID:      googleInfo.ID,
		Email:   googleInfo.Email,
		Name:    googleInfo.Name,
		Picture: googleInfo.Picture,
	}, nil
}
