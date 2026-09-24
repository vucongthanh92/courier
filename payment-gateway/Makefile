NGROK_DOMAIN ?= candied-corny-blatantly.ngrok-free.dev
MIGRATIONS_DIR ?= ../shared/migrations

migrate-file:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) ${filename}

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "${DATABASE_URL}" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "${DATABASE_URL}" down 1

test:
	go test -v ./...

wire:
	wire ./internal

run-local:
	@set -a; [ ! -f .env.local ] || . ./.env.local; set +a; go run main.go --config=./config/local/config.yaml

run-dev:
	@set -a; [ ! -f .env.local ] || . ./.env.local; set +a; go run main.go --config=./config/dev/config.yaml

# Helper target to pull environment variables from GitHub and write to .env.local.
# GitHub Actions secrets cannot be read back by the CLI; store values needed here as
# environment variables, or add them manually to the ignored .env.local file.
ENV_NAME ?= local
# Repo (owner/repo). Example: vucongthanh92/courier
REPO ?= https://github.com/vucongthanh92/courier
GH_VARS := PAYMENT_GATEWAY_SEPAY_MERCHANT_ID PAYMENT_GATEWAY_SEPAY_SECRET_KEY PAYMENT_GATEWAY_SEPAY_WEBHOOK_SECRET PAYMENT_GATEWAY_SEPAY_CHECKOUT_URL PAYMENT_GATEWAY_SEPAY_SUCCESS_URL PAYMENT_GATEWAY_SEPAY_ERROR_URL PAYMENT_GATEWAY_SEPAY_CANCEL_URL

get-env:
	@test -n "$(REPO)" || (echo "REPO is required (owner/repo)"; exit 1)
	@echo "# generated from GitHub env=$(ENV_NAME) at $$(date)" > .env.local
	@for v in $(GH_VARS); do \
		val=$$(gh variable get $$v --env $(ENV_NAME) --repo $(REPO) 2>/dev/null); \
		if [ -z "$$val" ]; then \
			echo "# $$v (not found, empty, or stored as a GitHub secret)" >> .env.local; \
		else \
			echo "$$v=$$val" >> .env.local; \
		fi; \
	done
	@echo "Wrote .env.local. It is ignored by Git; fill any commented values manually."

# Exposes the local HTTP server through the reserved SePay Sandbox webhook domain.
ngrok:
	ngrok http --url=$(NGROK_DOMAIN) 5003
