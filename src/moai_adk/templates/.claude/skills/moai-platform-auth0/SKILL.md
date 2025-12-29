---
name: moai-platform-auth0
description: Auth0 security specialist covering attack protection, multi-factor authentication, token security, sender constraining, and compliance. Use when implementing Auth0 security features, configuring attack defenses, setting up MFA, or meeting regulatory requirements.
version: 1.0.0
category: security
tags: [auth0, security, mfa, attack-protection, tokens, dpop, mtls, compliance, fapi, gdpr]
updated: 2025-12-24
status: active
allowed-tools: Read, Write, Edit, Grep, Glob, WebFetch, WebSearch, Bash
context7-libraries: /auth0/docs
---

# Auth0 Security Specialist

Comprehensive security skill for Auth0 implementations covering attack protection, multi-factor authentication, token security, sender constraining (DPoP/mTLS), and regulatory compliance (FAPI, GDPR, HIPAA).

## Quick Reference

### Security Feature Categories

Attack Protection:
- Bot Detection: CAPTCHA challenges for suspicious traffic
- Breached Password Detection: Blocks compromised credentials
- Brute Force Protection: Limits failed login attempts per account
- Suspicious IP Throttling: Rate limits high-velocity attacks

Multi-Factor Authentication:
- Push notifications via Auth0 Guardian
- One-time passwords (TOTP)
- WebAuthn with security keys and biometrics
- SMS/voice verification and Adaptive MFA

Token Security:
- JWT structure and validation
- Access token management with scopes
- Refresh token rotation and expiration
- Token revocation strategies

Sender Constraining:
- DPoP: Application-layer token binding
- mTLS: Transport-layer certificate binding

Compliance: FAPI, GDPR, HIPAA/HITECH, PCI DSS, ISO 27001, SOC 2

### Dashboard Navigation

Attack Protection: Dashboard > Security > Attack Protection
MFA Configuration: Dashboard > Security > Multi-factor Auth
Security Center: Dashboard > Security > Security Center

### Essential Setup Checklist

1. Enable Bot Detection with appropriate sensitivity
2. Activate Breached Password Detection
3. Configure Brute Force Protection thresholds
4. Enable Suspicious IP Throttling
5. Set up at least one MFA factor
6. Configure token expiration policies

---

## Implementation Guide

### Attack Protection

Bot Detection: Navigate to Dashboard > Security > Attack Protection > Bot Detection. Configure sensitivity (Low/Medium/High) and response type (Auth Challenge recommended, Simple CAPTCHA, or third-party). IP AllowList supports up to 100 addresses/CIDR ranges.

Supported flows: Universal Login, Classic Login, Lock.js v12.4.0+, native apps. Unsupported: Enterprise connections, social login, cross-origin authentication.

Breached Password Detection: Enable for signup and login. Response actions include blocking compromised credentials and user/admin notifications. Standard Detection has 7-13 months detection time; Credential Guard (Enterprise) reduces to 12-36 hours. Test with passwords starting with AUTH0-TEST-.

Brute Force Protection: Default threshold is 10 failed attempts (configurable 1-100). Protection mechanisms include IP-based blocking and account lockout. Blocks remove after 30 days, password change, admin removal, or user unblock link.

Suspicious IP Throttling: Velocity-based detection for high-volume attacks. Responds with HTTP 429. Configure separate thresholds for login (daily) and signup (per minute) attempts.

For details: modules/attack-protection-overview.md

### Multi-Factor Authentication

Factor Configuration: Navigate to Dashboard > Security > Multi-factor Auth.

Independent Factors (at least one required):
- WebAuthn with FIDO Security Keys
- One-time Password (OTP/TOTP)
- Push Notifications via Auth0 Guardian
- Phone Message (SMS/Voice)
- Cisco Duo Security

Dependent Factors: WebAuthn Biometrics, Email, Recovery codes

MFA Policies: Never, Use Adaptive MFA (Enterprise), Always

WebAuthn: Provides passwordless MFA with security keys or biometrics. Single interaction for multi-factor authentication, phishing-resistant.

Adaptive MFA (Enterprise): Evaluates risk signals per transaction:
- NewDevice: Device not used in past 30 days
- ImpossibleTravel: Geographic anomalies
- UntrustedIP: Suspicious activity history

High-risk transactions require verification regardless of existing MFA sessions.

Step-Up Authentication: Enhanced verification for sensitive operations. APIs use scopes; web apps verify ID token claims.

For details: modules/mfa-overview.md, modules/adaptive-mfa.md

### Token Security

JWT Fundamentals: RFC 7519 standard. Auth0 issues signed JWTs (JWS). Structure includes Header, Payload (claims), and Signature. Always validate signatures, never store sensitive data in payloads, use HTTPS only.

Access Tokens: Authorize API access with scopes. Types: Opaque (require introspection) and JWT (self-contained). Key claims: iss, sub, aud, scope, exp. Default lifetime: 86400 seconds (24 hours).

Refresh Tokens: Enable session continuity. Maximum 200 active per user per application. Security features: Rotation (invalidates predecessor), expiring tokens (idle/absolute), revocation via Management API.

