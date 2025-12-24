package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"

	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/gitlab"
)

type GitLabProvider struct {
	config      *oauth2.Config
	userInfoURL string
}

type gitlabUserInfo struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func NewGitLabProvider(clientID, clientSecret, redirectURL, userInfoURL string, scopes []string) OAuthProvider {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     gitlab.Endpoint,
	}

	return &GitLabProvider{
		config:      oauthConfig,
		userInfoURL: userInfoURL,
	}
}

func (g *GitLabProvider) GetProviderName() string {
	return ProviderGitLab
}

func (g *GitLabProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state)
}

func (g *GitLabProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		slog.Error("GitLabProvider.Exchange: exchange code for token", "error", err)
		return nil, domain.NewExternalServiceError("GitLab OAuth", err)
	}
	return token, nil
}

func (g *GitLabProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error) {
	client := g.config.Client(ctx, token)
	resp, err := client.Get(g.userInfoURL)
	if err != nil {
		slog.Error("GitLabProvider.GetUserInfo: get user info from API", "error", err)
		return nil, domain.NewExternalServiceError("GitLab API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("GitLabProvider.GetUserInfo: non-200 status code", "status_code", resp.StatusCode)
		return nil, domain.NewExternalServiceError("GitLab API", fmt.Errorf("status code %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("GitLabProvider.GetUserInfo: read response body", "error", err)
		return nil, domain.NewInternalError("read response", err)
	}

	var gitlabInfo gitlabUserInfo
	if err := json.Unmarshal(data, &gitlabInfo); err != nil {
		slog.Error("GitLabProvider.GetUserInfo: unmarshal response", "error", err)
		return nil, domain.NewInternalError("parse response", err)
	}

	return &domain.OAuthUserInfo{
		ID:      strconv.FormatInt(gitlabInfo.ID, 10),
		Email:   gitlabInfo.Email,
		Name:    gitlabInfo.Name,
		Picture: gitlabInfo.AvatarURL,
	}, nil
}
