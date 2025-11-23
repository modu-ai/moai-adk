---
name: moai-domain-security
description: Enterprise-grade security expertise with production-ready patterns for OWASP Top 10 2021, zero-trust architecture, threat modeling, secure SDLC, DevSecOps automation
version: 1.1.0
modularized: true
---

## 📊 Skill Metadata

**Name**: moai-domain-security
**Domain**: Enterprise Security Architecture
**Freedom Level**: high
**Target Users**: Security engineers, architects, DevSecOps specialists
**Invocation**: Skill("moai-domain-security")
**Progressive Disclosure**: SKILL.md (core) → modules/ (detailed guides)
**Last Updated**: 2025-11-23
**Modularized**: true

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Enterprise security expertise with OWASP compliance and zero-trust architecture.

**Key Capabilities**:
- OWASP Top 10 2021 vulnerability protection
- Zero-trust authentication & authorization
- Threat modeling (STRIDE, PASTA methodologies)
- DevSecOps pipeline automation
- Cryptography and encryption standards
- Compliance frameworks (SOC 2, ISO 27001, GDPR, CCPA)

**Core Tools**:
- OWASP ZAP (vulnerability scanning)
- Bandit (Python security testing)
- ThreatDragon (threat modeling)
- HashiCorp Vault (secrets management)
- SonarQube (code quality/security)

---

## 📚 Core Patterns (5-10 minutes)

### Pattern 1: OWASP Top 10 2021 Protection

**Key Concept**: Defend against the most critical web vulnerabilities.

**Approach**:
```python
# Security middleware for OWASP protection
class SecurityMiddleware:
    def __init__(self, app):
        self.app = app
        app.before_request(self.before_request_handler)
        app.after_request(self.after_request_handler)

    def before_request_handler(self, request):
        # A01: Broken Access Control
        self._verify_access_control(request)
        # A03: Injection prevention
        self._prevent_injection_attacks(request)

    def after_request_handler(self, response):
        # Security headers for A05: Cross-Site Request Forgery (CSRF)
        response.headers['X-Content-Type-Options'] = 'nosniff'
        response.headers['X-Frame-Options'] = 'DENY'
        response.headers['X-XSS-Protection'] = '1; mode=block'
        response.headers['Strict-Transport-Security'] = 'max-age=31536000'
        return response

    def _prevent_injection_attacks(self, request):
        # Parameterized queries prevent SQL injection (A03)
        # Input validation prevents cross-site scripting (A07)
        pass
```

**Use Case**: Protecting REST APIs from injection attacks, XSS, CSRF vulnerabilities.

### Pattern 2: Zero-Trust Architecture

**Key Concept**: Never trust, always verify authentication & authorization.

**Approach**:
```python
# Zero-trust authentication with risk scoring
class ZeroTrustAuth:
    def authenticate_user(self, credentials: dict, context: dict) -> dict:
        user = self._verify_credentials(credentials)

        # Calculate risk based on context
        risk_score = self._calculate_risk_score(user, context)

        # Create token with trust level
        token_claims = {
            'user_id': user['id'],
            'trust_level': self._determine_trust_level(risk_score),
            'risk_score': risk_score,
            'device_fingerprint': context.get('device_fingerprint'),
            'ip_address': context.get('ip_address')
        }

        token = jwt.encode(token_claims, self.secret_key)
        return {'token': token, 'trust_level': token_claims['trust_level']}

    def _calculate_risk_score(self, user: dict, context: dict) -> int:
        risk = 0
        risk += 20 if self._is_unusual_location(user, context) else 0
        risk += 15 if self._is_new_device(user, context) else 0
        risk += 10 if self._is_unusual_time(user) else 0
        return min(risk, 100)
```

**Use Case**: Implementing adaptive authentication for sensitive applications with risk-based MFA.

### Pattern 3: Threat Modeling (STRIDE)

**Key Concept**: Systematically identify and mitigate security threats.

**Approach**:
```python
# STRIDE threat modeling framework
class ThreatModelAnalyzer:
    def analyze_system(self, system_architecture: dict) -> list:
        threats = []

        for component_name, config in system_architecture.items():
            # Identify threats by category:
            # S: Spoofing, T: Tampering, R: Repudiation
            # I: Information Disclosure, D: Denial of Service
            # E: Elevation of Privilege
            component_threats = self._analyze_component(component_name, config)
            threats.extend(component_threats)

        return self._generate_threat_report(threats)

    def _analyze_component(self, name: str, config: dict) -> list:
        threats = []
        if config['type'] == 'web_application':
            threats.append({
                'category': 'Spoofing',
                'description': 'Attacker impersonates legitimate user',
                'mitigation': ['Implement MFA', 'Use CSRF tokens', 'Session management']
            })
        return threats
```

