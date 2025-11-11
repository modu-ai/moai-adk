---
name: moai-baas-railway-ext
version: 2.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: Railway deployment platform integration, container orchestration, and production environment management. Use when deploying applications to Railway, managing microservices, or implementing CI/CD pipelines.
keywords: ['railway', 'deployment', 'containers', 'microservices', 'production']
allowed-tools:
  - Read
  - Bash
  - WebFetch
---

# Railway Extension Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-railway-ext |
| **Version** | 2.0.0 (2025-11-11) |
| **Allowed tools** | Read, Bash, WebFetch |
| **Auto-load** | On demand when Railway deployment detected |
| **Tier** | BaaS Extension |

---

## What It Does

Railway deployment platform integration, container orchestration, and production environment management.

**Key capabilities**:
- ✅ Container-based deployment
- ✅ Environment management
- ✅ CI/CD pipeline integration
- ✅ Microservices orchestration
- ✅ Production monitoring

---

## When to Use

- ✅ Deploying applications to Railway
- ✅ Managing microservices
- ✅ Implementing CI/CD pipelines
- ✅ Production environment setup

---

## Core Railway Patterns

### Deployment Architecture
1. **Container Builder**: Docker container creation
2. **Environment Variables**: Configuration management
3. **Service Discovery**: Inter-service communication
4. **Health Checks**: Service monitoring
5. **Auto-scaling**: Resource optimization

### Environment Management
- **Production**: Live production environment
- **Staging**: Pre-production testing
- **Development**: Feature development
- **Preview**: Pull request environments
- **Branch Deployments**: Automatic deployments per branch

---

## Dependencies

- Railway account and project
- Docker containerization
- Environment configuration
- CI/CD pipeline setup

---

## Works Well With

- `moai-baas-foundation` (BaaS patterns)
- `moai-domain-devops` (DevOps patterns)
- `moai-domain-backend` (Backend deployment)

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, deployment patterns
- **v1.0.0** (2025-10-22): Initial Railway integration

---

**End of Skill** | Updated 2025-11-11
