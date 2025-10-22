# DevOps - Working Examples

> Real-world CI/CD, Docker, Kubernetes, and Infrastructure as Code examples

---

## Example 1: Dockerize a Node.js Application

### Project Structure
```
my-app/
├── Dockerfile
├── .dockerignore
├── package.json
├── src/
│   └── index.js
└── tests/
    └── app.test.js
```

### Dockerfile (Multi-stage Production Build)
```dockerfile
# Stage 1: Build
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build && npm test

# Stage 2: Production
FROM node:20-alpine
RUN addgroup -g 1001 appgroup && \
    adduser -D -u 1001 -G appgroup appuser
WORKDIR /app
COPY --from=builder --chown=appuser:appgroup /app/dist ./dist
COPY --from=builder --chown=appuser:appgroup /app/node_modules ./node_modules
USER appuser
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=3s \
  CMD node -e "require('http').get('http://localhost:3000/health', (r) => {process.exit(r.statusCode === 200 ? 0 : 1)})"
CMD ["node", "dist/index.js"]
```

### .dockerignore
```
node_modules
npm-debug.log
.git
.env
*.md
tests
coverage
.dockerignore
Dockerfile
```

### Build and Run
```bash
# Build image
docker build -t myapp:1.0.0 .

# Run container
docker run -d -p 3000:3000 --name myapp myapp:1.0.0

# View logs
docker logs -f myapp

# Stop and remove
docker stop myapp && docker rm myapp
```

---

## Example 2: Kubernetes Deployment with ConfigMap and Secrets

### Directory Structure
```
k8s/
├── namespace.yaml
├── configmap.yaml
├── secret.yaml
├── deployment.yaml
├── service.yaml
└── ingress.yaml
```

### namespace.yaml
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: production
```

### configmap.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: production
data:
  APP_ENV: "production"
  LOG_LEVEL: "info"
  MAX_CONNECTIONS: "100"
```

### secret.yaml
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: production
type: Opaque
stringData:
  DATABASE_URL: "postgresql://user:pass@db.example.com:5432/appdb"
  API_KEY: "your-secret-api-key"
```

### deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: production
  labels:
    app: web
    version: "1.0"
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
        version: "1.0"
    spec:
      containers:
      - name: app
        image: myregistry.io/myapp:1.0.0
        ports:
        - containerPort: 3000
          name: http
        envFrom:
        - configMapRef:
            name: app-config
        - secretRef:
            name: app-secrets
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
```

### service.yaml
```yaml
apiVersion: v1
kind: Service
metadata:
  name: web-service
  namespace: production
spec:
  selector:
    app: web
  ports:
  - name: http
    protocol: TCP
    port: 80
    targetPort: 3000
  type: ClusterIP
```

### Deploy to Kubernetes
```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Apply all resources
kubectl apply -f k8s/

# Check status
kubectl get all -n production

# Scale deployment
kubectl scale deployment web-app --replicas=5 -n production

# View logs
kubectl logs -f deployment/web-app -n production

# Port forward for local testing
kubectl port-forward -n production service/web-service 8080:80
```

---

## Example 3: Complete GitHub Actions CI/CD Pipeline

### .github/workflows/ci-cd.yml
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
    tags: ['v*']
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}
  NODE_VERSION: '20'

jobs:
  lint-and-test:
    name: Lint and Test
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run linter
        run: npm run lint

      - name: Run unit tests
        run: npm test -- --coverage

      - name: Enforce coverage threshold
        run: |
          COVERAGE=$(jq -r '.total.lines.pct' coverage/coverage-summary.json)
          echo "Coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 85" | bc -l) )); then
            echo "ERROR: Coverage $COVERAGE% is below 85%"
            exit 1
          fi

      - name: Upload coverage reports
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage/coverage-final.json
          flags: unittests
          name: codecov-umbrella

  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload Trivy results to GitHub Security tab
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'

  build-and-push:
    name: Build and Push Docker Image
    needs: [lint-and-test, security-scan]
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix={{branch}}-

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Scan image for vulnerabilities
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
          format: 'table'
          exit-code: '1'
          severity: 'CRITICAL,HIGH'

  deploy-staging:
    name: Deploy to Staging
    needs: build-and-push
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    environment:
      name: staging
      url: https://staging.example.com
    steps:
      - uses: actions/checkout@v4

      - name: Configure kubectl
        run: |
          mkdir -p $HOME/.kube
          echo "${{ secrets.KUBE_CONFIG_STAGING }}" | base64 -d > $HOME/.kube/config

      - name: Deploy to staging
        run: |
          kubectl set image deployment/web-app \
            app=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:develop-${{ github.sha }} \
            -n staging
          kubectl rollout status deployment/web-app -n staging --timeout=5m

      - name: Run smoke tests
        run: |
          curl -f https://staging.example.com/health || exit 1

  deploy-production:
    name: Deploy to Production
    needs: build-and-push
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://example.com
    steps:
      - uses: actions/checkout@v4

      - name: Configure kubectl
        run: |
          mkdir -p $HOME/.kube
          echo "${{ secrets.KUBE_CONFIG_PROD }}" | base64 -d > $HOME/.kube/config

      - name: Deploy to production
        run: |
          kubectl set image deployment/web-app \
            app=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:main-${{ github.sha }} \
            -n production
          kubectl rollout status deployment/web-app -n production --timeout=10m

      - name: Run smoke tests
        run: |
          curl -f https://example.com/health || exit 1

      - name: Notify Slack
        if: always()
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Production deployment ${{ job.status }}'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

