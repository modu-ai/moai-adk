# Advanced DevOps Patterns

> **Version**: 4.0.0
> **Last Updated**: 2025-11-22
> **Focus**: CI/CD automation, GitOps, service mesh, advanced Kubernetes patterns

---

## Advanced CI/CD Patterns

### Matrix Testing with Dependencies

```yaml
name: Matrix CI/CD Pipeline

on: [push, pull_request]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
        node: [18, 20, 22]
        include:
          # Add dependencies for specific combinations
          - os: ubuntu-latest
            node: 20
            postgres: '15'
            redis: 'latest'
            coverage: true

    services:
      postgres:
        image: postgres:${{ matrix.postgres }}
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:${{ matrix.redis }}
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4
      - name: Use Node.js ${{ matrix.node }}
        uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node }}
          cache: npm

      - run: npm ci
      - run: npm run test

      - name: Upload coverage
        if: matrix.coverage
        uses: codecov/codecov-action@v3
```

### Artifact Caching and Distribution

```yaml
name: Build Cache Optimization

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # Cache node modules
      - name: Restore dependencies cache
        uses: actions/cache@v3
        with:
          path: ~/.npm
          key: ${{ runner.os }}-npm-${{ hashFiles('**/package-lock.json') }}
          restore-keys: |
            ${{ runner.os }}-npm-

      # Cache build artifacts
      - name: Restore build cache
        uses: actions/cache@v3
        with:
          path: dist/
          key: ${{ runner.os }}-build-${{ github.sha }}
          restore-keys: |
            ${{ runner.os }}-build-

      - run: npm ci
      - run: npm run build

      # Upload to artifact repository
      - name: Publish to Artifactory
        run: |
          curl -u "${{ secrets.ARTIFACTORY_USER }}:${{ secrets.ARTIFACTORY_PASSWORD }}" \
            -T "dist/app-*.tar.gz" \
            "https://artifactory.example.com/artifactory/builds/${{ github.ref_name }}/"
```

### Dynamic Secret Rotation

```yaml
name: Secret Rotation Pipeline

on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly

jobs:
  rotate-secrets:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: write

    steps:
      - uses: actions/checkout@v4

      - name: Assume AWS Role
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::${{ secrets.AWS_ACCOUNT }}:role/GitHubSecretsRotation
          aws-region: us-east-1

      - name: Rotate database password
        run: |
          NEW_PASSWORD=$(aws secretsmanager get-random-password --query RandomPassword --output text)

          # Update database
          psql -h $DB_HOST -U postgres -d postgres \
            -c "ALTER USER app_user WITH PASSWORD '$NEW_PASSWORD';"

          # Store in AWS Secrets Manager
          aws secretsmanager update-secret \
            --secret-id prod/db/password \
            --secret-string "$NEW_PASSWORD"

      - name: Rotate API keys
        run: |
          # Rotate all API keys
          ./scripts/rotate-api-keys.sh

          # Commit changes
          git config user.name "Secret Rotator"
          git config user.email "bot@example.com"
          git add -A
          git commit -m "chore: rotate secrets [skip ci]" || true
          git push
```

---

## GitOps Patterns

### Progressive Delivery with Flagger

```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: myapp
  namespace: production

spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp

  service:
    port: 8080
    targetPort: 8080

  # Progressive delivery config
  skipAnalysis: false

  analysis:
    interval: 1m
    threshold: 10
    maxWeight: 50
    stepWeight: 5
    stepWeightPromotion: 10

    # Metrics to track
    metrics:
    - name: error-rate
      thresholdRange:
        max: 5
      interval: 1m

    - name: latency
      thresholdRange:
        max: 500
      interval: 1m

    # Webhooks for custom checks
    webhooks:
    - name: smoke-tests
      url: http://flagger-loadtester:80/
      metadata:
        type: smoke
        cmd: "curl -s http://myapp-canary:8080/api/health"
        logCmdOutput: "true"

    - name: load-test
      url: http://flagger-loadtester:80/
      metadata:
        type: load
        cmd: "hey -z 1m -q 10 -c 2 http://myapp-canary:8080/"
        logCmdOutput: "false"
```

