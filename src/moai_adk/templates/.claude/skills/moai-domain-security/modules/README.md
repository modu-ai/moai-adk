# Security Modules - Navigation Index

**Parent Skill**: moai-domain-security
**Version**: 2.0.0
**Last Updated**: 2025-11-24
**Module Count**: 13

---

## Module Directory Structure

```
modules/
├── README.md (this file)
├── access-control.md - RBAC, ABAC, and policy-based access control
├── advanced-patterns.md - Advanced security architecture patterns
├── cryptography-advanced.md - Advanced cryptographic techniques and protocols
├── cryptography-standards.md - Modern encryption and hashing standards
├── devsecops-automation.md - Security CI/CD integration patterns
├── owasp-compliance.md - OWASP Top 10 2021 detailed patterns (A01-A10)
├── optimization.md - Security performance optimization
├── reference.md - API reference, compliance checklists, tool guides
├── secure-architecture-patterns.md - Secure system design patterns
├── secure-coding-patterns.md - Language-specific secure coding practices
├── threat-modeling-advanced.md - Advanced threat modeling techniques
├── threat-modeling.md - STRIDE, PASTA, LINDDUN methodologies
└── zero-trust-architecture.md - Zero-trust implementation with adaptive authentication
```

---

## Module Descriptions

### Core Security Modules

#### **owasp-compliance.md** (Critical - Start Here)
**Purpose**: Complete coverage of OWASP Top 10 2021 vulnerabilities (A01-A10) with detailed remediation patterns.

**Coverage**:
- A01: Broken Access Control - BOLA/IDOR prevention
- A02: Cryptographic Failures - AES-256, TLS 1.3, bcrypt implementation
- A03: Injection - SQL, NoSQL, command injection prevention
- A04: Insecure Design - Threat modeling and secure design patterns
- A05: Security Misconfiguration - Hardening guides and automated checks
- A06: Vulnerable Components - Dependency scanning and SCA tools
- A07: Authentication Failures - MFA, password policies, session management
- A08: Software/Data Integrity - Code signing, SBOM, supply chain security
- A09: Logging & Monitoring - SIEM integration and audit trails
- A10: SSRF - URL validation and allowlist patterns

**When to Use**: Implementing web application security, API protection, compliance audits

**Prerequisites**: Understanding of HTTP/HTTPS, web application architecture

**Example Use Cases**:
- Building REST API with OWASP compliance
- Securing user authentication system
- Preventing SQL injection in database queries
- Implementing proper access control checks

**Learning Path**: Start here → cryptography-standards.md → zero-trust-architecture.md

---

#### **zero-trust-architecture.md** (Essential)
**Purpose**: Comprehensive zero-trust security model implementation with adaptive authentication.

**Coverage**:
- Zero-trust principles (never trust, always verify)
- Adaptive risk-based authentication
- Multi-factor authentication (MFA) strategies
- Continuous validation and re-authentication
- Device fingerprinting and behavioral analysis
- Risk scoring algorithms

**When to Use**: Building enterprise applications, financial systems, healthcare platforms

**Prerequisites**: owasp-compliance.md (authentication patterns)

**Example Use Cases**:
- Banking application with adaptive MFA
- SaaS platform with risk-based authentication
- Healthcare system with continuous verification
- Enterprise identity management

**Learning Path**: owasp-compliance.md → zero-trust-architecture.md → access-control.md

---

#### **threat-modeling.md** (Strategic)
**Purpose**: Systematic threat identification using STRIDE, PASTA, and LINDDUN methodologies.

**Coverage**:
- STRIDE methodology (Spoofing, Tampering, Repudiation, Information Disclosure, DoS, Elevation of Privilege)
- PASTA methodology (Process for Attack Simulation and Threat Analysis)
- LINDDUN methodology (privacy threat modeling)
- Automated threat analysis
- Risk matrix generation
- Mitigation planning

**When to Use**: Design phase, architecture reviews, security assessments

