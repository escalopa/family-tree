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
	"golang.org/x/oauth2/github"
)

type GitHubProvider struct {
	config      *oauth2.Config
	userInfoURL string
	emailURL    string
}

type githubUserInfo struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Location  string `json:"location"`
}

type githubEmailInfo struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

func NewGitHubProvider(clientID, clientSecret, redirectURL, userInfoURL string, scopes []string) OAuthProvider {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     github.Endpoint,
	}

	return &GitHubProvider{
		config:      oauthConfig,
		userInfoURL: userInfoURL,
		emailURL:    "https://api.github.com/user/emails",
	}
}

func (g *GitHubProvider) GetProviderName() string {
	return ProviderGitHub
}

func (g *GitHubProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state)
}

func (g *GitHubProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		slog.Error("GitHubProvider.Exchange: exchange code for token", "error", err)
		return nil, domain.NewExternalServiceError(err)
	}
	return token, nil
}

func (g *GitHubProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error) {
	client := g.config.Client(ctx, token)

	// Get user info
	resp, err := client.Get(g.userInfoURL)
	if err != nil {
		slog.Error("GitHubProvider.GetUserInfo: get user info from API", "error", err)
		return nil, domain.NewExternalServiceError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("GitHubProvider.GetUserInfo: non-200 status code", "status_code", resp.StatusCode)
		return nil, domain.NewExternalServiceError(fmt.Errorf("status code %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("GitHubProvider.GetUserInfo: read response body", "error", err)
		return nil, domain.NewInternalError(err)
	}

	var githubInfo githubUserInfo
	if err := json.Unmarshal(data, &githubInfo); err != nil {
		slog.Error("GitHubProvider.GetUserInfo: unmarshal response", "error", err)
		return nil, domain.NewInternalError(err)
	}

	email := githubInfo.Email

	// If email is not public, fetch from emails endpoint
	if email == "" {
		emailResp, err := client.Get(g.emailURL)
		if err == nil {
			defer emailResp.Body.Close()
			if emailResp.StatusCode == 200 {
				emailData, err := io.ReadAll(emailResp.Body)
				if err == nil {
					var emails []githubEmailInfo
					if err := json.Unmarshal(emailData, &emails); err == nil {
						// Find primary verified email
						for _, e := range emails {
							if e.Primary && e.Verified {
								email = e.Email
								break
							}
						}
						// If no primary verified, take first verified
						if email == "" {
							for _, e := range emails {
								if e.Verified {
									email = e.Email
									break
								}
							}
						}
						// If no verified, take first email
						if email == "" && len(emails) > 0 {
							email = emails[0].Email
						}
					}
				}
			}
		}
	}

	displayName := githubInfo.Name
	if displayName == "" {
		displayName = githubInfo.Login
	}

	// GitHub avatar URLs support size parameter (s=800) for higher resolution
	avatarURL := githubInfo.AvatarURL
	if avatarURL != "" {
		avatarURL = fmt.Sprintf("%s?s=800", avatarURL)
	}

	return &domain.OAuthUserInfo{
		ID:      strconv.FormatInt(githubInfo.ID, 10),
		Email:   email,
		Name:    displayName,
		Picture: avatarURL,
	}, nil
}
