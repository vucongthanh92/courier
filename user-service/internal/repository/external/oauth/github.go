package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type GitHubClient struct {
	HTTPClient   *http.Client
	APIBase      string
	ClientID     string
	ClientSecret string
}

func NewGitHubClient(apiBase, clientID, clientSecret string) *GitHubClient {
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	return &GitHubClient{
		HTTPClient:   http.DefaultClient,
		APIBase:      apiBase,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// Verify validates the GitHub access token and extracts the user's profile information.
func (c *GitHubClient) Verify(ctx context.Context, token string) (models.ProviderProfile, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.APIBase+"/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return models.ProviderProfile{}, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return models.ProviderProfile{}, errors.New("github token invalid")
	}

	var u struct {
		ID     int64  `json:"id"`
		Avatar string `json:"avatar_url"`
		Name   string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return models.ProviderProfile{}, err
	}

	email, err := c.primaryVerifiedEmail(ctx, token)
	if err != nil {
		return models.ProviderProfile{}, err
	}

	return models.ProviderProfile{
		Provider:      "github",
		ProviderUID:   strconv.FormatInt(u.ID, 10),
		Email:         strings.ToLower(email),
		EmailVerified: true,
		Name:          u.Name,
		AvatarURL:     u.Avatar,
	}, nil
}

// GitHub's API requires a separate call to get the user's email addresses, and we need to find the primary verified one.
func (c *GitHubClient) primaryVerifiedEmail(ctx context.Context, token string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.APIBase+"/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("github email scope missing or token invalid")
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	return "", errors.New("no verified primary email")
}

// ExchangeCode exchanges the authorization code for an access token using GitHub's OAuth API.
func (c *GitHubClient) ExchangeCode(ctx context.Context, code, redirectURI string) (string, error) {

	// GitHub expects the parameters in the request body as JSON
	payload := map[string]string{
		"client_id":     c.ClientID,
		"client_secret": c.ClientSecret,
		"code":          code,
	}

	if redirectURI != "" {
		payload["redirect_uri"] = redirectURI
	}

	// Make the POST request to GitHub's token endpoint
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", bytes.NewReader(b))
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}

	// GitHub returns the access token in the response body as JSON
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("exchange failed status=%d", resp.StatusCode)
	}

	// The response JSON will contain the access token in the "access_token" field
	var res struct {
		AccessToken string `json:"access_token"`
	}

	// Decode the response and extract the access token
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if res.AccessToken == "" {
		return "", errors.New("access_token empty")
	}

	return res.AccessToken, nil
}
