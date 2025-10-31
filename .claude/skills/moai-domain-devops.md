# moai-domain-devops

CI/CD pipelines, monitoring, infrastructure patterns, and deployment automation.

## Quick Start

DevOps encompasses the practices, tools, and culture of building, testing, and deploying software reliably. Use this skill when setting up CI/CD pipelines, configuring monitoring, implementing infrastructure as code, or automating deployments.

## Core Patterns

### Pattern 1: GitHub Actions CI/CD Pipeline

**Pattern**: Automate testing, building, and deploying on every git push.

```yaml
# .github/workflows/ci.yml - Complete CI/CD pipeline
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: ['3.11', '3.12', '3.13']

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Set up Python ${{ matrix.python-version }}
        uses: actions/setup-python@v4
        with:
          python-version: ${{ matrix.python-version }}

      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install -r requirements-dev.txt

      - name: Lint with ruff
        run: ruff check .

      - name: Type check with mypy
        run: mypy src/ --ignore-missing-imports

      - name: Run tests with pytest
        run: pytest tests/ -v --cov=src --cov-report=xml
        env:
          DATABASE_URL: postgresql://postgres:postgres@localhost/test_db

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.xml

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run security scan with bandit
        run: |
          pip install bandit
          bandit -r src/ -f json -o bandit.json || true

      - name: Dependency vulnerability check
        run: |
          pip install pip-audit
          pip-audit || true

  deploy:
    needs: [test, security]
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Deploy to Render
        run: |
          curl -X POST https://api.render.com/deploy/srv-${{ secrets.RENDER_SERVICE_ID }}?key=${{ secrets.RENDER_DEPLOY_KEY }}
```

**When to use**:
- Running tests automatically on every push
- Checking code quality and security
- Building and deploying applications
- Preventing broken code from reaching production

**Key benefits**:
- Catches bugs before they reach users
- Consistent testing across team
- Faster feedback loop
- Fully automated deployments

### Pattern 2: Monitoring & Observability

**Pattern**: Track application health, performance, and errors in production.

```python
# app/monitoring.py - Application monitoring setup
import logging
from prometheus_client import Counter, Histogram
import time
from functools import wraps

# Prometheus metrics
request_count = Counter(
    'http_requests_total',
    'Total HTTP requests',
    ['method', 'endpoint', 'status'],
)

request_duration = Histogram(
    'http_request_duration_seconds',
    'HTTP request duration',
    ['method', 'endpoint'],
)

error_count = Counter(
    'errors_total',
    'Total errors',
    ['type', 'endpoint'],
)

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
)

logger = logging.getLogger(__name__)

# FastAPI middleware for monitoring
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

def setup_monitoring(app: FastAPI):
    @app.middleware("http")
    async def track_metrics(request: Request, call_next):
        start_time = time.time()

        try:
            response = await call_next(request)
            duration = time.time() - start_time

            # Record metrics
            request_count.labels(
                method=request.method,
                endpoint=request.url.path,
                status=response.status_code,
            ).inc()

            request_duration.labels(
                method=request.method,
                endpoint=request.url.path,
            ).observe(duration)

            logger.info(
                f"{request.method} {request.url.path} - "
                f"Status: {response.status_code} - "
                f"Duration: {duration:.3f}s"
            )

            return response

        except Exception as e:
            error_count.labels(
                type=type(e).__name__,
                endpoint=request.url.path,
            ).inc()

            logger.error(
                f"Error in {request.method} {request.url.path}: {str(e)}"
            )

            return JSONResponse(
                {"error": "Internal server error"},
                status_code=500,
            )

    # Add health check endpoint
    @app.get("/metrics")
    async def metrics():
        """Prometheus metrics endpoint"""
        from prometheus_client import generate_latest
        return generate_latest()

    return app
```

**When to use**:
- Monitoring API response times
- Tracking error rates
- Detecting performance degradation
- Alerting on critical issues

**Key benefits**:
- Proactive problem detection
- Understanding user impact
- Performance optimization data
- SLA compliance tracking

### Pattern 3: Infrastructure as Code (IaC)

**Pattern**: Define infrastructure declaratively so it can be version-controlled and reproduced.

```yaml
# infrastructure/docker-compose.yml - Local development stack
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8000:8000"
    environment:
      DATABASE_URL: postgresql://user:password@db:5432/api_db
      REDIS_URL: redis://redis:6379
    depends_on:
      - db
      - redis
    volumes:
      - .:/app
    command: uvicorn app.main:app --host 0.0.0.0 --reload

  db:
    image: postgres:15
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: api_db
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

**When to use**:
- Setting up reproducible development environments
- Defining production infrastructure
- Managing multiple deployment targets
- Versioning infrastructure changes

**Key benefits**:
- Consistency across environments
- Easy onboarding for team members
- Infrastructure versioning with git
- Disaster recovery capability

## Progressive Disclosure

### Level 1: Basic CI/CD
- Run tests automatically on push
- Check code style with linters
- Deploy to staging environment
- Manual approval for production

### Level 2: Advanced Monitoring
- Set up metrics and alerts
- Implement distributed tracing
- Log aggregation and analysis
- Performance dashboards

### Level 3: Expert Infrastructure
- Infrastructure as code (Terraform, Pulumi)
- Multi-region deployments
- Canary releases and blue-green deployments
- Cost optimization and autoscaling

## Works Well With

- **GitHub**: Source control and Actions for CI/CD
- **Render**: Deploy applications and monitor performance
- **Vercel**: Frontend CI/CD and performance monitoring
- **Supabase**: Database backups and monitoring
- **Prometheus**: Metrics collection and alerting
- **ELK Stack**: Log aggregation and analysis

## References

- **GitHub Actions**: https://docs.github.com/en/actions
- **CI/CD Best Practices**: https://www.atlassian.com/continuous-delivery
- **Monitoring Patterns**: https://www.site24x7.com/blog/infrastructure-monitoring.html
- **Docker Best Practices**: https://docs.docker.com/develop/dev-best-practices/
- **Kubernetes**: https://kubernetes.io/docs/concepts/overview/
