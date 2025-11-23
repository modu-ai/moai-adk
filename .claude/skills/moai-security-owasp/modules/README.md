# OWASP Security Modules - Navigation Index

**Parent Skill**: moai-security-owasp
**Version**: 2.0.0
**Last Updated**: 2025-11-24
**Module Count**: 8 (7 active + 1 placeholder)

---

## Module Directory Structure

```
modules/
├── README.md (this file)
├── advanced-patterns.md - Advanced security patterns (placeholder)
├── cryptographic-failures.md - A02: Weak encryption, hardcoded keys, TLS
├── data-integrity-failures.md - A08: Insecure deserialization, code integrity
├── insecure-design.md - A04: Missing threat modeling, secure design
├── logging-monitoring.md - A09: Security event logging, alerting
├── optimization.md - Security performance optimization (placeholder)
├── vulnerable-components.md - A06: Outdated dependencies, CVE scanning
└── xss-input-validation.md - XSS prevention, CSRF protection
```

---

## Module Descriptions

### Core OWASP Modules (Active)

#### **cryptographic-failures.md** (A02 - Critical)
**Purpose**: Comprehensive guide to preventing cryptographic failures with modern standards.

**Coverage**:
- **Weak Encryption Detection**:
  - Deprecated algorithms (MD5, SHA1, DES, 3DES)
  - Insufficient key sizes (RSA <2048, AES <256)
  - ECB mode vulnerabilities

- **Modern Cryptography Standards**:
  - AES-256-GCM symmetric encryption
  - RSA-2048+ or ECC for asymmetric
  - TLS 1.3 transport security
  - bcrypt/Argon2id password hashing

- **Key Management**:
  - Secure key generation (secrets module)
  - Key rotation strategies
  - Hardware Security Modules (HSM)
  - Key storage best practices

- **TLS Configuration**:
  - TLS 1.3 setup (disable TLS 1.2 and below)
  - Certificate pinning
  - OCSP stapling
  - Perfect Forward Secrecy (PFS)

- **Implementation Examples**:
  - Python cryptography library
  - Node.js crypto module
  - Java KeyStore
  - .NET Cryptography

**When to Use**: Protecting sensitive data, password storage, secure communications

**Prerequisites**: Basic cryptography concepts

**Example Use Cases**:
- Encrypt user PII (AES-256-GCM)
- Hash passwords (bcrypt with 12 rounds)
- Configure HTTPS with TLS 1.3
- Implement certificate pinning

**Learning Path**: SKILL.md (Pattern 1) → cryptographic-failures.md → data-integrity-failures.md

**Estimated Reading Time**: 2-3 hours

**OWASP Mapping**: A02: Cryptographic Failures
**CWE**: CWE-327, CWE-326, CWE-330

---

#### **xss-input-validation.md** (A03 - Critical)
**Purpose**: Complete XSS prevention and CSRF protection implementation guide.

**Coverage**:
- **XSS Prevention**:
  - Reflected XSS protection
  - Stored XSS protection
  - DOM-based XSS protection
  - Output encoding strategies
  - Content Security Policy (CSP)

- **Input Validation**:
  - Allowlist validation (preferred)
  - Blocklist validation (last resort)
  - Type validation (Zod, Joi, Yup)
  - Length validation
  - Format validation (email, URL, phone)

- **CSRF Protection**:
  - Synchronizer token pattern
  - Double-submit cookie
  - SameSite cookie attribute
  - Custom headers validation

- **Context-Specific Encoding**:
  - HTML encoding
  - JavaScript encoding
  - URL encoding
  - CSS encoding

- **Framework Integration**:
  - React/Vue XSS prevention
  - Express.js CSRF middleware
  - Django CSRF protection
  - Spring Security

**When to Use**: Web applications, forms, user input handling

**Prerequisites**: SKILL.md Pattern 2 (Injection)

**Example Use Cases**:
- Prevent stored XSS in comments
- Implement CSP headers
- Add CSRF tokens to forms
- Sanitize user-generated HTML