**Prerequisites**: Basic security concepts, system architecture knowledge

**Example Use Cases**:
- Pre-implementation threat analysis
- Security architecture review
- Compliance documentation
- Risk assessment for new features

**Learning Path**: threat-modeling.md → threat-modeling-advanced.md → secure-architecture-patterns.md

---

### Cryptography Modules

#### **cryptography-standards.md** (Foundation)
**Purpose**: Modern cryptographic standards and best practices for data protection.

**Coverage** (50-100 words):
- AES-256-GCM symmetric encryption
- RSA-2048+ asymmetric encryption
- bcrypt password hashing (2^12 iterations)
- TLS 1.3 transport security
- Digital signatures (RSA-PSS, ECDSA)
- Key derivation (PBKDF2, Argon2id)
- Secure random generation

**When to Use**: Protecting sensitive data, password storage, secure communications

**Prerequisites**: None (beginner-friendly)

**Example Use Cases**:
- Encrypting user PII data
- Hashing and verifying passwords
- Implementing HTTPS/TLS
- Generating cryptographic signatures

---

#### **cryptography-advanced.md** (Expert)
**Purpose**: Advanced cryptographic techniques for specialized security requirements.

**Coverage**:
- Homomorphic encryption
- Post-quantum cryptography
- Elliptic curve cryptography (ECC)
- Key management systems (KMS)
- Hardware security modules (HSM)
- Certificate pinning and OCSP
- Cryptographic protocol design

**When to Use**: Advanced security requirements, regulatory compliance, research

**Prerequisites**: cryptography-standards.md

---

### Implementation Modules

#### **devsecops-automation.md** (Pipeline Integration)
**Purpose**: Automated security testing and validation in CI/CD pipelines.

**Coverage**:
- SAST (Static Application Security Testing)
- DAST (Dynamic Application Security Testing)
- IAST (Interactive Application Security Testing)
- Dependency vulnerability scanning
- Container image scanning
- Infrastructure as Code (IaC) security
- Security quality gates

**When to Use**: CI/CD pipeline setup, DevOps automation, continuous security

**Prerequisites**: Basic DevOps knowledge, CI/CD experience

**Example Use Cases**:
- GitHub Actions security pipeline
- GitLab CI security automation
- Jenkins security scanning
- Kubernetes security validation

---

#### **secure-coding-patterns.md** (Developer Guide)
**Purpose**: Language-specific secure coding practices across 25+ programming languages.

**Coverage**:
- Input validation and sanitization
- Output encoding patterns
- Secure error handling
- Memory safety
- SQL injection prevention
- XSS prevention
- CSRF token implementation
- Language-specific security libraries

**When to Use**: Writing secure code, code reviews, security training

**Prerequisites**: Programming experience in target language

---

#### **access-control.md** (Authorization)
**Purpose**: Comprehensive access control implementation with RBAC, ABAC, and policy-based patterns.

**Coverage**:
- Role-Based Access Control (RBAC)
- Attribute-Based Access Control (ABAC)
- Policy-Based Access Control (PBAC)
- Permission management
- Multi-tenant isolation
- Fine-grained authorization
- Access control lists (ACL)

**When to Use**: User authorization, multi-tenant applications, enterprise systems

**Prerequisites**: owasp-compliance.md (A01: Broken Access Control)

---

### Advanced Modules

#### **threat-modeling-advanced.md** (Expert)
**Purpose**: Advanced threat modeling techniques with automated analysis and AI-enhanced detection.

**Coverage**:
- Attack tree analysis
- Threat intelligence integration
- Machine learning for threat detection
- Automated vulnerability assessment
- Red team/blue team exercises
- Advanced STRIDE patterns

**When to Use**: Complex architectures, high-security environments, research

**Prerequisites**: threat-modeling.md, security architecture experience

---

#### **secure-architecture-patterns.md** (Architecture)
**Purpose**: Proven secure system design patterns for enterprise applications.

