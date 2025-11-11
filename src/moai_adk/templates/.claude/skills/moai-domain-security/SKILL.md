---
name: moai-domain-security
version: 2.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: Security best practices, vulnerability prevention, and application security patterns. Use when implementing authentication, securing APIs, or conducting security audits.
keywords: ['security', 'authentication', 'vulnerability-prevention', 'encryption', 'audit']
allowed-tools:
  - Read
  - Bash
  - WebFetch
---

# Security Development Domain Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-domain-security |
| **Version** | 2.0.0 (2025-11-11) |
| **Allowed tools** | Read, Bash, WebFetch |
| **Auto-load** | On demand when security patterns detected |
| **Tier** | Domain (Expert) |

---

## What It Does

Security best practices, vulnerability prevention, and application security patterns.

**Key capabilities**:
- ✅ Authentication and authorization
- ✅ Data encryption and protection
- ✅ Security vulnerability prevention
- ✅ Security audit procedures
- ✅ Compliance frameworks

---

## When to Use

- ✅ Implementing authentication systems
- ✅ Securing application APIs
- ✅ Conducting security audits
- ✅ Planning security architecture

---

## Core Security Patterns

### Authentication & Authorization
1. **JWT Implementation**: Secure token-based auth
2. **OAuth 2.0**: Third-party authentication
3. **RBAC**: Role-based access control
4. **Session Management**: Secure session handling
5. **Multi-Factor Auth**: Enhanced security layers

### Data Protection
- **Encryption**: Data at rest and in transit
- **Hashing**: Secure password storage
- **Input Validation**: Prevent injection attacks
- **CORS Configuration**: Cross-origin security
- **Security Headers**: HTTP security implementations

---

## Dependencies

- Authentication libraries
- Encryption tools
- Security scanning frameworks
- Compliance documentation

---

## Works Well With

- `moai-domain-backend` (API security)
- `moai-baas-auth0-ext` (Auth0 integration)
- `moai-essentials-debug` (Security debugging)

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, security patterns
- **v1.0.0** (2025-10-22): Initial security domain

---

**End of Skill** | Updated 2025-11-11
