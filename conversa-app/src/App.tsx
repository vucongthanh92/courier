import { FormEvent, type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  APP_OAUTH_CALLBACK_BASE_URL,
  GITHUB_OAUTH_CLIENT_ID,
  GITHUB_OAUTH_REDIRECT_URI,
  GOOGLE_OAUTH_CLIENT_ID,
  GOOGLE_OAUTH_REDIRECT_URI
} from "./config";
import { AUTH_UNAUTHORIZED_EVENT, authApi, chatApi, userApi } from "./lib/api";
import { RealtimeClient, type RealtimeStatus } from "./lib/realtime";
import { clearSession, readSession, saveSession, type Session } from "./lib/session";
import { cacheUserProfiles, readUserProfileCache } from "./lib/userProfileCache";
import type { Conversation, ListMessagesResponse, Message, OAuthProvider, RealtimeEvent, SearchUserResult, UserProfile } from "./types";

type AuthMode = "login" | "signup" | "verify";
const OAUTH_STATE_KEY = "conversa.oauth.state";

export function App() {
  const [session, setSession] = useState<Session | null>(() => readSession());

  useEffect(() => {
    function handleUnauthorized() {
      clearSession();
      setSession(null);
    }

    window.addEventListener(AUTH_UNAUTHORIZED_EVENT, handleUnauthorized);
    return () => window.removeEventListener(AUTH_UNAUTHORIZED_EVENT, handleUnauthorized);
  }, []);

  if (!session) {
    return <AuthScreen onAuthenticated={setSession} />;
  }

  return (
    <MessengerShell
      session={session}
      onLogout={() => {
        clearSession();
        setSession(null);
      }}
    />
  );
}

