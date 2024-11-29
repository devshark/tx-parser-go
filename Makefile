.PHONY: vendor build test run-server run-client run

vendor:
	go mod tidy && go mod vendor

build:
	go build -o build/http ./cmd/http/

test:
	go test -v ./...

run-server:
	go run app/cmd/main.go

run-client:
	go run client/cmd/main.go

# doesn't work because run-server doesn't exit
run: run-server run-client