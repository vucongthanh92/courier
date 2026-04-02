package oauth

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"google.golang.org/api/idtoken"
)

type GoogleClient struct {
	audience string
}

func NewGoogleClient(aud string) *GoogleClient {
	return &GoogleClient{audience: aud}
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
