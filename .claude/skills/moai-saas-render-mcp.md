# moai-saas-render-mcp

Deploy FastAPI applications to Render with PostgreSQL, environment management, and production optimization.

## Quick Start

Render is a cloud platform for deploying web services, background workers, cron jobs, and databases. Use this skill when deploying FastAPI backends, setting up PostgreSQL databases, configuring environment variables, or creating scheduled jobs.

## Core Patterns

### Pattern 1: FastAPI Web Service Deployment

**Pattern**: Deploy a production-ready FastAPI application with health checks and environment configuration.

```yaml
# render.yaml - Infrastructure as Code for Render
services:
  - type: web
    name: api-service
    runtime: python
    region: oregon
    plan: standard
    numInstances: 2
    healthCheckPath: /health
    envVars:
      - key: PYTHON_VERSION
        value: 3.12
      - key: DATABASE_URL
        fromDatabase:
          name: postgres-db
          property: connectionString
      - key: SECRET_KEY
        sync: false # Set manually in dashboard
    buildCommand: pip install -r requirements.txt
    startCommand: gunicorn app.main:app --workers 4 --timeout 60

databases:
  - name: postgres-db
    databaseName: api_db
    user: api_user
    region: oregon
    plan: standard
```

**When to use**:
- Deploying REST APIs built with FastAPI
- Running production services with auto-scaling
- Integrating with Postgres databases
- Managing environment-specific configurations

**Key benefits**:
- Declarative infrastructure configuration
- Automatic SSL/TLS certificates
- Built-in load balancing and auto-scaling
- Simple git-based deployments

### Pattern 2: Background Jobs & Cron Schedulers

**Pattern**: Schedule periodic background tasks (email sending, data processing, cleanup).

```python
# app/tasks.py - Background job definitions
from celery import Celery
from app.config import settings
import asyncio
from sqlalchemy import select
from app.models import User

celery_app = Celery(
    'tasks',
    broker=settings.REDIS_URL,
    backend=settings.REDIS_URL,
)

@celery_app.task
def send_weekly_digest():
    """Send digest email to all active users"""
    from app.email import send_digest_email

    # Fetch users
    users = query_active_users()
    for user in users:
        send_digest_email(user.email)

    return f'Sent digests to {len(users)} users'

@celery_app.task
def cleanup_expired_sessions():
    """Remove sessions older than 30 days"""
    from datetime import timedelta, datetime
    from app.db import SessionLocal

    db = SessionLocal()
    cutoff = datetime.utcnow() - timedelta(days=30)

    deleted = db.query(Session).filter(
        Session.created_at < cutoff
    ).delete()

    db.commit()
    db.close()

    return f'Deleted {deleted} expired sessions'

# Celery beat schedule for periodic tasks
celery_app.conf.beat_schedule = {
    'send-weekly-digest': {
        'task': 'app.tasks.send_weekly_digest',
        'schedule': crontab(day_of_week=1, hour=9, minute=0),  # Monday 9 AM
    },
    'cleanup-sessions-daily': {
        'task': 'app.tasks.cleanup_expired_sessions',
        'schedule': crontab(hour=2, minute=0),  # Daily at 2 AM
    },
}
```

**When to use**:
- Sending batch emails
- Processing large data files
- Running periodic maintenance tasks
- Generating reports

**Key benefits**:
- Tasks run reliably even if API crashes
- Automatic retry on failure
- Scheduled execution at specific times
- No additional infrastructure needed

### Pattern 3: Environment Management & Secrets

**Pattern**: Configure environment variables and secrets for development, staging, and production.

```python
# config.py - Environment configuration
from pydantic_settings import BaseSettings
from typing import Optional

class Settings(BaseSettings):
    # API Configuration
    API_TITLE: str = "My API"
    API_VERSION: str = "1.0.0"
    DEBUG: bool = False

    # Database
    DATABASE_URL: str
    DATABASE_POOL_SIZE: int = 20
    DATABASE_MAX_OVERFLOW: int = 0

    # Authentication
    SECRET_KEY: str
    JWT_ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 30

    # AWS S3 (optional)
    AWS_ACCESS_KEY_ID: Optional[str] = None
    AWS_SECRET_ACCESS_KEY: Optional[str] = None
    AWS_S3_BUCKET: Optional[str] = None

    # Email configuration
    SMTP_HOST: str
    SMTP_PORT: int = 587
    SMTP_USER: str
    SMTP_PASSWORD: str

    # Redis (optional, for caching/celery)
    REDIS_URL: Optional[str] = None

    class Config:
        env_file = ".env"
        case_sensitive = True

# Usage in FastAPI app
from fastapi import FastAPI
from config import Settings

settings = Settings()

app = FastAPI(
    title=settings.API_TITLE,
    version=settings.API_VERSION,
    debug=settings.DEBUG,
)

@app.get("/health")
async def health_check():
    """Health check endpoint for Render"""
    return {
        "status": "ok",
        "version": settings.API_VERSION,
    }
```

**When to use**:
- Managing configuration across environments
- Protecting sensitive credentials
- Deploying the same code to dev/staging/production
- Rotating secrets without redeploying

**Key benefits**:
- Secrets never hardcoded in source
- Environment-specific settings
- Easy configuration management
- Audit trail of secret access

## Progressive Disclosure

### Level 1: Basic Deployment
- Connect GitHub repository
- Deploy FastAPI app with single command
- Access live API URL
- View deployment logs

### Level 2: Advanced Configuration
- Set up PostgreSQL database
- Configure environment variables
- Enable auto-scaling
- Monitor performance metrics

### Level 3: Expert Optimization
- Deploy background workers
- Create cron job schedulers
- Implement custom health checks
- Optimize database connection pooling

## Works Well With

- **FastAPI 0.120+**: Official recommended hosting
- **Python 3.13**: Full async/await support
- **PostgreSQL 15**: Managed database instances
- **Supabase**: Use Supabase database from Render backend
- **Vercel**: Next.js frontend consuming Render backend API
- **GitHub**: Automatic deployments on git push

## References

- **Official Documentation**: https://render.com/docs
- **FastAPI Deployment**: https://fastapi.tiangolo.com/deployment/
- **PostgreSQL Hosting**: https://render.com/docs/databases
- **Environment Variables**: https://render.com/docs/environment-variables
- **Background Jobs**: https://render.com/docs/cron-jobs