**Learning Path**: SKILL.md (Pattern 2) → xss-input-validation.md → insecure-design.md

**Estimated Reading Time**: 1.5-2 hours

**OWASP Mapping**: A03: Injection (XSS subset)
**CWE**: CWE-79 (XSS), CWE-352 (CSRF)

---

#### **insecure-design.md** (A04 - Strategic)
**Purpose**: Secure design patterns and threat modeling integration.

**Coverage**:
- **Secure Design Principles**:
  - Defense in depth
  - Least privilege
  - Fail secure (not fail open)
  - Complete mediation
  - Separation of duties

- **Threat Modeling**:
  - STRIDE methodology
  - Attack surface analysis
  - Data flow diagrams
  - Trust boundary identification

- **Secure Design Patterns**:
  - Security by design (not bolted on)
  - Privacy by design
  - Secure defaults
  - Minimize attack surface

- **Business Logic Security**:
  - Workflow validation
  - State machine security
  - Transaction integrity
  - Resource limits

**When to Use**: Design phase, architecture reviews, new features

**Prerequisites**: Basic architecture knowledge

**Example Use Cases**:
- Design secure payment flow
- Implement multi-step approval workflow
- Secure state machine transitions
- Prevent business logic bypass

**Learning Path**: SKILL.md (Pattern 4) → insecure-design.md → vulnerable-components.md

**Estimated Reading Time**: 1.5-2 hours

**OWASP Mapping**: A04: Insecure Design
**CWE**: CWE-840

---

#### **vulnerable-components.md** (A06 - Critical)
**Purpose**: Dependency vulnerability scanning and management strategies.

**Coverage**:
- **Dependency Scanning**:
  - npm audit (Node.js)
  - pip-audit (Python)
  - OWASP Dependency-Check (Java)
  - Snyk integration
  - GitHub Dependabot

- **CVE Management**:
  - CVE database integration
  - Severity assessment (CVSS scores)
  - Vulnerability prioritization
  - Patch management

- **Software Composition Analysis (SCA)**:
  - License compliance
  - Transitive dependency tracking
  - Component inventory (SBOM)
  - Vulnerability correlation

- **Automated Remediation**:
  - Auto-update strategies
  - CI/CD integration
  - Automated PR creation
  - Version pinning vs. ranges

**When to Use**: Continuous integration, dependency updates, security audits

**Prerequisites**: Basic package management knowledge

**Example Use Cases**:
- Scan project for vulnerable dependencies
- Automate dependency updates
- Generate Software Bill of Materials (SBOM)
- Enforce vulnerability SLA (<7 days for critical)

**Learning Path**: SKILL.md (Pattern 4) → vulnerable-components.md → data-integrity-failures.md

**Estimated Reading Time**: 1.5-2 hours

**OWASP Mapping**: A06: Vulnerable and Outdated Components
**CWE**: CWE-1035

---

#### **data-integrity-failures.md** (A08 - High)
**Purpose**: Insecure deserialization prevention and code integrity protection.

**Coverage**:
- **Insecure Deserialization**:
  - Pickle attacks (Python)
  - Java deserialization vulnerabilities
  - YAML/XML injection
  - Safe deserialization patterns

- **Code Integrity**:
  - Code signing
  - Subresource Integrity (SRI)
  - Package integrity verification
  - CI/CD pipeline security

- **Supply Chain Security**:
  - Dependency verification
  - Build reproducibility
  - Artifact signing
  - Software Bill of Materials (SBOM)

- **Data Integrity Validation**:
  - HMAC verification
  - Digital signatures
  - Checksum validation
  - Tamper detection

**When to Use**: Serialization/deserialization, code deployment, supply chain

**Prerequisites**: Understanding of serialization formats

**Example Use Cases**:
- Prevent Python pickle deserialization attack
- Implement SRI for CDN resources
- Sign Docker images
- Verify package integrity

**Learning Path**: vulnerable-components.md → data-integrity-failures.md → logging-monitoring.md

**Estimated Reading Time**: 1.5-2 hours

