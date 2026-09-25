import { useCallback, useEffect, useMemo, useState } from "react";
import { conversaDestination } from "./config";
import { submitCheckout, walletApi } from "./lib/api";
import { readAccessToken, readDisplayName } from "./lib/session";
import type { CheckoutInstruction, NavigationItem, WalletBalance } from "./types";

const DEMO_BALANCE: WalletBalance = {
  wallet_id: "FLK-08A3",
  currency: "VND",
  status: "active",
  available_minor: 1280000,
  pending_minor: 0,
  held_minor: 0,
  updated_at: new Date().toISOString()
};

const activity = [
  { name: "Top up via SePay", detail: "Today · 10:42", value: "+500,000 ₫", icon: "↓", tone: "mint" },
  { name: "Ride · District 1", detail: "Yesterday · 18:26", value: "−68,000 ₫", icon: "↗", tone: "blue" },
  { name: "Courier Plus", detail: "12 Sep · Subscription", value: "−99,000 ₫", icon: "✦", tone: "violet" }
];

const navItems: Array<{ id: NavigationItem; label: string; icon: string }> = [
  { id: "home", label: "Overview", icon: "⌂" },
  { id: "activity", label: "Activity", icon: "◷" },
  { id: "cards", label: "Payment methods", icon: "▣" },
  { id: "settings", label: "Settings", icon: "⚙" }
];

export function App() {
  const [token] = useState(readAccessToken);
  const [name] = useState(readDisplayName);
  const [activeNav, setActiveNav] = useState<NavigationItem>("home");
  const [balance, setBalance] = useState<WalletBalance | null>(token ? null : DEMO_BALANCE);
  const [loading, setLoading] = useState(Boolean(token));
  const [error, setError] = useState("");
  const [topUpOpen, setTopUpOpen] = useState(false);
  const [roadmapOpen, setRoadmapOpen] = useState<"transfer" | "withdraw" | null>(null);

  const refreshBalance = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError("");
    try {
      setBalance(await walletApi.getBalance(token));
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Could not load your wallet.");
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    void refreshBalance();
  }, [refreshBalance]);

  const initials = useMemo(
    () => name.split(" ").filter(Boolean).slice(0, 2).map((word) => word[0]).join("").toUpperCase() || "CM",
    [name]
  );

  function openConversa() {
    window.location.assign(conversaDestination());
  }

  return (
    <main className="flikk-shell">
      <div className="grid-overlay" />
      <div className="orb orb-blue" />
      <div className="orb orb-mint" />
      <div className="orb orb-amber" />

      <aside className="side-rail glass-panel">
        <div className="brand-lockup" aria-label="flikk home">
          <div className="brand-mark">f</div>
          <span>flikk</span>
        </div>

        <nav aria-label="Wallet navigation">
          {navItems.map((item) => (
            <button
              className={`nav-item ${activeNav === item.id ? "active" : ""}`}
              key={item.id}
              onClick={() => setActiveNav(item.id)}
              type="button"
            >
              <span>{item.icon}</span>
              {item.label}
            </button>
          ))}
        </nav>

        <div className="rail-footer">
          <div className="security-badge"><span>◈</span> Protected by Courier</div>
          <button className="conversa-switch" type="button" onClick={openConversa}>
            <span>C</span>
            <b>Back to Conversa</b>
            <i>↗</i>
          </button>
          <button className="profile-chip" type="button">
            <span className="avatar">{initials}</span>
            <span><b>{name}</b><small>Personal wallet</small></span>
            <i>⌄</i>
          </button>
        </div>
      </aside>

      <section className="wallet-workspace">
        <header className="topbar">
          <div>
            <p className="eyebrow">Courier wallet</p>
            <h1>{activeNav === "home" ? "Your money, in motion." : navItems.find((item) => item.id === activeNav)?.label}</h1>
          </div>
          <div className="topbar-actions">
            <button className="conversa-topbar" type="button" onClick={openConversa}><span>C</span> Conversa</button>
            <button className="icon-button" onClick={() => void refreshBalance()} type="button" aria-label="Refresh balance">↻</button>
            <button className="icon-button" type="button" aria-label="Notifications">◌<em /></button>
          </div>
        </header>

        {!token && <div className="demo-notice">Demo preview · Sign in through Conversa or set <code>VITE_COURIER_ACCESS_TOKEN</code> to connect your live wallet.</div>}
        {error && <div className="error-notice"><span>!</span>{error}<button type="button" onClick={() => void refreshBalance()}>Try again</button></div>}

        {activeNav === "home" ? (
          <Overview
            balance={balance}
            loading={loading}
            onTopUp={() => setTopUpOpen(true)}
            onTransfer={() => setRoadmapOpen("transfer")}
            onWithdraw={() => setRoadmapOpen("withdraw")}
          />
        ) : (
          <RoadmapPage page={activeNav} />
        )}
      </section>

      {topUpOpen && <TopUpModal token={token} onClose={() => setTopUpOpen(false)} onSuccess={() => void refreshBalance()} />}
      {roadmapOpen && <RoadmapModal feature={roadmapOpen} onClose={() => setRoadmapOpen(null)} />}
    </main>
  );
}

