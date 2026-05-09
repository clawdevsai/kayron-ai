# Node.js Client Setup for MT5 gRPC Service

This guide walks through setting up a Node.js client to connect to the MT5 gRPC service.

## Prerequisites

- Node.js 14+ installed
- npm or yarn package manager
- `protoc` compiler installed

## Installation

### 1. Initialize Node.js Project

```bash
mkdir mt5-client-node
cd mt5-client-node
npm init -y
```

### 2. Install Dependencies

```bash
npm install @grpc/grpc-js @grpc/proto-loader
npm install --save-dev grpc-tools
```

### 3. Generate Node.js Code from Proto Files

```bash
npx grpc_tools_node_protoc \
  --js_out=import_style=commonjs,binary:. \
  --grpc_out=grpc_js:. \
  --plugin=protoc-gen-grpc=`which grpc_tools_node_protoc_plugin` \
  proto/mt5_messages.proto \
  proto/mt5_service.proto
```

### 4. Create Node.js Client

```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const path = require('path');

const PROTO_PATH = path.join(__dirname, 'proto/mt5_service.proto');

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

const mt5Proto = grpc.loadPackageDefinition(packageDefinition);
const mt5 = mt5Proto.mt5;

// Connect to server
const client = new mt5.MT5Service(
  'localhost:50051',
  grpc.credentials.createInsecure()
);

// Get account info
client.getAccountInfo({}, (err, response) => {
  if (err) {
    console.error('Error:', err);
  } else {
    console.log('Account Info:', response);
  }
});

// Health check
client.checkHealth({}, (err, response) => {
  if (err) {
    console.error('Error:', err);
  } else {
    console.log('Health:', response);
  }
});
```

### 5. Run Client

```bash
node client.js
```

## Example: Place Order

```javascript
const request = {
  symbol: 'EURUSD',
  type: 'BUY',
  volume: 1.0,
  price: 1.0950,
};

client.executeOrderOperation(request, (err, response) => {
  if (err) {
    console.error('Order error:', err);
  } else {
    console.log('Order response:', response);
  }
});
```

## Streaming Callbacks (Async)

For bidirectional streaming, use async/await pattern:

```javascript
const executeOrderAsync = async () => {
  return new Promise((resolve, reject) => {
    const stream = client.executeOrderOperation(request);

    stream.on('data', (response) => {
      console.log('Callback update:', response);
    });

    stream.on('error', (err) => {
      reject(err);
    });

    stream.on('end', () => {
      resolve();
    });
  });
};

executeOrderAsync().catch(console.error);
```

## Error Handling

gRPC errors include status code and details:

```javascript
client.executeOrderOperation(request, (err, response) => {
  if (err) {
    console.error('Status Code:', err.code);
    console.error('Message:', err.message);
    console.error('Details:', err.details);
  } else {
    console.log('Success:', response);
  }
});
```

## References

- [gRPC Node.js Documentation](https://grpc.io/docs/languages/node/)
- [Protocol Buffers JavaScript Guide](https://developers.google.com/protocol-buffers/docs/reference/javascript-generated)
- [MT5 gRPC Service](../README.md)