Best Practices:
- Treat signing keys as critical credentials
- Prefer RS256 over HS256 for public key validation
- Store tokens server-side when possible
- Cache and reuse until expiration

For details: modules/tokens-overview.md, modules/token-best-practices.md

### Sender Constraining

DPoP (Application Layer): Binds tokens to client-generated asymmetric key pairs.

Steps: Generate key pair (ES256 recommended), create DPoP Proof JWT, send via DPoP header, include updated proof with each API request.

Proof JWT Structure:
- Header: typ (dpop+jwt), alg, jwk (public key)
- Payload: jti, htm, htu, iat, ath (for API calls)

Public clients must handle use_dpop_nonce errors.

mTLS (Transport Layer): Binds tokens to X.509 certificates.

Process: Client establishes mTLS connection, Auth0 calculates certificate SHA-256 thumbprint, embeds in token cnf claim as x5t#S256. Resource server validates thumbprint.

Requirements: Confidential clients only, Enterprise Plan with HRI add-on, PKI infrastructure.

For details: modules/dpop-implementation.md, modules/mtls-sender-constraining.md

### Compliance

Highly Regulated Identity (Enterprise + HRI add-on):
- Strong Customer Authentication: Minimum two independent factors
- Dynamic Linking: Transaction details in authorization
- PAR: Pushed Authorization Requests
- JAR: JWT-Secured Authorization Requests
- JWE: Access token encryption
- Private Key JWT and mTLS authentication

GDPR Compliance:
- Customer as Data Controller, Auth0 as Data Processor
- User rights: Access, portability (JSON export), erasure, consent management
- Security: Profile encryption, breach detection, brute-force protection

Certifications: ISO 27001/27017/27018, SOC 2 Type 2, CSA STAR, FAPI 1 Advanced OP, HIPAA BAA available, PCI DSS compliant models

For details: modules/highly-regulated-identity.md, modules/gdpr-compliance.md

---

## Advanced Patterns

### Security Center Monitoring

Access from Dashboard > Security > Security Center.

Threat Categories:
- Credential Stuffing: Machine-driven compromise attempts
- Signup Attacks: Automated account creation
- MFA Bypass: Circumvention attempts

Filtering: Time period (up to 14 days), applications, connections. Auto-aggregation by minute/hour/day.

Metrics: Bot detection counts, IP throttling events, brute force triggers, breached password alerts, MFA success/failure rates.

### Application Credentials

Client Secret (Default): Symmetric, simple but vulnerable to interception.

Private Key JWT (Enterprise): Asymmetric key pairs, private key never transmitted, short-lived assertions. Recommended for enhanced security.

mTLS for OAuth (HRI): X.509 certificates, strongest protection.

Key Management: Register up to two public keys for zero-downtime rotation. Algorithms: RS256, RS384, PS256.

### Continuous Session Protection

Use Auth0 Actions for session context during token refresh events.

Capabilities: IP/ASN monitoring, device tracking, expiration management, anomaly detection.

Dynamic management: Customize lifetimes by user attributes, organization, or role.

---

## Module Reference

Attack Protection:
- modules/attack-protection-overview.md
- modules/bot-detection.md
- modules/breached-password-detection.md
- modules/brute-force-protection.md
- modules/suspicious-ip-throttling.md
- modules/akamai-integration.md
- modules/attack-protection-log-events.md
- modules/state-parameters.md

MFA:
- modules/mfa-overview.md
- modules/mfa-factors.md
- modules/webauthn-fido.md
- modules/adaptive-mfa.md
- modules/guardian-configuration.md
- modules/step-up-authentication.md
- modules/mfa-api-management.md
- modules/customize-mfa.md
- modules/ropg-flow-mfa.md

Tokens:
- modules/tokens-overview.md
- modules/jwt-fundamentals.md
- modules/id-tokens.md
- modules/access-tokens.md
- modules/delegation-tokens.md
- modules/refresh-tokens.md
- modules/token-revocation.md
- modules/token-best-practices.md

Sender Constraining:
- modules/dpop-implementation.md
- modules/mtls-sender-constraining.md

Compliance:
- modules/compliance-overview.md
- modules/fapi-implementation.md
- modules/highly-regulated-identity.md
- modules/gdpr-compliance.md
- modules/certifications.md
- modules/tenant-access-control.md
- modules/customer-managed-keys.md

Security Operations:
- modules/security-center.md
- modules/application-credentials.md
- modules/continuous-session-protection.md
- modules/security-guidance.md
- modules/mdl-verification.md

---

## Usage Guide

This skill provides comprehensive Auth0 security guidance. Use it for:
- Attack Protection configuration
- Multi-Factor Authentication setup
- Token security implementation
- Sender constraining (DPoP/mTLS)
- Compliance verification (FAPI, GDPR, HIPAA)

For comprehensive security reviews, use the expert-security agent included in this plugin.

---

## Resources

Official Documentation:
- https://auth0.com/docs/secure
- https://auth0.com/docs/secure/attack-protection
- https://auth0.com/docs/secure/multi-factor-authentication
- https://auth0.com/docs/secure/tokens
- https://auth0.com/docs/secure/sender-constraining
- https://auth0.com/docs/secure/data-privacy-and-compliance
