package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type GitHubClient struct {
	HTTPClient *http.Client
	APIBase    string
}

func NewGitHubClient(apiBase string) *GitHubClient {
	return &GitHubClient{
		HTTPClient: http.DefaultClient,
		APIBase:    apiBase,
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
