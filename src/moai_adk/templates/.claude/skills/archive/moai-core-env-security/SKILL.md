---
name: moai-core-env-security
description: Environment variable security, secret management, and credential protection patterns
version: 1.0.0
modularized: false
tags:
  - env
  - enterprise
  - framework
  - architecture
  - security
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: security, moai, core, env  


# Environment Variable Security & Secret Management

## Quick Reference (30 seconds)

Environment variable security is an essential security pattern for safely managing application configuration and ensuring sensitive credentials (API keys, database passwords, JWT tokens) are not exposed in version control, logs, or memory dumps. Implement multi-layered defense using `.env` files, cloud secret managers, and IAM roles.

**Core principles**:
- Never commit `.env` files to version control
- Use dedicated secret managers in production environments
- Rotate secrets every 30-90 days
- Apply least privilege principle (IAM policies)


## Implementation Guide

### 1. Development Environment Setup

**`.env` file pattern**:
```bash
# .env (development - exclude from version control)
DATABASE_URL=postgresql://user:password@localhost:5432/devdb
API_KEY=sk_dev_abc123xyz789
JWT_SECRET=dev-secret-key-not-for-production
DEBUG=true
```

**`.env.example` (template - include in version control)**:
```bash
# .env.example (no actual values, field definitions only)
DATABASE_URL=
API_KEY=
JWT_SECRET=
DEBUG=
```

**`.gitignore` configuration**:
```bash
# Protect secret files
.env
.env.local
.env.*.local
.env.development
.env.test.local
.env.production

# Other sensitive files
*.pem
*.key
secrets.yaml
credentials.json
.aws/
.gcloud/
```

### 2. Python Environment Variable Loading

**Using python-dotenv**:
```python
from dotenv import load_dotenv
import os

# Load .env file
load_dotenv()

# Access environment variables
DATABASE_URL = os.getenv('DATABASE_URL')
API_KEY = os.getenv('API_KEY', default='fallback-value')
```

**Required variable validation**:
```python
import os
from typing import List

def validate_env_variables(required: List[str]) -> None:
    """Validate required environment variables"""
    missing = [var for var in required if not os.getenv(var)]
    if missing:
        raise ValueError(f"Missing required environment variables: {missing}")

# Usage
validate_env_variables([
    'DATABASE_URL',
    'API_KEY',
    'JWT_SECRET'
])
```

**Type-safe validation using Pydantic**:
```python
from pydantic import BaseSettings, Field

class Settings(BaseSettings):
    database_url: str = Field(..., description="PostgreSQL connection string")
    api_key: str = Field(..., min_length=20)
    jwt_secret: str = Field(..., min_length=32)
    debug: bool = Field(default=False)

    class Config:
        env_file = ".env"

settings = Settings()  # Auto validation
```

### 3. Cloud Secret Management (Production)

**AWS Secrets Manager**:
```python
import boto3
import json

def get_aws_secret(secret_name: str, region: str = 'us-east-1') -> dict:
    """Retrieve secret from AWS"""
    client = boto3.client('secretsmanager', region_name=region)
    try:
        response = client.get_secret_value(SecretId=secret_name)
        return json.loads(response['SecretString'])
    except Exception as e:
        raise RuntimeError(f"Failed to retrieve secret: {e}")

# Usage
db_creds = get_aws_secret('prod/database')
DATABASE_URL = db_creds['url']
```

**Google Cloud Secret Manager**:
```python
from google.cloud import secretmanager

def get_gcp_secret(project_id: str, secret_id: str, version: str = 'latest') -> str:
    """Retrieve secret from Google Cloud"""
    client = secretmanager.SecretManagerServiceClient()
    name = f"projects/{project_id}/secrets/{secret_id}/versions/{version}"
    response = client.access_secret_version(request={"name": name})
    return response.payload.data.decode("UTF-8")

# Usage
api_key = get_gcp_secret('my-project', 'api-key')
```

**HashiCorp Vault**:
```python
import hvac

def get_vault_secret(path: str, vault_addr: str, token: str) -> dict:
    """Retrieve secret from Vault"""
    client = hvac.Client(url=vault_addr, token=token)
    response = client.secrets.kv.read_secret_version(path)
    return response['data']['data']

# Usage
db_secret = get_vault_secret('secret/db/prod', 'https://vault.example.com', token)
```

### 4. Secret Rotation

**Automatic rotation pattern**:
```python
import os
from datetime import datetime, timedelta

class SecretRotationManager:
    """Secret rotation management"""

    def __init__(self, rotation_days: int = 90):
        self.rotation_days = rotation_days

    def should_rotate(self, secret_name: str) -> bool:
        """Check if rotation is needed"""
        last_rotated = self._get_last_rotation_date(secret_name)
        if not last_rotated:
            return True

        days_since = (datetime.now() - last_rotated).days
        return days_since >= self.rotation_days

    def rotate_secret(self, secret_name: str) -> str:
        """Execute secret rotation"""
        new_secret = self._generate_new_secret()
        self._update_secret_in_manager(secret_name, new_secret)
        self._record_rotation_timestamp(secret_name)
        return new_secret

    def _generate_new_secret(self) -> str:
        import secrets
        return secrets.token_urlsafe(32)

    def _update_secret_in_manager(self, name: str, value: str):
        # Update in AWS Secrets Manager, Vault, etc.
        pass

    def _record_rotation_timestamp(self, name: str):
        # Record rotation timestamp
        pass
```

### 5. Secret Detection and Prevention

**Git Pre-commit Hook** (git-secrets):
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Install git-secrets
git secrets --install

# Define patterns
git secrets --register-aws

# Scan all commits
git secrets --scan
```

**Python Pre-commit validation**:
```python
import re

SECRET_PATTERNS = [
    r'(?i)(password|passwd|pwd)\s*=\s*[\'"]?([^\s\'\"]+)',
    r'(?i)(api_?key|api_?token)\s*=\s*[\'"]?([^\s\'\"]+)',
    r'(?i)(secret|token)\s*=\s*[\'"]?([^\s\'\"]+)',
    r'Bearer\s+[A-Za-z0-9\-\._~\+\/]+=*',
]

def detect_secrets_in_code(code: str) -> list:
    """Detect secret patterns in code"""
    found = []
    for pattern in SECRET_PATTERNS:
        matches = re.finditer(pattern, code)
        found.extend([m.group() for m in matches])
    return found
```


## Best Practices

### ✅ DO
- Use `.env.example` templates (no actual values)
- Separate secrets by environment (dev, staging, prod)
- Avoid default values when loading environment variables
- Validate required environment variables at startup
- Log password/token access monitoring
- Use IAM roles (instead of long-term credentials)
- Minimize secret access permissions

### ❌ DON'T
- Output environment variables to logs
- Share `.env` files via email/Slack
- Use production secrets in development environments
- Unencrypted transmission (HTTP instead of HTTPS)
- Hardcoded passwords/keys
- Ignore secret expiration policies
- Skip secret access audit logging


## Works Well With

- `moai-domain-security` (overall security architecture)
- `moai-lang-python` (Python environment variable patterns)
- `moai-domain-devops` (deployment secret management)
- `moai-baas-foundation` (BaaS secret integration)


**Version**: 2.0.0 | **Last Updated**: 2025-11-21 | **Lines**: 180
