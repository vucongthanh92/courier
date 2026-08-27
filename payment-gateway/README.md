# Payment Gateway

`payment-gateway` owns Courier wallet balances, the immutable double-entry ledger,
top-up intents and payment-provider integrations. Its PostgreSQL objects live only
in the quoted `"payment-gateway"` schema and refer to users by `user_id`; there is
no cross-service foreign key.

Its database migration is version `017_create_payment_gateway_core` in the shared
repository migration sequence at `../shared/migrations/`.

The initial integration is SePay Sandbox hosted checkout. Start locally with
`PAYMENT_GATEWAY_SEPAY_MERCHANT_ID` and `PAYMENT_GATEWAY_SEPAY_SECRET_KEY` set,
then run `make run-local`. The current top-up endpoint is documented in
[`docs/sepay-topup-flow.md`](docs/sepay-topup-flow.md).

The `X-User-ID` header is development scaffolding only. Before production, it is
replaced by user-service/JWT authorization and the signed SePay IPN handler will
be responsible for crediting the wallet ledger.