### ArgoCD Multi-Repo Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  namespace: argocd

data:
  # Enable multi-repo support
  url: https://argocd.example.com

  # Account configuration
  accounts.dev: apiKey,login
  accounts.dev.capabilities: api,login

  # RBAC configuration
  policy.default: role:readonly
  policy.csv: |
    p, role:admin, *, *, */*, allow
    p, role:dev, applications, get, development-*, allow
    p, role:dev, applications, sync, development-*, allow
    p, role:readonly, *, get, */*, allow

    g, developers, role:dev
    g, admins, role:admin

---
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: myapp-environments

spec:
  generators:
  # Generate applications for each environment
  - list:
      elements:
      - cluster: staging
        url: https://staging-api.example.com
      - cluster: production
        url: https://api.example.com

  template:
    metadata:
      name: myapp-{{ cluster }}

    spec:
      project: default

      source:
        repoURL: https://github.com/org/gitops-repo
        targetRevision: main
        path: apps/myapp/overlays/{{ cluster }}

      destination:
        server: https://kubernetes.default.svc
        namespace: myapp

      syncPolicy:
        automated:
          prune: true
          selfHeal: true

        syncOptions:
        - CreateNamespace=true
```

---

## Service Mesh Patterns

### Istio Traffic Management

```yaml
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: api-gateway
  namespace: production

spec:
  selector:
    istio: ingressgateway

  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
      credentialName: api-cert
    hosts:
    - "api.example.com"

---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: api-vs
  namespace: production

spec:
  hosts:
  - "api.example.com"

  gateways:
  - api-gateway

  http:
  # Route by path
  - match:
    - uri:
        prefix: /v2/
    route:
    - destination:
        host: api-service
        subset: v2
      weight: 100

  # Route by header
  - match:
    - headers:
        user-type:
          exact: beta-tester
    route:
    - destination:
        host: api-service
        subset: v2
      weight: 100

  # Default route
  - route:
    - destination:
        host: api-service
        subset: v1
      weight: 100

  timeout: 30s

  retries:
    attempts: 3
    perTryTimeout: 10s

---
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: api-dr
  namespace: production

spec:
  host: api-service

  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 100
        http2MaxRequests: 1000

    outlierDetection:
      consecutiveErrors: 5
      interval: 30s
      baseEjectionTime: 30s

  subsets:
  - name: v1
    labels:
      version: v1

  - name: v2
    labels:
      version: v2
    trafficPolicy:
      connectionPool:
        http:
          http1MaxPendingRequests: 50
```

---

## Kubernetes Advanced Patterns

### Pod Disruption Budgets and Scaling

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: myapp-pdb
  namespace: production

spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: myapp

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: myapp-hpa
  namespace: production

spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp

  minReplicas: 3
  maxReplicas: 10

  metrics:
  # CPU-based scaling
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70

  # Memory-based scaling
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80

  # Custom metric scaling
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "1k"

  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 15

    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
      - type: Pods
        value: 4
        periodSeconds: 15
      selectPolicy: Max
```

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: myapp-network-policy
  namespace: production

spec:
  podSelector:
    matchLabels:
      app: myapp

  policyTypes:
  - Ingress
  - Egress

  ingress:
  # Allow from ingress controller
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080

  # Allow from monitoring
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 9090

  egress:
  # Allow DNS
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: UDP
      port: 53

  # Allow to database
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432

  # Allow external HTTPS
  - to:
    - ipBlock:
        cidr: 0.0.0.0/0
        except:
        - 169.254.169.254/32  # Block metadata service
    ports:
    - protocol: TCP
      port: 443
```

---

**Best Practices**:
- Implement canary deployments for risk reduction
- Use GitOps for infrastructure as code
- Implement proper network policies for security
- Monitor and optimize resource usage
- Use pod disruption budgets for reliability
- Implement proper RBAC and secrets management

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
