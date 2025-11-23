---
name: moai-security-authentication
description: Authentication and identity management (JWT, OAuth2, SSO, MFA, RBAC)
version: 1.0.0
modularized: true
tags:
  - enterprise
  - unified
  - development
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**Name**: moai-security-authentication
**Domain**: Unified Capability
**Freedom Level**: high
**Target Users**: Developers, engineers, technical teams
**Invocation**: Skill("moai-security-authentication")
**Progressive Disclosure**: SKILL.md (core) → modules/ (detailed patterns)
**Last Updated**: 2025-11-24
**Modularized**: true
**Replaces**: moai-security-auth, moai-security-identity

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Authentication and identity management (JWT, OAuth2, SSO, MFA, RBAC)

**Core Capabilities**:
- **JWT Authentication**: Token-based auth with JSON Web Tokens
- **OAuth2 Flows**: Authorization code, client credentials, PKCE
- **Multi-Factor Authentication**: TOTP, SMS, biometric MFA
- **Single Sign-On (SSO)**: SAML, OpenID Connect integration
- **Role-Based Access Control**: Permissions and authorization

**When to Use**:
- Implementing jwt authentication
- Working with Authentication and identity management (JWT
- Enterprise-grade security

---

## 📚 Core Patterns (5-10 minutes each)


### Pattern 1: JWT Authentication

**Concept**: Token-based auth with JSON Web Tokens

**Implementation**: [Detailed implementation patterns in modules/]

---

### Pattern 2: OAuth2 Flows

**Concept**: Authorization code, client credentials, PKCE

**Implementation**: [Detailed implementation patterns in modules/]

---

### Pattern 3: Multi-Factor Authentication

**Concept**: TOTP, SMS, biometric MFA

**Implementation**: [Detailed implementation patterns in modules/]

---

### Pattern 4: Single Sign-On (SSO)

**Concept**: SAML, OpenID Connect integration

**Implementation**: [Detailed implementation patterns in modules/]

---

### Pattern 5: Role-Based Access Control

**Concept**: Permissions and authorization

**Implementation**: [Detailed implementation patterns in modules/]

---


## 📖 Advanced Documentation

For detailed patterns and implementation strategies:

- **modules/jwt-authentication.md** - Token-based auth with JSON Web Tokens
- **modules/oauth2-flows.md** - Authorization code, client credentials, PKCE
- **modules/multi-factor-authentication.md** - TOTP, SMS, biometric MFA
- **modules/single-sign-on-(sso).md** - SAML, OpenID Connect integration
- **modules/role-based-access-control.md** - Permissions and authorization

---

## ✅ Best Practices

### DO
- ✅ Follow established patterns and conventions
- ✅ Use latest best practices from Context7 integration
- ✅ Implement comprehensive error handling
- ✅ Document all configurations and decisions
- ✅ Test implementations thoroughly

### DON'T
- ❌ Skip validation and testing steps
- ❌ Ignore security best practices
- ❌ Use deprecated patterns or libraries
- ❌ Hardcode configuration values
- ❌ Neglect documentation updates

---

## 🔗 Works Well With

- `moai-context7-integration` - Latest patterns and best practices
- `moai-foundation-trust` - TRUST 5 quality framework
- `moai-core-personas` - Adaptive communication

---

## 📈 Integration Workflow

**Typical Workflow**:
```
1. Initialize (Pattern 1)
   ↓
2. Configure (Pattern 2)
   ↓
3. Implement (Pattern 3)
   ↓
4. Validate (Pattern 4)
   ↓
5. Deploy (Pattern 5)
```

---

## 📊 Success Metrics

- **Implementation Quality**: ≥90% code quality score
- **Test Coverage**: ≥85% coverage
- **Security Compliance**: 100% OWASP compliance
- **Documentation Coverage**: 100% API documentation
- **Performance**: Meets defined SLAs

---

## 🔄 Changelog

- **v1.0.0** (2025-11-24): Initial unified skill combining 2 source skills

---

**Status**: Production Ready (Enterprise)
**Generated with**: MoAI-ADK Skill Factory
**Modular Architecture**: SKILL.md + 5 pattern modules
**Tier**: Tier 3: Security