function AuthScreen({ onAuthenticated }: { onAuthenticated: (session: Session) => void }) {
  const [mode, setMode] = useState<AuthMode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [phoneNumber, setPhoneNumber] = useState("");
  const [token, setToken] = useState("");
  const [loading, setLoading] = useState(false);
  const [oauthLoading, setOauthLoading] = useState<OAuthProvider | null>(null);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const hashParams = new URLSearchParams(window.location.hash.replace(/^#/, ""));
    const providerFromPath = window.location.pathname.match(/\/oauth\/callback\/(google|github)$/)?.[1] as OAuthProvider | undefined;
    const provider = providerFromPath ?? (params.get("provider") as OAuthProvider | null);
    const code = params.get("code");
    const state = params.get("state") ?? hashParams.get("state");
    const oauthResult = hashParams.get("oauth_result");
    const normalizedState = normalizeOAuthState(state);
    const storedState = sessionStorage.getItem(OAUTH_STATE_KEY) ?? localStorage.getItem(OAUTH_STATE_KEY);

    if (!provider || !["google", "github"].includes(provider)) return;

    if (!normalizedState || !storedState || normalizedState !== storedState) {
      console.warn("OAuth state mismatch", {
        hasState: Boolean(state),
        hasStoredState: Boolean(storedState),
        sameState: normalizedState === storedState
      });
      setError("OAuth session is invalid. Please try again.");
      return;
    }

    clearOAuthState();
    setOauthLoading(provider);
    setError("");
    setNotice("");

    if (oauthResult) {
      try {
        const response = JSON.parse(decodeBase64Url(oauthResult));
        onAuthenticated(saveSession(response));
      } catch {
        setError("OAuth response is invalid. Please try again.");
      } finally {
        window.history.replaceState({}, document.title, "/");
        setOauthLoading(null);
      }
      return;
    }

    if (!code) return;
    authApi.oauthCallback(provider, code, getOAuthRedirectUri(provider))
      .then((response) => {
        onAuthenticated(saveSession(response));
        window.history.replaceState({}, document.title, window.location.pathname);
      })
      .catch((err) => {
        setError(err instanceof Error ? err.message : "OAuth login failed");
        window.history.replaceState({}, document.title, window.location.pathname);
      })
      .finally(() => setOauthLoading(null));
  }, [onAuthenticated]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setLoading(true);
    setError("");
    setNotice("");

    try {
      if (mode === "login") {
        const response = await authApi.login({ email, password });
        onAuthenticated(saveSession(response));
        return;
      }
      if (mode === "signup") {
        await authApi.signup({
          email,
          password,
          display_name: displayName,
          phone_number: phoneNumber
        });
        setMode("verify");
        setNotice("Account created. Check your email and enter the verification token.");
        return;
      }
      await authApi.verifyEmail({ email, token });
      setMode("login");
      setNotice("Email verified. You can sign in now.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
    } finally {
      setLoading(false);
    }
  }

  async function resend() {
    setLoading(true);
    setError("");
    setNotice("");
    try {
      await authApi.resendVerifyEmail(email);
      setNotice("Verification email resent.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not resend verification email");
    } finally {
      setLoading(false);
    }
  }

  function loginWithOAuth(provider: OAuthProvider) {
    const clientId = provider === "google" ? GOOGLE_OAUTH_CLIENT_ID : GITHUB_OAUTH_CLIENT_ID;
    if (!clientId) {
      setError(`${provider === "google" ? "Google" : "GitHub"} OAuth client id is not configured.`);
      return;
    }

    const nonce = crypto.randomUUID();
    const state = encodeOAuthState(nonce, `${APP_OAUTH_CALLBACK_BASE_URL}/${provider}`);
    saveOAuthState(state);

    const redirectUri = getOAuthRedirectUri(provider);

    const authUrl =
      provider === "google"
        ? new URL("https://accounts.google.com/o/oauth2/v2/auth")
        : new URL("https://github.com/login/oauth/authorize");

    authUrl.searchParams.set("client_id", clientId);
    authUrl.searchParams.set("redirect_uri", redirectUri);
    authUrl.searchParams.set("state", state);

    if (provider === "google") {
      authUrl.searchParams.set("response_type", "code");
      authUrl.searchParams.set("scope", "openid email profile");
      authUrl.searchParams.set("prompt", "select_account");
    } else {
      authUrl.searchParams.set("scope", "read:user user:email");
    }

    setOauthLoading(provider);
    window.location.assign(authUrl.toString());
  }

  return (
    <main className="auth-page">
      <div className="aurora aurora-a" />
      <div className="aurora aurora-b" />
      <section className="auth-card">
        <ConversaLogo />
        <h1>Conversa</h1>
        <p className="muted">Spatial messaging for Courier teams.</p>

        <div className="oauth-actions" aria-label="OAuth login options">
          <button type="button" onClick={() => loginWithOAuth("google")} disabled={loading || oauthLoading !== null}>
            <GoogleLogo />
            <span>{oauthLoading === "google" ? "Connecting..." : "Google"}</span>
          </button>
          <button type="button" onClick={() => loginWithOAuth("github")} disabled={loading || oauthLoading !== null}>
            <GitHubLogo />
            <span>{oauthLoading === "github" ? "Connecting..." : "GitHub"}</span>
          </button>
        </div>

        <div className="auth-tabs" role="tablist">
          <button className={mode === "login" ? "active" : ""} onClick={() => setMode("login")}>
            Login
          </button>
          <button className={mode === "signup" ? "active" : ""} onClick={() => setMode("signup")}>
            Sign up
          </button>
          <button className={mode === "verify" ? "active" : ""} onClick={() => setMode("verify")}>
            Verify
          </button>
        </div>

        <form className="auth-form" onSubmit={submit}>
          <label>
            Email
            <input
              autoComplete="email"
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              required
            />
          </label>

          {mode === "signup" && (
            <>
              <label>
                Display name
                <input
                  autoComplete="name"
                  value={displayName}
                  onChange={(event) => setDisplayName(event.target.value)}
                  required
                />
              </label>
              <label>
                Phone number
                <input
                  autoComplete="tel"
                  value={phoneNumber}
                  onChange={(event) => setPhoneNumber(event.target.value)}
                  required
                />
              </label>
            </>
          )}

          {mode !== "verify" ? (
            <label>
              Password
              <input
                autoComplete={mode === "login" ? "current-password" : "new-password"}
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
              />
            </label>
          ) : (
            <label>
              Verification token
              <input value={token} onChange={(event) => setToken(event.target.value)} required />
            </label>
          )}

          {error && <div className="alert error">{error}</div>}
          {notice && <div className="alert success">{notice}</div>}

          <button className="primary-button" disabled={loading}>
            {loading ? "Please wait..." : mode === "login" ? "Login" : mode === "signup" ? "Create account" : "Verify email"}
          </button>

          {mode === "verify" && (
            <button className="text-button" type="button" onClick={resend} disabled={!email || loading}>
              Resend verification email
            </button>
          )}
        </form>
      </section>
    </main>
  );
}

function getOAuthRedirectUri(provider: OAuthProvider) {
  return provider === "google" ? GOOGLE_OAUTH_REDIRECT_URI : GITHUB_OAUTH_REDIRECT_URI;
}

function saveOAuthState(state: string) {
  sessionStorage.setItem(OAUTH_STATE_KEY, state);
  localStorage.setItem(OAUTH_STATE_KEY, state);
}

function clearOAuthState() {
  sessionStorage.removeItem(OAUTH_STATE_KEY);
  localStorage.removeItem(OAUTH_STATE_KEY);
}

function normalizeOAuthState(state: string | null) {
  if (!state) return null;
  let normalized = state;
  for (let index = 0; index < 2 && normalized.includes("%"); index += 1) {
    try {
      normalized = decodeURIComponent(normalized);
    } catch {
      return normalized;
    }
  }
  return normalized;
}

function encodeOAuthState(nonce: string, returnUri: string) {
  return `courier_oauth:${nonce}:${encodeBase64Url(returnUri)}`;
}

function encodeBase64Url(value: string) {
  return window.btoa(value).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

function decodeBase64Url(value: string) {
  const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
  return window.atob(padded);
}

function MessengerShell({ session, onLogout }: { session: Session; onLogout: () => void }) {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [members, setMembers] = useState<ListMessagesResponse["members"]>([]);
  const [userProfilesById, setUserProfilesById] = useState<Map<string, UserProfile>>(() => readUserProfileCache());
  const [hasOlderMessages, setHasOlderMessages] = useState(false);
  const [nextBeforeMessageId, setNextBeforeMessageId] = useState<string | undefined>();
  const [sidebarLoading, setSidebarLoading] = useState(true);
  const [messagesLoading, setMessagesLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [realtimeStatus, setRealtimeStatus] = useState<RealtimeStatus>("disconnected");
  const [unreadConversationIds, setUnreadConversationIds] = useState<Set<string>>(() => new Set());
  const [draft, setDraft] = useState("");
  const [error, setError] = useState("");
  const [conversationQuery, setConversationQuery] = useState("");
  const [conversationFilter, setConversationFilter] = useState<"all" | "unread" | "groups">("all");
  const [inspectorVisible, setInspectorVisible] = useState(true);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [newConversationName, setNewConversationName] = useState("");
  const [userSearchQuery, setUserSearchQuery] = useState("");
  const [userSearchResults, setUserSearchResults] = useState<SearchUserResult[]>([]);
  const [selectedUsers, setSelectedUsers] = useState<SearchUserResult[]>([]);
  const [userSearchLoading, setUserSearchLoading] = useState(false);
  const [creatingConversation, setCreatingConversation] = useState(false);
  const [createConversationError, setCreateConversationError] = useState("");
  const messageViewportRef = useRef<HTMLDivElement | null>(null);
  const selectedIdRef = useRef<string | null>(null);

  const scrollToMessage = useCallback((messageId?: string) => {
    const viewport = messageViewportRef.current;
    if (!viewport) return;

    if (messageId) {
      const target = viewport.querySelector<HTMLElement>(`[data-message-id="${CSS.escape(messageId)}"]`);
      if (target) {
        target.scrollIntoView({ block: "end", behavior: "smooth" });
        return;
      }
    }

    viewport.scrollTo({ top: viewport.scrollHeight, behavior: "smooth" });
  }, []);

  const selectedConversation = conversations.find((conversation) => conversation.id === selectedId) ?? null;
  const selectedConversationIcon = systemConversationIcon(selectedConversation);
  const selectedConversationIsNotification = isNotificationConversation(selectedConversation);
  const memberByUserId = useMemo(
    () => new Map(members.map((member) => [member.user_id, member])),
    [members]
  );
  const filteredConversations = useMemo(
    () =>
      conversations.filter((conversation) => {
        const title = conversationTitle(conversation).toLowerCase();
        const preview = (conversation.last_message?.body ?? "").toLowerCase();
        const query = conversationQuery.trim().toLowerCase();
        const matchesQuery = !query || title.includes(query) || preview.includes(query);
        const matchesFilter =
          conversationFilter === "all" ||
          (conversationFilter === "unread" && unreadConversationIds.has(conversation.id)) ||
          (conversationFilter === "groups" && conversation.type.toLowerCase() === "group");
        return matchesQuery && matchesFilter;
      }),
    [conversationFilter, conversationQuery, conversations, unreadConversationIds]
  );
  const sharedItems = useMemo(
    () =>
      messages
        .filter((message) => message.metadata && Object.keys(message.metadata).length > 0)
        .slice(-3)
        .reverse(),
    [messages]
  );

  const loadConversations = useCallback(async () => {
    setSidebarLoading(true);
    setError("");
    try {
      const response = await chatApi.listConversations(session.access_token);
      setConversations(response.conversations);
      setSelectedId((current) => current ?? response.conversations[0]?.id ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load conversations");
    } finally {
      setSidebarLoading(false);
    }
  }, [session.access_token]);

  const loadMessages = useCallback(async (conversationId: string, beforeMessageId?: string) => {
    const viewport = messageViewportRef.current;
    const previousScrollHeight = viewport?.scrollHeight ?? 0;
    setMessagesLoading(true);
    setError("");
    try {
      const response = await chatApi.listMessages(session.access_token, conversationId, beforeMessageId);
      if (response.members.length > 0) {
        setMembers((current) => mergeMembers(current, response.members));
      }
      setHasOlderMessages(response.pagination.has_more);
      setNextBeforeMessageId(response.pagination.next_before_message_id);
      setMessages((current) =>
        beforeMessageId ? mergeMessages(response.messages, current) : mergeMessages(response.messages)
      );
      window.setTimeout(() => {
        const currentViewport = messageViewportRef.current;
        if (!currentViewport) return;
        if (beforeMessageId) {
          currentViewport.scrollTop = currentViewport.scrollHeight - previousScrollHeight;
          return;
        }
        scrollToMessage(response.messages[response.messages.length - 1]?.id);
      }, 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load messages");
    } finally {
      setMessagesLoading(false);
    }
  }, [scrollToMessage, session.access_token]);

  const loadConversationMembers = useCallback(async (conversationId: string) => {
    setError("");
    try {
      const response = await chatApi.listConversationMembers(session.access_token, conversationId);
      setMembers(response.members);
      const profiles = response.members.map((member) => member.profile).filter(Boolean) as UserProfile[];
      setUserProfilesById(cacheUserProfiles(profiles));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load conversation members");
    }
  }, [session.access_token]);

  const handleRealtimeEvent = useCallback((event: RealtimeEvent) => {
    if (event.type !== "message.created") return;
    const message = event.message;

    setConversations((current) =>
      sortConversationsByActivity(
        current.map((conversation) =>
          conversation.id === event.conversation_id
            ? {
                ...conversation,
                last_message_id: message.id,
                last_message_at: message.created_at,
                last_message: message,
                updated_at: message.updated_at
              }
            : conversation
        )
      )
    );

    if (message.sender_id !== session.user_id && selectedIdRef.current !== event.conversation_id) {
      setUnreadConversationIds((current) => {
        const next = new Set(current);
        next.add(event.conversation_id);
        return next;
      });
    }

    if (selectedIdRef.current !== event.conversation_id) return;
    setMessages((current) => mergeMessages(current, [message]));
    window.setTimeout(() => scrollToMessage(message.id), 0);
  }, [scrollToMessage, session.user_id]);

  async function sendMessage(event: FormEvent) {
    event.preventDefault();
    const text = draft.trim();
    if (!text || !selectedId) return;

    setSending(true);
    setError("");
    try {
      const message = await chatApi.createMessage(session.access_token, selectedId, text);
      setMessages((current) => mergeMessages(current, [message]));
      setDraft("");
      await loadConversations();
      window.setTimeout(() => scrollToMessage(message.id), 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not send message");
    } finally {
      setSending(false);
    }
  }

  function addSelectedUser(user: SearchUserResult) {
    setSelectedUsers((current) => {
      if (current.some((item) => item.user_id === user.user_id)) return current;
      return [...current, user];
    });
    setCreateConversationError("");
  }

  function removeSelectedUser(userId: string) {
    setSelectedUsers((current) => current.filter((item) => item.user_id !== userId));
  }

  function closeCreateDialog() {
    setCreateDialogOpen(false);
    setNewConversationName("");
    setUserSearchQuery("");
    setUserSearchResults([]);
    setSelectedUsers([]);
    setCreateConversationError("");
  }

  async function createConversation(event: FormEvent) {
    event.preventDefault();
    if (selectedUsers.length === 0) {
      setCreateConversationError("Choose at least one user.");
      return;
    }

    setCreatingConversation(true);
    setCreateConversationError("");
    try {
      const created = await chatApi.createConversation(session.access_token, {
        name: newConversationName.trim() || undefined,
        member_user_ids: selectedUsers.map((user) => user.user_id)
      });
      setConversations((current) => sortConversationsByActivity(mergeConversations(current, [created])));
      setSelectedId(created.id);
      closeCreateDialog();
      await loadConversations();
      setSelectedId(created.id);
    } catch (err) {
      setCreateConversationError(err instanceof Error ? err.message : "Could not create conversation");
    } finally {
      setCreatingConversation(false);
    }
  }

  useEffect(() => {
    void loadConversations();
  }, [loadConversations]);

  useEffect(() => {
    if (!createDialogOpen) return;
    const searchKey = userSearchQuery.trim();
    if (searchKey.length < 2) {
      setUserSearchResults([]);
      setUserSearchLoading(false);
      return;
    }

    const controller = new AbortController();
    const timer = window.setTimeout(() => {
      setUserSearchLoading(true);
      userApi.searchUsers(session.access_token, searchKey)
        .then((users) => {
          if (controller.signal.aborted) return;
          setUserSearchResults(users);
          setCreateConversationError("");
        })
        .catch((err) => {
          if (controller.signal.aborted) return;
          setCreateConversationError(err instanceof Error ? err.message : "Could not search users");
        })
        .finally(() => {
          if (!controller.signal.aborted) setUserSearchLoading(false);
        });
    }, 300);

    return () => {
      controller.abort();
      window.clearTimeout(timer);
    };
  }, [createDialogOpen, session.access_token, userSearchQuery]);

  useEffect(() => {
    selectedIdRef.current = selectedId;
  }, [selectedId]);

  useEffect(() => {
    if (!selectedId) return;
    setMessages([]);
    setMembers([]);
    void loadConversationMembers(selectedId);
    void loadMessages(selectedId);
  }, [selectedId, loadConversationMembers, loadMessages]);

  useEffect(() => {
    const client = new RealtimeClient({
      token: session.access_token,
      onEvent: handleRealtimeEvent,
      onStatusChange: setRealtimeStatus
    });
    client.connect();
    return () => client.disconnect();
  }, [handleRealtimeEvent, session.access_token]);

  useEffect(() => {
    if (realtimeStatus !== "connected") return;
    void loadConversations();
    if (selectedIdRef.current) {
      void loadMessages(selectedIdRef.current);
    }
  }, [loadConversations, loadMessages, realtimeStatus]);

  function selectConversation(conversationId: string) {
    setSelectedId(conversationId);
    setUnreadConversationIds((current) => {
      if (!current.has(conversationId)) return current;
      const next = new Set(current);
      next.delete(conversationId);
      return next;
    });
  }

  return (
    <main className={`messenger-shell ${inspectorVisible ? "" : "inspector-collapsed"}`}>
      <div className="aurora aurora-a" />
      <div className="aurora aurora-b" />
      <div className="aurora aurora-c" />

      <aside className="space-rail" aria-label="Primary navigation">
        <ConversaLogo compact />
        <button className="rail-action active" title="Messages">C</button>
        <button className="rail-action" title="Friends">+</button>
        <button className="rail-action" title="Groups">#</button>
        <button className="rail-action" title="Notifications">.</button>
        <button className="rail-action bottom" title="Logout" onClick={onLogout}>
          Q
        </button>
      </aside>

      <aside className="conversation-panel glass-card">
        <header className="panel-title">
          <div>
            <p>Courier Conversa</p>
            <h1>Spaces</h1>
          </div>
          <button title="Create conversation" onClick={() => setCreateDialogOpen(true)}>+</button>
        </header>

        <section className="profile-glance">
          <Avatar label={session.display_name || `User ${session.user_id ?? "?"}`} imageUrl={session.avatar_url} tone="me" />
          <div>
            <strong>{session.display_name || `User #${session.user_id ?? "?"}`}</strong>
            <small>{realtimeStatus} · spatial glass</small>
          </div>
          <span className={`live-dot status-${realtimeStatus}`} />
        </section>

        <label className="space-search">
          <span>Search</span>
          <input
            value={conversationQuery}
            onChange={(event) => setConversationQuery(event.target.value)}
            placeholder="Find friends, groups, messages..."
          />
        </label>

        <nav className="smart-filters" aria-label="Conversation filters">
          <button className={conversationFilter === "all" ? "active" : ""} onClick={() => setConversationFilter("all")}>
            All
          </button>
          <button className={conversationFilter === "unread" ? "active" : ""} onClick={() => setConversationFilter("unread")}>
            Unread
          </button>
          <button className={conversationFilter === "groups" ? "active" : ""} onClick={() => setConversationFilter("groups")}>
            Groups
          </button>
        </nav>

        <div className="conversation-list">
          {sidebarLoading && <div className="empty-state">Loading conversations...</div>}
          {!sidebarLoading && conversations.length === 0 && <div className="empty-state">No conversations yet.</div>}
          {!sidebarLoading && conversations.length > 0 && filteredConversations.length === 0 && (
            <div className="empty-state">No matching conversations.</div>
          )}
          {filteredConversations.map((conversation) => {
            const isUnread = unreadConversationIds.has(conversation.id);
            const icon = systemConversationIcon(conversation);
            return (
              <button
                key={conversation.id}
                className={`conversation-item ${conversation.id === selectedId ? "selected" : ""} ${isUnread ? "unread" : ""} type-${conversation.type}`}
                onClick={() => selectConversation(conversation.id)}
              >
                <Avatar
                  label={conversationTitle(conversation)}
                  tone={conversation.type}
                  systemIcon={icon}
                />
                <span className="conversation-copy">
                  <strong>
                    <span className="conversation-title">{conversationTitle(conversation)}</span>
                    <ConversationTypeBadge type={conversation.type} />
                  </strong>
                  <small>{conversation.last_message?.body ?? "No messages yet"}</small>
                </span>
                <span className="conversation-meta">
                  <time>{formatConversationTime(conversation.last_message_at ?? conversation.updated_at)}</time>
                  {isUnread && <b>{unreadConversationIds.has(conversation.id) ? 1 : ""}</b>}
                </span>
              </button>
            );
          })}
        </div>
      </aside>

      <section className="chat-panel glass-card">
        {selectedConversation ? (
          <>
            <header className="chat-header">
              <div className="room-identity">
                <Avatar
                  label={conversationTitle(selectedConversation)}
                  tone={selectedConversation.type}
                  systemIcon={selectedConversationIcon}
                />
                <div>
                  <h2>{conversationTitle(selectedConversation)}</h2>
                  <span>{members.length || "?"} members · {selectedConversation.type} · {realtimeStatus}</span>
                </div>
              </div>
              <div className="chat-actions">
                <button title="Audio call">Call</button>
                <button title="Search messages">Find</button>
                <button
                  className={`icon-button split-toggle ${inspectorVisible ? "active" : ""}`}
                  type="button"
                  title={inspectorVisible ? "Hide conversation info" : "Show conversation info"}
                  onClick={() => setInspectorVisible((current) => !current)}
                >
                  <SplitPanelIcon />
                </button>
              </div>
            </header>

            {error && <div className="inline-error">{error}</div>}

            <div className="message-viewport" ref={messageViewportRef}>
              {hasOlderMessages && (
                <button
                  className="load-older-button"
                  disabled={messagesLoading}
                  onClick={() => selectedId && nextBeforeMessageId && loadMessages(selectedId, nextBeforeMessageId)}
                >
                  {messagesLoading ? "Loading..." : "Load older messages"}
                </button>
              )}
              {!messagesLoading && messages.length === 0 && <div className="empty-state">No messages in this conversation.</div>}
              {messages.map((message) => {
                const isMine = message.sender_id === session.user_id;
                const messageIcon = systemMessageIcon(message, selectedConversation);
                const isAssistant = messageIcon === "bot";
                const isNotification = messageIcon === "notification";
                const senderProfile = userProfilesById.get(message.sender_id) ?? memberByUserId.get(message.sender_id)?.profile;
                const senderName = senderDisplayName(message, selectedConversation, memberByUserId.get(message.sender_id), senderProfile);
                return (
                  <div
                    key={message.id}
                    className={`message-row ${isMine ? "mine" : ""} ${isAssistant ? "assistant" : ""} ${isNotification ? "notification" : ""}`}
                    data-message-id={message.id}
                  >
                    {!isMine && (
                      <Avatar
                        label={senderName}
                        imageUrl={senderProfile?.avatar_url}
                        systemIcon={messageIcon}
                      />
                    )}
                    <div className="message-bubble">
                      {!isMine && (
                        <header>
                          <strong>{senderName}</strong>
                        </header>
                      )}
                      <MessageBody message={message} />
                      <footer>
                        <time>{formatTime(message.created_at)}</time>
                        {isMine && <span>Sent</span>}
                      </footer>
                    </div>
                  </div>
                );
              })}
            </div>

            {!selectedConversationIsNotification && (
              <form className="composer" onSubmit={sendMessage}>
                <button type="button" title="Attach file">
                  +
                </button>
                <button type="button" title="Emoji">
                  :)
                </button>
                <input
                  value={draft}
                  onChange={(event) => setDraft(event.target.value)}
                  placeholder="Write a message..."
                />
                <button type="button" title="Voice message">
                  Mic
                </button>
                <button className="primary-button" disabled={sending || !draft.trim()}>
                  Send
                </button>
              </form>
            )}
          </>
        ) : (
          <div className="no-selection">
            <ConversaLogo />
            <h2>Select a conversation</h2>
            <p className="muted">Your conversations will appear on the left.</p>
          </div>
        )}
      </section>

      {createDialogOpen && (
        <div className="modal-backdrop" role="presentation" onMouseDown={closeCreateDialog}>
          <form className="create-conversation-dialog" onSubmit={createConversation} onMouseDown={(event) => event.stopPropagation()}>
            <span className="sheet-handle" aria-hidden="true" />
            <header className="create-dialog-header">
              <h2>New Conversation</h2>
              <button className="dialog-close" type="button" title="Close" onClick={closeCreateDialog}>×</button>
            </header>

            <div className="dialog-form-stack">
              <label className="dialog-field name-field">
                <span>Name</span>
                <input
                  value={newConversationName}
                  onChange={(event) => setNewConversationName(event.target.value)}
                  placeholder="Optional conversation name"
                  maxLength={255}
                />
              </label>

              <label className="dialog-field add-users-field">
                <span>Add users</span>
                <div className="search-input-shell">
                  <input
                    value={userSearchQuery}
                    onChange={(event) => setUserSearchQuery(event.target.value)}
                    placeholder="Search by name, phone, or email"
                    autoFocus
                  />
                  <b>Add</b>
                </div>
              </label>
            </div>

            {selectedUsers.length > 0 && (
              <div className="selected-user-list" aria-label="Selected users">
                {selectedUsers.map((user) => (
                  <button key={user.user_id} className="selected-user-chip" type="button" onClick={() => removeSelectedUser(user.user_id)}>
                    <Avatar label={user.display_name} imageUrl={user.avatar} />
                    <span>{user.display_name}</span>
                    <b>×</b>
                  </button>
                ))}
              </div>
            )}

            <div className="user-search-results">
              {userSearchLoading && <div className="empty-state">Searching users...</div>}
              {!userSearchLoading && userSearchQuery.trim().length > 0 && userSearchQuery.trim().length < 2 && (
                <div className="empty-state">Type at least 2 characters.</div>
              )}
              {!userSearchLoading && userSearchQuery.trim().length >= 2 && userSearchResults.length === 0 && (
                <div className="empty-state">No verified users found.</div>
              )}
              {!userSearchLoading && userSearchResults.map((user) => {
                const selected = selectedUsers.some((item) => item.user_id === user.user_id);
                return (
                  <button key={user.user_id} className={selected ? "selected" : ""} type="button" disabled={selected} onClick={() => addSelectedUser(user)}>
                    <Avatar label={user.display_name} imageUrl={user.avatar} />
                    <span>
                      <strong>{user.display_name}</strong>
                      <small>{user.email || user.phone_number}</small>
                    </span>
                    <b>{selected ? "Added" : "Add"}</b>
                  </button>
                );
              })}
            </div>

            {createConversationError && <div className="inline-error">{createConversationError}</div>}

            <footer>
              <button type="button" onClick={closeCreateDialog}>Cancel</button>
              <button className="primary-button" disabled={creatingConversation || selectedUsers.length === 0}>
                {creatingConversation ? "Creating..." : "Create"}
              </button>
            </footer>
          </form>
        </div>
      )}

      {inspectorVisible && <aside className="inspector-panel glass-card">
        {selectedConversation ? (
          <>
            <section className="room-card">
              <Avatar
                label={conversationTitle(selectedConversation)}
                tone={selectedConversation.type}
                large
                systemIcon={selectedConversationIcon}
              />
              <h2>{conversationTitle(selectedConversation)}</h2>
              <small>{selectedConversation.type} conversation</small>
              <div className="member-pile">
                {members.slice(0, 4).map((member) => (
                  <Avatar
                    key={member.id}
                    label={member.profile?.display_name || userProfilesById.get(member.user_id)?.display_name || `User ${member.user_id}`}
                    imageUrl={member.profile?.avatar_url || userProfilesById.get(member.user_id)?.avatar_url}
                  />
                ))}
                {members.length === 0 && (
                  <>
                    <Avatar label="T" tone="me" />
                    <Avatar label="C" tone="system" />
                  </>
                )}
              </div>
            </section>

            <section className="inspector-section">
              <h3>Conversation management</h3>
              <button>Members and roles</button>
              <button>Notifications and mute</button>
              <button>Pin topic</button>
            </section>

            <section className="inspector-section">
              <h3>Shared context</h3>
              {sharedItems.length === 0 ? (
                <article className="shared-item">
                  <b>No shared metadata yet</b>
                  <small>Attachments and links can appear here later.</small>
                </article>
              ) : (
                sharedItems.map((message) => (
                  <article key={message.id} className="shared-item">
                    <b>{message.type} message</b>
                    <small>{formatTime(message.created_at)} · metadata available</small>
                  </article>
                ))
              )}
            </section>

            <section className="inspector-section">
              <h3>User search</h3>
              <label className="mini-search">
                <span>Search</span>
                <input placeholder="Invite user to this space" />
              </label>
              <button>Invite user</button>
            </section>
          </>
        ) : (
          <div className="empty-state">Select a conversation to inspect its space.</div>
        )}
      </aside>}
    </main>
  );
}

function ConversationTypeBadge({ type }: { type: Conversation["type"] }) {
  const normalizedType = type.toLowerCase();
  return (
    <small className={`conversation-type-badge type-${normalizedType}`}>
      {normalizedType}
    </small>
  );
}

function ConversaLogo({ compact = false }: { compact?: boolean }) {
  return (
    <div className={`conversa-logo ${compact ? "compact" : ""}`} aria-label="Conversa logo">
      <img src={compact ? "/brand/conversa-icon.png" : "/brand/conversa-wordmark.png"} alt="" />
    </div>
  );
}

function GoogleLogo() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#4285f4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09Z" />
      <path fill="#34a853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23Z" />
      <path fill="#fbbc05" d="M5.84 14.1c-.22-.66-.35-1.36-.35-2.1s.13-1.44.35-2.1V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l3.66-2.84Z" />
      <path fill="#ea4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06L5.84 9.9C6.71 7.31 9.14 5.38 12 5.38Z" />
    </svg>
  );
}

function GitHubLogo() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path
        fill="currentColor"
        d="M12 .5A11.5 11.5 0 0 0 8.36 22.9c.58.11.79-.25.79-.56v-2.02c-3.22.7-3.9-1.38-3.9-1.38-.53-1.34-1.29-1.7-1.29-1.7-1.05-.72.08-.7.08-.7 1.16.08 1.77 1.19 1.77 1.19 1.04 1.77 2.71 1.26 3.37.96.1-.75.4-1.26.73-1.55-2.57-.29-5.28-1.29-5.28-5.73 0-1.27.45-2.3 1.19-3.11-.12-.29-.52-1.47.11-3.06 0 0 .98-.31 3.18 1.19A10.95 10.95 0 0 1 12 6.04c.98 0 1.96.13 2.88.39 2.2-1.5 3.17-1.19 3.17-1.19.64 1.59.24 2.77.12 3.06.74.81 1.18 1.84 1.18 3.11 0 4.46-2.71 5.43-5.29 5.72.42.36.79 1.07.79 2.16v3.05c0 .31.21.67.8.56A11.5 11.5 0 0 0 12 .5Z"
      />
    </svg>
  );
}

function MessageBody({ message }: { message: Message }) {
  if (isAssistantMessage(message)) {
    return <div className="message-body markdown-body">{renderAssistantMarkdown(message.body)}</div>;
  }

  return <div className="message-body">{message.body}</div>;
}

function Avatar({
  label,
  imageUrl,
  tone,
  large = false,
  systemIcon
}: {
  label: string;
  imageUrl?: string;
  tone?: string;
  large?: boolean;
  systemIcon?: SystemIcon;
}) {
  const letter = label.trim().charAt(0).toUpperCase() || "C";
  const normalizedTone = tone?.toLowerCase().replace(/[^a-z0-9_-]/g, "") || "default";
  if (systemIcon) {
    return (
      <span className={`avatar system-avatar ${systemIcon}-avatar tone-system ${large ? "large" : ""}`} aria-label={label}>
        {systemIcon === "bot" ? <BotAvatarIcon /> : <NotificationAvatarIcon />}
      </span>
    );
  }

  return (
    <span className={`avatar tone-${normalizedTone} ${large ? "large" : ""}`}>
      {imageUrl ? <img src={imageUrl} alt="" /> : letter}
    </span>
  );
}

function SplitPanelIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="4" y="5" width="16" height="14" rx="2.5" />
      <path d="M14 5v14" />
    </svg>
  );
}

