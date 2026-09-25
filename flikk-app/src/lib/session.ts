import { STANDALONE_ACCESS_TOKEN } from "../config";

const CONVERSA_SESSION_KEY = "conversa.session";

type ConversaSession = {
  access_token?: string;
  display_name?: string;
  user?: { display_name?: string };
};

export function readAccessToken() {
  try {
    const stored = localStorage.getItem(CONVERSA_SESSION_KEY);
    if (stored) {
      const session = JSON.parse(stored) as ConversaSession;
      if (session.access_token) return session.access_token;
    }
  } catch {
    localStorage.removeItem(CONVERSA_SESSION_KEY);
  }

  return STANDALONE_ACCESS_TOKEN;
}

export function readDisplayName() {
  try {
    const stored = localStorage.getItem(CONVERSA_SESSION_KEY);
    if (!stored) return "Courier member";
    const session = JSON.parse(stored) as ConversaSession;
    return session.user?.display_name ?? session.display_name ?? "Courier member";
  } catch {
    return "Courier member";
  }
}
