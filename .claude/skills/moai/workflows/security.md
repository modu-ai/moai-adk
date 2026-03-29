---
name: moai-workflow-security
description: >
  Dedicated OWASP security audit with SQL injection, XSS, CSRF, secrets scan,
  authentication checks, and dependency vulnerability analysis. Deeper than
  review --security. Use before deployment or when auth/database/API changes are made.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-03-29"
  tags: "security, audit, owasp, injection, xss, csrf, secrets, vulnerability"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["security", "audit", "owasp", "vulnerability", "injection", "xss", "csrf", "secrets"]
  agents: ["expert-security", "manager-quality"]
  phases: ["security"]
---

# Security Workflow Orchestration

## Purpose

Dedicated OWASP security audit providing deeper analysis than `/moai review --security`. Covers injection attacks, authentication flaws, secrets exposure, dependency vulnerabilities, and data isolation.

## Difference from Review --security

| Aspect | `/moai review --security` | `/moai security` |
|--------|--------------------------|-------------------|
| Scope | Security perspective within multi-perspective review | Dedicated full security audit |
| Depth | Surface-level OWASP check | Deep analysis with dependency scanning |
| Dependencies | Skipped | Full vulnerability scan (go vuln, npm audit, etc.) |
| Secrets scan | Basic pattern match | Comprehensive (env files, config, git history) |
| Data isolation | Not checked | Cross-tenant, permission boundary analysis |
| Output | Section in review report | Standalone security report with CVSS-like scoring |

## Input

- $ARGUMENTS: Optional flags and scope
  - --full: Scan entire codebase (default: changed files only)
  - --deps: Focus on dependency vulnerability analysis
  - --secrets: Focus on secrets and credential exposure
  - --file PATH: Audit specific file(s) only
  - --branch BRANCH: Compare against branch for differential audit

## Phase 1: Scope Detection

Determine audit scope:

If --full: All source files in project
If --branch: Files changed between current and target branch
If --file: Specified file(s) only
If no flag: Files changed since last commit (`git diff HEAD~1`)

Collect:
- List of files to audit with their categories (auth, database, API, config, frontend)
- Technology stack detection (frameworks, ORMs, template engines)
- Existing security configurations (.env.example, CORS config, auth middleware)

## Phase 2: OWASP Top 10 Analysis

Agent: expert-security subagent

Analyze each OWASP category with language-specific patterns:

### A01: Broken Access Control
- Missing authentication checks on endpoints
- Insecure direct object references (IDOR)
- CORS misconfiguration
- Missing rate limiting
- Privilege escalation paths

### A02: Cryptographic Failures
- Weak hashing algorithms (MD5, SHA1 for passwords)
- Hardcoded encryption keys
- Missing TLS enforcement
- Insecure random number generation
- Cleartext storage of sensitive data

### A03: Injection
- SQL injection (raw queries, string concatenation)
- Command injection (exec, system calls with user input)
- XSS (unescaped output in templates, innerHTML)
- LDAP injection, XML injection, NoSQL injection
- Path traversal

### A04: Insecure Design
- Missing input validation at trust boundaries
- Business logic flaws (race conditions in transactions)
- Missing anti-automation controls
- Insecure API design patterns

### A05: Security Misconfiguration
- Default credentials or configurations
- Unnecessary features enabled (debug mode, verbose errors)
- Missing security headers (CSP, X-Frame-Options, HSTS)
- Overly permissive CORS or file permissions

### A06: Vulnerable Components
- Known CVEs in dependencies
- Outdated packages with security patches
- Unmaintained dependencies

Language-specific dependency scanning:
- Go: `govulncheck ./...`
- Python: `pip audit` or `safety check`
- Node.js: `npm audit` or `yarn audit`
- Rust: `cargo audit`
- Ruby: `bundle audit`
- Java: `mvn dependency-check:check`
- PHP: `composer audit`

### A07: Authentication Failures
- Weak password policies
- Missing MFA support
- Insecure session management
- JWT vulnerabilities (none algorithm, weak signing)
- Missing account lockout

### A08: Data Integrity Failures
- Missing integrity checks on updates
- Insecure deserialization
- Missing CSRF protection on state-changing operations
- Unsigned or unverified software updates

