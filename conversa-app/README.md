# Conversa

Conversa is a small React + TypeScript messaging client for the Courier chat-service MVP.

## Local setup

```bash
cd conversa-app
cp .env.example .env.local
npm install
npm run dev
```

The dev server uses `http://localhost:8080`, which matches the current local CORS settings in `user-service` and `chat-service`.

## Backend assumptions

- user-service: `http://localhost:5001/api/v1`
- chat-service: `http://localhost:5002/api/v1`

The app uses the real routes currently registered in backend `routes.go` files.

## OAuth Login

The login screen includes Google and GitHub buttons.

Required local env values:

```env
VITE_GOOGLE_OAUTH_CLIENT_ID=
VITE_GITHUB_OAUTH_CLIENT_ID=
VITE_GOOGLE_OAUTH_REDIRECT_URI=http://localhost:5001/api/v1/auth/identity/google/callback
VITE_GITHUB_OAUTH_REDIRECT_URI=http://localhost:5001/api/v1/auth/identity/github/callback
VITE_APP_OAUTH_CALLBACK_BASE_URL=http://localhost:8080/oauth/callback
```

Keep real values in `.env.local`; this file is ignored by Git.

Flow:

- The app redirects to Google or GitHub with a Courier OAuth state containing the return URL.
- The provider redirects to `user-service`.
- `user-service` exchanges the code and redirects back to `conversa-app` with the Courier JWT session in the URL fragment.
- The app saves the session and enters the chat shell.

## New Conversation Modal

The plus button in the left rail opens the "New Conversation" modal.

Behavior:

- Search users through `GET /api/v1/user/search`.
- Add one or more verified users by display name, phone number, or email.
- Send selected user IDs as strings to avoid large-ID precision loss.
- Let `chat-service` infer conversation type:
  - one selected user: `direct`
  - two or more selected users: `group`
- Use the custom name if one is entered; otherwise the backend generates a participant-based name.
- Open the created conversation automatically after success.

The modal uses the app's glass visual language and keeps the conversation type hidden from the user.

## Expired Sessions

All authenticated API calls share the same API wrapper. If any token-backed request receives `401 Unauthorized`, the app clears local session data and returns to the login screen.
