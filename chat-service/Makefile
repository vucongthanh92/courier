migrate-file:
	migrate create -ext sql -dir migrations ${filename}

install:
	sh install.sh

wire:
	wire ./internal

swagger:
	swag init --parseDependency --parseInternal

test:
	go test -v ./...

run-local:
	go run main.go start --config=./config/local/config.yaml
