# moai-domain-security: Reference & Standards (2024-2025)

## OWASP Top 10 2024 (Latest)

### Critical Changes from 2021
- **New**: A11:2024 - LLM & AI Security
- **New**: A12:2024 - Supply Chain Security
- **Updated**: A03:2024 - Injection (includes LLM Prompt Injection)

### OWASP Top 10 2024 Complete List

**A01:2024 - Broken Access Control**
- CWE-22 (Path Traversal), CWE-284 (Improper Access Control)
- **Mitigations**: RBAC, ABAC, principle of least privilege
- **Tools**: OWASP ZAP, Burp Suite

**A02:2024 - Cryptographic Failures**
- CWE-259 (Hard-coded Password), CWE-327 (Weak Crypto)
- **Mitigations**: TLS 1.3, AES-256-GCM, proper key management
- **Tools**: SSL Labs, testssl.sh

**A03:2024 - Injection (including LLM Prompt Injection)**
- CWE-79 (XSS), CWE-89 (SQL Injection), CWE-1336 (Prompt Injection)
- **New in 2024**: LLM Prompt Injection, AI Model Poisoning
- **Mitigations**: Parameterized queries, input validation, output encoding
- **Tools**: SQLMap, OWASP ZAP, Semgrep

**A04:2024 - Insecure Design**
- CWE-209 (Information Exposure), CWE-256 (Plaintext Storage)
- **Mitigations**: Threat modeling (STRIDE, PASTA), secure design principles
- **Tools**: Microsoft Threat Modeling Tool, OWASP Threat Dragon

**A05:2024 - Security Misconfiguration**
- CWE-16 (Configuration), CWE-611 (XML External Entity)
- **Mitigations**: Security headers, disable default accounts, patch management
- **Tools**: Nessus, OpenVAS, Lynis

**A06:2024 - Vulnerable and Outdated Components**
- CWE-1035 (Third-party Vulnerabilities), CWE-1104 (Unmaintained Deps)
- **Mitigations**: SBOM, SCA tools, dependency scanning
- **Tools**: Snyk, Dependency-Check, Dependabot

**A07:2024 - Authentication Failures**
- CWE-287 (Improper Authentication), CWE-307 (Brute Force)
- **Mitigations**: MFA, rate limiting, WebAuthn/Passkeys
- **Tools**: Hydra, OWASP ZAP Authentication Scanner

**A08:2024 - Software and Data Integrity Failures**
- CWE-502 (Deserialization), CWE-829 (Untrusted Control Sphere)
- **Mitigations**: Code signing, integrity verification, supply chain security
- **Tools**: Sigstore, in-toto, SLSA framework

**A09:2024 - Security Logging and Monitoring Failures**
- CWE-778 (Insufficient Logging), CWE-223 (Omission of Security Data)
- **Mitigations**: Centralized logging, SIEM, real-time alerting
- **Tools**: ELK Stack, Splunk, Datadog Security Monitoring

**A10:2024 - Server-Side Request Forgery (SSRF)**
- CWE-918 (SSRF)
- **Mitigations**: Input validation, network segmentation, allowlist-based access
- **Tools**: Burp Collaborator, OWASP ZAP SSRF Scanner

**A11:2024 - LLM & AI Security (NEW)**
- CWE-1336 (Prompt Injection), CWE-1337 (Model Poisoning)
- **Mitigations**: Input sanitization, output validation, model versioning
- **Tools**: LLM Fuzzing Tools, AI Red Teaming

**A12:2024 - Supply Chain Security (NEW)**
- CWE-1395 (Dependency Confusion), CWE-494 (Download of Code Without Integrity)
- **Mitigations**: SBOM, SCA, dependency pinning, signed packages
- **Tools**: Syft, Grype, SLSA verification

## LLM & AI Security (2024-2025)

