---
name: moai-baas-railway-ext
description: Enterprise Railway Full-Stack Platform with AI-powered container orchestration,
---

## Quick Reference (30 seconds)

# Enterprise Railway Full-Stack Platform Expert 

---

## When to Use

**Automatic triggers**:
- Railway deployment architecture and container orchestration discussions
- Full-stack application development and database integration
- CI/CD pipeline setup and automated deployment strategies
- Multi-region deployment and scaling optimization

**Manual invocation**:
- Designing enterprise Railway architectures with optimal container configuration
- Implementing automated CI/CD pipelines with GitHub integration
- Planning full-stack application migrations to Railway
- Optimizing costs and auto-scaling configuration

---

# Quick Reference (Level 1)

## Advanced Scaling Strategies

```python
class RailwayScalingManager:
    def __init__(self):
        self.railway_client = RailwayClient()
        self.metrics_analyzer = MetricsAnalyzer()
        self.cost_optimizer = CostOptimizer()
    
    async def implement_intelligent_scaling(self, 
                                          project_id: str,
                                          scaling_config: ScalingConfiguration) -> ScalingImplementation:
        """Implement intelligent auto-scaling for Railway services."""
        
        # Analyze current usage patterns
        usage_analysis = await self.metrics_analyzer.analyze_usage_patterns(
            project_id, timeframe="7d"
        )
        
        # Configure predictive scaling
        predictive_config = self._configure_predictive_scaling(
            usage_analysis,
            scaling_config
        )
        
        # Set up cost optimization
        cost_optimization = self.cost_optimizer.optimize_scaling_costs(
            predictive_config,
            scaling_config.budget_constraints
        )
        
        return ScalingImplementation(
            scaling_rules=self._create_scaling_rules(predictive_config),
            monitoring_setup=self._setup_scaling_monitoring(),
            cost_controls=cost_optimization,
            performance_alerts=self._configure_performance_alerts()
        )
```

### Database Optimization Patterns

```typescript
// Railway PostgreSQL optimization with connection pooling
import { Pool } from 'pg';

// Production-ready database configuration
const productionPool = new Pool({
  connectionString: process.env.DATABASE_URL,
  ssl: { rejectUnauthorized: false },
  
  // Optimized for Railway's containerized environment
  max: 20, // Maximum connections in pool
  min: 5,  // Minimum connections to maintain
  idleTimeoutMillis: 30000, // Close idle connections after 30s
  connectionTimeoutMillis: 2000, // Give up connecting after 2s
  statement_timeout: 10000, // Kill slow queries after 10s
  
  // Connection retry logic
  retry: 3,
  retryDelay: 1000,
});

// Application database service
export class DatabaseService {
  private pool = productionPool;
  
  async query<T>(text: string, params?: any[]): Promise<T[]> {
    const start = Date.now();
    const client = await this.pool.connect();
    
    try {
      const result = await client.query(text, params);
      const duration = Date.now() - start;
      
      // Log slow queries for optimization
      if (duration > 1000) {
        console.warn('Slow query detected:', {
          query: text,
          duration,
          rowCount: result.rowCount
        });
      }
      
      return result.rows;
    } finally {
      client.release();
    }
  }
  
  async transaction<T>(callback: (client: any) => Promise<T>): Promise<T> {
    const client = await this.pool.connect();
    
    try {
      await client.query('BEGIN');
      const result = await callback(client);
      await client.query('COMMIT');
      return result;
    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally {
      client.release();
    }
  }
}

// Health check endpoint
app.get('/api/health', async (req, res) => {
  try {
    await db.query('SELECT 1');
    res.status(200).json({ 
      status: 'healthy',
      timestamp: new Date().toISOString(),
      database: 'connected'
    });
  } catch (error) {
    res.status(503).json({ 
      status: 'unhealthy',
      timestamp: new Date().toISOString(),
      database: 'disconnected'
    });
  }
});
```

### Environment Management

```python
class RailwayEnvironmentManager:
    def __init__(self):
        self.railway_client = RailwayClient()
        self.config_manager = ConfigurationManager()
    
    def setup_production_environment(self, project_id: str, config: EnvironmentConfig) -> EnvironmentSetup:
        """Configure production environment with best practices."""
        
        # Set up production variables
        production_vars = {
            # Application configuration
            'NODE_ENV': 'production',
            'LOG_LEVEL': 'info',
            
            # Database configuration
            'DATABASE_URL': config.database_url,
            'DATABASE_POOL_SIZE': '20',
            'DATABASE_TIMEOUT': '10000',
            
            # Security configuration
            'JWT_SECRET': config.jwt_secret,
            'ENCRYPTION_KEY': config.encryption_key,
            'CORS_ORIGIN': config.frontend_url,
            
            # External services
            'REDIS_URL': config.redis_url,
            'EMAIL_SERVICE_API_KEY': config.email_api_key,
            
            # Monitoring and observability
            'SENTRY_DSN': config.sentry_dsn,
            'LOGTAIL_SOURCE_TOKEN': config.logtail_token
        }
        
        # Configure environment variables
        env_setup = self.railway_client.set_environment_variables(
            project_id, production_vars
        )
        
        return EnvironmentSetup(
            variables=production_vars,
            security_config=self._configure_security(),
            monitoring_config=self._configure_monitoring(),
            backup_config=self._configure_backups()
        )
```

---

# Reference & Integration (Level 4)

---

## Core Implementation

## What It Does

Enterprise Railway Full-Stack Platform expert with AI-powered container orchestration, Context7 integration, and intelligent deployment automation for scalable modern applications.

