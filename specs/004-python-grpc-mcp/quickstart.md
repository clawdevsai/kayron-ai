# Quickstart: Python + gRPC MT5 MCP

Complete setup guide and example clients for MT5 gRPC service.

---

## 1. Prerequisites

- **MT5 Terminal**: Installed and running on the server machine
- **Python 3.8+**: For server and example client
- **Go 1.19+** (optional): For Go client example
- **Node.js 18+** (optional): For Node.js client example

---

## 2. Server Setup

### Install Dependencies

```bash
cd mcp-server
pip install -r requirements.txt
```

**requirements.txt** contains:
```
grpcio==1.50.0
grpcio-tools==1.50.0
metatrader5==5.0.45
asyncio==3.4.3
pydantic==1.10.0
sqlalchemy==2.0.0
```

### Generate Python Code from Proto

```bash
python -m grpc_tools.protoc \
  -I./proto \
  --python_out=./src \
  --grpc_python_out=./src \
  proto/mt5_messages.proto proto/mt5_service.proto
```

### Configure Server

Create `config.yaml`:
```yaml
server:
  host: "0.0.0.0"
  port: 50051
  max_concurrent_streams: 100

mt5:
  terminal_path: "C:\\Program Files\\MetaTrader 5\\terminal64.exe"  # Windows
  login: 12345678                 # Demo account
  password: "demo_password"       # Demo password
  server: "MetaQuotes-Demo"       # Broker server name

auth:
  api_keys:
    - key: "agent-001-key"
      agent_name: "agent-001"
    - key: "agent-002-key"
      agent_name: "agent-002"

queue:
  db_path: "./data/operations.db"
  max_retries: 10
  retry_base_delay: 1             # seconds
  operation_timeout: 300          # seconds

logging:
  level: "INFO"
  format: "json"
```

### Start Server

```bash
python src/server.py --config config.yaml
```

Server will:
1. Initialize MT5 connection pool
2. Load API keys from config
3. Start gRPC service on port 50051
4. Begin listening for agent connections

**Output**:
```
[INFO] MT5Service running on 0.0.0.0:50051
[INFO] MT5 terminal connected (login: 12345678)
[INFO] Operation queue initialized with SQLite backend
[INFO] Ready to accept agent connections
```

---

## 3. Python Client Example

```python
import grpc
import asyncio
from mt5_service_pb2_grpc import MT5ServiceStub
from mt5_service_pb2 import (
    OrderOperationRequest, PlaceOrderRequest, OrderType,
    GetAccountInfoRequest
)
from google.protobuf.empty_pb2 import Empty

# Create secure channel with API key
def get_channel(api_key: str):
    credentials = grpc.ssl_channel_credentials()  # Use TLS in production
    channel = grpc.secure_channel("localhost:50051", credentials)
    return channel

async def place_order_example():
    api_key = "agent-001-key"
    
    with get_channel(api_key) as channel:
        stub = MT5ServiceStub(channel)
        
        # Create order request
        place_order = PlaceOrderRequest(
            symbol="EURUSD",
            type=OrderType.BUY,
            volume=1.0,
            price=1.0850,  # Limit price
            stop_loss=1.0800,
            take_profit=1.0900,
            comment="Test order"
        )
        
        operation = OrderOperationRequest(
            operation_id="op-001",
            place_order=place_order
        )
        
        # Send operation and listen for callbacks
        metadata = [("api-key", api_key)]
        call = stub.ExecuteOrderOperation(
            operation,
            metadata=metadata
        )
        
        async for response in call:
            print(f"Status: {response.status}")
            if response.status == OperationStatus.QUEUED:
                print(f"  → Queued, waiting for MT5 slot")
            elif response.status == OperationStatus.EXECUTING:
                print(f"  → Executing on MT5")
            elif response.status == OperationStatus.COMPLETED:
                print(f"  → Completed: ticket {response.order_result.ticket}")
                break
            elif response.status == OperationStatus.FAILED:
                print(f"  → Failed: {response.error.message}")
                break

async def get_account_info_example():
    api_key = "agent-001-key"
    
    with get_channel(api_key) as channel:
        stub = MT5ServiceStub(channel)
        
        metadata = [("api-key", api_key)]
        call = stub.GetAccountInfo(Empty(), metadata=metadata)
        
        async for response in call:
            if response.status == OperationStatus.COMPLETED:
                account = response.account_info
                print(f"Account: {account.login}")
                print(f"Balance: {account.balance}")
                print(f"Equity: {account.equity}")
                print(f"Margin: {account.margin} / {account.margin_free}")
                break

# Run examples
if __name__ == "__main__":
    asyncio.run(place_order_example())
    asyncio.run(get_account_info_example())
```

---

## 4. Go Client Example

