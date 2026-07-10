package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/escalopa/family-tree/internal/config"
	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
)

type GenericProvider struct {
	name         string
	config       *oauth2.Config
	userInfoURL  string
	idField      string
	emailField   string
	nameField    string
	pictureField string
}

func NewGenericProvider(providerName string, providerCfg config.OAuthProviderConfig, redirectURL string) OAuthProvider {
	oauthConfig := &oauth2.Config{
		ClientID:     providerCfg.ClientID,
		ClientSecret: providerCfg.ClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       providerCfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  providerCfg.AuthURL,
			TokenURL: providerCfg.TokenURL,
		},
	}

	return &GenericProvider{
		name:         providerName,
		config:       oauthConfig,
		userInfoURL:  providerCfg.UserInfoURL,
		idField:      firstNonEmpty(providerCfg.IDField, "id,sub"),
		emailField:   firstNonEmpty(providerCfg.EmailField, "email"),
		nameField:    firstNonEmpty(providerCfg.NameField, "name,display_name,login"),
		pictureField: firstNonEmpty(providerCfg.PictureField, "picture,avatar_url"),
	}
}

func (g *GenericProvider) GetProviderName() string {
	return g.name
}

func (g *GenericProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (g *GenericProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		slog.Error("GenericProvider.Exchange: exchange code for token", "provider", g.name, "error", err)
		return nil, domain.NewExternalServiceError(err)
	}
	return token, nil
}

func (g *GenericProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error) {
	client := g.config.Client(ctx, token)
	resp, err := client.Get(g.userInfoURL)
	if err != nil {
		slog.Error("GenericProvider.GetUserInfo: get user info from API", "provider", g.name, "error", err)
		return nil, domain.NewExternalServiceError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("GenericProvider.GetUserInfo: non-200 status code", "provider", g.name, "status_code", resp.StatusCode)
		return nil, domain.NewExternalServiceError(fmt.Errorf("status code %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("GenericProvider.GetUserInfo: read response body", "provider", g.name, "error", err)
		return nil, domain.NewInternalError(err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		slog.Error("GenericProvider.GetUserInfo: unmarshal response", "provider", g.name, "error", err)
		return nil, domain.NewInternalError(err)
	}

	userInfo := &domain.OAuthUserInfo{
		ID:      valueFromPayload(payload, g.idField),
		Email:   valueFromPayload(payload, g.emailField),
		Name:    valueFromPayload(payload, g.nameField),
		Picture: valueFromPayload(payload, g.pictureField),
	}
	if userInfo.ID == "" || userInfo.Email == "" {
		return nil, domain.NewExternalServiceError(fmt.Errorf("provider %s user info missing id or email", g.name))
	}
	if userInfo.Name == "" {
		userInfo.Name = userInfo.Email
	}

	return userInfo, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func valueFromPayload(payload map[string]any, fields string) string {
	for _, field := range strings.Split(fields, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if value := stringValue(payload[field]); value != "" {
			return value
		}
	}
	return ""
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case int:
		return strconv.Itoa(typed)
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}
