export const PAYMENT_GATEWAY_API_BASE_URL =
  import.meta.env.VITE_PAYMENT_GATEWAY_API_BASE_URL ?? "http://localhost:5003/api/v1";

export const STANDALONE_ACCESS_TOKEN = import.meta.env.VITE_COURIER_ACCESS_TOKEN ?? "";

export const CONVERSA_APP_URL = import.meta.env.VITE_CONVERSA_APP_URL ?? "http://localhost:8080";

export function conversaDestination() {
  const configured = new URL(CONVERSA_APP_URL);
  const returnTo = new URLSearchParams(window.location.search).get("return_to");
  if (!returnTo) return configured.toString();

  try {
    const candidate = new URL(returnTo);
    return candidate.origin === configured.origin ? candidate.toString() : configured.toString();
  } catch {
    return configured.toString();
  }
}
