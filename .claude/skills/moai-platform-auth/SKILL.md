---
name: moai-platform-auth
description: >
  Authentication and authorization specialist covering Auth0, Clerk, and Firebase Auth.
  Use when implementing authentication, authorization, MFA, SSO, passkeys, WebAuthn,
  social login, or security features. Supports enterprise (Auth0), modern UX (Clerk),
  and mobile-first (Firebase) patterns.
license: MIT
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Grep Glob Bash(npm:*) Bash(npx:*) Bash(firebase:*) Bash(curl:*) WebFetch WebSearch mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.0.0"
  category: "platform"
  status: "active"
  updated: "2026-02-09"
  modularized: "false"
  platforms: "Auth0, Clerk, Firebase Auth"
  tags: "auth0, clerk, firebase, authentication, authorization, mfa, sso, passkeys, webauthn, social-login, security"
  context7-libraries: "/auth0/docs, /clerk/clerk-docs, /firebase/firebase-docs"
  related-skills: "moai-platform-supabase, moai-platform-vercel, moai-lang-typescript, moai-domain-backend, moai-expert-security"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4500

# MoAI Extension: Triggers
triggers:
  keywords: ["auth0", "clerk", "firebase auth", "authentication", "authorization", "mfa", "sso", "passkeys", "webauthn", "social login", "user management", "attack protection", "auth ui", "passwordless", "oauth", "identity", "jwt", "token security"]
  agents: ["expert-backend", "expert-security", "expert-frontend"]
  phases: ["run"]
---

# Authentication Platform Specialist

Comprehensive authentication and authorization guidance covering three major platforms: Auth0 (enterprise security), Clerk (modern UX), and Firebase Auth (mobile-first).

## Quick Platform Selection

### Auth0 - Enterprise Security

Enterprise-grade identity platform focused on security compliance and attack protection.

Best For: Enterprise applications requiring strong compliance (FAPI, GDPR, HIPAA), sophisticated attack protection, token security with sender constraining (DPoP/mTLS), multi-tenant B2B SaaS.

Key Strengths: Advanced attack protection (bot detection, breached passwords, brute force), adaptive MFA, compliance certifications (ISO 27001, SOC 2, FAPI), token security (DPoP, mTLS), extensive security monitoring.

Cost Model: Priced per monthly active user with enterprise features at higher tiers.

Context7 Library: /auth0/docs

### Clerk - Modern User Experience

Modern authentication with beautiful pre-built UI components and WebAuthn support.

Best For: Modern web applications prioritizing developer experience and user experience, Next.js applications, applications requiring social login with minimal setup, passwordless authentication.

Key Strengths: Drop-in React components with beautiful UI, WebAuthn and passkeys support, seamless Next.js integration, organization management, simple API with excellent DX.

Cost Model: Free tier available, priced per monthly active user with generous limits.

Context7 Library: /clerk/clerk-docs

### Firebase Auth - Mobile-First Integration

Google ecosystem authentication with seamless Firebase services integration.

Best For: Mobile applications (iOS, Android, Flutter), Google ecosystem integration, serverless Cloud Functions, applications requiring anonymous auth with upgrade path, small to medium web applications.

Key Strengths: Native mobile SDKs for iOS/Android/Flutter, Google Sign-In integration, Firebase services integration (Firestore, Storage, Cloud Functions), phone authentication, free tier with generous limits.

Cost Model: Free tier with generous limits, pay-as-you-go for higher volumes.

Context7 Library: /firebase/firebase-docs

## Quick Decision Guide

Choose Auth0 when:
- Enterprise security and compliance requirements are critical
- Need sophisticated attack protection and security monitoring
- Implementing sender-constrained tokens (DPoP, mTLS)
- Supporting complex B2B multi-tenant scenarios
- FAPI, GDPR, HIPAA, or PCI DSS compliance required

Choose Clerk when:
- Building modern Next.js or React applications
- Developer experience and beautiful UI are priorities
- Need passwordless or WebAuthn authentication quickly
- Want minimal authentication code in your application
- Organization management with role-based access

Choose Firebase Auth when:
- Building mobile-first applications
- Already using Firebase ecosystem (Firestore, Storage, Functions)
- Need Google Sign-In or Google ecosystem integration
- Want anonymous authentication with upgrade path
- Prefer serverless architecture with Cloud Functions

## Common Authentication Patterns

### Universal Patterns

These patterns apply across all three platforms with platform-specific implementations.

**Session Management:**

All platforms support session persistence, refresh tokens, and session invalidation. Auth0 uses refresh token rotation, Clerk uses session tokens with automatic refresh, Firebase uses ID token refresh with custom claims.

**Multi-Factor Authentication:**