function Overview({ balance, loading, onTopUp, onTransfer, onWithdraw }: {
  balance: WalletBalance | null;
  loading: boolean;
  onTopUp: () => void;
  onTransfer: () => void;
  onWithdraw: () => void;
}) {
  const displayBalance = balance ?? DEMO_BALANCE;
  return <div className="overview-grid">
    <section className="balance-card glass-panel">
      <div className="balance-card-glow" />
      <div className="balance-heading"><span>Available balance</span><span className={`status-dot ${displayBalance.status}`}><i />{displayBalance.status === "not_created" ? "Ready to start" : displayBalance.status}</span></div>
      <div className={`balance-value ${loading ? "loading-value" : ""}`}>{formatVnd(displayBalance.available_minor)}</div>
      <p className="balance-note">{displayBalance.wallet_id ? `Wallet • ${displayBalance.wallet_id}` : "Your VND wallet will be created with your first top-up."}</p>
      <div className="balance-actions">
        <button className="primary-action" type="button" onClick={onTopUp}><span>＋</span> Add money</button>
        <button className="soft-action" type="button" onClick={onTransfer}><span>↗</span> Transfer</button>
        <button className="soft-action" type="button" onClick={onWithdraw}><span>↓</span> Withdraw</button>
      </div>
    </section>

    <section className="insight-card glass-panel">
      <p className="eyebrow">This month</p>
      <h2>Keep every move clear.</h2>
      <div className="insight-row"><span>Added to wallet</span><b>500,000 ₫</b></div>
      <div className="insight-row"><span>Spent with Courier</span><b>167,000 ₫</b></div>
      <div className="mini-chart" aria-label="Wallet activity chart"><i /><i /><i /><i /><i className="tall" /><i /><i className="medium" /></div>
    </section>

    <section className="activity-card glass-panel">
      <div className="section-heading"><div><p className="eyebrow">Latest activity</p><h2>Transaction story</h2></div><button type="button">View all <span>→</span></button></div>
      <div className="activity-list">
        {activity.map((item) => <article className="activity-row" key={item.name}>
          <span className={`activity-icon ${item.tone}`}>{item.icon}</span>
          <span className="activity-copy"><b>{item.name}</b><small>{item.detail}</small></span>
          <strong>{item.value}</strong>
        </article>)}
      </div>
    </section>

    <section className="security-card glass-panel">
      <div className="shield">◈</div>
      <div><p className="eyebrow">Wallet safety</p><h2>Your balance is protected.</h2><p>Every wallet credit is reconciled against a signed payment event and immutable ledger entry.</p></div>
      <button type="button">Security details <span>→</span></button>
    </section>
  </div>;
}

