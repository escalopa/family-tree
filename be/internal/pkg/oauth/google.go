package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProvider struct {
	config *oauth2.Config
}

type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func NewGoogleProvider(clientID, clientSecret, redirectURL string, scopes []string) *GoogleProvider {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	return &GoogleProvider{config: config}
}

func (g *GoogleProvider) GetProviderName() string {
	return "google"
}

func (g *GoogleProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (g *GoogleProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		log.Printf("Google OAuth exchange failed: %v", err)
		return nil, domain.NewExternalServiceError("Google OAuth", err)
	}
	return token, nil
}

func (g *GoogleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Failed to get user info from Google: %v", err)
		return nil, domain.NewExternalServiceError("Google API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Google API returned status code %d", resp.StatusCode)
		return nil, domain.NewExternalServiceError("Google API", fmt.Errorf("status code %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read Google API response: %v", err)
		return nil, domain.NewInternalError("failed to read response", err)
	}

	var googleInfo googleUserInfo
	if err := json.Unmarshal(data, &googleInfo); err != nil {
		log.Printf("Failed to unmarshal Google user info: %v", err)
		return nil, domain.NewInternalError("failed to parse response", err)
	}

	return &UserInfo{
		ID:            googleInfo.ID,
		Email:         googleInfo.Email,
		VerifiedEmail: googleInfo.VerifiedEmail,
		Name:          googleInfo.Name,
		GivenName:     googleInfo.GivenName,
		FamilyName:    googleInfo.FamilyName,
		Picture:       googleInfo.Picture,
		Locale:        googleInfo.Locale,
	}, nil
}