**Revolutionary  capabilities**:
- 🤖 **AI-Powered Railway Architecture** using Context7 MCP for latest deployment patterns
- 📊 **Intelligent Container Orchestration** with automated scaling and optimization
- 🚀 **Real-time Performance Analytics** with AI-driven deployment insights
- 🔗 **Enterprise CI/CD Integration** with zero-configuration pipeline automation
- 📈 **Predictive Cost Analysis** with usage forecasting and resource optimization

---

## Railway Full-Stack Platform (November 2025)

### Core Features Overview
- **Container Deployment**: Docker container deployment from GitHub
- **Database Provisioning**: PostgreSQL, MongoDB, Redis with automatic setup
- **Multi-Region Deployment**: 4+ global regions for optimal latency
- **Git-Based CI/CD**: Automated deployments from Git commits
- **Background Jobs**: Scheduled tasks and job processing
- **One-Click Rollback**: Instant deployment history and rollback

### Supported Services
- **Applications**: Node.js, Python, Ruby, Go, Rust, Java, PHP
- **Databases**: PostgreSQL, MongoDB, Redis, MySQL
- **Static Sites**: Next.js, React, Vue, Angular, Hugo
- **Background Workers**: Bull queue, Celery, Sidekiq integration
- **File Storage**: Integrated with cloud storage providers

### Key Benefits
- **Zero Infrastructure Management**: No server configuration required
- **Developer-Friendly**: Focus on code, not deployment complexity
- **Auto-Scaling**: Automatic scaling based on traffic and load
- **Cost Controls**: Built-in spending limits and monitoring

### Performance Characteristics
- **Cold Start**: < 2 seconds for container spin-up
- **Scaling**: Instant horizontal and vertical scaling
- **Database Performance**: Optimized configurations for each database type
- **Global Latency**: < 50ms in major regions

---

# Core Implementation (Level 2)

## Railway Architecture Intelligence

```python
# AI-powered Railway architecture optimization with Context7
class RailwayArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.container_analyzer = ContainerAnalyzer()
        self.scaling_optimizer = ScalingOptimizer()
    
    async def design_optimal_railway_architecture(self, 
                                                requirements: ApplicationRequirements) -> RailwayArchitecture:
        """Design optimal Railway architecture using AI analysis."""
        
        # Get latest Railway and containerization documentation via Context7
        railway_docs = await self.context7_client.get_library_docs(
            context7_library_id='/railway/docs',
            topic="container deployment ci-cd scaling optimization 2025",
            tokens=3000
        )
        
        containerization_docs = await self.context7_client.get_library_docs(
            context7_library_id='/docker/docs',
            topic="optimization best practices orchestration 2025",
            tokens=2000
        )
        
        # Optimize container configuration
        container_optimization = self.container_analyzer.optimize_configuration(
            requirements.application_stack,
            containerization_docs
        )
        
        # Design scaling strategy
        scaling_strategy = self.scaling_optimizer.design_scaling_strategy(
            requirements.traffic_patterns,
            requirements.performance_requirements,
            railway_docs
        )
        
        return RailwayArchitecture(
            application_services=self._design_application_services(requirements),
            database_services=self._design_database_services(requirements),
            container_configuration=container_optimization,
            scaling_strategy=scaling_strategy,
            deployment_pipeline=self._design_cicd_pipeline(requirements),
            monitoring_setup=self._setup_monitoring(),
            cost_analysis=self._analyze_pricing_model(requirements)
        )
```

## Multi-Service Deployment Configuration

```yaml
# Railway multi-service application configuration
version: "1.0"
services:
  frontend:
    build:
      dockerfile: Dockerfile.frontend
      context: .
    environment:
      - NEXT_PUBLIC_API_URL=${API_URL}
      - NODE_ENV=production
    deploy:
      replicas: 2
      memory: 512Mi
      cpu: 0.5
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  backend:
    build:
      dockerfile: Dockerfile.backend
      context: .
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
    deploy:
      replicas: 3
      memory: 1Gi
      cpu: 1.0
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  worker:
    build:
      dockerfile: Dockerfile.worker
      context: .
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
    deploy:
      replicas: 1
      memory: 512Mi
      cpu: 0.5

  postgres:
    image: "postgres:16-alpine"
    environment:
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=${POSTGRES_DB}
    volume_mounts:
      - mountPath: /var/lib/postgresql/data
        name: postgres-data
    deploy:
      memory: 2Gi
      cpu: 1.0

  redis:
    image: "redis:7-alpine"
    deploy:
      memory: 512Mi
      cpu: 0.5
```

## CI/CD Pipeline Integration

```yaml
# GitHub Actions workflow for Railway deployment
name: Deploy to Railway

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run tests
        run: npm test
      
      - name: Run E2E tests
        run: npm run test:e2e

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Install Railway CLI
        run: npm install -g @railway/cli
      
      - name: Deploy to Railway
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
        run: |
          railway login --token $RAILWAY_TOKEN
          railway up --service frontend,backend,worker
```

---

# Advanced Implementation (Level 3)



---

## Reference & Resources

See [reference.md](reference.md) for detailed API reference and official documentation.


---

## Context7 Integration

### Related Libraries & Tools
- [Railway](/railwayapp/nixpacks): Infrastructure deployment

### Official Documentation
- [Documentation](https://docs.railway.app/)
- [API Reference](https://docs.railway.app/reference/)

### Version-Specific Guides
Latest stable version: Latest
- [Release Notes](https://railway.app/changelog)
- [Migration Guide](https://docs.railway.app/guides/migrate-from-heroku)
