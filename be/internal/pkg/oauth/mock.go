package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
)

type MockProvider struct {
	clientID    string
	redirectURL string
	authURL     string
	userInfoURL string
}

type mockUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func NewMockProvider(clientID, _ string, redirectURL, authURL, _ string, userInfoURL string, _ []string) OAuthProvider {
	return &MockProvider{
		clientID:    clientID,
		redirectURL: redirectURL,
		authURL:     authURL,
		userInfoURL: userInfoURL,
	}
}

func (m *MockProvider) GetProviderName() string {
	return ProviderMock
}

func (m *MockProvider) GetAuthURL(state string) string {
	authURL, err := url.Parse(m.authURL)
	if err != nil {
		slog.Error("MockProvider.GetAuthURL: parse auth URL", "error", err)
		return m.redirectURL
	}

	query := authURL.Query()
	query.Set("client_id", m.clientID)
	query.Set("redirect_uri", m.redirectURL)
	query.Set("response_type", "code")
	query.Set("state", state)
	authURL.RawQuery = query.Encode()

	return authURL.String()
}

func (m *MockProvider) Exchange(_ context.Context, code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, domain.NewExternalServiceError(fmt.Errorf("mock provider returned empty code"))
	}

	return &oauth2.Token{
		AccessToken: code,
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(5 * time.Minute),
	}, nil
}

func (m *MockProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error) {
	userInfoURL, err := url.Parse(m.userInfoURL)
	if err != nil {
		return nil, domain.NewExternalServiceError(err)
	}

	query := userInfoURL.Query()
	query.Set("code", token.AccessToken)
	userInfoURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL.String(), nil)
	if err != nil {
		return nil, domain.NewInternalError(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("MockProvider.GetUserInfo: get user info from mock SSO", "error", err)
		return nil, domain.NewExternalServiceError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.NewExternalServiceError(fmt.Errorf("status code %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, domain.NewInternalError(err)
	}

	var info mockUserInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, domain.NewInternalError(err)
	}

	return &domain.OAuthUserInfo{
		ID:      info.ID,
		Email:   info.Email,
		Name:    info.Name,
		Picture: info.Picture,
	}, nil
}
