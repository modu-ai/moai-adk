---
name: moai-core-env-security
description: Environment variable security, secret management, and credential protection patterns
allowed-tools: [Read, Bash]
---

# Environment Variable Security

## Quick Reference

Secure environment variable management using `.env` files, secret managers (AWS Secrets Manager, HashiCorp Vault), and 12-factor app principles to prevent credential leaks in version control and logs.

**Security Principles**:
- Never commit `.env` files to version control
- Use secret managers for production credentials
- Rotate secrets regularly (30-90 day cycles)
- Implement least-privilege access (IAM policies)

**Tools** (November 2025):
- `python-dotenv`: Python .env loading
- `AWS Secrets Manager`: Cloud secret storage
- `HashiCorp Vault`: Enterprise secret management
- `git-secrets`: Pre-commit credential scanning

---

## Implementation Guide

**.gitignore Protection**:
```
.env*
!.env.example
secrets.yaml
credentials.json
*.pem
*.key
```

**AWS Secrets Manager**:
```python
import boto3
client = boto3.client('secretsmanager')
secret = client.get_secret_value(SecretId='prod/db')
```

**Environment Validation**:
```python
import os
required_vars = ['DATABASE_URL', 'API_KEY', 'JWT_SECRET']
missing = [v for v in required_vars if not os.getenv(v)]
if missing:
    raise ValueError(f"Missing: {missing}")
```

---

## Best Practices

### ✅ DO
- Use `.env.example` templates (no real values)
- Implement secret rotation policies
- Enable audit logging for secret access
- Use IAM roles instead of long-term credentials

### ❌ DON'T
- Log environment variables (PII leaks)
- Share `.env` files via email/Slack
- Use production secrets in development
- Skip encryption for secrets at rest

---

**Version**: 1.0.0 | **Last Updated**: 2025-11-21
