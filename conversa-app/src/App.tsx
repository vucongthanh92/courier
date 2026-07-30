import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { authApi, chatApi } from "./lib/api";
import { RealtimeClient, type RealtimeStatus } from "./lib/realtime";
import { clearSession, readSession, saveSession, type Session } from "./lib/session";
import type { Conversation, ListMessagesResponse, Message, RealtimeEvent } from "./types";

type AuthMode = "login" | "signup" | "verify";

export function App() {
  const [session, setSession] = useState<Session | null>(() => readSession());

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
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

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

  return (
    <main className="auth-page">
      <section className="auth-card">
        <div className="brand-mark">C</div>
        <h1>Conversa</h1>
        <p className="muted">A lightweight Courier messaging client.</p>

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

function MessengerShell({ session, onLogout }: { session: Session; onLogout: () => void }) {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [members, setMembers] = useState<ListMessagesResponse["members"]>([]);
  const [hasOlderMessages, setHasOlderMessages] = useState(false);
  const [nextBeforeMessageId, setNextBeforeMessageId] = useState<string | undefined>();
  const [sidebarLoading, setSidebarLoading] = useState(true);
  const [messagesLoading, setMessagesLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [realtimeStatus, setRealtimeStatus] = useState<RealtimeStatus>("disconnected");
  const [unreadConversationIds, setUnreadConversationIds] = useState<Set<string>>(() => new Set());
  const [draft, setDraft] = useState("");
  const [error, setError] = useState("");
  const messageViewportRef = useRef<HTMLDivElement | null>(null);
  const selectedIdRef = useRef<string | null>(null);

  const selectedConversation = conversations.find((conversation) => conversation.id === selectedId) ?? null;
  const memberByUserId = useMemo(
    () => new Map(members.map((member) => [member.user_id, member])),
    [members]
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
      setMembers(response.members);
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
        currentViewport.scrollTop = currentViewport.scrollHeight;
      }, 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load messages");
    } finally {
      setMessagesLoading(false);
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
    window.setTimeout(() => {
      const viewport = messageViewportRef.current;
      if (!viewport) return;
      const distanceFromBottom = viewport.scrollHeight - viewport.scrollTop - viewport.clientHeight;
      if (distanceFromBottom < 160 || message.sender_id === session.user_id) {
        viewport.scrollTop = viewport.scrollHeight;
      }
    }, 0);
  }, [session.user_id]);

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
      window.setTimeout(() => {
        const viewport = messageViewportRef.current;
        if (viewport) viewport.scrollTop = viewport.scrollHeight;
      }, 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not send message");
    } finally {
      setSending(false);
    }
  }

  useEffect(() => {
    void loadConversations();
  }, [loadConversations]);

  useEffect(() => {
    selectedIdRef.current = selectedId;
  }, [selectedId]);

  useEffect(() => {
    if (!selectedId) return;
    setMessages([]);
    setMembers([]);
    void loadMessages(selectedId);
  }, [selectedId, loadMessages]);

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
    <main className="messenger-shell">
      <aside className="sidebar">
        <div className="sidebar-header">
          <div>
            <h1>Conversa</h1>
            <span>User #{session.user_id ?? "?"} · {realtimeStatus}</span>
          </div>
          <button className="text-button" onClick={onLogout}>
            Logout
          </button>
        </div>

        <div className="conversation-list">
          {sidebarLoading && <div className="empty-state">Loading conversations...</div>}
          {!sidebarLoading && conversations.length === 0 && <div className="empty-state">No conversations yet.</div>}
          {conversations.map((conversation) => {
            const isUnread = unreadConversationIds.has(conversation.id);
            return (
              <button
                key={conversation.id}
                className={`conversation-item ${conversation.id === selectedId ? "selected" : ""} ${isUnread ? "unread" : ""} type-${conversation.type}`}
                onClick={() => selectConversation(conversation.id)}
              >
                <Avatar label={conversation.name || conversation.type || String(conversation.id)} />
                <span>
                  <strong>
                    {conversationTitle(conversation)}
                    <ConversationTypeBadge type={conversation.type} />
                  </strong>
                  <small>{conversation.last_message?.body ?? "No messages yet"}</small>
                </span>
              </button>
            );
          })}
        </div>
      </aside>

      <section className="chat-panel">
        {selectedConversation ? (
          <>
            <header className="chat-header">
              <Avatar label={conversationTitle(selectedConversation)} />
              <div>
                <h2>{conversationTitle(selectedConversation)}</h2>
                <span>{selectedConversation.type}</span>
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
                const label = isMine ? "Me" : `U${message.sender_id}`;
                return (
                  <div key={message.id} className={`message-row ${isMine ? "mine" : ""}`}>
                    {!isMine && <Avatar label={memberByUserId.get(message.sender_id)?.role ?? label} />}
                    <div className="message-bubble">
                      <span>{message.body}</span>
                      <time>{formatTime(message.created_at)}</time>
                    </div>
                    {isMine && <Avatar label="Me" />}
                  </div>
                );
              })}
            </div>

            <form className="composer" onSubmit={sendMessage}>
              <input
                value={draft}
                onChange={(event) => setDraft(event.target.value)}
                placeholder="Write a message..."
              />
              <button className="primary-button" disabled={sending || !draft.trim()}>
                Send
              </button>
            </form>
          </>
        ) : (
          <div className="no-selection">
            <div className="brand-mark">C</div>
            <h2>Select a conversation</h2>
            <p className="muted">Your conversations will appear on the left.</p>
          </div>
        )}
      </section>
    </main>
  );
}

function ConversationTypeBadge({ type }: { type: Conversation["type"] }) {
  const normalizedType = type.toLowerCase();
  const icon = normalizedType === "group" ? "👥" : normalizedType === "system" ? "⚙️" : "💬";
  return (
    <small className={`conversation-type-badge type-${normalizedType}`}>
      <span aria-hidden="true">{icon}</span>
      {normalizedType}
    </small>
  );
}

function Avatar({ label }: { label: string }) {
  const letter = label.trim().charAt(0).toUpperCase() || "C";
  return <span className="avatar">{letter}</span>;
}

function conversationTitle(conversation: Conversation) {
  return conversation.name || `${conversation.type} #${conversation.id}`;
}

function formatTime(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
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
