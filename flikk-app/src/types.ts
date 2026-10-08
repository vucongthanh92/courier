export type WalletBalance = {
  wallet_id: string | null;
  currency: string;
  status: "active" | "restricted" | "closed" | "not_created";
  available_minor: number;
  pending_minor: number;
  held_minor: number;
  updated_at: string | null;
};

export type CheckoutInstruction = {
  topup_id: string;
  invoice_number: string;
  payment_code: string;
  expires_at: string;
  checkout_action: string;
  checkout_fields: Record<string, string>;
};

export type ApiResponse<T> = {
  success: boolean;
  data: T | null;
  errors: Array<{ message: string; code: string; field?: string }> | null;
};

export type AuthenticatedUser = {
  id: string;
  display_name: string;
  avatar_url?: string;
};

export type JwtTokenResponse = {
  access_token: string;
  expires_in: number;
  refresh_token: string;
  refresh_expires_in: number;
  id_token?: string;
  token_type: string;
  need_password_setup?: boolean;
  user?: AuthenticatedUser;
};

export type SsoTokenRequest = {
  grant_type: "authorization_code";
  client_id: string;
  code: string;
  redirect_uri: string;
  code_verifier: string;
};

export type NavigationItem = "home" | "activity" | "cards" | "settings";
