import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { authApi, chatApi } from "./lib/api";
import { RealtimeClient, type RealtimeStatus } from "./lib/realtime";
import { clearSession, readSession, saveSession, type Session } from "./lib/session";
import { cacheUserProfiles, readUserProfileCache } from "./lib/userProfileCache";
import type { Conversation, ListMessagesResponse, Message, RealtimeEvent, UserProfile } from "./types";

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
      <div className="aurora aurora-a" />
      <div className="aurora aurora-b" />
      <section className="auth-card">
        <ConversaLogo />
        <h1>Conversa</h1>
        <p className="muted">Spatial messaging for Courier teams.</p>

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
  const messageViewportRef = useRef<HTMLDivElement | null>(null);
  const selectedIdRef = useRef<string | null>(null);

  const selectedConversation = conversations.find((conversation) => conversation.id === selectedId) ?? null;
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
        currentViewport.scrollTop = currentViewport.scrollHeight;
      }, 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load messages");
    } finally {
      setMessagesLoading(false);
    }
  }, [session.access_token]);

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
    <main className="messenger-shell">
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
          <button title="Create conversation">+</button>
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
            return (
              <button
                key={conversation.id}
                className={`conversation-item ${conversation.id === selectedId ? "selected" : ""} ${isUnread ? "unread" : ""} type-${conversation.type}`}
                onClick={() => selectConversation(conversation.id)}
              >
                <Avatar label={conversation.name || conversation.type || String(conversation.id)} tone={conversation.type} />
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
                <Avatar label={conversationTitle(selectedConversation)} tone={selectedConversation.type} />
                <div>
                  <h2>{conversationTitle(selectedConversation)}</h2>
                  <span>{members.length || "?"} members · {selectedConversation.type} · {realtimeStatus}</span>
                </div>
              </div>
              <div className="chat-actions">
                <button title="Audio call">Call</button>
                <button title="Search messages">Find</button>
                <button title="Conversation info">Info</button>
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
                const senderProfile = userProfilesById.get(message.sender_id) ?? memberByUserId.get(message.sender_id)?.profile;
                const senderName = senderDisplayName(message, memberByUserId.get(message.sender_id), senderProfile);
                return (
                  <div key={message.id} className={`message-row ${isMine ? "mine" : ""}`}>
                    {!isMine && <Avatar label={senderName} imageUrl={senderProfile?.avatar_url} />}
                    <div className="message-bubble">
                      {!isMine && (
                        <header>
                          <strong>{senderName}</strong>
                        </header>
                      )}
                      <span>{message.body}</span>
                      <footer>
                        <time>{formatTime(message.created_at)}</time>
                        {isMine && <span>Sent</span>}
                      </footer>
                    </div>
                  </div>
                );
              })}
            </div>

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
          </>
        ) : (
          <div className="no-selection">
            <ConversaLogo />
            <h2>Select a conversation</h2>
            <p className="muted">Your conversations will appear on the left.</p>
          </div>
        )}
      </section>

      <aside className="inspector-panel glass-card">
        {selectedConversation ? (
          <>
            <section className="room-card">
              <Avatar label={conversationTitle(selectedConversation)} tone={selectedConversation.type} large />
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
      </aside>
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

function Avatar({ label, imageUrl, tone, large = false }: { label: string; imageUrl?: string; tone?: string; large?: boolean }) {
  const letter = label.trim().charAt(0).toUpperCase() || "C";
  const normalizedTone = tone?.toLowerCase().replace(/[^a-z0-9_-]/g, "") || "default";
  return (
    <span className={`avatar tone-${normalizedTone} ${large ? "large" : ""}`}>
      {imageUrl ? <img src={imageUrl} alt="" /> : letter}
    </span>
  );
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

function senderDisplayName(message: Message, member?: ListMessagesResponse["members"][number], profile?: UserProfile) {
  if (profile?.display_name) return profile.display_name;
  if (member?.profile?.display_name) return member.profile.display_name;
  const metadataName = readMetadataString(message.metadata, "sender_display_name", "sender_name", "display_name", "name");
  if (metadataName) return metadataName;
  if (member?.role && member.role !== "member") return `${member.role} · User ${message.sender_id}`;
  return `User ${message.sender_id}`;
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
