---
name: moai-baas-auth0-ext
version: 2.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: Auth0 authentication service integration patterns, JWT token management, and security best practices. Use when implementing Auth0 authentication, managing user sessions, or securing APIs.
keywords: ['auth0', 'authentication', 'jwt', 'security', 'identity']
allowed-tools:
  - Read
  - Bash
  - WebFetch
---

# Auth0 Extension Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-auth0-ext |
| **Version** | 2.0.0 (2025-11-11) |
| **Allowed tools** | Read, Bash, WebFetch |
| **Auto-load** | On demand when Auth0 integration detected |
| **Tier** | BaaS Extension |

---

## What It Does

Auth0 authentication service integration patterns, JWT token management, and security best practices.

**Key capabilities**:
- ✅ Auth0 SDK integration
- ✅ JWT token handling
- ✅ User session management
- ✅ Security best practices
- ✅ Multi-provider authentication

---

## When to Use

- ✅ Implementing Auth0 authentication
- ✅ Managing user sessions
- ✅ Securing APIs with Auth0
- ✅ Setting up social logins

---

## Core Auth0 Patterns

### Authentication Flow
1. **Universal Login**: Auth0-hosted login page
2. **SDK Integration**: React, Vue, Angular, Node.js
3. **Token Management**: Access tokens, refresh tokens
4. **API Protection**: JWT validation middleware
5. **User Management**: Profile synchronization

### Security Configuration
- **Password Policies**: Strong password requirements
- **MFA**: Multi-factor authentication
- **Rate Limiting**: Prevent brute force attacks
- **CORS**: Cross-origin resource sharing
- **HTTPS**: TLS encryption enforcement

---

## Dependencies

- Auth0 account and tenant
- Auth0 SDK for your platform
- Secure storage for tokens
- CORS configuration

---

## Works Well With

- `moai-baas-foundation` (BaaS patterns)
- `moai-domain-security` (Security practices)
- `moai-domain-backend` (API security)

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, security patterns, JWT handling
- **v1.0.0** (2025-10-22): Initial Auth0 integration

---

**End of Skill** | Updated 2025-11-11
