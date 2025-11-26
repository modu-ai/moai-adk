# MCP Server Deployment

Deployment strategies for MCP servers in production environments.

---

## Local Development

### Running Server Locally

```bash
# Basic startup
python server.py

# With environment variables
API_KEY=your-key LOG_LEVEL=debug python server.py

# Using uv (recommended)
uv run python server.py
```

### Environment Configuration

**File: `.env`**:

```bash
# Server configuration
SERVER_NAME=my-mcp-server
SERVER_VERSION=1.0.0
LOG_LEVEL=info

# Authentication
API_KEY=your-api-key-here
OAUTH_CLIENT_ID=your-client-id
OAUTH_CLIENT_SECRET=your-client-secret

# Database
DATABASE_URL=sqlite:///data.db

# External services
EXTERNAL_API_URL=https://api.example.com
EXTERNAL_API_KEY=your-external-key
```

**Load environment**:

```python
import os
from dotenv import load_dotenv

load_dotenv()
api_key = os.getenv("API_KEY")
db_url = os.getenv("DATABASE_URL", "sqlite:///data.db")
```

---

## Docker Deployment

### Dockerfile

**File: `Dockerfile`**:

```dockerfile
# Build stage
FROM python:3.11-slim AS builder

WORKDIR /build
COPY requirements.txt .
RUN pip install --user --no-cache-dir -r requirements.txt

# Runtime stage
FROM python:3.11-slim

WORKDIR /app

# Copy Python dependencies from builder
COPY --from=builder /root/.local /root/.local

# Copy application
COPY server.py .
COPY . .

# Set environment
ENV PATH=/root/.local/bin:$PATH
ENV PYTHONUNBUFFERED=1
ENV LOG_LEVEL=info

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8000/health')" || exit 1

EXPOSE 8000

# Run server
CMD ["python", "server.py"]
```

### Build and Run

```bash
# Build image
docker build -t mcp-server:latest .

# Run container
docker run -p 8000:8000 \
  -e API_KEY=your-key \
  -e LOG_LEVEL=debug \
  mcp-server:latest

# Run with mounted volumes
docker run -p 8000:8000 \
  -v $(pwd)/data:/app/data \
  -e DATABASE_URL=sqlite:////app/data/app.db \
  mcp-server:latest

# Run with .env file
docker run -p 8000:8000 \
  --env-file .env \
  mcp-server:latest
```

### Docker Compose

**File: `docker-compose.yml`**:

```yaml
version: "3.8"

services:
  mcp-server:
    build: .
    ports:
      - "8000:8000"
    environment:
      API_KEY: ${API_KEY}
      LOG_LEVEL: info
      DATABASE_URL: postgresql://user:password@db:5432/mcp
    depends_on:
      - db
    volumes:
      - ./data:/app/data
    healthcheck:
      test: ["CMD", "python", "-c", "import requests; requests.get('http://localhost:8000/health')"]
      interval: 30s
      timeout: 10s
      retries: 3

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: mcp
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

**Run with Compose**:

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f mcp-server

# Stop services
docker-compose down
```

---

## Kubernetes Deployment

### Basic Deployment

**File: `kubernetes/deployment.yaml`**:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-server
  namespace: default
  labels:
    app: mcp-server
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  selector:
    matchLabels:
      app: mcp-server
  template:
    metadata:
      labels:
        app: mcp-server
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8000"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: mcp-server
      containers:
      - name: mcp-server
        image: mcp-server:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8000
          name: http
          protocol: TCP
        env:
        - name: LOG_LEVEL
          value: "info"
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: mcp-secrets
              key: api-key
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: mcp-secrets
              key: database-url
        resources:
          requests:
            cpu: "250m"
            memory: "256Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
              - ALL
        volumeMounts:
        - name: tmp
          mountPath: /tmp
        - name: cache
          mountPath: /app/cache
      volumes:
      - name: tmp
        emptyDir: {}
      - name: cache
        emptyDir: {}
```

### Service Exposure

**File: `kubernetes/service.yaml`**:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: mcp-server
  namespace: default
  labels:
    app: mcp-server
spec:
  type: LoadBalancer
  selector:
    app: mcp-server
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8000
    name: http
  sessionAffinity: None
```

### ConfigMap and Secrets

**File: `kubernetes/secrets.yaml`**:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mcp-secrets
  namespace: default
type: Opaque
stringData:
  api-key: your-api-key-here
  database-url: postgresql://user:password@postgres:5432/mcp
```

### Deploy to Kubernetes

```bash
# Create namespace
kubectl create namespace mcp-apps

# Create secrets
kubectl apply -f kubernetes/secrets.yaml -n mcp-apps

# Deploy application
kubectl apply -f kubernetes/deployment.yaml -n mcp-apps
kubectl apply -f kubernetes/service.yaml -n mcp-apps

