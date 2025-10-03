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
	docker build . -t go-base-project

proto-gen:
	protoc --go_out=pkg/grpc --go_opt=paths=source_relative \
        --go-grpc_out=pkg/grpc --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false \
        -I pkg/grpc/proto \
        pkg/grpc/proto/*.proto

run-esc:
	go run -gcflags="-m" main.go start --config=./config/local/config.yaml