build:
	go build -o ./bin/blockchainX

run: build
	./bin/blockchainX

test:
	go test ./...