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
- 도메인 관련 설계 문서 및 사용자 요구사항.
- 프로젝트 기술 스택 및 운영 제약.

## Outputs
- 도메인 특화 아키텍처 또는 구현 가이드라인.
- 연관 서브 에이전트/스킬 권장 목록.

## Failure Modes
- 도메인 근거 문서가 없거나 모호할 때.
- 프로젝트 전략이 미확정이라 구체화할 수 없을 때.

## Dependencies
- `.moai/project/` 문서와 최신 기술 브리핑이 필요합니다.

## References
- OWASP. "Top 10 Web Application Security Risks." https://owasp.org/www-project-top-ten/ (accessed 2025-03-29).
- NIST. "Secure Software Development Framework." https://csrc.nist.gov/publications/detail/sp/800-218/final (accessed 2025-03-29).

## Changelog
- 2025-03-29: 도메인 스킬에 대한 입력/출력 및 실패 대응을 명문화했습니다.

## Works well with

- alfred-trust-validation (security validation)
- web-api-expert (API security)
- devops-expert (secure deployments)

## Best Practices
- 도메인 결정 사항마다 근거 문서(버전/링크)를 기록합니다.
- 성능·보안·운영 요구사항을 초기 단계에서 동시에 검토하세요.
