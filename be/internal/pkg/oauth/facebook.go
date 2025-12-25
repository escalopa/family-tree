package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

type FacebookProvider struct {
	config      *oauth2.Config
	userInfoURL string
}

type facebookUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture struct {
		Data struct {
			Height       int    `json:"height"`
			IsSilhouette bool   `json:"is_silhouette"`
			URL          string `json:"url"`
			Width        int    `json:"width"`
		} `json:"data"`
	} `json:"picture"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func NewFacebookProvider(clientID, clientSecret, redirectURL, userInfoURL string, scopes []string) OAuthProvider {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     facebook.Endpoint,
	}

	return &FacebookProvider{
		config:      oauthConfig,
		userInfoURL: userInfoURL,
	}
}

func (f *FacebookProvider) GetProviderName() string {
	return ProviderFacebook
}

func (f *FacebookProvider) GetAuthURL(state string) string {
	return f.config.AuthCodeURL(state)
}

func (f *FacebookProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := f.config.Exchange(ctx, code)
	if err != nil {
		slog.Error("FacebookProvider.Exchange: exchange code for token", "error", err)
		return nil, domain.NewExternalServiceError(err)
	}
	return token, nil
}

func (f *FacebookProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error) {
	client := f.config.Client(ctx, token)

	// Facebook Graph API requires fields parameter
	// Request high-resolution profile picture (800x800)
	url := fmt.Sprintf("%s?fields=id,email,name,first_name,last_name,picture.width(800).height(800)", f.userInfoURL)
	resp, err := client.Get(url)
	if err != nil {
		slog.Error("FacebookProvider.GetUserInfo: get user info from API", "error", err)
		return nil, domain.NewExternalServiceError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("FacebookProvider.GetUserInfo: non-200 status code", "status_code", resp.StatusCode)
		return nil, domain.NewExternalServiceError(fmt.Errorf("status code %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("FacebookProvider.GetUserInfo: read response body", "error", err)
		return nil, domain.NewInternalError(err)
	}

	var fbInfo facebookUserInfo
	if err := json.Unmarshal(data, &fbInfo); err != nil {
		slog.Error("FacebookProvider.GetUserInfo: unmarshal response", "error", err)
		return nil, domain.NewInternalError(err)
	}

	picture := ""
	if !fbInfo.Picture.Data.IsSilhouette {
		picture = fbInfo.Picture.Data.URL
	}

	return &domain.OAuthUserInfo{
		ID:      fbInfo.ID,
		Email:   fbInfo.Email,
		Name:    fbInfo.Name,
		Picture: picture,
	}, nil
}
