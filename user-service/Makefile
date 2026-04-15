migrate-file:
	migrate create -ext sql -dir migrations ${filename}

swagger:
	swag init --parseDependency --parseInternal

install:
	sh install.sh

wire:
	wire ./internal

compose:
	docker compose -f docker-compose.local.yaml up --build -d

compose-dev:
	docker compose -f docker-compose.dev.yaml up -d

test:
	go test -v ./...

build-image:
	docker build . -t courier/user-service

proto-gen:
	protoc --go_out=pkg/grpc --go_opt=paths=source_relative \
        --go-grpc_out=pkg/grpc --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false \
        -I pkg/grpc/proto \
        pkg/grpc/proto/*.proto

run-local:
	go run -gcflags="-m" main.go start --config=./config/local/config.yaml

run-user:
	go run -gcflags="-m" main.go start --config=./config/local/config.yaml 2>&1 | tee -a ../logs/user-service-$$(date +%Y-%m-%d).log



# Helper target to pull environment variables from GitHub and write to .env.local (variables only; secrets must be filled manually).
ENV_NAME ?= local
# Repo (owner/repo). Example: vucongthanh92/courier
REPO ?= https://github.com/vucongthanh92/courier
# List of GitHub variables to pull (add more as needed)
GH_VARS := OAUTH_GITHUB_CLIENT_ID OAUTH_GITHUB_CLIENT_SECRET OAUTH_GOOGLE_CLIENT_ID OAUTH_GOOGLE_CLIENT_SECRET

get-env:
	@test -n "$(REPO)" || (echo "REPO is required (owner/repo)"; exit 1)
	@echo "# generated from GitHub env=$(ENV_NAME) at $$(date)" > .env.local
	@for v in $(GH_VARS); do \
		val=$$(gh variable get $$v --env $(ENV_NAME) --repo $(REPO) 2>/dev/null); \
		if [ -z "$$val" ]; then \
			echo "# $$v (not found or empty)" >> .env.local; \
		else \
			echo "$$v=$$val" >> .env.local; \
		fi; \
	done
	@echo "Wrote .env.local (variables only; secrets must be filled manually)."

# Usage: make get-env && source .env.local
# make get-env && source .env.local

load-env: .env.local
	@set -a; source .env.local; set +a