**Coverage**:
- Defense-in-depth architecture
- Network segmentation patterns
- Microservices security
- API gateway patterns
- Service mesh security
- Cloud security architectures

**When to Use**: System design, architecture reviews, migration planning

**Prerequisites**: Architecture experience, threat-modeling.md

---

### Reference Modules

#### **reference.md** (Quick Reference)
**Purpose**: Comprehensive API reference, compliance checklists, and tool guides.

**Coverage**:
- Security tool catalog
- Compliance framework mapping (SOC 2, ISO 27001, GDPR, HIPAA)
- CWE/CVE reference
- Security testing commands
- Best practice checklists
- Troubleshooting guides

**When to Use**: Quick lookups, compliance audits, tool selection

**Prerequisites**: None (reference material)

---

#### **optimization.md** (Performance)
**Purpose**: Security performance optimization without compromising security posture.

**Coverage**:
- Cryptographic performance optimization
- Authentication caching strategies
- Rate limiting optimization
- Security monitoring performance
- Log aggregation efficiency

**When to Use**: Performance tuning, scalability planning

**Prerequisites**: Security fundamentals

---

#### **advanced-patterns.md** (Future-Proofing)
**Purpose**: Cutting-edge security patterns and emerging threat protection.

**Coverage**:
- AI/ML security patterns
- Quantum-resistant cryptography
- Zero-knowledge proofs
- Blockchain security
- Container security advanced patterns
- Serverless security

**When to Use**: Research, future planning, innovative projects

**Prerequisites**: Advanced security knowledge

---

## Learning Paths

### Beginner Path (Web Application Security)
**Goal**: Build secure web applications with OWASP compliance

**Sequence**:
1. **Start**: owasp-compliance.md (2-3 hours)
   - Focus: A01, A03, A07 (access control, injection, authentication)
2. **Next**: cryptography-standards.md (1-2 hours)
   - Focus: Password hashing, data encryption
3. **Then**: secure-coding-patterns.md (1-2 hours)
   - Focus: Input validation, output encoding
4. **Finally**: reference.md (reference)
   - Focus: Checklists, tools

**Estimated Time**: 6-9 hours
**Outcome**: Build secure REST API with OWASP compliance

---

### Intermediate Path (Enterprise Security)
**Goal**: Implement enterprise-grade security architecture

**Sequence**:
1. **Start**: threat-modeling.md (2-3 hours)
   - Focus: STRIDE methodology
2. **Next**: zero-trust-architecture.md (2-3 hours)
   - Focus: Adaptive authentication
3. **Then**: access-control.md (1-2 hours)
   - Focus: RBAC/ABAC implementation
4. **Then**: devsecops-automation.md (2-3 hours)
   - Focus: CI/CD security integration
5. **Finally**: secure-architecture-patterns.md (1-2 hours)
   - Focus: Defense-in-depth

**Estimated Time**: 10-15 hours
**Outcome**: Design and implement enterprise security architecture

---

### Expert Path (Security Specialist)
**Goal**: Master advanced security techniques and compliance

**Sequence**:
1. **Start**: threat-modeling-advanced.md (3-4 hours)
   - Focus: Automated threat analysis
2. **Next**: cryptography-advanced.md (3-4 hours)
   - Focus: Post-quantum crypto, HSM
3. **Then**: advanced-patterns.md (2-3 hours)
   - Focus: AI/ML security, zero-knowledge proofs
4. **Finally**: All reference materials
   - Focus: Compliance frameworks, tool mastery

**Estimated Time**: 15-20 hours
**Outcome**: Design and audit complex security architectures, lead security teams

---

### Compliance-Focused Path (Regulatory Requirements)
**Goal**: Achieve and maintain security compliance

**Sequence**:
1. **Start**: reference.md (compliance section) (1 hour)
   - Focus: SOC 2, ISO 27001, GDPR requirements
2. **Next**: owasp-compliance.md (3-4 hours)
   - Focus: All A01-A10 compliance
3. **Then**: devsecops-automation.md (2-3 hours)
   - Focus: Automated compliance validation