function BotAvatarIcon() {
  return (
    <svg viewBox="0 0 32 32" aria-hidden="true">
      <path className="bot-antenna" d="M16 7v-3" />
      <circle className="bot-signal" cx="16" cy="3.5" r="1.8" />
      <rect className="bot-face" x="7" y="9" width="18" height="15" rx="6" />
      <path className="bot-mouth" d="M13 19.5h6" />
      <circle className="bot-eye" cx="13" cy="15.5" r="1.7" />
      <circle className="bot-eye" cx="19" cy="15.5" r="1.7" />
      <path className="bot-ear" d="M7 16h-2M27 16h-2" />
    </svg>
  );
}

function NotificationAvatarIcon() {
  return (
    <svg viewBox="0 0 32 32" aria-hidden="true">
      <path className="notification-bell" d="M10 22h12l-1.4-2.5v-4.2a4.6 4.6 0 0 0-9.2 0v4.2L10 22Z" />
      <path className="notification-clapper" d="M14 24.5a2.4 2.4 0 0 0 4 0" />
      <path className="notification-spark" d="M8 11.5l-2-2M24 11.5l2-2" />
      <circle className="notification-dot" cx="22.5" cy="9.5" r="3" />
    </svg>
  );
}

function conversationTitle(conversation: Conversation) {
  if (isAssistantConversation(conversation)) return "Courier Assistant";
  if (isNotificationConversation(conversation)) return "Notification";
  return conversation.name || `${conversation.type} #${conversation.id}`;
}

