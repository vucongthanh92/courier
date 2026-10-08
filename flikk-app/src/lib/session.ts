import { STANDALONE_ACCESS_TOKEN } from "../config";
import type { JwtTokenResponse } from "../types";

const FLIKK_SESSION_KEY = "flikk.session";

export type FlikkSession = JwtTokenResponse & {
  saved_at: number;
  user_id?: string;
  display_name?: string;
  avatar_url?: string;
};

type StoredSession = {
  access_token?: string;
  display_name?: string;
  user?: { display_name?: string };
};

export function readSession(): FlikkSession | null {
  const raw = localStorage.getItem(FLIKK_SESSION_KEY);
  if (!raw) return null;
  try {
    const session = JSON.parse(raw) as FlikkSession;
    return {
      ...session,
      user_id: session.user?.id ?? parseJwtSubject(session.access_token),
      display_name: session.user?.display_name ?? session.display_name,
      avatar_url: session.user?.avatar_url ?? session.avatar_url
    };
  } catch {
    localStorage.removeItem(FLIKK_SESSION_KEY);
    return null;
  }
}

export function saveSession(tokens: JwtTokenResponse): FlikkSession {
  const session: FlikkSession = {
    ...tokens,
    saved_at: Date.now(),
    user_id: tokens.user?.id ?? parseJwtSubject(tokens.access_token),
    display_name: tokens.user?.display_name,
    avatar_url: tokens.user?.avatar_url
  };
  localStorage.setItem(FLIKK_SESSION_KEY, JSON.stringify(session));
  return session;
}

export function clearSession() {
  localStorage.removeItem(FLIKK_SESSION_KEY);
}

export function readAccessToken() {
  const flikkSession = readSession();
  if (flikkSession?.access_token) return flikkSession.access_token;

  try {
    const stored = localStorage.getItem(FLIKK_SESSION_KEY);
    if (stored) {
      const session = JSON.parse(stored) as StoredSession;
      if (session.access_token) return session.access_token;
    }
  } catch {
    localStorage.removeItem(FLIKK_SESSION_KEY);
  }

  return STANDALONE_ACCESS_TOKEN;
}

export function readDisplayName() {
  const flikkSession = readSession();
  if (flikkSession?.display_name) return flikkSession.display_name;

  try {
    const stored = localStorage.getItem(FLIKK_SESSION_KEY);
    if (!stored) return "Courier member";
    const session = JSON.parse(stored) as StoredSession;
    return session.user?.display_name ?? session.display_name ?? "Courier member";
  } catch {
    return "Courier member";
  }
}

function parseJwtSubject(token: string): string | undefined {
  const [, payload] = token.split(".");
  if (!payload) return undefined;
  try {
    const normalized = payload.replace(/-/g, "+").replace(/_/g, "/");
    const decoded = JSON.parse(window.atob(normalized));
    const sub = String(decoded.sub ?? "");
    return sub ? sub : undefined;
  } catch {
    return undefined;
  }
}
