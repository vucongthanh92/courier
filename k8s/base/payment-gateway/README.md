# payment-gateway Kubernetes configuration

Create `payment-gateway-config` manually with a complete `config.yaml` key.
The config may retain its existing `${PAYMENT_GATEWAY_*}` placeholders, but the
corresponding values must be injected only from an approved secret source. Do
not commit SePay, database, or Redis credentials.
