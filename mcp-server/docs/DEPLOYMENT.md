# MT5 gRPC Service Deployment Guide

This guide covers deploying the MT5 gRPC service via Docker and Kubernetes.

## Prerequisites

- MT5 terminal running on host machine or accessible via network
- MT5 API key and credentials
- Docker installed (for Docker deployment)
- kubectl and Kubernetes cluster (for K8s deployment)

## Docker Deployment

### 1. Build Docker Image

```bash
docker build -t mt5-grpc-server:latest .
```

### 2. Create .env File

```bash
cat > .env << EOF
MT5_LOGIN=123456789
MT5_PASSWORD=your_password
MT5_SERVER=MetaQuotes-Demo
GRPC_PORT=50051
LOG_LEVEL=INFO
EOF
```

### 3. Run Container

```bash
docker run -d \
  --name mt5-grpc-server \
  --env-file .env \
  -p 50051:50051 \
  -v mt5-data:/data \
  mt5-grpc-server:latest
```

### 4. Verify Deployment

```bash
# Check logs
docker logs mt5-grpc-server

# Test connection
python -c "
import grpc
channel = grpc.aio.insecure_channel('localhost:50051')
print('Connected to MT5 gRPC server')
"
```

### 5. Stop Container

```bash
docker stop mt5-grpc-server
docker rm mt5-grpc-server
```

## Kubernetes Deployment

### 1. Update Credentials

Edit `k8s/service.yaml` and update:

```yaml
stringData:
  login: "YOUR_MT5_LOGIN"
  password: "YOUR_MT5_PASSWORD"
```

### 2. Create Namespace (Optional)

```bash
kubectl create namespace mt5-grpc
```

### 3. Deploy Service

```bash
# Apply all manifests
kubectl apply -f k8s/

# Or apply individually
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/deployment.yaml
```

### 4. Verify Deployment

```bash
# Check deployment
kubectl get deployment mt5-grpc-server

# Check pods
kubectl get pods -l app=mt5-grpc-server

# Check logs
kubectl logs -l app=mt5-grpc-server

# Check service
kubectl get svc mt5-grpc-server
```

### 5. Port Forwarding

```bash
# Forward local port to service
kubectl port-forward svc/mt5-grpc-server 50051:50051

# In another terminal, test connection
python examples/client_python.py
```

### 6. Update Configuration

```bash
# Update ConfigMap
kubectl edit configmap mt5-config

# Update Secrets
kubectl edit secret mt5-credentials

# Restart pods to apply changes
kubectl rollout restart deployment/mt5-grpc-server
```

### 7. Scale (Limited by Single MT5 Connection)

```bash
# Note: Due to single MT5 connection constraint, replicas should be 1
# If needed for HA, must use external connection pooling

kubectl scale deployment mt5-grpc-server --replicas=1
```

### 8. Delete Deployment

```bash
kubectl delete -f k8s/
```

## Monitoring

### Health Check Endpoint

```bash
curl http://localhost:50052/health
```

Expected response:
```json
{
  "status": "SERVING",
  "checks": {
    "mt5_connection": {"status": "HEALTHY"},
    "database": {"status": "HEALTHY"},
    "sessions": {"status": "HEALTHY"}
  }
}
```

### Prometheus Metrics

```bash
# Scrape metrics
curl http://localhost:50051/metrics
```

### Kubernetes Health Checks

- Liveness probe: `/health` on port 50052
- Readiness probe: `/ready` on port 50052

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker logs mt5-grpc-server

# Common issues:
# - MT5 credentials invalid
# - MT5 terminal not running
# - Port already in use
```

### gRPC Connection Issues

```bash
# Test connection
grpcurl -plaintext localhost:50051 list

# If fails, check:
# - Port 50051 is open
# - Firewall rules
# - Network connectivity
```

### Database Issues

```bash
# Check database
sqlite3 mcp-server.db ".tables"

# Check logs for errors
docker logs mt5-grpc-server | grep -i error
```

### Performance Issues

```bash
# Monitor resource usage
docker stats mt5-grpc-server

# Check operation queue depth
sqlite3 mcp-server.db "SELECT COUNT(*) FROM queued_operations WHERE status='QUEUED';"

# Check for slow operations
sqlite3 mcp-server.db "SELECT * FROM operation_logs WHERE latency_ms > 1000;"
```

## Production Checklist

- [ ] MT5 credentials securely stored
- [ ] Database volume persistent and backed up
- [ ] Logging configured and monitored
- [ ] Health checks passing
- [ ] Performance baseline established
- [ ] Error handling validated
- [ ] Security audits completed
- [ ] Runbooks documented
- [ ] Alerting rules configured
- [ ] Disaster recovery plan in place

## References

- [MT5 gRPC Service README](../README.md)
- [Quickstart Guide](../quickstart.md)
- [Configuration Guide](../docs/CONFIG.md)
- [Troubleshooting Guide](../docs/TROUBLESHOOTING.md)
