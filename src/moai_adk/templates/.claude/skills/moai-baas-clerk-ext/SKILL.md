---
name: moai-baas-clerk-ext
version: 2.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: Clerk authentication service integration, modern user management, and session handling. Use when implementing Clerk auth, managing user profiles, or building authentication flows.
keywords: ['clerk', 'authentication', 'user-management', 'sessions', 'modern-auth']
allowed-tools:
  - Read
  - Bash
  - WebFetch
---

# Clerk Extension Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-clerk-ext |
| **Version** | 2.0.0 (2025-11-11) |
| **Allowed tools** | Read, Bash, WebFetch |
| **Auto-load** | On demand when Clerk integration detected |
| **Tier** | BaaS Extension |

---

## What It Does

Clerk authentication service integration, modern user management, and session handling.

**Key capabilities**:
- ✅ Clerk SDK integration
- ✅ Modern auth components
- ✅ User profile management
- ✅ Session handling
- ✅ Social authentication

---

## When to Use

- ✅ Implementing Clerk authentication
- ✅ Building user profile systems
- ✅ Managing user sessions
- ✅ Creating authentication UI

---

## Core Clerk Patterns

### Integration Approach
1. **Clerk Provider**: Wrap application with ClerkProvider
2. **Auth Components**: Pre-built sign-up/sign-in components
3. **User Hooks**: useUser, useSession, useAuth
4. **Route Protection**: Middleware and protected routes
5. **User Management**: Profile customization and preferences

### Modern Auth Features
- **Passwordless**: Email/SMS magic links
- **Social Providers**: Google, GitHub, Discord, etc.
- **Multi-tenant**: Organization management
- **Webhooks**: Real-time user events
- **Custom Flows**: Tailored authentication experiences

---

## Dependencies

- Clerk account and API keys
- Clerk SDK for your framework
- React/Next.js for components
- Webhook endpoints for events

---

## Works Well With

- `moai-baas-foundation` (BaaS patterns)
- `moai-domain-frontend` (Frontend integration)
- `moai-domain-security` (Security best practices)

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, modern auth patterns
- **v1.0.0** (2025-10-22): Initial Clerk integration

---

**End of Skill** | Updated 2025-11-11
