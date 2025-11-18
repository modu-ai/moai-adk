---
name: moai-core-env-security
version: 4.0.0
created: 2025-11-18
updated: '2025-11-18'
status: stable
description: Environment variable security, secrets management, and secure credential
  handling for MoAI-ADK projects
allowed-tools:
- Read
- Bash
- Write
stability: stable
---


# Environment Security & Secrets Management - 

**Secure environment variable and credentials management patterns for production systems**

> **Scope**: MoAI-ADK Security Infrastructure
> **Framework**: dotenv, environment isolation, secrets vault integration
> **Keywords**: env-security, secrets-management, credentials, dotenv, vault

## Level 1: Quick Reference

### Core Principles

- **Never commit secrets**: .env files in .gitignore
- **Environment separation**: dev, test, production isolation
- **Credential rotation**: Regular key updates
- **Access control**: Principle of least privilege
- **Audit logging**: Track secret access

### Environment Management

```bash
# Development
.env.local (gitignored, local overrides)
.env (defaults, can be in git with defaults only)
.env.example (template, safe to commit)

# Production
.env.production (via deployment platform)
Secret vault integration (AWS Secrets Manager, Vault, etc)
```

---

## Level 2: Implementation Patterns

### .env File Structure

```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost/db
DATABASE_POOL_SIZE=20

# API Keys
STRIPE_API_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Feature Flags
FEATURE_NEW_CHECKOUT=true
DEBUG_MODE=false
```

### Secret Rotation

- Automated rotation policies
- Deprecation warnings
- Migration guides
- Minimal downtime strategies

---

## Level 3: Enterprise Security

### Vault Integration

- HashiCorp Vault for centralized secrets
- Dynamic credentials generation
- Audit trail and compliance
- Multi-environment management

### Compliance

- OWASP guidelines
- SOC 2, ISO 27001 requirements
- GDPR data handling
- PCI-DSS for payment data

---

## References

- **12-Factor App**: https://12factor.net/config
- **OWASP Secrets Management**: https://cheatsheetseries.owasp.org/

---

**Last Updated**: 2025-11-18
**Format**: Markdown | **Language**: English
**Status**: Stable
**Version**: 4.0.0
