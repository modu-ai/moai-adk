---
name: moai-domain-security
description: OWASP Top 10, static analysis (SAST), dependency security, and secrets
  management
allowed-tools:
- Read
- Bash
---

# Security Expert

## What it does

Provides expertise in application security, including OWASP Top 10 vulnerabilities, static application security testing (SAST), dependency vulnerability scanning, and secrets management.

## When to use

- "보안 취약점 분석", "OWASP 검증", "시크릿 관리", "의존성 보안"
- Automatically invoked when security concerns arise
- Security SPEC implementation (`/alfred:2-build`)

## How it works

**OWASP Top 10 (2021)**:
1. **Broken Access Control**: Authorization checks
2. **Cryptographic Failures**: Encryption at rest/transit
3. **Injection**: SQL injection, XSS prevention
4. **Insecure Design**: Threat modeling
5. **Security Misconfiguration**: Secure defaults
6. **Vulnerable Components**: Dependency scanning
7. **Identification/Authentication Failures**: MFA, password policies
8. **Software/Data Integrity Failures**: Code signing
9. **Security Logging/Monitoring Failures**: Audit logs
10. **Server-Side Request Forgery (SSRF)**: Input validation

**Static Analysis (SAST)**:
- **Semgrep**: Multi-language static analysis
- **SonarQube**: Code quality + security
- **Bandit**: Python security linter
- **ESLint security plugins**: JavaScript security

**Dependency Security**:
- **Snyk**: Vulnerability scanning
- **Dependabot**: Automated dependency updates
- **npm audit**: Node.js vulnerabilities
- **safety**: Python dependency checker

**Secrets Management**:
- **Never commit secrets**: .gitignore for .env files
- **Vault**: Secrets storage (HashiCorp Vault)
- **Environment variables**: Runtime configuration
- **Secret scanning**: git-secrets, trufflehog

**Secure Coding Practices**:
- Input validation and sanitization
- Parameterized queries (SQL injection prevention)
- CSP (Content Security Policy) headers
- HTTPS enforcement

## Examples

### Example 1: OWASP compliance check
User: "/alfred:2-build SEC-001"
Claude: (creates RED security test, GREEN implementation with input validation, REFACTOR)

### Example 2: Dependency vulnerability scan
User: "의존성 보안 스캔"
Claude: (runs npm audit or snyk test and reports vulnerabilities)

## Works well with

- alfred-trust-validation (security validation)
- web-api-expert (API security)
- devops-expert (secure deployments)
