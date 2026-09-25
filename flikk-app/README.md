# flikk

`flikk-app` is the Courier wallet client. It follows Conversa's spatial-glass
visual language while focusing on money, balance and transaction clarity.

## Run locally

```bash
cd flikk-app
cp .env.example .env.local
npm install
npm run dev
```

The development server runs on `http://localhost:8081`.

## Current API integration

- `GET /api/v1/wallet/balance` renders the signed-in user's VND balance.
- `POST /api/v1/wallet/top-up` creates a SePay checkout instruction. Flikk then
  submits the returned form to SePay so the user can continue to QR payment.

The client can be opened directly, and has a two-way app switcher with Conversa.
`VITE_CONVERSA_APP_URL` and Conversa's `VITE_FLIKK_APP_URL` define the allowed
destinations. A `return_to` URL is accepted only when it has Conversa's configured
origin, preventing open redirects.

For standalone local testing, set `VITE_COURIER_ACCESS_TOKEN`. Browsers isolate
local storage by origin, so `localhost:8080` and `localhost:8081` cannot share a
Conversa session directly. Production seamless SSO should use a short-lived,
single-use handoff code minted by user-service; never put a JWT in the URL.

Money transfer, withdrawal and transaction history are intentionally presented as
roadmap templates. Their buttons do not submit money because the corresponding
payment-gateway APIs have not been implemented yet.
