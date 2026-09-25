# SePay Sandbox Top-up E2E Test

This runbook tests the complete currently implemented path:

```text
JWT user -> create top-up -> SePay checkout / simulated bank transfer
-> signed bank webhook -> ledger + wallet balance + outbox
```

## 1. Start local dependencies

Run the shared migrations once:

```bash
cd payment-gateway
DATABASE_URL='postgresql://dev:dev@127.0.0.1:5432/dev_db?sslmode=disable' make migrate-up
```

Export SePay Sandbox credentials. Do not put these values in Git:

```bash
export PAYMENT_GATEWAY_SEPAY_MERCHANT_ID='<sandbox-merchant-id>'
export PAYMENT_GATEWAY_SEPAY_SECRET_KEY='<sandbox-checkout-secret>'
export PAYMENT_GATEWAY_SEPAY_WEBHOOK_SECRET='<sandbox-webhook-hmac-secret>'
```

Start the API and the reserved-domain tunnel in separate terminals:

```bash
cd payment-gateway
make run-local
```

```bash
cd payment-gateway
make ngrok
```

The SePay Sandbox webhook must be configured as:

```text
POST https://candied-corny-blatantly.ngrok-free.dev/api/v1/webhooks/sepay
Content-Type: application/json
Authentication: HMAC-SHA256
```

Set the payment-code prefix in SePay Sandbox to `CRTOP_`, and configure the
same receiving test account number in `sepay.receivingAccountNumbers`.

## 2. Obtain a Courier JWT

Log in through `user-service` with a verified test user. Copy its access token
to a shell variable:

```bash
export COURIER_JWT='<access-token-from-user-service>'
```

`POST /api/v1/wallet/top-ups` reads the user from JWT claim `sub`; it no longer
accepts `X-User-ID`.

## 3. Create a top-up intent

Choose a small Sandbox amount, for example `1000` VND. Use a new UUID every new
business request; reuse exactly the same UUID only to test idempotency.

```bash
curl -sS -X POST http://localhost:5003/api/v1/wallet/top-ups \
  -H "Authorization: Bearer ${COURIER_JWT}" \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: 11111111-2222-4333-8444-555555555555' \
  -d '{"amount_minor":1000,"method":"bank_transfer","provider_name":"sepay"}'
```

Expected response: `201 Created` with a payload like:

```json
{
  "success": true,
  "data": {
    "topup_id": "...",
    "invoice_number": "CRTOP_...",
    "expires_at": "...",
    "checkout_action": "https://pay-sandbox.sepay.vn/v1/checkout/init",
    "checkout_fields": { "order_amount": "1000", "signature": "..." }
  }
}
```

Record `invoice_number`. It is the required SePay payment code for this test.

Retry the identical request with the same idempotency key. Expected result is
`200 OK` and the exact same `topup_id`/`invoice_number`. A changed request body
with that key must return `409 Conflict`.

## 4. Make the Sandbox payment

Use either path below.

### Hosted checkout

POST the returned `checkout_fields` to `checkout_action` from a browser or your
API client, then complete the Sandbox payment page.

### Bank-webhook simulation

In SePay Test mode, simulate an incoming transfer to the configured test account:

- Type: `in`
- Amount: exactly `1000`
- Content: include the recorded `CRTOP_...` code
- Webhook: the HMAC-SHA256 webhook configured in step 1

SePay signs and delivers the webhook automatically. Do not manually generate a
different JSON body after signing: signature verification uses raw body bytes.

## 5. Expected webhook outcome

The provider calls the public endpoint with:

```text
POST /api/v1/webhooks/sepay
X-SePay-Signature: sha256=<hmac>
X-SePay-Timestamp: <unix-seconds>
```

Expected HTTP response:

```json
{"success":true}
```

The payment-gateway verifies signature/timestamp, then validates `transferType`,
receiving account, payment code, pending intent, expiry and exact amount. A
successful first webhook writes one provider event, one provider transaction,
one balanced journal, two ledger entries, one wallet balance update and one
outbox event. A replay of the same SePay transaction `id` is acknowledged but
does not credit the wallet again.

## 6. Verify wallet balance

Read the authenticated user's balance after the successful webhook:

```bash
curl -sS http://localhost:5003/api/v1/wallet/balance \
  -H 'Authorization: Bearer <courier-jwt>'
```

For a credited wallet, `available_minor` increases by the top-up amount. A user
without a wallet receives `200` with `status: "not_created"` and a zero VND
balance; this read endpoint never creates a wallet.

For detailed reconciliation, inspect PostgreSQL:

```sql
SELECT id, status, amount_minor, payment_code, succeeded_at
FROM "payment-gateway".topup_intents
WHERE provider_invoice_number = '<CRTOP invoice number>';

SELECT available_minor, pending_minor, held_minor, version
FROM "payment-gateway".wallet_balances
WHERE wallet_id = <wallet_id from topup_intents>;

SELECT source_type, source_provider, reference_type, reference_id, status
FROM "payment-gateway".ledger_journals
WHERE reference_id = '<SePay transaction id>';

SELECT side, amount_minor, currency
FROM "payment-gateway".ledger_entries
WHERE journal_id = <journal_id>;
```

Success criteria:

- `topup_intents.status = succeeded`.
- `available_minor` increased by `1000` exactly once.
- Journal has `source_type = external_provider` and `source_provider = sepay`.
- Ledger contains equal debit and credit entries of `1000` VND.
- Audit logs include `wallet.topup.created` and `wallet.topup.succeeded` for a
  successful flow, or `wallet.topup.ignored` for an acknowledged invalid event.

## Negative checks

Test each case after the happy path:

| Case | Expected result |
| --- | --- |
| Reuse SePay transaction `id` | HTTP 200; no second credit |
| Wrong amount | no credit; provider event is ignored |
| Unknown `CRTOP_` code | no credit |
| Expired intent | no credit |
| Invalid HMAC / stale timestamp | HTTP 401 and no DB processing |
| Wrong receiving account / `transferType = out` | HTTP 200 acknowledgement; no credit |