**Use Case**: Designing secure architecture by identifying threats before implementation.

### Pattern 4: DevSecOps Automation

**Key Concept**: Integrate security checks into CI/CD pipeline.

**Approach**:
```yaml
# Security pipeline in CI/CD
jobs:
  security-scan:
    steps:
    - name: Static application security testing
      run: bandit -r src/ --json --output bandit-report.json

    - name: Dependency vulnerability scan
      run: safety check --json --output safety-report.json

    - name: Dynamic security testing
      run: |
        docker run owasp/zap2docker-stable \
          zap-baseline.py -t http://app-url

    - name: Threat modeling
      run: python threat_modeling.py --output threat-model.json

    - name: Compliance check
      run: |
        python compliance_check.py --framework soc2
        python compliance_check.py --framework gdpr
```

**Use Case**: Automated security enforcement preventing vulnerable code from reaching production.

### Pattern 5: Cryptography Standards

**Key Concept**: Use modern, proven cryptographic algorithms.

**Approach**:
```python
# Secure cryptography implementation
from cryptography.fernet import Fernet
from cryptography.hazmat.primitives import hashes
import bcrypt

# Data encryption (at-rest & in-transit)
cipher = Fernet(key)
encrypted_data = cipher.encrypt(plaintext.encode())

# Password hashing
hashed_password = bcrypt.hashpw(password.encode(), bcrypt.gensalt(rounds=12))

# Digital signatures for data integrity
from cryptography.hazmat.primitives.asymmetric import rsa, padding
private_key = rsa.generate_private_key(
    public_exponent=65537,
    key_size=2048
)
signature = private_key.sign(
    data,
    padding.PSS(
        mgf=padding.MGF1(hashes.SHA256()),
        salt_length=padding.PSS.MAX_LENGTH
    ),
    hashes.SHA256()
)
```

**Use Case**: Protecting sensitive data with AES-256 encryption, bcrypt password hashing, RSA digital signatures.

---

## 📖 Advanced Documentation

This Skill uses Progressive Disclosure. For detailed patterns:

- **[modules/owasp-compliance.md](modules/owasp-compliance.md)** - OWASP Top 10 2021 detailed patterns
- **[modules/zero-trust-architecture.md](modules/zero-trust-architecture.md)** - Zero-trust implementation patterns
- **[modules/threat-modeling.md](modules/threat-modeling.md)** - STRIDE & PASTA methodologies
- **[modules/devsecops-automation.md](modules/devsecops-automation.md)** - Security CI/CD integration
- **[modules/cryptography-standards.md](modules/cryptography-standards.md)** - Encryption & hashing patterns
- **[modules/reference.md](modules/reference.md)** - API reference, compliance checklists, tools

---

## 🎯 Security Assessment Workflow

**Step 1**: Identify system components (architecture)
**Step 2**: Apply threat modeling (STRIDE analysis)
**Step 3**: Map to OWASP Top 10 vulnerabilities
**Step 4**: Design security controls (zero-trust)
**Step 5**: Implement cryptography standards
**Step 6**: Automate with DevSecOps pipeline

---

## 🔗 Integration with Other Skills

**Complementary Skills**:
- Skill("moai-security-api") - API security patterns
- Skill("moai-security-identity") - Identity & access management
- Skill("moai-security-owasp") - OWASP compliance validation
- Skill("moai-security-zero-trust") - Zero-trust architecture
- Skill("moai-domain-cloud") - Cloud security patterns
- Skill("moai-domain-devops") - DevOps infrastructure

---

## 📈 Version History

**1.1.0** (2025-11-23)
- 🔄 Refactored with Progressive Disclosure
- ✨ 5 Core Patterns highlighted
- ✨ Modularized advanced content

**1.0.0** (2025-11-12)
- ✨ OWASP Top 10 compliance
- ✨ Zero-trust architecture
- ✨ Threat modeling (STRIDE)

---

**Maintained by**: alfred
**Domain**: Enterprise Security
**Generated with**: MoAI-ADK Skill Factory
