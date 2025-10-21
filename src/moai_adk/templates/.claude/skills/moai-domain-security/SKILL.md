---

name: moai-domain-security
description: OWASP Top 10, static analysis (SAST), dependency security, and secrets management. Use when working on security and compliance reviews scenarios.
allowed-tools:
  - Read
  - Bash
---

# Security Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when security keywords appear |
| Trigger cues | Threat modeling, OWASP findings, secrets management, compliance reviews. |
| Tier | 4 |

## What it does

Provides expertise in application security, including OWASP Top 10 vulnerabilities, static application security testing (SAST), dependency vulnerability scanning, and secrets management.

## When to use

- Engages when the team asks about security posture or mitigation steps.
- “Security vulnerability analysis”, “OWASP verification”, “Secret management”, “Dependency security”
- Automatically invoked when security concerns arise
- Security SPEC implementation (`/alfred:2-run`)

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
```markdown
- Run SAST/DAST tools and attach findings summary.
- Update risk matrix with severity/owner/ETA.
```

## Inputs
- Domain-related design documents and user requirements.
- Project technology stack and operational constraints.

## Outputs
- Domain-specific architecture or implementation guidelines.
- Recommended list of associated sub-agents/skills.

## Failure Modes
- When the domain document does not exist or is ambiguous.
- When the project strategy is unconfirmed and cannot be specified.

## Dependencies
- `.moai/project/` document and latest technical briefing are required.

## References
- OWASP. "Top 10 Web Application Security Risks." https://owasp.org/www-project-top-ten/ (accessed 2025-03-29).
- NIST. "Secure Software Development Framework." https://csrc.nist.gov/publications/detail/sp/800-218/final (accessed 2025-03-29).

## Changelog
- 2025-03-29: Codified input/output and failure responses for domain skills.

## Works well with

- alfred-trust-validation (security validation)
- web-api-expert (API security)
- devops-expert (secure deployments)

## Best Practices
- Record supporting documentation (version/link) for each domain decision.
- Review performance, security, and operational requirements simultaneously at an early stage.
