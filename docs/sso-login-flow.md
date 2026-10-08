# Courier SSO Login Flow

This document captures the agreed phase-1 SSO design for Courier apps. The initial apps are `conversa-app` and `flikk-app`; future web and mobile clients should reuse the same pattern.

## Login And Cross-App SSO

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant Browser
    participant Conversa as conversa-app
    participant UserSvc as user-service / SSO
    participant Google as Google/GitHub
    participant Flikk as flikk-app

    User->>Browser: Open conversa-app
    Browser->>Conversa: Load app

    Conversa->>Browser: Generate PKCE verifier/challenge + state
    Conversa->>Browser: Redirect to /sso/authorize

    Browser->>UserSvc: GET /api/v1/sso/authorize<br/>client_id=conversa-web<br/>redirect_uri=conversa callback<br/>code_challenge<br/>state

    alt No valid SSO session
        UserSvc->>Browser: Redirect to login UI in conversa-app
        Browser->>Conversa: Show login page
        User->>Conversa: Login with password or Google/GitHub

        alt Social login
            Conversa->>Browser: Redirect to Google/GitHub
            Browser->>Google: OAuth authorize
            Google->>Browser: Redirect to user-service callback with code
            Browser->>UserSvc: GET /auth/identity/{provider}/callback
            UserSvc->>Google: Exchange provider code
            Google-->>UserSvc: Provider token/profile
        else Password login
            Conversa->>UserSvc: POST /auth/login
        end

        UserSvc->>UserSvc: Validate user, create Courier SSO session
        UserSvc->>Browser: Set HttpOnly SSO cookie
        UserSvc->>Browser: Redirect back to original /sso/authorize
    else Valid SSO session exists
        UserSvc->>UserSvc: Validate SSO cookie/session
    end

    UserSvc->>UserSvc: Create single-use authorization code
    UserSvc->>Browser: Redirect to conversa callback<br/>?code=...&state=...

    Browser->>Conversa: Open callback URL
    Conversa->>UserSvc: POST /api/v1/sso/token<br/>code + code_verifier
    UserSvc->>UserSvc: Verify code + PKCE
    UserSvc-->>Conversa: access_token + refresh_token + id_token
    Conversa->>Browser: Save app session
    Conversa->>User: Show authenticated app

    User->>Browser: Open flikk-app
    Browser->>Flikk: Load app
    Flikk->>Browser: Generate PKCE + state
    Flikk->>Browser: Redirect to /sso/authorize

    Browser->>UserSvc: GET /api/v1/sso/authorize<br/>client_id=flikk-web

    UserSvc->>UserSvc: Validate existing SSO cookie
    UserSvc->>UserSvc: Create code for flikk-web
    UserSvc->>Browser: Redirect to flikk callback<br/>?code=...&state=...

    Browser->>Flikk: Open callback URL
    Flikk->>UserSvc: POST /api/v1/sso/token<br/>code + code_verifier
    UserSvc-->>Flikk: access_token + refresh_token + id_token
    Flikk->>User: Show authenticated app without login again
```

## Global Logout

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant App as conversa-app / flikk-app
    participant UserSvc as user-service / SSO
    participant DB as DB / Token Store

    User->>App: Click logout
    App->>UserSvc: POST /api/v1/sso/logout
    UserSvc->>DB: Revoke SSO session
    UserSvc->>DB: Revoke all app refresh tokens for user
    UserSvc->>UserSvc: Add access token/session denylist if needed
    UserSvc->>App: Clear SSO cookie response
    App->>App: Clear local app session
    App->>User: Return to login screen
```

## Short Version

```mermaid
flowchart LR
    A["conversa-app login"] --> B["user-service creates SSO cookie"]
    B --> C["conversa-app gets own tokens"]
    B --> D["flikk-app redirects to SSO"]
    D --> E["user-service sees SSO cookie"]
    E --> F["flikk-app gets own tokens without login again"]
```

## Phase-1 Decisions

- Local testing uses `localhost:<port>` for now. Gateway/load-balancer domain setup can be revisited later.
- Phase 1 includes an OIDC `id_token`.
- Global logout revokes the SSO session and all app refresh tokens for the user.
- The login UI remains in `conversa-app` for phase 1.
- The implementation should be additive and preserve existing login/OAuth APIs during migration.