### OWASP Top 10 for LLM Applications
- **LLM01**: Prompt Injection
- **LLM02**: Insecure Output Handling
- **LLM03**: Training Data Poisoning
- **LLM04**: Model Denial of Service
- **LLM05**: Supply Chain Vulnerabilities
- **LLM06**: Sensitive Information Disclosure
- **LLM07**: Insecure Plugin Design
- **LLM08**: Excessive Agency
- **LLM09**: Overreliance
- **LLM10**: Model Theft

### LLM Security Best Practices
- **Input Validation**: Sanitize user prompts
- **Output Validation**: Filter sensitive information
- **Model Monitoring**: Track usage patterns
- **Rate Limiting**: Prevent abuse
- **Access Control**: RBAC for model access

### Resources
- **OWASP LLM Top 10**: https://owasp.org/www-project-top-10-for-large-language-model-applications/
- **NIST AI Risk Management Framework**: https://www.nist.gov/itl/ai-risk-management-framework
- **AI Red Team Testing**: https://learn.microsoft.com/en-us/azure/ai-studio/concepts/red-teaming

## Supply Chain Security (2024-2025)

### SBOM (Software Bill of Materials)
- **CycloneDX 1.6**: https://cyclonedx.org/
- **SPDX 2.3**: https://spdx.dev/
- **SBOM Tools**:
  - Syft: https://github.com/anchore/syft
  - Trivy: https://github.com/aquasecurity/trivy
  - Grype: https://github.com/anchore/grype

### SCA (Software Composition Analysis)
- **Snyk**: https://snyk.io/
- **Dependency-Check (OWASP)**: https://owasp.org/www-project-dependency-check/
- **Dependabot**: https://github.com/dependabot
- **WhiteSource Renovate**: https://www.mend.io/renovate/

### SLSA Framework (Supply Chain Levels for Software Artifacts)
- **SLSA v1.0**: https://slsa.dev/
- **Levels**: 1-4 (increasing security guarantees)
- **Implementation**: Sigstore, in-toto attestations

## Security Tools (2024-2025)

### SAST (Static Application Security Testing)
- **SonarQube**: https://www.sonarsource.com/sonarqube/
- **Snyk Code**: https://snyk.io/product/snyk-code/
- **CodeQL**: https://codeql.github.com/
- **Semgrep**: https://semgrep.dev/

### DAST (Dynamic Application Security Testing)
- **OWASP ZAP**: https://www.zaproxy.org/
- **Burp Suite**: https://portswigger.net/burp
- **Acunetix**: https://www.acunetix.com/
- **Netsparker**: https://www.invicti.com/

### SCA (Software Composition Analysis)
- **Snyk Open Source**: https://snyk.io/product/open-source-security-management/
- **OWASP Dependency-Check**: https://owasp.org/www-project-dependency-check/
- **Trivy**: https://github.com/aquasecurity/trivy
- **Grype**: https://github.com/anchore/grype

### Container Security
- **Trivy**: https://github.com/aquasecurity/trivy
- **Clair**: https://github.com/quay/clair
- **Anchore**: https://anchore.com/
- **Snyk Container**: https://snyk.io/product/container-vulnerability-management/

### Secrets Scanning
- **GitGuardian**: https://www.gitguardian.com/
- **TruffleHog**: https://github.com/trufflesecurity/trufflehog
- **detect-secrets**: https://github.com/Yelp/detect-secrets
- **git-secrets**: https://github.com/awslabs/git-secrets

## Security Headers (2024)

### HTTP Security Headers
- **Content-Security-Policy (CSP) 3.0**: https://www.w3.org/TR/CSP3/
- **Permissions-Policy**: https://www.w3.org/TR/permissions-policy-1/
- **Cross-Origin-Opener-Policy (COOP)**: https://html.spec.whatwg.org/multipage/origin.html#cross-origin-opener-policies
- **Cross-Origin-Resource-Policy (CORP)**: https://fetch.spec.whatwg.org/#cross-origin-resource-policy-header
- **Cross-Origin-Embedder-Policy (COEP)**: https://html.spec.whatwg.org/multipage/origin.html#coep
- **Strict-Transport-Security (HSTS)**: https://tools.ietf.org/html/rfc6797
- **X-Frame-Options**: DENY or SAMEORIGIN
- **X-Content-Type-Options**: nosniff

