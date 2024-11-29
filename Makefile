.PHONY: vendor build test run-server run-client

vendor:
	go mod tidy && go mod vendor

build:
	go build -o build/http ./app/cmd/

test:
	go test -v ./...

run-server:
	PUBLIC_NODE_URL=https://ethereum-rpc.publicnode.com/ PORT=8081 START_BLOCK=21292394 JOB_SCHEDULE=1s go run app/cmd/main.go

run-build: build
	PUBLIC_NODE_URL=https://ethereum-rpc.publicnode.com/ PORT=8081 START_BLOCK=21292394 JOB_SCHEDULE=1s ./build/http

run-client:
	PARSER_URL=http://localhost:8081 fetchFrequency=10s SUBSCRIBE_ADDRESSES="0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5,0x6eaE5e2d47f1CbB5979734812521579921d37C9A" go run client/cmd/main.go

# doesn't work because run-server doesn't exit
# run: run-server run-client