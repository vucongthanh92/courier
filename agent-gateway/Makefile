run-local:
	go run main.go --config=./config/local/config.yaml

test:
	go test -v ./...

vendor:
	go mod tidy
	go mod vendor

test-vendor:
	go test -mod=vendor -v ./...

qdrant-up:
	docker compose up -d qdrant

qdrant-down:
	docker compose down

qdrant-logs:
	docker compose logs -f qdrant
