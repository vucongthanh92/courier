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

run-local:
	go run main.go --config=./config/local/config.yaml

run-dev:
	go run main.go --config=./config/dev/config.yaml

# Exposes the local HTTP server through the reserved SePay Sandbox webhook domain.
ngrok:
	ngrok http --url=$(NGROK_DOMAIN) 5003