All platforms support multiple MFA factors including TOTP, SMS, and push notifications. Auth0 provides WebAuthn and adaptive MFA, Clerk provides WebAuthn with passkeys, Firebase provides phone verification and custom MFA.

**Social Authentication:**

All platforms support major social providers (Google, Facebook, GitHub, Apple). Auth0 requires connection configuration per provider, Clerk provides pre-configured social login buttons, Firebase requires OAuth configuration and SDK setup.

**Role-Based Access Control:**

All platforms support custom claims or metadata for authorization. Auth0 uses custom claims in JWT tokens with Actions, Clerk uses organization roles and metadata, Firebase uses custom claims with Admin SDK.

**Token Management:**

All platforms issue JWT tokens for API authorization. Auth0 provides access tokens with scopes and refresh tokens, Clerk provides session tokens via getToken(), Firebase provides ID tokens with custom claims.

### Security Best Practices

Applicable to all platforms:

**Token Storage:**
- Never store tokens in localStorage on web (XSS vulnerability)
- Use httpOnly cookies when possible
- For SPAs, use memory storage with refresh token rotation
- Mobile apps use secure storage (Keychain, Keystore)

**HTTPS Enforcement:**
- Always use HTTPS in production
- Configure secure redirect URIs
- Enable HSTS headers

**Token Validation:**
- Always validate token signatures
- Verify token audience (aud claim)
- Check token expiration (exp claim)
- Validate issuer (iss claim)

**Password Policies:**
- Enforce strong password requirements
- Enable breached password detection
- Implement account lockout after failed attempts
- Use password strength indicators

**API Security:**
- Require authentication for all protected endpoints
- Implement rate limiting
- Use scopes or permissions for authorization
- Log authentication and authorization events

## Platform-Specific Implementation

For detailed platform-specific implementation guidance, see the reference files:

### Auth0 Implementation

File: reference/auth0.md

Covers attack protection configuration, MFA setup with WebAuthn and adaptive policies, token security with DPoP and mTLS sender constraining, compliance features for FAPI/GDPR/HIPAA, Security Center monitoring, and continuous session protection.

Key sections: Dashboard navigation, bot detection configuration, breached password detection, brute force protection, WebAuthn setup, token validation, DPoP implementation, mTLS certificate binding, compliance certifications.

### Clerk Implementation

File: reference/clerk.md

Covers ClerkProvider setup for Next.js, authentication components (SignIn, SignUp, UserButton), route protection with middleware, useAuth and useUser hooks, server-side authentication, organization management, and Core 2 migration.

Key sections: Environment variables, middleware configuration, protecting routes, accessing user data, organization switching, custom authentication flows, webhook integration.

### Firebase Auth Implementation

File: reference/firebase-auth.md

Covers Firebase SDK initialization across platforms (Web, Flutter, iOS, Android), social authentication setup, phone authentication with SMS verification, anonymous auth with account linking, custom claims for RBAC, and Security Rules integration.

Key sections: Project setup, SDK initialization, Google Sign-In, Facebook Login, phone verification, custom claims management, Firestore and Storage rules, Cloud Functions triggers.

### Platform Comparison

File: reference/comparison.md

Provides detailed comparison matrix covering features, pricing, use cases, migration considerations, and integration complexity.

Key sections: Feature comparison table, pricing breakdown, use case decision matrix, platform migration strategies, ecosystem integration, developer experience comparison.

## Navigation Guide

When working with authentication features:

1. Start with Quick Platform Selection (above) if choosing a platform
2. Review Common Authentication Patterns for universal concepts
3. Open platform-specific reference file for implementation details
4. Refer to comparison.md when evaluating multiple platforms
5. Use Context7 tools to access latest platform documentation

## Context7 Documentation Access

Access up-to-date platform documentation using Context7 MCP:

**Auth0:**
- Use resolve-library-id with "auth0" to get library ID
- Use get-library-docs with topic "attack-protection", "mfa", "tokens", "compliance"

**Clerk:**
- Use resolve-library-id with "clerk" to get library ID
- Use get-library-docs with topic "nextjs", "react", "authentication"

**Firebase Auth:**
- Use resolve-library-id with "firebase" to get library ID
- Use get-library-docs with topic "authentication", "security-rules"

## Works Well With

- moai-platform-supabase: Database with auth integration
- moai-platform-vercel: Deployment with edge authentication
- moai-lang-typescript: TypeScript patterns for auth SDKs
- moai-domain-backend: Backend architecture with authentication
- moai-domain-frontend: React/Next.js frontend integration
- moai-expert-security: Security audit and threat modeling

---

Status: Active
Version: 2.0.0 (Consolidated Platform Coverage)
Last Updated: 2026-02-09
Platforms: Auth0, Clerk, Firebase Auth
