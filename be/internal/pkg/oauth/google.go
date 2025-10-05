package oauth

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleOAuthClient struct {
	config *oauth2.Config
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func NewGoogleOAuthClient(clientID, clientSecret, redirectURL string) *GoogleOAuthClient {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleOAuthClient{
		config: config,
	}
}

func (c *GoogleOAuthClient) GetAuthURL(state string) string {
	// TODO: Implementation
	return ""
}

func (c *GoogleOAuthClient) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	// TODO: Implementation
	return nil, nil
}

func (c *GoogleOAuthClient) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	// TODO: Implementation
	return nil, nil
}
