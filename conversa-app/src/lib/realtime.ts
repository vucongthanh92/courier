import { CHAT_WS_URL } from "../config";
import type { RealtimeEvent } from "../types";

type RealtimeClientOptions = {
  token: string;
  onEvent: (event: RealtimeEvent) => void;
  onStatusChange?: (status: RealtimeStatus) => void;
};

export type RealtimeStatus = "connecting" | "connected" | "disconnected";

export class RealtimeClient {
  private socket: WebSocket | null = null;
  private reconnectTimer: number | undefined;
  private reconnectAttempt = 0;
  private closed = false;

  constructor(private readonly options: RealtimeClientOptions) {}

  connect() {
    this.closed = false;
    this.setStatus("connecting");
    const url = new URL(CHAT_WS_URL);
    url.searchParams.set("access_token", this.options.token);

    this.socket = new WebSocket(url.toString());
    this.socket.onopen = () => {
      this.reconnectAttempt = 0;
      this.setStatus("connected");
    };
    this.socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data) as RealtimeEvent;
        if (payload.type === "message.created") {
          this.options.onEvent(payload);
        }
      } catch {
        // Ignore malformed realtime events; REST recovery remains the source of truth.
      }
    };
    this.socket.onclose = () => this.scheduleReconnect();
    this.socket.onerror = () => {
      this.socket?.close();
    };
  }

  disconnect() {
    this.closed = true;
    window.clearTimeout(this.reconnectTimer);
    this.socket?.close();
    this.socket = null;
    this.setStatus("disconnected");
  }

  private scheduleReconnect() {
    if (this.closed) return;
    this.setStatus("disconnected");
    window.clearTimeout(this.reconnectTimer);
    const delay = Math.min(1000 * 2 ** this.reconnectAttempt, 10000);
    this.reconnectAttempt += 1;
    this.reconnectTimer = window.setTimeout(() => this.connect(), delay);
  }

  private setStatus(status: RealtimeStatus) {
    this.options.onStatusChange?.(status);
  }
}