type SystemIcon = "bot" | "notification";

function systemConversationIcon(conversation?: Conversation | null): SystemIcon | undefined {
  if (!conversation) return undefined;
  if (isAssistantConversation(conversation)) return "bot";
  if (isNotificationConversation(conversation)) return "notification";
  return undefined;
}

function systemMessageIcon(message: Message, conversation?: Conversation | null): SystemIcon | undefined {
  if (String(message.sender_id) !== "0") return undefined;
  return systemConversationIcon(conversation);
}

function isAssistantConversation(conversation?: Conversation | null) {
  return conversation?.type.toLowerCase() === "system" && (conversation.name ?? "").trim().toLowerCase() === "assistant";
}

function isNotificationConversation(conversation?: Conversation | null) {
  return conversation?.type.toLowerCase() === "system" && (conversation.name ?? "").trim().toLowerCase() === "notification";
}

function formatTime(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}

function formatConversationTime(value: string) {
  const date = new Date(value);
  const now = new Date();
  const sameDay = date.toDateString() === now.toDateString();
  if (sameDay) return formatTime(value);
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "2-digit"
  }).format(date);
}

function senderDisplayName(
  message: Message,
  conversation?: Conversation | null,
  member?: ListMessagesResponse["members"][number],
  profile?: UserProfile
) {
  if (String(message.sender_id) === "0" && isNotificationConversation(conversation)) return "Courier Notification";
  if (isAssistantMessage(message)) return "Courier Assistant";
  if (profile?.display_name) return profile.display_name;
  if (member?.profile?.display_name) return member.profile.display_name;
  const metadataName = readMetadataString(message.metadata, "sender_display_name", "sender_name", "display_name", "name");
  if (metadataName) return metadataName;
  if (member?.role && member.role !== "member") return `${member.role} · User ${message.sender_id}`;
  return `User ${message.sender_id}`;
}

