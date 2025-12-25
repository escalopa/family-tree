package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/escalopa/family-tree/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

type VKProvider struct {
	config      *oauth2.Config
	userInfoURL string
}

type vkUserInfo struct {
	Response []struct {
		ID         int64  `json:"id"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Photo200   string `json:"photo_200"`
		PhotoMax   string `json:"photo_max_orig"` // High resolution original photo
	} `json:"response"`
}

func NewVKProvider(clientID, clientSecret, redirectURL, userInfoURL string, scopes []string) OAuthProvider {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     vk.Endpoint,
	}

	return &VKProvider{
		config:      oauthConfig,
		userInfoURL: userInfoURL,
	}
}

func (v *VKProvider) GetProviderName() string {
	return ProviderVK
}

func (v *VKProvider) GetAuthURL(state string) string {
	return v.config.AuthCodeURL(state)
}

func (v *VKProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := v.config.Exchange(ctx, code)
	if err != nil {
		slog.Error("VKProvider.Exchange: exchange code for token", "error", err)
		return nil, domain.NewExternalServiceError(err)
	}
	return token, nil
}

func (v *VKProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*domain.OAuthUserInfo, error) {
	client := v.config.Client(ctx, token)

	// VK API requires access token and API version
	// Request photo_max_orig for high-resolution profile picture
	url := fmt.Sprintf("%s?access_token=%s&v=5.131&fields=photo_200,photo_max_orig", v.userInfoURL, token.AccessToken)
	resp, err := client.Get(url)
	if err != nil {
		slog.Error("VKProvider.GetUserInfo: get user info from API", "error", err)
		return nil, domain.NewExternalServiceError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("VKProvider.GetUserInfo: non-200 status code", "status_code", resp.StatusCode)
		return nil, domain.NewExternalServiceError(fmt.Errorf("status code %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("VKProvider.GetUserInfo: read response body", "error", err)
		return nil, domain.NewInternalError(err)
	}

	var vkInfo vkUserInfo
	if err := json.Unmarshal(data, &vkInfo); err != nil {
		slog.Error("VKProvider.GetUserInfo: unmarshal response", "error", err)
		return nil, domain.NewInternalError(err)
	}

	if len(vkInfo.Response) == 0 {
		return nil, domain.NewInternalError(fmt.Errorf("VK API: empty response"))
	}

	user := vkInfo.Response[0]
	displayName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	// VK doesn't provide email in user.get API, it's returned in the token exchange
	// Extract email from token extras
	email := ""
	if extras := token.Extra("email"); extras != nil {
		if emailStr, ok := extras.(string); ok {
			email = emailStr
		}
	}

	// Use high-resolution photo if available, fallback to photo_200
	picture := user.PhotoMax
	if picture == "" {
		picture = user.Photo200
	}

	return &domain.OAuthUserInfo{
		ID:      fmt.Sprintf("%d", user.ID),
		Email:   email,
		Name:    displayName,
		Picture: picture,
	}, nil
}
