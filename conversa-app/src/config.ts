export const USER_API_BASE_URL =
  import.meta.env.VITE_USER_API_BASE_URL ?? "http://localhost:5001/api/v1";

export const CHAT_API_BASE_URL =
  import.meta.env.VITE_CHAT_API_BASE_URL ?? "http://localhost:5002/api/v1";

function toWebSocketUrl(httpUrl: string) {
  const url = new URL(httpUrl);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = `${url.pathname.replace(/\/$/, "")}/ws`;
  url.search = "";
  return url.toString();
}

export const CHAT_WS_URL =
  import.meta.env.VITE_CHAT_WS_URL ?? toWebSocketUrl(CHAT_API_BASE_URL);

export const GOOGLE_OAUTH_CLIENT_ID = import.meta.env.VITE_GOOGLE_OAUTH_CLIENT_ID ?? "";

export const GITHUB_OAUTH_CLIENT_ID = import.meta.env.VITE_GITHUB_OAUTH_CLIENT_ID ?? "";

export const GOOGLE_OAUTH_REDIRECT_URI =
  import.meta.env.VITE_GOOGLE_OAUTH_REDIRECT_URI ?? `${USER_API_BASE_URL}/auth/identity/google/callback`;

export const GITHUB_OAUTH_REDIRECT_URI =
  import.meta.env.VITE_GITHUB_OAUTH_REDIRECT_URI ?? `${USER_API_BASE_URL}/auth/identity/github/callback`;

export const APP_OAUTH_CALLBACK_BASE_URL =
  import.meta.env.VITE_APP_OAUTH_CALLBACK_BASE_URL ?? `${window.location.origin}/oauth/callback`;
