# User Service

`user-service` owns Courier user identity, authentication, profile lookup, email verification, OAuth identity linking, and JWT issuance.

## Local Development

```bash
make run-local
make test
```

Local API base URL:

```text
http://localhost:5001/api/v1
```

## Verified User Search

`conversa-app` uses `user-service` to find users when creating a new conversation.

Endpoint:

```text
GET /api/v1/user/search?search_key=<text>
Authorization: Bearer <access_token>
```

Important behavior:

- Requires a valid user token.
- Excludes the current authenticated user from results.
- Returns only users with `status = verified`.
- Searches by `display_name`, `phone_number`, and `email`.
- Returns limited public fields only: `user_id`, `display_name`, `phone_number`, `email`, and `avatar`.

## OAuth Login

Google and GitHub OAuth are supported through `user-service` identity callback routes.

Callback routes:

```text
GET /api/v1/auth/identity/google/callback
GET /api/v1/auth/identity/github/callback
```

Flow:

- `conversa-app` redirects users to Google or GitHub with the provider callback URL pointing at `user-service`.
- The provider redirects back to `user-service` with `code` and `state`.
- `user-service` exchanges the code, validates provider identity, creates or links the Courier user, and issues Courier JWT tokens.
- If the OAuth `state` includes a Courier app return URL, `user-service` redirects back to `conversa-app` with the JWT session payload in the URL fragment.
- If no Courier app return URL is present, the callback keeps the API behavior and returns the normal JSON success response.

For Google, the token exchange uses the same redirect URI received from the OAuth callback request, so the Google Cloud Console authorized redirect URI must match the app config exactly.

Local `conversa-app` values usually look like:

```env
VITE_GOOGLE_OAUTH_REDIRECT_URI=http://localhost:5001/api/v1/auth/identity/google/callback
VITE_GITHUB_OAUTH_REDIRECT_URI=http://localhost:5001/api/v1/auth/identity/github/callback
VITE_APP_OAUTH_CALLBACK_BASE_URL=http://localhost:8080/oauth/callback
```