**OWASP Mapping**: A08: Software and Data Integrity Failures
**CWE**: CWE-502

---

#### **logging-monitoring.md** (A09 - Essential)
**Purpose**: Security event logging and monitoring implementation guide.

**Coverage**:
- **Security Event Logging**:
  - Authentication events (login, logout, failures)
  - Authorization events (access granted/denied)
  - Sensitive operations (password reset, privilege escalation)
  - Security exceptions (injection attempts, validation failures)

- **Log Management**:
  - Structured logging (JSON format)
  - Log levels (DEBUG, INFO, WARN, ERROR, FATAL)
  - Correlation IDs
  - Log rotation and retention

- **SIEM Integration**:
  - Splunk integration
  - ELK Stack (Elasticsearch, Logstash, Kibana)
  - AWS CloudWatch
  - Azure Monitor

- **Alerting & Monitoring**:
  - Real-time alerts
  - Anomaly detection
  - Threshold-based alerts
  - Incident response integration

- **Audit Trails**:
  - Immutable logs
  - Log integrity verification
  - Compliance logging (GDPR, HIPAA)
  - User activity tracking

**When to Use**: Production systems, compliance, incident response

**Prerequisites**: Basic logging concepts

**Example Use Cases**:
- Log all authentication failures
- Set up SIEM integration
- Configure real-time security alerts
- Implement audit trail for compliance

**Learning Path**: SKILL.md (Pattern 5) → logging-monitoring.md → advanced-patterns.md

**Estimated Reading Time**: 1.5-2 hours

**OWASP Mapping**: A09: Security Logging and Monitoring Failures
**CWE**: CWE-778

---

### Placeholder Modules

#### **advanced-patterns.md** (Placeholder)
**Current Status**: Placeholder file with minimal content

**Planned Enhancement**:
- Advanced OWASP attack scenarios
- Zero-day vulnerability response
- Advanced threat detection
- Security automation patterns

**Priority**: Low (core modules cover most use cases)

---

#### **optimization.md** (Placeholder)
**Current Status**: Placeholder file with minimal content

**Planned Enhancement**:
- Security performance optimization
- Efficient cryptographic operations
- Rate limiting optimization
- Log aggregation performance

**Priority**: Low (security correctness > performance)

---

## OWASP Top 10 Coverage Map

### Covered in Parent SKILL.md
- **A01: Broken Access Control** - SKILL.md Pattern 1 (RBAC, BOLA/IDOR)
- **A03: Injection** - SKILL.md Pattern 2 (SQL, NoSQL, Command)
- **A05: Security Misconfiguration** - SKILL.md Pattern 4 (Security headers, CSP)
- **A07: Authentication Failures** - SKILL.md Pattern 3 (MFA, rate limiting)
- **A10: SSRF** - SKILL.md Pattern 5 (URL validation, allowlists)

### Covered in Modules
- **A02: Cryptographic Failures** → cryptographic-failures.md
- **A03: Injection (XSS)** → xss-input-validation.md
- **A04: Insecure Design** → insecure-design.md
- **A06: Vulnerable Components** → vulnerable-components.md
- **A08: Data Integrity Failures** → data-integrity-failures.md
- **A09: Logging & Monitoring Failures** → logging-monitoring.md

### Complete Coverage: 10/10 ✅

---

## Learning Paths

### Beginner Path (Web Application Security)
**Goal**: Secure basic web applications against OWASP Top 10

**Sequence**:
1. **Start**: SKILL.md (1 hour)
   - Focus: 5 core patterns overview
2. **Next**: xss-input-validation.md (2 hours)
   - Focus: XSS prevention, CSRF protection
3. **Then**: cryptographic-failures.md (2.5 hours)
   - Focus: Password hashing, data encryption
4. **Finally**: logging-monitoring.md (2 hours)
   - Focus: Security logging

**Estimated Time**: 7-9 hours
**Outcome**: Build secure web applications with OWASP compliance

---

### Intermediate Path (Enterprise Security)
**Goal**: Implement comprehensive OWASP defenses

