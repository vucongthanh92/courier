package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"google.golang.org/api/idtoken"
)

type GoogleClient struct {
	httpClient   *http.Client
	audience     string
	clientID     string
	clientSecret string
	redirectURI  string
}

func NewGoogleClient(audience, clientID, clientSecret, redirectURI string) *GoogleClient {
	return &GoogleClient{
		httpClient:   http.DefaultClient,
		audience:     audience,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

// Verify validates the Google ID token and extracts the user's profile information.
func (c *GoogleClient) Verify(ctx context.Context, token string) (models.ProviderProfile, error) {

	// validate the token with Google's library, which also checks the signature and expiry
	payload, err := idtoken.Validate(ctx, token, c.audience)
	if err != nil {
		return models.ProviderProfile{}, err
	}

	// extract relevant info from the token payload
	email, _ := payload.Claims["email"].(string)
	sub, _ := payload.Claims["sub"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)
	pic, _ := payload.Claims["picture"].(string)

	// for Google, we can trust the email_verified field since it's from their token,
	// but for other providers we might need to verify it ourselves
	return models.ProviderProfile{
		Provider:      "google",
		ProviderUID:   sub,
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
		AvatarURL:     pic,
	}, nil
}

// ExchangeCode exchanges the authorization code for an ID token by making a POST request to Google's token endpoint.
func (c *GoogleClient) ExchangeCode(ctx context.Context, code, redirectURI string) (string, error) {

	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", c.redirectURI)
	if redirectURI != "" {
		form.Set("redirect_uri", redirectURI)
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out struct {
		IDToken string `json:"id_token"`
		Error   string `json:"error"`
		Desc    string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK || out.IDToken == "" {
		return "", errors.New("exchange failed: " + out.Error + " " + out.Desc)
	}
	return out.IDToken, nil
}
