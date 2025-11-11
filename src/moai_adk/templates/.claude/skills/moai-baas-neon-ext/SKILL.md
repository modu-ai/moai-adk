---
name: moai-baas-neon-ext
version: 2.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: Neon serverless PostgreSQL database integration, branching workflows, and connection management. Use when implementing serverless databases, database branching for development, or auto-scaling PostgreSQL instances.
keywords: ['neon', 'postgresql', 'serverless', 'database-branching', 'auto-scaling']
allowed-tools:
  - Read
  - Bash
  - WebFetch
---

# Neon Extension Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-neon-ext |
| **Version** | 2.0.0 (2025-11-11) |
| **Allowed tools** | Read, Bash, WebFetch |
| **Auto-load** | On demand when Neon integration detected |
| **Tier** | BaaS Extension |

---

## What It Does

Neon serverless PostgreSQL database integration, branching workflows, and connection management.

**Key capabilities**:
- ✅ Serverless PostgreSQL
- ✅ Database branching
- ✅ Auto-scaling instances
- ✅ Connection pooling
- ✅ Development workflows

---

## When to Use

- ✅ Implementing serverless databases
- ✅ Database branching for development
- ✅ Auto-scaling PostgreSQL instances
- ✅ Creating development workflows

---

## Core Neon Patterns

### Serverless Architecture
1. **Auto-scaling**: Scale based on demand
2. **Branching**: Git-like database versioning
3. **Connection Pooling**: Efficient connection management
4. **Multi-region**: Global database deployment
5. **Backup & Restore**: Point-in-time recovery

### Development Workflow
- **Feature Branches**: Isolated development databases
- **CI/CD Integration**: Automated testing environments
- **Preview Databases**: Temporary instances for PRs
- **Production Promotion**: Seamless branch merging
- **Rollback**: Quick database reversion

---

## Dependencies

- Neon account and project
- PostgreSQL client libraries
- Connection string management
- Branching workflow tools

---

## Works Well With

- `moai-baas-foundation` (BaaS patterns)
- `moai-domain-database` (Database patterns)
- `moai-cc-skills` (Development workflows)

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, serverless patterns
- **v1.0.0** (2025-10-22): Initial Neon integration

---

**End of Skill** | Updated 2025-11-11
