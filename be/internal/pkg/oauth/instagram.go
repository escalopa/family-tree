package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/instagram"
)

type InstagramProvider struct {
	config      *oauth2.Config
	userInfoURL string
}

type instagramUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func NewInstagramProvider(clientID, clientSecret, redirectURL, userInfoURL string, scopes []string) OAuthProvider {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     instagram.Endpoint,
	}

	return &InstagramProvider{
		config:      oauthConfig,
		userInfoURL: userInfoURL,
	}
}

func (i *InstagramProvider) GetProviderName() string {
	return ProviderInstagram
}

func (i *InstagramProvider) GetAuthURL(state string) string {
	return i.config.AuthCodeURL(state)
}

func (i *InstagramProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := i.config.Exchange(ctx, code)
	if err != nil {
		slog.Error("InstagramProvider.Exchange: exchange code for token", "error", err)
		return nil, domain.NewExternalServiceError("Instagram OAuth", err)
	}
	return token, nil
}

func (i *InstagramProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error) {
	client := i.config.Client(ctx, token)

	// Instagram Basic Display API
	url := fmt.Sprintf("%s?fields=id,username&access_token=%s", i.userInfoURL, token.AccessToken)
	resp, err := client.Get(url)
	if err != nil {
		slog.Error("InstagramProvider.GetUserInfo: get user info from API", "error", err)
		return nil, domain.NewExternalServiceError("Instagram API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("InstagramProvider.GetUserInfo: non-200 status code", "status_code", resp.StatusCode)
		return nil, domain.NewExternalServiceError("Instagram API", fmt.Errorf("status code %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("InstagramProvider.GetUserInfo: read response body", "error", err)
		return nil, domain.NewInternalError("read response", err)
	}

	var instagramInfo instagramUserInfo
	if err := json.Unmarshal(data, &instagramInfo); err != nil {
		slog.Error("InstagramProvider.GetUserInfo: unmarshal response", "error", err)
		return nil, domain.NewInternalError("parse response", err)
	}

	// Instagram Basic Display API doesn't provide email
	// For email access, you need Instagram Graph API with additional permissions
	return &domain.OAuthUserInfo{
		ID:      instagramInfo.ID,
		Email:   "", // Not available in Basic Display API
		Name:    instagramInfo.Username,
		Picture: "",
	}, nil
}
