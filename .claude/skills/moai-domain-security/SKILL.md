---
name: moai-domain-security
description: OWASP Top 10, static analysis (SAST), dependency security, and secrets management
allowed-tools:
  - Read
  - Bash
tier: 2
auto-load: "true"
---

# Security Expert

## What it does

Provides expertise in application security, including OWASP Top 10 vulnerabilities, static application security testing (SAST), dependency vulnerability scanning, and secrets management.

## When to use

- "보안 취약점 분석", "OWASP 검증", "시크릿 관리", "의존성 보안", "SQL 인젝션", "XSS", "인증", "암호화", "SAST", "취약점 스캔"
- "Security analysis", "OWASP Top 10", "Secrets management", "Vulnerability scanning", "Authentication", "Encryption"
- Automatically invoked when security concerns arise
- Security SPEC implementation (`/alfred:2-run`)

- "보안 취약점 분석", "OWASP 검증", "시크릿 관리", "의존성 보안"
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

### Example 1: SQL Injection Prevention

**❌ Before (Vulnerable)**:
```python
# @CODE:INJECTION-001: SQL 인젝션 위험
email = request.args.get('email')
query = f"SELECT * FROM users WHERE email = '{email}'"
result = db.execute(query)

# 공격: email = "' OR '1'='1"
# 쿼리: SELECT * FROM users WHERE email = '' OR '1'='1'
# 결과: 모든 사용자 반환 💥
```

**✅ After (Parameterized Query)**:
```python
# @CODE:INJECTION-001: 매개변수화된 쿼리
email = request.args.get('email')
query = "SELECT * FROM users WHERE email = %s"
result = db.execute(query, (email,))

# SQL: email 값이 자동으로 이스케이프됨
# 안전성: 100% ✅
```

### Example 2: XSS Prevention

**❌ Before (Vulnerable)**:
```javascript
// @CODE:XSS-001: XSS 취약점
const userInput = req.query.comment;
response.send(`<p>${userInput}</p>`);

// 공격: comment = "<img src=x onerror='alert(1)'>"
// 결과: 스크립트 실행 💥
```

**✅ After (Escaped Output)**:
```javascript
// @CODE:XSS-001: HTML 이스케이프
const userInput = req.query.comment;
const escaped = escapeHtml(userInput);
response.send(`<p>${escaped}</p>`);

// 결과: <p>&lt;img src=x onerror=&#39;alert(1)&#39;&gt;</p>
// 안전성: 100% ✅
```

### Example 3: OWASP Top 10 Checklist

```markdown
1. Broken Access Control
   - [ ] 사용자별 권한 검증
   - [ ] 리소스 접근 제어 테스트

2. Cryptographic Failures
   - [ ] HTTPS 사용
   - [ ] 비밀번호 bcrypt 해싱
   - [ ] 민감한 데이터 암호화

3. Injection
   - [ ] SQL: 매개변수화된 쿼리
   - [ ] NoSQL: 입력 검증
   - [ ] OS: 커맨드 필터링

4. Insecure Design
   - [ ] 위협 모델 작성
   - [ ] 보안 설계 리뷰

5. Security Misconfiguration
   - [ ] 기본값 변경 (포트, 암호)
   - [ ] 보안 헤더 설정

6. Vulnerable Components
   - [ ] 의존성 최신 버전 유지
   - [ ] npm audit / snyk 실행

7. Identification/Authentication Failures
   - [ ] MFA 구현
   - [ ] 강력한 암호 정책
   - [ ] JWT 타임아웃 설정

8. Software/Data Integrity Failures
   - [ ] 코드 서명
   - [ ] 백업 암호화

9. Security Logging/Monitoring Failures
   - [ ] 감사 로그 기록
   - [ ] 이상 탐지

10. SSRF (Server-Side Request Forgery)
    - [ ] URL 입력 검증
    - [ ] 내부 IP 차단
```

### Example 4: Secrets Management

**❌ Before (Hardcoded Secrets)**:
```python
# @CODE:SECRETS-001: 위험 (절대 금지!)
API_KEY = "sk_test_123456789abcdef"
DB_PASSWORD = "admin123"

# 문제: Git 히스토리에 영구 저장 💥
```

**✅ After (Environment Variables)**:
```python
# @CODE:SECRETS-001: 환경 변수
import os

API_KEY = os.environ.get('API_KEY')
DB_PASSWORD = os.environ.get('DB_PASSWORD')

# .env 파일 (.gitignore에 포함):
# API_KEY=sk_test_123456789abcdef
# DB_PASSWORD=admin123

# 안전성: Git에 저장 안 됨 ✅
```

## Keywords

"보안", "OWASP", "SQL 인젝션", "XSS", "인증", "암호화", "시크릿 관리", "의존성 보안", "SAST", "threat modeling", "vulnerability scanning"

## Reference

- OWASP Top 10: `.moai/memory/development-guide.md#OWASP-Top-10`
- Security best practices: CLAUDE.md#보안-기본-원칙
- Vulnerability management: `.moai/memory/development-guide.md#취약점-관리`

## Works well with

- moai-domain-backend (서버 보안)
- moai-domain-web-api (API 보안)
- moai-domain-devops (배포 보안)
