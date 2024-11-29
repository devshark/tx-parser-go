.PHONY: vendor build

vendor:
	go mod tidy && go mod vendor

build:
	go build -o build/http ./cmd/http/

test:
	go test -v ./...