function isAssistantMessage(message: Message) {
  return String(message.sender_id) === "0";
}

function readMetadataString(metadata: Record<string, unknown>, ...keys: string[]) {
  for (const key of keys) {
    const value = metadata[key];
    if (typeof value === "string" && value.trim()) return value.trim();
  }
  return "";
}

function sortMessagesByCreatedAt(messages: Message[]) {
  return [...messages].sort((left, right) => {
    const leftTime = new Date(left.created_at).getTime();
    const rightTime = new Date(right.created_at).getTime();
    if (leftTime !== rightTime) return leftTime - rightTime;
    return compareSnowflakeIds(left.id, right.id);
  });
}

function mergeMessages(...messageGroups: Message[][]) {
  const messagesById = new Map<string, Message>();
  for (const messages of messageGroups) {
    for (const message of messages) {
      messagesById.set(message.id, message);
    }
  }
  return sortMessagesByCreatedAt([...messagesById.values()]);
}

function renderAssistantMarkdown(body: string) {
  const blocks: ReactNode[] = [];
  const lines = body.split(/\r?\n/);
  let paragraph: string[] = [];
  let bullets: string[] = [];

  const flushParagraph = () => {
    if (paragraph.length === 0) return;
    blocks.push(
      <p key={`p-${blocks.length}`}>
        {renderInlineMarkdown(paragraph.join(" "))}
      </p>
    );
    paragraph = [];
  };

  const flushBullets = () => {
    if (bullets.length === 0) return;
    blocks.push(
      <ul key={`ul-${blocks.length}`}>
        {bullets.map((item, index) => (
          <li key={`${index}-${item}`}>{renderInlineMarkdown(item)}</li>
        ))}
      </ul>
    );
    bullets = [];
  };

  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (!line) {
      flushParagraph();
      flushBullets();
      continue;
    }

    const bullet = line.match(/^[-*]\s+(.+)$/);
    if (bullet) {
      flushParagraph();
      bullets.push(bullet[1]);
      continue;
    }

    flushBullets();
    paragraph.push(line);
  }

  flushParagraph();
  flushBullets();

  return blocks.length > 0 ? blocks : <p>{body}</p>;
}

