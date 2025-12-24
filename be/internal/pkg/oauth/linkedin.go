package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
)

type LinkedInProvider struct {
	config      *oauth2.Config
	userInfoURL string
	emailURL    string
}

type linkedinUserInfo struct {
	ID                 string `json:"id"`
	LocalizedFirstName string `json:"localizedFirstName"`
	LocalizedLastName  string `json:"localizedLastName"`
}

type linkedinEmailInfo struct {
	Elements []struct {
		Handle struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"handle~"`
	} `json:"elements"`
}

func NewLinkedInProvider(clientID, clientSecret, redirectURL, userInfoURL string, scopes []string) OAuthProvider {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     linkedin.Endpoint,
	}

	return &LinkedInProvider{
		config:      oauthConfig,
		userInfoURL: userInfoURL,
		emailURL:    "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))",
	}
}

func (l *LinkedInProvider) GetProviderName() string {
	return ProviderLinkedIn
}

func (l *LinkedInProvider) GetAuthURL(state string) string {
	return l.config.AuthCodeURL(state)
}

func (l *LinkedInProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := l.config.Exchange(ctx, code)
	if err != nil {
		slog.Error("LinkedInProvider.Exchange: exchange code for token", "error", err)
		return nil, domain.NewExternalServiceError("LinkedIn OAuth", err)
	}
	return token, nil
}

func (l *LinkedInProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error) {
	client := l.config.Client(ctx, token)

	// Get user profile
	resp, err := client.Get(l.userInfoURL)
	if err != nil {
		slog.Error("LinkedInProvider.GetUserInfo: get user info from API", "error", err)
		return nil, domain.NewExternalServiceError("LinkedIn API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("LinkedInProvider.GetUserInfo: non-200 status code", "status_code", resp.StatusCode)
		return nil, domain.NewExternalServiceError("LinkedIn API", fmt.Errorf("status code %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("LinkedInProvider.GetUserInfo: read response body", "error", err)
		return nil, domain.NewInternalError("read response", err)
	}

	var linkedinInfo linkedinUserInfo
	if err := json.Unmarshal(data, &linkedinInfo); err != nil {
		slog.Error("LinkedInProvider.GetUserInfo: unmarshal response", "error", err)
		return nil, domain.NewInternalError("parse response", err)
	}

	// Get email
	email := ""
	emailResp, err := client.Get(l.emailURL)
	if err == nil {
		defer emailResp.Body.Close()
		if emailResp.StatusCode == 200 {
			emailData, err := io.ReadAll(emailResp.Body)
			if err == nil {
				var emailInfo linkedinEmailInfo
				if err := json.Unmarshal(emailData, &emailInfo); err == nil {
					if len(emailInfo.Elements) > 0 {
						email = emailInfo.Elements[0].Handle.EmailAddress
					}
				}
			}
		}
	}

	displayName := fmt.Sprintf("%s %s", linkedinInfo.LocalizedFirstName, linkedinInfo.LocalizedLastName)

	return &domain.OAuthUserInfo{
		ID:      linkedinInfo.ID,
		Email:   email,
		Name:    displayName,
		Picture: "",
	}, nil
}