---

## Example 4: Terraform AWS Infrastructure

### Project Structure
```
terraform/
├── main.tf
├── variables.tf
├── outputs.tf
├── vpc.tf
├── ecs.tf
└── rds.tf
```

### main.tf
```hcl
terraform {
  required_version = ">= 1.10"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  backend "s3" {
    bucket         = "myapp-terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}

provider "aws" {
  region = var.aws_region
  default_tags {
    tags = {
      Environment = var.environment
      Project     = "myapp"
      ManagedBy   = "terraform"
    }
  }
}
```

### vpc.tf
```hcl
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.environment}-vpc"
  }
}

resource "aws_subnet" "public" {
  count                   = 2
  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 8, count.index)
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.environment}-public-subnet-${count.index + 1}"
    Type = "public"
  }
}

resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index + 10)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "${var.environment}-private-subnet-${count.index + 1}"
    Type = "private"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.environment}-igw"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "${var.environment}-public-rt"
  }
}

resource "aws_route_table_association" "public" {
  count          = 2
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}
```

### ecs.tf
```hcl
resource "aws_ecs_cluster" "main" {
  name = "${var.environment}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

resource "aws_ecs_task_definition" "app" {
  family                   = "${var.environment}-app"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.task_cpu
  memory                   = var.task_memory
  execution_role_arn       = aws_iam_role.ecs_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode([
    {
      name  = "app"
      image = var.app_image
      portMappings = [
        {
          containerPort = var.app_port
          protocol      = "tcp"
        }
      ]
      environment = [
        {
          name  = "NODE_ENV"
          value = var.environment
        }
      ]
      secrets = [
        {
          name      = "DATABASE_URL"
          valueFrom = aws_secretsmanager_secret.db_url.arn
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.app.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:${var.app_port}/health || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 60
      }
    }
  ])
}

resource "aws_ecs_service" "app" {
  name            = "${var.environment}-app-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = var.app_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = [aws_security_group.ecs_tasks.id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn
    container_name   = "app"
    container_port   = var.app_port
  }

  depends_on = [aws_lb_listener.app]
}

resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/${var.environment}-app"
  retention_in_days = 30
}
```

### Terraform Workflow
```bash
# Initialize
terraform init

# Select workspace (environment)
terraform workspace select production || terraform workspace new production

# Validate
terraform validate

# Plan changes
terraform plan -out=tfplan

# Apply changes
terraform apply tfplan

# Show outputs
terraform output

# Destroy (careful!)
terraform destroy
```

---

## Example 5: Docker Compose Development Environment

### docker-compose.yml
```yaml
version: '3.9'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
      args:
        NODE_VERSION: 20
    ports:
      - "3000:3000"
    volumes:
      - .:/app
      - /app/node_modules
      - /app/.next
    environment:
      - NODE_ENV=development
      - DATABASE_URL=postgresql://app:secret@db:5432/appdb
      - REDIS_URL=redis://redis:6379
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
    networks:
      - app-network
    command: npm run dev

  db:
    image: postgres:17.2-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=app
      - POSTGRES_PASSWORD=secret
      - POSTGRES_DB=appdb
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./docker/postgres/init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U app"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network

  redis:
    image: redis:7.4.0-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    networks:
      - app-network

  nginx:
    image: nginx:1.27-alpine
    ports:
      - "80:80"
    volumes:
      - ./docker/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - app
    networks:
      - app-network

volumes:
  postgres_data:
  redis_data:

networks:
  app-network:
    driver: bridge
```

### Dockerfile.dev
```dockerfile
FROM node:20-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY package*.json ./

RUN npm install

COPY . .

EXPOSE 3000

CMD ["npm", "run", "dev"]
```

### Usage
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Run migrations
docker-compose exec app npm run migrate

# Run tests
docker-compose exec app npm test

# Access database
docker-compose exec db psql -U app -d appdb

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

---

## Best Practices Summary

1. **Docker**: Use multi-stage builds, specific versions, non-root users, .dockerignore
2. **Kubernetes**: Define resource limits, health checks, use ConfigMaps/Secrets, enable rolling updates
3. **CI/CD**: Enforce quality gates (coverage ≥85%), scan for vulnerabilities, automate deployments
4. **Terraform**: Use remote state, modules, workspaces, version constraints
5. **Security**: Scan images, rotate secrets, use least privilege, enable audit logs

---

**Last Updated**: 2025-10-22
**Tools**: Docker 27.4.0, Kubernetes 1.32.0, Terraform 1.10.0, GitHub Actions
