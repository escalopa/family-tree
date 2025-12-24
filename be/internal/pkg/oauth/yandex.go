package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

type YandexProvider struct {
	config      *oauth2.Config
	userInfoURL string
}

type yandexUserInfo struct {
	ID              string   `json:"id"`
	Login           string   `json:"login"`
	DefaultEmail    string   `json:"default_email"`
	Emails          []string `json:"emails"`
	DisplayName     string   `json:"display_name"`
	RealName        string   `json:"real_name"`
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	DefaultAvatarID string   `json:"default_avatar_id"`
	IsAvatarEmpty   bool     `json:"is_avatar_empty"`
}

func NewYandexProvider(clientID, clientSecret, redirectURL, userInfoURL string, scopes []string) OAuthProvider {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     yandex.Endpoint,
	}

	return &YandexProvider{
		config:      oauthConfig,
		userInfoURL: userInfoURL,
	}
}

func (y *YandexProvider) GetProviderName() string {
	return ProviderYandex
}

func (y *YandexProvider) GetAuthURL(state string) string {
	return y.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (y *YandexProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := y.config.Exchange(ctx, code)
	if err != nil {
		slog.Error("YandexProvider.Exchange: exchange code for token", "error", err)
		return nil, domain.NewExternalServiceError("Yandex OAuth", err)
	}
	return token, nil
}

func (y *YandexProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error) {
	client := y.config.Client(ctx, token)
	resp, err := client.Get(y.userInfoURL)
	if err != nil {
		slog.Error("YandexProvider.GetUserInfo: get user info from API", "error", err)
		return nil, domain.NewExternalServiceError("Yandex API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("YandexProvider.GetUserInfo: non-200 status code", "status_code", resp.StatusCode)
		return nil, domain.NewExternalServiceError("Yandex API", fmt.Errorf("status code %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("YandexProvider.GetUserInfo: read response body", "error", err)
		return nil, domain.NewInternalError("read response", err)
	}

	var yandexInfo yandexUserInfo
	if err := json.Unmarshal(data, &yandexInfo); err != nil {
		slog.Error("YandexProvider.GetUserInfo: unmarshal response", "error", err)
		return nil, domain.NewInternalError("parse response", err)
	}

	email := yandexInfo.DefaultEmail
	if email == "" && len(yandexInfo.Emails) > 0 {
		email = yandexInfo.Emails[0]
	}

	displayName := yandexInfo.DisplayName
	if displayName == "" {
		displayName = yandexInfo.RealName
	}

	picture := ""
	if !yandexInfo.IsAvatarEmpty && yandexInfo.DefaultAvatarID != "" {
		picture = fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", yandexInfo.DefaultAvatarID)
	}

	return &domain.OAuthUserInfo{
		ID:      yandexInfo.ID,
		Email:   email,
		Name:    displayName,
		Picture: picture,
	}, nil
}
