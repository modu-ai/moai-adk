# DevOps Implementation Examples

> **Version**: 4.0.0
> **Last Updated**: 2025-11-22
> **Focus**: CI/CD pipelines, GitHub Actions, Terraform IaC, Kubernetes deployments, GitOps

---

## CI/CD Pipeline Examples

### GitHub Actions Complete Workflow

```yaml
name: Production CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - name: Install and test
        run: |
          npm ci --legacy-peer-deps
          npm run lint
          npm run test:coverage

  docker-build:
    needs: build-test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build and push image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest

  deploy-production:
    needs: [ build-test, docker-build ]
    runs-on: ubuntu-latest
    environment: production
    steps:
      - name: Deploy to production
        run: |
          kubectl set image deployment/app \
            app=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest
```

### Multi-Environment Terraform Deployment

```yaml
name: Infrastructure Deployment

on:
  workflow_dispatch:
    inputs:
      environment:
        type: choice
        options: [dev, staging, production]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Deploy with Terraform
        run: |
          terraform init
          terraform plan -var-file="envs/${{ inputs.environment }}.tfvars"
          terraform apply -auto-approve -var-file="envs/${{ inputs.environment }}.tfvars"
```

---

## Kubernetes Deployment Examples

### Blue-Green Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-blue
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
      version: blue
  template:
    metadata:
      labels:
        app: myapp
        version: blue
    spec:
      containers:
      - name: app
        image: myapp:v1.0.0
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10

---
apiVersion: v1
kind: Service
metadata:
  name: myapp-service
spec:
  selector:
    app: myapp
    version: blue
  ports:
  - port: 80
    targetPort: 8080
```

---

## Terraform Infrastructure

### EKS Cluster Configuration

```hcl
resource "aws_eks_cluster" "main" {
  name            = "production-cluster"
  role_arn        = aws_iam_role.cluster.arn
  version         = "1.31"

  vpc_config {
    subnet_ids = aws_subnet.private[*].id
  }
}

resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "main-node-group"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.private[*].id

  scaling_config {
    desired_size = 3
    max_size     = 10
    min_size     = 3
  }

  instance_types = ["t3.medium"]
}
```

### RDS Database

```hcl
resource "aws_db_instance" "main" {
  identifier         = "production-postgres"
  engine             = "postgres"
  engine_version     = "15.5"
  instance_class     = "db.r6i.xlarge"
  allocated_storage  = 100
  storage_encrypted  = true

  db_name  = "proddb"
  username = "postgres"
  password = random_password.db_password.result

  multi_az              = true
  backup_retention_period = 30
  skip_final_snapshot   = false
}
```

---

## GitOps with ArgoCD

### Application Definition

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp

spec:
  project: default
  source:
    repoURL: https://github.com/org/gitops-repo
    targetRevision: main
    path: apps/myapp/overlays/production

  destination:
    server: https://kubernetes.default.svc
    namespace: production

  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

---

## Best Practices

- Use infrastructure as code for all deployments
- Implement comprehensive testing before deployment
- Use blue-green deployments for zero downtime
- Store secrets securely with appropriate access control
- Monitor all deployments with proper observability
- Maintain detailed documentation and runbooks
- Practice disaster recovery procedures regularly

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