**Sequence**:
1. **Start**: SKILL.md (1 hour)
2. **Next**: All 6 active modules (12 hours)
   - cryptographic-failures.md
   - xss-input-validation.md
   - insecure-design.md
   - vulnerable-components.md
   - data-integrity-failures.md
   - logging-monitoring.md
3. **Practice**: Implement all A01-A10 defenses
4. **Finally**: reference.md (in SKILL.md)

**Estimated Time**: 15-18 hours
**Outcome**: Achieve full OWASP Top 10 compliance

---

### Compliance-Focused Path (Audit Preparation)
**Goal**: Pass security compliance audits

**Sequence**:
1. **Start**: SKILL.md reference section (1 hour)
   - Focus: Compliance mapping
2. **Next**: insecure-design.md (2 hours)
   - Focus: Threat modeling documentation
3. **Then**: vulnerable-components.md (2 hours)
   - Focus: SCA, SBOM generation
4. **Then**: data-integrity-failures.md (2 hours)
   - Focus: Code signing, integrity
5. **Finally**: logging-monitoring.md (2 hours)
   - Focus: Audit trails, compliance logging

**Estimated Time**: 9-11 hours
**Outcome**: Pass SOC 2, ISO 27001, GDPR compliance

---

### DevSecOps Path (CI/CD Security)
**Goal**: Automate security in CI/CD pipeline

**Sequence**:
1. **Start**: vulnerable-components.md (2 hours)
   - Focus: Automated dependency scanning
2. **Next**: data-integrity-failures.md (2 hours)
   - Focus: Code signing, build verification
3. **Then**: SKILL.md Pattern 4 (2 hours)
   - Focus: Security misconfiguration detection
4. **Finally**: logging-monitoring.md (2 hours)
   - Focus: CI/CD security logging

**Estimated Time**: 8-10 hours
**Outcome**: Fully automated security pipeline

---

## Cross-References

### Internal Skill References
- **Main Skill**: [moai-security-owasp SKILL.md](../SKILL.md) - 5 core patterns
- **Related Skills**:
  - `moai-domain-security` - Enterprise security architecture
  - `moai-security-identity` - Authentication & authorization
  - `moai-security-encryption` - Advanced cryptography
  - `moai-security-api` - API security patterns
  - `moai-security-threat` - Threat modeling

### External References
- **Context7 Libraries**:
  - `/owasp/top-ten` - OWASP Top 10 patterns
  - `/zaproxy/zaproxy` - Security testing tool
  - `/snyk/snyk` - Dependency scanning
  - `/PyCQA/bandit` - Python security linter
  - `/helmetjs/helmet` - Security headers