# View deployment
kubectl get deployments -n mcp-apps
kubectl get pods -n mcp-apps

# View logs
kubectl logs -f deployment/mcp-server -n mcp-apps

# Scale deployment
kubectl scale deployment/mcp-server --replicas=5 -n mcp-apps

# Update image
kubectl set image deployment/mcp-server \
  mcp-server=mcp-server:v1.1.0 -n mcp-apps
```

---

## Health Checks

### Liveness Probe

Checks if server is running:

```python
@server.resource("health://live")
def liveness_probe() -> dict:
    """Kubernetes liveness probe."""
    return {"status": "alive", "timestamp": datetime.now().isoformat()}
```

### Readiness Probe

Checks if server can handle requests:

```python
@server.resource("health://ready")
def readiness_probe() -> dict:
    """Kubernetes readiness probe."""
    # Check dependencies
    db_ready = check_database_connection()
    external_api_ready = check_external_api()

    if db_ready and external_api_ready:
        return {"status": "ready"}
    else:
        raise ValueError("Service dependencies not ready")
```

---

## Monitoring & Observability

### Prometheus Metrics

```python
from prometheus_client import Counter, Histogram, Gauge
import time

# Define metrics
request_count = Counter(
    'mcp_requests_total',
    'Total requests',
    ['tool_name', 'status']
)

request_duration = Histogram(
    'mcp_request_duration_seconds',
    'Request duration',
    ['tool_name']
)

active_requests = Gauge(
    'mcp_active_requests',
    'Active requests'
)

# Instrument tool
@server.tool()
def monitored_tool(param: str) -> dict:
    """Tool with metrics."""
    active_requests.inc()
    start = time.time()

    try:
        result = execute_tool(param)
        request_count.labels(tool_name="monitored_tool", status="success").inc()
        return result

    except Exception as e:
        request_count.labels(tool_name="monitored_tool", status="error").inc()
        raise

    finally:
        duration = time.time() - start
        request_duration.labels(tool_name="monitored_tool").observe(duration)
        active_requests.dec()
```

### Logging

```python
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

logger = logging.getLogger(__name__)

@server.tool()
def logged_tool(param: str) -> dict:
    """Tool with structured logging."""
    logger.info(f"Tool called with param: {param}")

    try:
        result = execute_tool(param)
        logger.info(f"Tool succeeded: {result}")
        return result

    except Exception as e:
        logger.error(f"Tool failed: {str(e)}", exc_info=True)
        raise
```

---

## Scaling Considerations

### Stateless Design

- No server-local state
- Use external databases for persistence
- Use distributed caches for caching
- Enable horizontal scaling

### Load Balancing

```yaml
# Kubernetes automatically load balances across replicas
# Each request goes to a different pod
spec:
  replicas: 3  # 3 instances for redundancy
```

### Rate Limiting

```python
from slowapi import Limiter
from slowapi.util import get_remote_address

limiter = Limiter(key_func=get_remote_address)

@server.tool()
@limiter.limit("10/minute")
def rate_limited_tool(param: str) -> dict:
    """Tool with rate limiting."""
    return execute_tool(param)
```

---

## Database Considerations

### Connection Pooling

```python
from sqlalchemy import create_engine
from sqlalchemy.pool import QueuePool

# Create pooled connection
engine = create_engine(
    DATABASE_URL,
    poolclass=QueuePool,
    pool_size=10,
    max_overflow=20,
    pool_recycle=3600
)
```

### Migrations

```bash
# Use Alembic for database migrations
alembic init alembic
alembic revision --autogenerate -m "Initial migration"
alembic upgrade head
```

---

## Backup & Recovery

### Backup Strategy

```bash
# Docker: Backup database
docker-compose exec db pg_dump -U user -d mcp > backup.sql

# Restore from backup
docker-compose exec -T db psql -U user -d mcp < backup.sql
```

### Kubernetes: Persistent Volumes

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mcp-data-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

---

## Security Best Practices

### Runtime Security

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL
```

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: mcp-server-netpol
spec:
  podSelector:
    matchLabels:
      app: mcp-server
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: default
    ports:
    - protocol: TCP
      port: 8000
```

---

## Troubleshooting Deployment

### Check Pod Status

```bash
kubectl describe pod <pod-name> -n mcp-apps
kubectl logs <pod-name> -n mcp-apps
```

### Port Forwarding for Testing

```bash
kubectl port-forward svc/mcp-server 8000:80 -n mcp-apps
curl http://localhost:8000/health
```

### View Metrics

```bash
kubectl top nodes
kubectl top pods -n mcp-apps
```

---

**Last Updated**: 2025-11-27 | Production-grade deployment patterns