4. **Finally**: threat-modeling.md (2-3 hours)
   - Focus: Compliance documentation

**Estimated Time**: 10-12 hours
**Outcome**: Pass security compliance audits

---

## Cross-References

### Internal Skill References
- **Main Skill**: [moai-domain-security SKILL.md](../SKILL.md) - Quick reference and 5 core patterns
- **Related Skills**:
  - `moai-security-owasp` - OWASP Top 10 validation and testing
  - `moai-security-identity` - Identity and access management (IAM)
  - `moai-security-api` - API security patterns
  - `moai-security-zero-trust` - Zero-trust architecture deep dive
  - `moai-security-threat` - Advanced threat modeling
  - `moai-domain-cloud` - Cloud security (AWS, GCP, Azure)
  - `moai-domain-devops` - DevOps infrastructure security

### External References
- **Context7 Libraries**:
  - `/owasp/top-ten` - OWASP Top 10 patterns
  - `/nist/zero-trust` - Zero-trust guidance
  - `/cryptography/hazmat` - Python cryptography
  - `/owasp/zap` - Dynamic security testing
  - `/pycqa/bandit` - Python security linting

---

## Search & Index

### Quick Topic Lookup

**Access Control**:
- RBAC → access-control.md
- ABAC → access-control.md
- BOLA/IDOR → owasp-compliance.md (A01)

**Authentication**:
- MFA → owasp-compliance.md (A07)
- Zero-trust → zero-trust-architecture.md
- Session management → owasp-compliance.md (A07)

**Cryptography**:
- AES-256 → cryptography-standards.md
- bcrypt → cryptography-standards.md
- TLS 1.3 → cryptography-standards.md
- Post-quantum → cryptography-advanced.md

**Injection**:
- SQL injection → owasp-compliance.md (A03)
- NoSQL injection → owasp-compliance.md (A03)
- Command injection → secure-coding-patterns.md

**DevSecOps**:
- CI/CD security → devsecops-automation.md
- SAST/DAST → devsecops-automation.md
- Container scanning → devsecops-automation.md

**Threat Modeling**:
- STRIDE → threat-modeling.md
- PASTA → threat-modeling.md
- Attack trees → threat-modeling-advanced.md

**Compliance**:
- SOC 2 → reference.md
- ISO 27001 → reference.md
- GDPR → reference.md
- HIPAA → reference.md

---

## Module Statistics

| Module | Lines | Complexity | Est. Reading Time |
|--------|-------|------------|-------------------|
| owasp-compliance.md | 800+ | High | 2-3 hours |
| zero-trust-architecture.md | 600+ | High | 1.5-2 hours |
| threat-modeling.md | 500+ | Medium | 1-1.5 hours |
| cryptography-standards.md | 400+ | Medium | 1 hour |
| devsecops-automation.md | 500+ | Medium | 1-1.5 hours |
| secure-coding-patterns.md | 600+ | Medium | 1.5-2 hours |
| access-control.md | 400+ | Medium | 1 hour |
| cryptography-advanced.md | 500+ | High | 1.5-2 hours |
| threat-modeling-advanced.md | 500+ | High | 1.5-2 hours |
| secure-architecture-patterns.md | 400+ | High | 1-1.5 hours |
| reference.md | 600+ | Low | Reference |
| optimization.md | 300+ | Medium | 45 min |
| advanced-patterns.md | 400+ | High | 1-1.5 hours |

**Total Estimated Learning Time**: 25-35 hours (all modules)

---

## Contribution Guidelines

When adding new security modules:
1. Follow Progressive Disclosure structure
2. Include code examples in 3+ languages
3. Map to OWASP Top 10 categories
4. Add Context7 references
5. Provide compliance mapping
6. Include validation checklists

---

**Last Updated**: 2025-11-24
**Maintained By**: MoAI-ADK Security Team
**Status**: Production Ready
**Module Architecture**: Progressive Disclosure
