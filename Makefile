.PHONY: vendor build test run-server run-client

vendor:
	go mod tidy && go mod vendor

build:
	CGO_ENABLED=0 go build  -ldflags \
		"-w -s" \
		-o build/http \
		-tags netgo \
		-a ./app/cmd/

test:
	go test -v ./...

run-server:
	PUBLIC_NODE_URL=https://ethereum-rpc.publicnode.com/ PORT=8080 JOB_SCHEDULE=1s go run app/cmd/main.go

run-build: build
	PUBLIC_NODE_URL=https://ethereum-rpc.publicnode.com/ PORT=8080 JOB_SCHEDULE=1s ./build/http

run-client:
	PARSER_URL=http://localhost:8080 FETCH_FREQUENCY=10s SUBSCRIBE_ADDRESSES="0xdAC17F958D2ee523a2206206994597C13D831ec7,0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599" go run client/cmd/main.go

# doesn't work because run-server doesn't exit
# run: run-server run-client