function renderInlineMarkdown(text: string) {
  const nodes: ReactNode[] = [];
  const parts = text.split(/(\*\*[^*]+\*\*)/g);

  for (const part of parts) {
    if (!part) continue;
    if (part.startsWith("**") && part.endsWith("**")) {
      nodes.push(<strong key={nodes.length}>{part.slice(2, -2)}</strong>);
      continue;
    }
    nodes.push(part);
  }

  return nodes;
}

function mergeMembers(...memberGroups: ListMessagesResponse["members"][]) {
  const membersByUserID = new Map<string, ListMessagesResponse["members"][number]>();
  for (const members of memberGroups) {
    for (const member of members) {
      const current = membersByUserID.get(member.user_id);
      membersByUserID.set(member.user_id, {
        ...current,
        ...member,
        profile: member.profile ?? current?.profile
      });
    }
  }
  return [...membersByUserID.values()];
}

function mergeConversations(...conversationGroups: Conversation[][]) {
  const conversationsById = new Map<string, Conversation>();
  for (const conversations of conversationGroups) {
    for (const conversation of conversations) {
      conversationsById.set(conversation.id, {
        ...conversationsById.get(conversation.id),
        ...conversation
      });
    }
  }
  return [...conversationsById.values()];
}

function sortConversationsByActivity(conversations: Conversation[]) {
  return [...conversations].sort((left, right) => {
    const leftTime = new Date(left.last_message_at ?? left.created_at).getTime();
    const rightTime = new Date(right.last_message_at ?? right.created_at).getTime();
    if (leftTime !== rightTime) return rightTime - leftTime;
    return compareSnowflakeIds(right.id, left.id);
  });
}

function compareSnowflakeIds(left: string, right: string) {
  const leftId = BigInt(left);
  const rightId = BigInt(right);
  return leftId < rightId ? -1 : leftId > rightId ? 1 : 0;
}
