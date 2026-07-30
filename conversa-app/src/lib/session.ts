import type { JwtTokenResponse } from "../types";

const SESSION_KEY = "conversa.session";

export type Session = JwtTokenResponse & {
  saved_at: number;
  user_id?: string;
};

export function readSession(): Session | null {
  const raw = localStorage.getItem(SESSION_KEY);
  if (!raw) return null;
  try {
    const session = JSON.parse(raw) as Session;
    return {
      ...session,
      user_id: parseJwtSubject(session.access_token)
    };
  } catch {
    localStorage.removeItem(SESSION_KEY);
    return null;
  }
}

export function saveSession(tokens: JwtTokenResponse): Session {
  const session: Session = {
    ...tokens,
    saved_at: Date.now(),
    user_id: parseJwtSubject(tokens.access_token)
  };
  localStorage.setItem(SESSION_KEY, JSON.stringify(session));
  return session;
}

export function clearSession() {
  localStorage.removeItem(SESSION_KEY);
}

export function parseJwtSubject(token: string): string | undefined {
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