```go
package main

import (
    "context"
    "fmt"
    "io"
    "log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"

    pb "github.com/kayron-ai/mt5-mcp/pb"
)

func main() {
    conn, err := grpc.Dial(
        "localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewMT5ServiceClient(conn)
    apiKey := "agent-001-key"
    ctx := metadata.AppendToOutgoingContext(
        context.Background(),
        "api-key", apiKey,
    )

    // Example: PlaceOrder
    req := &pb.OrderOperationRequest{
        OperationId: "op-go-001",
        Operation: &pb.OrderOperationRequest_PlaceOrder{
            PlaceOrder: &pb.PlaceOrderRequest{
                Symbol:     "EURUSD",
                Type:       pb.OrderType_BUY,
                Volume:     1.0,
                Price:      1.0850,
                StopLoss:   1.0800,
                TakeProfit: 1.0900,
                Comment:    "Go client test",
            },
        },
    }

    stream, err := client.ExecuteOrderOperation(ctx, req)
    if err != nil {
        log.Fatalf("Failed to execute: %v", err)
    }

    for {
        response, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalf("Stream error: %v", err)
        }

        fmt.Printf("Status: %v\n", response.Status)
        if response.Status == pb.OperationStatus_COMPLETED {
            fmt.Printf("Order placed: %d\n", response.GetOrderResult().Ticket)
            break
        }
    }
}
```

---

## 5. Node.js Client Example

```javascript
const grpc = require("@grpc/grpc-js");
const protoLoader = require("@grpc/proto-loader");

const packageDef = protoLoader.loadSync("proto/mt5_service.proto", {
    includeDirs: ["proto"],
});

const proto = grpc.loadPackageDefinition(packageDef);
const MT5Service = proto.mt5.MT5Service;

const client = new MT5Service(
    "localhost:50051",
    grpc.credentials.createInsecure()
);

const apiKey = "agent-001-key";
const metadata = new grpc.Metadata();
metadata.add("api-key", apiKey);

// PlaceOrder example
const request = {
    operationId: "op-js-001",
    placeOrder: {
        symbol: "EURUSD",
        type: "BUY",
        volume: 1.0,
        price: 1.0850,
        stopLoss: 1.0800,
        takeProfit: 1.0900,
        comment: "Node.js client test",
    },
};

const stream = client.executeOrderOperation(request, metadata);

stream.on("data", (response) => {
    console.log(`Status: ${response.status}`);
    if (response.status === "COMPLETED") {
        console.log(`Order placed: ${response.orderResult.ticket}`);
    }
});

stream.on("error", (err) => {
    console.error(`Stream error: ${err.message}`);
});

stream.on("end", () => {
    console.log("Stream ended");
});
```

---

## 6. Testing & Debugging

### Health Check

```bash
grpcurl -plaintext \
  localhost:50051 \
  mt5.MT5Service/CheckHealth
```

**Response**:
```json
{
  "status": "SERVING",
  "message": "MT5 terminal connected (login: 12345678)"
}
```

### List Available Methods

```bash
grpcurl -plaintext \
  localhost:50051 \
  list mt5.MT5Service
```

### Test with Example Request

```bash
grpcurl -plaintext \
  -H "api-key: agent-001-key" \
  localhost:50051 \
  mt5.MT5Service/GetAccountInfo \
  {}
```

### View Server Logs

```bash
tail -f logs/mt5-service.log | jq .
```

---

## 7. Deployment

### Docker

```dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY src/ ./src
COPY proto/ ./proto
COPY config.yaml .

RUN python -m grpc_tools.protoc \
    -I./proto \
    --python_out=./src \
    --grpc_python_out=./src \
    proto/mt5_messages.proto proto/mt5_service.proto

EXPOSE 50051

CMD ["python", "src/server.py", "--config", "config.yaml"]
```

**Build & Run**:
```bash
docker build -t mt5-mcp:latest .
docker run -p 50051:50051 \
  -v /path/to/config.yaml:/app/config.yaml \
  mt5-mcp:latest
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mt5-mcp
spec:
  replicas: 1  # Single replica (single MT5 terminal)
  selector:
    matchLabels:
      app: mt5-mcp
  template:
    metadata:
      labels:
        app: mt5-mcp
    spec:
      containers:
      - name: server
        image: mt5-mcp:latest
        ports:
        - containerPort: 50051
        env:
        - name: CONFIG_PATH
          value: /config/config.yaml
        volumeMounts:
        - name: config
          mountPath: /config
        - name: data
          mountPath: /app/data
        livenessProbe:
          exec:
            command:
            - grpcurl
            - -plaintext
            - localhost:50051
            - mt5.MT5Service/CheckHealth
          initialDelaySeconds: 10
          periodSeconds: 10
      volumes:
      - name: config
        configMap:
          name: mt5-config
      - name: data
        persistentVolumeClaim:
          claimName: mt5-data
---
apiVersion: v1
kind: Service
metadata:
  name: mt5-mcp
spec:
  type: ClusterIP
  ports:
  - port: 50051
    targetPort: 50051
  selector:
    app: mt5-mcp
```

---

## 8. Next Steps

- Read `data-model.md` for entity details
- Review `contracts/` for proto specifications
- Implement server modules (see `plan.md` for Phase 2 tasks)
- Run integration tests against real MT5 terminal
