import { CHAT_API_BASE_URL, USER_API_BASE_URL } from "../config";
import type {
  ApiResponse,
  ListConversationMembersResponse,
  ConversationListResponse,
  CreateConversationRequest,
  JwtTokenResponse,
  ListMessagesResponse,
  LoginRequest,
  Message,
  OAuthProvider,
  SearchUserResult,
  SignupRequest,
  VerifyEmailRequest
} from "../types";

type RequestOptions = {
  token?: string;
  method?: string;
  body?: unknown;
};

export const AUTH_UNAUTHORIZED_EVENT = "conversa:auth-unauthorized";

async function request<T>(baseUrl: string, path: string, options: RequestOptions = {}): Promise<T> {
  const response = await fetch(`${baseUrl}${path}`, {
    method: options.method ?? "GET",
    headers: {
      "Content-Type": "application/json",
      ...(options.token ? { Authorization: `Bearer ${options.token}` } : {})
    },
    body: options.body ? JSON.stringify(options.body) : undefined
  });

  const payload = (await response.json().catch(() => null)) as ApiResponse<T> | null;
  if (response.status === 401 && options.token) {
    window.dispatchEvent(new CustomEvent(AUTH_UNAUTHORIZED_EVENT));
  }
  if (!response.ok || !payload?.success) {
    const message =
      payload?.errors?.map((error) => error.message || error.code).filter(Boolean).join(", ") ||
      `Request failed with status ${response.status}`;
    throw new Error(message);
  }
  if (payload.data === null) {
    throw new Error("Response data is empty");
  }
  return payload.data;
}

export const authApi = {
  login(body: LoginRequest) {
    return request<JwtTokenResponse>(USER_API_BASE_URL, "/auth/login", {
      method: "POST",
      body
    });
  },
  signup(body: SignupRequest) {
    return request<unknown>(USER_API_BASE_URL, "/auth/sign-up", {
      method: "POST",
      body
    });
  },
  verifyEmail(body: VerifyEmailRequest) {
    return request<unknown>(USER_API_BASE_URL, "/auth/verify-email", {
      method: "POST",
      body
    });
  },
  resendVerifyEmail(email: string) {
    return request<unknown>(USER_API_BASE_URL, "/auth/verify-email/resend", {
      method: "PUT",
      body: { email }
    });
  },
  oauthCallback(provider: OAuthProvider, code: string, redirectUri: string) {
    const query = new URLSearchParams({ code, redirect_uri: redirectUri });
    return request<JwtTokenResponse>(USER_API_BASE_URL, `/auth/identity/${provider}/callback?${query.toString()}`);
  }
};

export const chatApi = {
  listConversations(token: string) {
    return request<ConversationListResponse>(CHAT_API_BASE_URL, "/conversations?limit=30", {
      token
    });
  },
  listMessages(token: string, conversationId: string, beforeMessageId?: string) {
    const query = new URLSearchParams({ limit: "30" });
    if (beforeMessageId) query.set("before_message_id", String(beforeMessageId));
    return request<ListMessagesResponse>(
      CHAT_API_BASE_URL,
      `/conversation/${conversationId}/messages?${query.toString()}`,
      { token }
    );
  },
  listConversationMembers(token: string, conversationId: string) {
    return request<ListConversationMembersResponse>(CHAT_API_BASE_URL, `/conversation/${conversationId}/members`, {
      token
    });
  },
  createConversation(token: string, body: CreateConversationRequest) {
    return request<ConversationListResponse["conversations"][number]>(CHAT_API_BASE_URL, "/conversation/create", {
      method: "POST",
      token,
      body
    });
  },
  createMessage(token: string, conversationId: string, body: string) {
    return request<Message>(CHAT_API_BASE_URL, `/conversation/${conversationId}/messages/create`, {
      method: "POST",
      token,
      body: {
        type: "text",
        body,
        client_message_id: crypto.randomUUID(),
        metadata: { source: "conversa-app" }
      }
    });
  }
};

export const userApi = {
  searchUsers(token: string, searchKey: string) {
    const query = new URLSearchParams({ search_key: searchKey });
    return request<SearchUserResult[]>(USER_API_BASE_URL, `/user/search?${query.toString()}`, {
      token
    });
  }
};
