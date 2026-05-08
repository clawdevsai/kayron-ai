.PHONY: help test build lint clean proto

help:
	@echo "MT5 MCP Server Build Targets"
	@echo "  make test        - Run all tests"
	@echo "  make build       - Build the binary"
	@echo "  make lint        - Run linters"
	@echo "  make proto       - Generate gRPC stubs from proto"
	@echo "  make clean       - Clean build artifacts"

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

build:
	@echo "Building MT5 MCP Server..."
	mkdir -p ./bin
	go build -o ./bin/mcp-mt5-server ./cmd/mcp-mt5-server

lint:
	@echo "Running go vet..."
	go vet ./...
	@echo "Running golangci-lint..."
	golangci-lint run ./... || true

proto:
	@echo "Generating gRPC stubs..."
	protoc --go_out=. --go-grpc_out=. ./api/mt5.proto

clean:
	@echo "Cleaning..."
	rm -f ./bin/mcp-mt5-server
	rm -f ./mt5_queue.db
	rm -f coverage.txt
	rm -f api/*.pb.go