### CSP 3.0 Example
```
Content-Security-Policy: 
  default-src 'self'; 
  script-src 'self' 'nonce-{random}'; 
  style-src 'self' 'nonce-{random}'; 
  img-src 'self' data: https:; 
  font-src 'self' data:; 
  connect-src 'self' https://api.example.com; 
  frame-ancestors 'none'; 
  base-uri 'self'; 
  form-action 'self';
```

## Zero Trust Architecture (2024)

### NIST Zero Trust Architecture (SP 800-207)
- **Document**: https://csrc.nist.gov/publications/detail/sp/800-207/final
- **Core Principles**:
  - Never trust, always verify
  - Assume breach
  - Explicit verification
  - Least privileged access

### Zero Trust Components
- **Identity and Access Management (IAM)**
- **Multi-Factor Authentication (MFA)**
- **Micro-Segmentation**
- **Continuous Monitoring**
- **Data Encryption** (at rest and in transit)

## Threat Modeling (2024)

### STRIDE Model
- **S**poofing
- **T**ampering
- **R**epudiation
- **I**nformation Disclosure
- **D**enial of Service
- **E**levation of Privilege

### PASTA (Process for Attack Simulation and Threat Analysis)
- **Stage 1**: Define business objectives
- **Stage 2**: Define technical scope
- **Stage 3**: Application decomposition
- **Stage 4**: Threat analysis
- **Stage 5**: Vulnerability analysis
- **Stage 6**: Attack modeling
- **Stage 7**: Risk and impact analysis

### Threat Modeling Tools
- **Microsoft Threat Modeling Tool**: https://www.microsoft.com/en-us/securityengineering/sdl/threatmodeling
- **OWASP Threat Dragon**: https://owasp.org/www-project-threat-dragon/
- **IriusRisk**: https://www.iriusrisk.com/
- **ThreatModeler**: https://threatmodeler.com/

## Compliance Frameworks (2024)

### SOC 2 Type II
- **Trust Service Criteria**: Security, Availability, Confidentiality, Processing Integrity, Privacy
- **Controls**: Access controls, encryption, logging, monitoring
- **Auditors**: AICPA-approved auditors

### ISO 27001:2022
- **Domains**: 93 controls across 4 themes
- **Certification**: External audit required
- **Recertification**: Every 3 years

### GDPR (General Data Protection Regulation)
- **Article 32**: Security of processing
- **Article 33**: Notification of personal data breach
- **Article 35**: Data protection impact assessment
- **Fines**: Up to €20 million or 4% of annual global turnover

### CCPA (California Consumer Privacy Act)
- **Rights**: Right to know, right to delete, right to opt-out
- **Requirements**: Privacy notice, data access, deletion mechanisms
- **Fines**: $2,500 per violation, $7,500 for intentional violations

### PCI DSS 4.0 (2024)
- **12 Requirements**: Build and maintain secure network, protect cardholder data, etc.
- **Key Changes from 3.2**: Customized approach, expanded authentication requirements
- **Compliance Deadline**: March 31, 2025

## Related MoAI Skills

- **moai-security-auth**: Authentication and authorization patterns
- **moai-security-encryption**: Cryptography and key management
- **moai-security-api**: API security best practices
- **moai-security-compliance**: Compliance frameworks and auditing
- **moai-security-owasp**: OWASP Top 10 implementation
- **moai-domain-cloud**: Cloud security patterns
- **moai-domain-backend**: Backend security hardening

---

**Last Updated**: 2025-11-22  
**Compliance**: OWASP Top 10 2024, SOC 2, ISO 27001:2022, GDPR, CCPA, PCI DSS 4.0