### A09: Logging and Monitoring Failures
- Missing audit logging for security events
- Sensitive data in logs (passwords, tokens, PII)
- Missing error handling that could leak information
- No alerting for suspicious activities

### A10: Server-Side Request Forgery (SSRF)
- Unvalidated URLs in fetch/request operations
- Missing URL allowlist for external calls
- Internal service exposure through user-controlled URLs

## Phase 3: Secrets Scan

Scan for exposed secrets and credentials:

### Pattern Detection
- API keys, tokens, passwords in source code
- .env files with sensitive values committed
- Configuration files with embedded credentials
- Private keys or certificates in repository

### Git History Scan (when --secrets or --full)
- Check recent commits for accidentally committed secrets
- `git log --all --diff-filter=A -- '*.env' '*.key' '*.pem'`
- Pattern match in git diff history for common secret patterns

### Secret Patterns by Language
- Go: Hardcoded strings matching `password|secret|key|token` in variable assignments
- Python: Django SECRET_KEY, database passwords in settings.py
- Node.js: .env values referenced but .env not in .gitignore
- General: Base64-encoded credentials, JWT tokens, AWS keys (AKIA*), SSH keys

## Phase 4: Data Isolation Analysis

Analyze data access boundaries:

- Multi-tenant data isolation: Are tenant IDs consistently enforced in queries?
- Permission boundaries: Are authorization checks applied before data access?
- API endpoint protection: Do all mutation endpoints require authentication?
- Database query safety: Are parameterized queries used consistently?
- File access control: Are uploaded files properly sandboxed?

## Phase 5: Security Report

### Report Structure

```markdown
## Security Audit Report

### Executive Summary
- Risk Level: CRITICAL / HIGH / MEDIUM / LOW
- Total Findings: N
- Critical: N | High: N | Medium: N | Low: N

### OWASP Findings

#### CRITICAL
- [A03:Injection] src/db/queries.go:45 — SQL injection via string concatenation
  Impact: Full database read/write access
  Fix: Use parameterized queries

#### HIGH
- [A01:Access] src/api/handler.go:23 — Missing auth check on DELETE endpoint
  Impact: Unauthorized data deletion
  Fix: Add authentication middleware

#### MEDIUM
- [A05:Config] src/server.go:12 — Debug mode enabled in production config
  Impact: Information disclosure
  Fix: Set DEBUG=false in production environment

#### LOW
- [A09:Logging] src/auth/login.go:67 — Password attempt count not logged
  Impact: Reduced incident response capability
  Fix: Add audit logging for failed login attempts

### Dependency Vulnerabilities
| Package | Version | CVE | Severity | Fixed In |
|---------|---------|-----|----------|----------|
| example | 1.2.3   | CVE-2024-XXXX | HIGH | 1.2.4 |

### Secrets Scan
- Exposed secrets found: N
- Files with potential secrets: [list]

### Data Isolation
- Tenant isolation: PASS/FAIL
- Permission boundaries: PASS/FAIL
- API protection: PASS/FAIL
```

## Phase 6: Next Steps

Present via AskUserQuestion:

- Auto-fix Critical Issues (Recommended): Delegate to expert-security subagent to fix critical and high findings. Re-run audit after fixes.
- Create Fix Tasks: Generate TaskList items for each finding, prioritized by severity.
- Export Report: Save to `.moai/reports/security-audit-{timestamp}.md`
- Dismiss: Acknowledge findings without immediate action.

## Integration with Sync Workflow

When invoked as Phase 0.55 within sync.md (not standalone):
- Only CRITICAL findings block the sync pipeline
- HIGH findings are reported as warnings in PR description
- MEDIUM and LOW findings are logged in sync report
- Dependency scan runs only if package files changed

## Completion Criteria

- All OWASP categories analyzed for in-scope files
- Dependency vulnerability scan completed (language-specific)
- Secrets scan completed (source + optional git history)
- Data isolation analysis completed (if applicable)
- Report generated with severity-ranked findings
- Auto-fix offered for critical issues

---

Version: 1.0.0
Updated: 2026-03-29
