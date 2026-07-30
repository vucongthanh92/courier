export type ApiResponse<T> = {
  success: boolean;
  data: T | null;
  errors: Array<{ code?: string; message?: string; field?: string }>;
};

export type JwtTokenResponse = {
  access_token: string;
  expires_in: number;
  refresh_token: string;
  refresh_expires_in: number;
  token_type: string;
  need_password_setup?: boolean;
};

export type LoginRequest = {
  email: string;
  password: string;
};

export type SignupRequest = {
  phone_number: string;
  email: string;
  password: string;
  display_name: string;
};

export type VerifyEmailRequest = {
  email: string;
  token: string;
};

export type Conversation = {
  id: string;
  type: "direct" | "group" | string;
  direct_key?: string;
  name?: string;
  created_by: string;
  last_message_id?: string;
  last_message_at?: string;
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  last_message?: Message;
};

export type ConversationListResponse = {
  conversations: Conversation[];
  pagination: {
    limit: number;
    has_more: boolean;
    next_before_last_message_at?: string;
    next_before_conversation_id?: string;
  };
};

export type Member = {
  id: string;
  conversation_id: string;
  user_id: string;
  role: string;
  status: string;
  joined_at: string;
  last_read_message_id?: string;
  last_read_at?: string;
};

export type Message = {
  id: string;
  conversation_id: string;
  sender_id: string;
  type: "text" | string;
  body: string;
  reply_to_message_id?: string;
  client_message_id?: string;
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  edited_at?: string;
};

export type ListMessagesResponse = {
  conversation_id: string;
  messages: Message[];
  members: Member[];
  pagination: {
    limit: number;
    next_before_message_id?: string;
    has_more: boolean;
  };
};

export type MessageCreatedEvent = {
  type: "message.created";
  conversation_id: string;
  recipient_user_ids: number[];
  message: Message;
  event_at: string;
};

export type RealtimeEvent = MessageCreatedEvent;