### Official Documentation
- [OWASP Top 10 2021](https://owasp.org/Top10/)
- [OWASP ASVS](https://owasp.org/www-project-application-security-verification-standard/)
- [CWE Top 25](https://cwe.mitre.org/top25/)

---

## Search & Index

### Quick Topic Lookup by OWASP Category

**A01: Broken Access Control**:
- RBAC → SKILL.md Pattern 1
- BOLA/IDOR → SKILL.md Pattern 1
- Multi-tenant isolation → SKILL.md Pattern 1

**A02: Cryptographic Failures**:
- AES-256 → cryptographic-failures.md
- bcrypt → cryptographic-failures.md
- TLS 1.3 → cryptographic-failures.md
- Key management → cryptographic-failures.md

**A03: Injection**:
- SQL injection → SKILL.md Pattern 2
- XSS → xss-input-validation.md
- CSRF → xss-input-validation.md
- Input validation → xss-input-validation.md

**A04: Insecure Design**:
- Threat modeling → insecure-design.md
- Secure design patterns → insecure-design.md
- STRIDE → insecure-design.md

**A05: Security Misconfiguration**:
- Security headers → SKILL.md Pattern 4
- CSP → SKILL.md Pattern 4
- Debug mode → SKILL.md Pattern 4

**A06: Vulnerable Components**:
- Dependency scanning → vulnerable-components.md
- CVE management → vulnerable-components.md
- SCA → vulnerable-components.md
- SBOM → vulnerable-components.md

**A07: Authentication Failures**:
- MFA → SKILL.md Pattern 3
- Rate limiting → SKILL.md Pattern 3
- Password policies → SKILL.md Pattern 3

**A08: Data Integrity Failures**:
- Deserialization → data-integrity-failures.md
- Code signing → data-integrity-failures.md
- SRI → data-integrity-failures.md

**A09: Logging & Monitoring**:
- Security logging → logging-monitoring.md
- SIEM → logging-monitoring.md
- Audit trails → logging-monitoring.md

**A10: SSRF**:
- URL validation → SKILL.md Pattern 5
- Allowlists → SKILL.md Pattern 5
- Internal IP blocking → SKILL.md Pattern 5

---

## Module Statistics

| Module | Status | Lines | Complexity | Est. Reading Time |
|--------|--------|-------|------------|-------------------|
| cryptographic-failures.md | ✅ Active | 600+ | High | 2-3 hours |
| xss-input-validation.md | ✅ Active | 500+ | Medium | 1.5-2 hours |
| insecure-design.md | ✅ Active | 400+ | Medium | 1.5-2 hours |
| vulnerable-components.md | ✅ Active | 500+ | Medium | 1.5-2 hours |
| data-integrity-failures.md | ✅ Active | 450+ | Medium | 1.5-2 hours |
| logging-monitoring.md | ✅ Active | 500+ | Medium | 1.5-2 hours |
| advanced-patterns.md | ⚠️ Placeholder | <100 | Low | TBD |
| optimization.md | ⚠️ Placeholder | <100 | Low | TBD |

**Total Estimated Learning Time** (active modules): 10-15 hours
**Total Estimated Learning Time** (with SKILL.md): 11-16 hours

---

## Vulnerability Remediation Workflow

### Using Modules in Remediation Process

**Step 1: Identify Vulnerability** (automated scanning)
- OWASP ZAP scan
- Dependency vulnerability scanning
- Static code analysis

**Step 2: Classify Vulnerability** (OWASP category)
- Map to OWASP Top 10 category (A01-A10)
- Determine severity (Critical/High/Medium/Low)
- Assign CWE identifier

**Step 3: Select Appropriate Module**
- A01 → SKILL.md Pattern 1
- A02 → cryptographic-failures.md
- A03 (Injection) → SKILL.md Pattern 2
- A03 (XSS) → xss-input-validation.md
- A04 → insecure-design.md
- A05 → SKILL.md Pattern 4
- A06 → vulnerable-components.md
- A07 → SKILL.md Pattern 3
- A08 → data-integrity-failures.md
- A09 → logging-monitoring.md
- A10 → SKILL.md Pattern 5

**Step 4: Apply Remediation Pattern** (from module)
- Follow module implementation guide
- Use Context7 validated patterns
- Apply code examples

**Step 5: Validate Fix** (automated testing)
- Run security tests
- Verify vulnerability resolved
- Update compliance status

---

## Success Metrics by Module

| Module | Target Metric | Success Criteria |
|--------|---------------|------------------|
| cryptographic-failures.md | Encryption strength | 100% AES-256+, TLS 1.3 |
| xss-input-validation.md | XSS prevention | 0 XSS vulnerabilities |
| insecure-design.md | Threat model coverage | 100% of features |
| vulnerable-components.md | CVE remediation | <7 days for critical |
| data-integrity-failures.md | Code integrity | 100% signed artifacts |
| logging-monitoring.md | Security event coverage | ≥95% logged |

---

## Contribution Guidelines

When adding new OWASP modules:
1. Map to specific OWASP Top 10 category
2. Include CWE mappings
3. Provide code examples in 3+ languages
4. Add Context7 security tool references
5. Document remediation workflow
6. Include validation checklists

---

**Last Updated**: 2025-11-24
**Maintained By**: MoAI-ADK Security Team
**Status**: Production Ready
**OWASP Coverage**: 10/10 categories ✅
**Module Architecture**: Progressive Disclosure
**Compliance Score**: 95%