function TopUpModal({ token, onClose, onSuccess }: { token: string; onClose: () => void; onSuccess: () => void }) {
  const [amount, setAmount] = useState(100000);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [checkout, setCheckout] = useState<CheckoutInstruction | null>(null);

  async function createTopUp() {
    if (!token) {
      setError("Sign in with a Courier session before creating a live top-up.");
      return;
    }
    if (!Number.isSafeInteger(amount) || amount < 1000) {
      setError("Enter an amount of at least 1,000 ₫.");
      return;
    }
    setLoading(true);
    setError("");
    try {
      setCheckout(await walletApi.createTopUp(token, amount));
      onSuccess();
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Could not create the top-up.");
    } finally {
      setLoading(false);
    }
  }

  return <div className="modal-backdrop" role="presentation" onMouseDown={onClose}>
    <section className="topup-modal glass-panel" role="dialog" aria-modal="true" aria-labelledby="topup-title" onMouseDown={(event) => event.stopPropagation()}>
      <button className="close-button" type="button" onClick={onClose}>×</button>
      {!checkout ? <>
        <div className="modal-icon">＋</div>
        <p className="eyebrow">Add money</p>
        <h2 id="topup-title">Top up your wallet</h2>
        <p className="modal-copy">Pay securely with a SePay QR transfer. Your balance updates after the bank confirms your payment.</p>
        <label className="amount-field"><span>Amount</span><div><input type="number" min="1000" step="1000" value={amount} onChange={(event) => setAmount(Number(event.target.value))} /><b>₫</b></div></label>
        <div className="preset-row">{[50000, 100000, 200000, 500000].map((value) => <button className={amount === value ? "selected" : ""} type="button" key={value} onClick={() => setAmount(value)}>{formatCompactVnd(value)}</button>)}</div>
        <div className="provider-line"><span className="sepay-mark">S</span><span><b>SePay</b><small>Bank transfer · QR payment</small></span><i>✓</i></div>
        {error && <p className="form-error">{error}</p>}
        <button className="primary-action wide" type="button" disabled={loading} onClick={() => void createTopUp()}>{loading ? "Preparing secure payment…" : `Continue with ${formatVnd(amount)}`}</button>
      </> : <>
        <div className="modal-icon success">✓</div>
        <p className="eyebrow">Payment request ready</p>
        <h2 id="topup-title">One more secure step.</h2>
        <p className="modal-copy">SePay will show a QR code and transfer details for <b>{formatVnd(amount)}</b>. Use the payment code exactly as shown.</p>
        <div className="checkout-code"><span>Payment code</span><b>{checkout.payment_code}</b><small>Expires {new Date(checkout.expires_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}</small></div>
        <button className="primary-action wide" type="button" onClick={() => submitCheckout(checkout)}>Open secure QR payment <span>→</span></button>
        <button className="text-action" type="button" onClick={onClose}>I’ll pay later</button>
      </>}
    </section>
  </div>;
}

function RoadmapModal({ feature, onClose }: { feature: "transfer" | "withdraw"; onClose: () => void }) {
  const label = feature === "transfer" ? "Money transfer" : "Withdraw money";
  return <div className="modal-backdrop" role="presentation" onMouseDown={onClose}>
    <section className="roadmap-modal glass-panel" role="dialog" aria-modal="true" onMouseDown={(event) => event.stopPropagation()}>
      <button className="close-button" type="button" onClick={onClose}>×</button>
      <div className="roadmap-orbit">{feature === "transfer" ? "↗" : "↓"}</div>
      <p className="eyebrow">On the roadmap</p><h2>{label} is being designed.</h2>
      <p>Flikk keeps this action visible so the wallet model feels complete, but it will only unlock once Courier has the required ledger, risk and payout APIs.</p>
      <button className="primary-action wide" type="button" onClick={onClose}>Got it</button>
    </section>
  </div>;
}

function RoadmapPage({ page }: { page: NavigationItem }) {
  const labels: Record<Exclude<NavigationItem, "home">, string> = {
    activity: "Transaction history will appear here once payment-gateway exposes paginated ledger-backed history.",
    cards: "Saved bank accounts and payment methods will appear here after account-linking APIs are available.",
    settings: "Wallet limits, security controls and notification preferences will live here."
  };
  return <section className="empty-page glass-panel"><div className="roadmap-orbit">✦</div><p className="eyebrow">Flikk template</p><h2>Designed for the next money flow.</h2><p>{labels[page as Exclude<NavigationItem, "home">]}</p></section>;
}

function formatVnd(value: number) {
  return `${new Intl.NumberFormat("vi-VN").format(value)} ₫`;
}

function formatCompactVnd(value: number) {
  return value >= 1000000 ? `${value / 1000000}m` : `${value / 1000}k`;
}
