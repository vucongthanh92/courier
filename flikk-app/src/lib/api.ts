import { PAYMENT_GATEWAY_API_BASE_URL } from "../config";
import type { ApiResponse, CheckoutInstruction, WalletBalance } from "../types";

async function request<T>(path: string, token: string, init: RequestInit = {}) {
  const response = await fetch(`${PAYMENT_GATEWAY_API_BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...init.headers
    }
  });

  const payload = (await response.json().catch(() => null)) as ApiResponse<T> | null;
  if (!response.ok || !payload?.success || payload.data === null) {
    const message = payload?.errors?.map((error) => error.message).filter(Boolean).join(", ");
    throw new Error(message || `Request failed with status ${response.status}`);
  }

  return payload.data;
}

export const walletApi = {
  getBalance(token: string) {
    return request<WalletBalance>("/wallet/balance", token);
  },
  createTopUp(token: string, amountMinor: number) {
    return request<CheckoutInstruction>("/wallet/top-up", token, {
      method: "POST",
      headers: { "Idempotency-Key": crypto.randomUUID() },
      body: JSON.stringify({
        amount_minor: amountMinor,
        method: "bank_transfer",
        provider_name: "sepay"
      })
    });
  }
};

export function submitCheckout(instruction: CheckoutInstruction) {
  const form = document.createElement("form");
  form.method = "POST";
  form.action = instruction.checkout_action;

  for (const [name, value] of Object.entries(instruction.checkout_fields)) {
    const input = document.createElement("input");
    input.type = "hidden";
    input.name = name;
    input.value = value;
    form.append(input);
  }

  document.body.append(form);
  form.submit();
}
