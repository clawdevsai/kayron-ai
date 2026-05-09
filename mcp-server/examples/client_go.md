# Go Client Setup for MT5 gRPC Service

This guide walks through setting up a Go client to connect to the MT5 gRPC service.

## Prerequisites

- Go 1.16+ installed
- `protoc` compiler installed
- `protoc-gen-go` and `protoc-gen-go-grpc` installed

## Installation

### 1. Install Protocol Buffer Compiler

```bash
# macOS
brew install protobuf

# Linux (Ubuntu/Debian)
sudo apt-get install protobuf-compiler

# Windows (Chocolatey)
choco install protoc
```

### 2. Install Go Protocol Buffer Plugins

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 3. Generate Go Code from Proto Files

```bash
cd mcp-server

protoc --go_out=. --go-grpc_out=. \
  proto/mt5_messages.proto \
  proto/mt5_service.proto
```

### 4. Create Go Client

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "github.com/kayron-ai/mt5-grpc/pb"
)

func main() {
    // Connect to server
    conn, err := grpc.DialContext(
        context.Background(),
        "localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewMT5ServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Get account info
    resp, err := client.GetAccountInfo(ctx, &pb.GetAccountInfoRequest{})
    if err != nil {
        log.Fatalf("failed to get account info: %v", err)
    }

    fmt.Printf("Account Info: %v\n", resp)

    // Health check
    health, err := client.CheckHealth(ctx, &pb.CheckHealthRequest{})
    if err != nil {
        log.Fatalf("health check failed: %v", err)
    }

    fmt.Printf("Health: %v\n", health)
}
```

### 5. Build and Run

```bash
go mod init mt5-client
go mod tidy
go run main.go
```

## Example: Place Order

```go
orderResp, err := client.ExecuteOrderOperation(ctx, &pb.ExecuteOrderRequest{
    Symbol: "EURUSD",
    Type: "BUY",
    Volume: 1.0,
    Price: 1.0950,
})
if err != nil {
    log.Fatalf("order failed: %v", err)
}

fmt.Printf("Order: %v\n", orderResp)
```

## Error Handling

gRPC errors are returned with status codes. Use `status.Code()` to check:

```go
resp, err := client.ExecuteOrderOperation(ctx, req)
if err != nil {
    st := status.Convert(err)
    fmt.Printf("Error: %v (Code: %v)\n", st.Message(), st.Code())
}
```

## References

- [gRPC Go Documentation](https://grpc.io/docs/languages/go/)
- [Protocol Buffers Go Guide](https://developers.google.com/protocol-buffers/docs/gotutorial)
- [MT5 gRPC Service](../README.md)
