---
name: moai-domain-security
description: Enterprise security with OWASP Top 10 2021, zero-trust architecture, threat modeling (STRIDE/PASTA), secure SDLC, DevSecOps automation, and compliance frameworks (SOC 2, ISO 27001, GDPR). Use when implementing security controls, conducting threat assessments, or building secure applications.
allowed-tools: Read, WebFetch, Bash, Grep, Glob
---

## 📊 Skill Metadata

**Name**: moai-domain-security
**Version**: 2.0.0
**Domain**: Enterprise Security Architecture
**Freedom Level**: high
**Target Users**: Security engineers, architects, DevSecOps specialists
**Last Updated**: 2025-11-24
**Modularized**: true
**Author**: MoAI-ADK Security Team
**Compliance Score**: 95%

**Auto-Trigger Keywords**: security, owasp, zero-trust, threat-modeling, devsecops, vulnerability, encryption, authentication, authorization, compliance

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Enterprise security expertise with OWASP compliance, zero-trust architecture, and threat modeling for production-ready applications.

**Key Capabilities**:
- ✅ OWASP Top 10 2021 vulnerability protection (A01-A10)
- ✅ Zero-trust authentication with adaptive risk scoring
- ✅ Threat modeling (STRIDE, PASTA, LINDDUN methodologies)
- ✅ DevSecOps pipeline automation (SAST, DAST, IAST)
- ✅ Cryptography standards (AES-256, RSA-2048, bcrypt)
- ✅ Compliance frameworks (SOC 2, ISO 27001, GDPR, CCPA, HIPAA)
- ✅ Security testing automation with Context7 latest patterns

**Core Framework**: Defense-in-Depth
```
Perimeter Security (Firewall, WAF)
  ↓
Network Security (Segmentation, Zero-Trust)
  ↓
Application Security (OWASP Top 10, Secure Coding)
  ↓
Data Security (Encryption, DLP)
  ↓
Identity Security (IAM, MFA, RBAC)
```

**When to Use**:
- Designing secure application architectures
- Implementing authentication/authorization systems
- Conducting threat modeling and risk assessments
- Building DevSecOps CI/CD pipelines
- Ensuring OWASP Top 10 compliance
- Meeting regulatory compliance requirements
- Responding to security incidents

---

## 📚 Core Patterns (5-10 minutes each)

### Pattern 1: OWASP Top 10 2021 Comprehensive Protection

**Key Concept**: Defend against the most critical web application vulnerabilities with layered security controls.

**OWASP Top 10 2021 Coverage**:
```
A01: Broken Access Control       → RBAC + ABAC patterns
A02: Cryptographic Failures      → AES-256, TLS 1.3, bcrypt
A03: Injection                   → Parameterized queries, input validation
A04: Insecure Design             → Threat modeling, secure design review
A05: Security Misconfiguration   → Hardening guides, automated checks
A06: Vulnerable Components       → Dependency scanning, SCA tools
A07: Authentication Failures     → MFA, password policies, session management
A08: Software/Data Integrity     → Code signing, SBOM, supply chain security
A09: Logging & Monitoring        → SIEM integration, audit trails
A10: Server-Side Request Forgery → URL validation, allowlist patterns
```

**Implementation**:
```python
# Multi-layer security middleware with OWASP protection
class OWASPSecurityMiddleware:
    """Enterprise security middleware covering OWASP Top 10 2021."""

    def __init__(self, app, config: SecurityConfig):
        self.app = app
        self.config = config
        self._setup_security_layers()

    def _setup_security_layers(self):
        """Initialize all security layers."""
        # A01: Access Control
        self.access_controller = RBACController(self.config.rbac_rules)

        # A02: Cryptographic Standards
        self.crypto_manager = CryptoManager(
            encryption_key=self.config.encryption_key,
            algorithm='AES-256-GCM'
        )

        # A03: Injection Prevention
        self.input_validator = InputValidator(
            sql_patterns=SQL_INJECTION_PATTERNS,
            xss_patterns=XSS_PATTERNS,
            command_patterns=COMMAND_INJECTION_PATTERNS
        )

    def process_request(self, request: Request) -> Response:
        """Process request with security validation."""
        # Step 1: Verify authentication (A07)
        user = self._verify_authentication(request)

        # Step 2: Check access control (A01)
        if not self.access_controller.has_permission(user, request.resource):
            raise ForbiddenError("Access denied")

        # Step 3: Validate and sanitize inputs (A03)
        validated_data = self.input_validator.validate(request.data)

        # Step 4: Apply rate limiting (A09)
        self._apply_rate_limiting(user, request)

        # Step 5: Log security event (A09)
        self._log_security_event(user, request, 'allowed')

        return self.app.handle(request, validated_data)

    def _verify_authentication(self, request: Request) -> User:
        """Verify JWT token and session (A07)."""
        token = request.headers.get('Authorization', '').replace('Bearer ', '')
        if not token:
            raise UnauthorizedError("Missing authentication token")

        try:
            claims = jwt.decode(token, self.config.jwt_secret, algorithms=['HS256'])
            return User.from_claims(claims)
        except jwt.ExpiredSignatureError:
            raise UnauthorizedError("Token expired")
        except jwt.InvalidTokenError:
            raise UnauthorizedError("Invalid token")
```

**Use Case**: Protecting REST APIs and web applications from OWASP Top 10 vulnerabilities with automated enforcement.

---

### Pattern 2: Zero-Trust Architecture with Adaptive Authentication

**Key Concept**: Never trust, always verify every access request with risk-based authentication.

**Zero-Trust Principles**:
1. **Verify Every Access**: Authenticate and authorize every request
2. **Least Privilege**: Grant minimum required permissions
3. **Assume Breach**: Design assuming network compromise
4. **Verify Explicitly**: Use identity, device, location, behavior for decisions
5. **Continuous Validation**: Re-verify throughout session

**Implementation**:
```python
# Zero-trust authentication engine with adaptive risk scoring
class ZeroTrustAuthEngine:
    """Adaptive authentication with risk-based MFA."""

    def authenticate(self, credentials: dict, context: dict) -> AuthResult:
        """Authenticate user with adaptive risk assessment."""
        # Step 1: Verify credentials (something you know)
        user = self._verify_credentials(credentials)
        if not user:
            raise AuthenticationError("Invalid credentials")

        # Step 2: Calculate risk score based on context
        risk_score = self._calculate_risk_score(user, context)

        # Step 3: Determine required authentication factors
        required_factors = self._determine_auth_factors(risk_score)

        # Step 4: Create token with trust level
        token_claims = {
            'user_id': user.id,
            'trust_level': self._calculate_trust_level(risk_score),
            'risk_score': risk_score,
            'device_fingerprint': context.get('device_fingerprint'),
            'ip_address': context.get('ip_address'),
            'session_id': str(uuid.uuid4()),
            'issued_at': datetime.utcnow().isoformat(),
            'expires_at': (datetime.utcnow() + timedelta(hours=8)).isoformat()
        }

        # Step 5: Issue token with appropriate TTL
        token = jwt.encode(token_claims, self.config.jwt_secret, algorithm='HS256')

        return AuthResult(
            token=token,
            trust_level=token_claims['trust_level'],
            required_factors=required_factors,
            risk_score=risk_score
        )

    def _calculate_risk_score(self, user: User, context: dict) -> int:
        """Calculate risk score from multiple signals."""
        risk = 0

        # Geolocation risk (0-30 points)
        if self._is_unusual_location(user.id, context.get('ip_address')):
            risk += 30

        # Device risk (0-25 points)
        if self._is_new_device(user.id, context.get('device_fingerprint')):
            risk += 25

        # Time-based risk (0-15 points)
        if self._is_unusual_time(user.id):
            risk += 15

        # Behavioral risk (0-20 points)
        if self._is_anomalous_behavior(user.id, context):
            risk += 20

        # Velocity risk (0-10 points)
        if self._is_impossible_travel(user.id, context.get('ip_address')):
            risk += 10

        return min(risk, 100)  # Cap at 100

    def _determine_auth_factors(self, risk_score: int) -> List[str]:
        """Determine required authentication factors based on risk."""
        if risk_score < 20:
            return ['password']  # Low risk: password only
        elif risk_score < 50:
            return ['password', 'totp']  # Medium risk: password + TOTP
        else:
            return ['password', 'totp', 'hardware_key']  # High risk: all factors
```

**Use Case**: Implementing adaptive authentication for sensitive applications (banking, healthcare, enterprise SaaS) with risk-based MFA.

---

### Pattern 3: Threat Modeling with STRIDE Methodology

**Key Concept**: Systematically identify and mitigate security threats before implementation.

**STRIDE Categories**:
```
S - Spoofing:           Identity impersonation
T - Tampering:          Data modification
R - Repudiation:        Deny actions
I - Information Disclosure: Leak sensitive data
D - Denial of Service:  Resource exhaustion
E - Elevation of Privilege: Unauthorized access
```

**Implementation**:
```python
# Automated threat modeling with STRIDE analysis
class STRIDEThreatAnalyzer:
    """Automated threat identification and mitigation planning."""

    def analyze_architecture(self, architecture: dict) -> ThreatReport:
        """Analyze system architecture for security threats."""
        threats = []

        for component_name, component_config in architecture.items():
            component_threats = self._analyze_component(
                component_name, component_config
            )
            threats.extend(component_threats)

        return ThreatReport(
            threats=threats,
            risk_matrix=self._generate_risk_matrix(threats),
            mitigations=self._generate_mitigations(threats),
            compliance_mapping=self._map_to_compliance(threats)
        )

    def _analyze_component(self, name: str, config: dict) -> List[Threat]:
        """Analyze individual component for STRIDE threats."""
        threats = []
        component_type = config.get('type')

        # Web Application Threats
        if component_type == 'web_application':
            threats.extend([
                Threat(
                    category='Spoofing',
                    severity='high',
                    description='Attacker impersonates legitimate user',
                    attack_vector='Session hijacking, CSRF, phishing',
                    mitigation=[
                        'Implement MFA with TOTP or hardware keys',
                        'Use CSRF tokens in all state-changing operations',
                        'Implement secure session management',
                        'Apply SameSite cookie attribute'
                    ],
                    owasp_mapping=['A07: Authentication Failures', 'A01: Broken Access Control']
                ),
                Threat(
                    category='Injection',
                    severity='critical',
                    description='SQL injection or command injection attacks',
                    attack_vector='Unvalidated user input in queries',
                    mitigation=[
                        'Use parameterized queries (prepared statements)',
                        'Apply input validation with allowlists',
                        'Implement WAF with SQL injection rules',
                        'Use ORM frameworks with query builders'
                    ],
                    owasp_mapping=['A03: Injection']
                )
            ])

        # API Threats
        elif component_type == 'rest_api':
            threats.extend([
                Threat(
                    category='Information Disclosure',
                    severity='high',
                    description='Sensitive data exposed through API',
                    attack_vector='Insecure direct object references, excessive data exposure',
                    mitigation=[
                        'Implement object-level authorization',
                        'Use data minimization (only return required fields)',
                        'Apply rate limiting to prevent enumeration',
                        'Implement API key rotation'
                    ],
                    owasp_mapping=['A01: Broken Access Control']
                )
            ])

        # Database Threats
        elif component_type == 'database':
            threats.extend([
                Threat(
                    category='Tampering',
                    severity='critical',
                    description='Unauthorized data modification',
                    attack_vector='SQL injection, privilege escalation',
                    mitigation=[
                        'Apply principle of least privilege for DB accounts',
                        'Enable audit logging for all data modifications',
                        'Encrypt data at rest (AES-256)',
                        'Implement row-level security'
                    ],
                    owasp_mapping=['A02: Cryptographic Failures', 'A03: Injection']
                )
            ])

        return threats
```

**Use Case**: Conducting threat modeling during design phase to identify security risks before implementation.

---

### Pattern 4: DevSecOps Pipeline Automation

**Key Concept**: Integrate security testing and validation into CI/CD pipeline for continuous security assurance.

**Security Testing Layers**:
```
1. Pre-Commit:   IDE plugins, linters
2. Commit:       Git hooks, secret detection
3. Build:        SAST, dependency scanning
4. Test:         DAST, IAST, penetration testing
5. Deploy:       Infrastructure scanning, compliance checks
6. Production:   Runtime monitoring, RASP
```

**Implementation**:
```yaml
# Complete DevSecOps pipeline with security gates
name: DevSecOps Security Pipeline

on: [push, pull_request]

jobs:
  security-scan:
    runs-on: ubuntu-latest
    steps:
      # Step 1: Secret Detection (A02: Cryptographic Failures)
      - name: Detect Secrets
        run: |
          detect-secrets scan --all-files --force-use-all-plugins \
            --baseline .secrets.baseline
        continue-on-error: false

      # Step 2: Static Application Security Testing (SAST)
      - name: SAST with Bandit (Python)
        run: |
          bandit -r src/ \
            -f json -o bandit-report.json \
            -ll  # Only high/medium severity
        continue-on-error: false

      # Step 3: Dependency Vulnerability Scanning (A06)
      - name: Dependency Scanning
        run: |
          pip-audit --format json --output pip-audit-report.json
          safety check --json --output safety-report.json
        continue-on-error: false

      # Step 4: Container Image Scanning
      - name: Scan Docker Image
        run: |
          trivy image --severity HIGH,CRITICAL \
            --format json --output trivy-report.json \
            myapp:latest
        continue-on-error: false

      # Step 5: Infrastructure as Code (IaC) Scanning
      - name: IaC Security Scan
        run: |
          checkov --directory . --output json --output-file checkov-report.json
        continue-on-error: false

      # Step 6: Dynamic Application Security Testing (DAST)
      - name: DAST with OWASP ZAP
        run: |
          docker run -v $(pwd):/zap/wrk:rw \
            owasp/zap2docker-stable \
            zap-baseline.py -t https://staging.myapp.com \
            -J zap-report.json
        continue-on-error: false

      # Step 7: Compliance Validation
      - name: Compliance Check (SOC 2, GDPR)
        run: |
          python scripts/compliance_validator.py \
            --frameworks soc2,gdpr,hipaa \
            --output compliance-report.json
        continue-on-error: false

      # Step 8: Security Report Aggregation
      - name: Aggregate Security Reports
        run: |
          python scripts/aggregate_security_reports.py \
            --reports bandit-report.json,pip-audit-report.json,trivy-report.json \
            --output security-summary.json

      # Step 9: Quality Gate Enforcement
      - name: Enforce Security Quality Gate
        run: |
          python scripts/security_quality_gate.py \
            --report security-summary.json \
            --max-critical 0 \
            --max-high 2
        continue-on-error: false
```

**Use Case**: Automated security enforcement in CI/CD preventing vulnerable code from reaching production.

---

### Pattern 5: Modern Cryptography Standards

**Key Concept**: Use proven, modern cryptographic algorithms for data protection at rest and in transit.

**Cryptography Standards** (2025):
```
Encryption:           AES-256-GCM (symmetric), RSA-2048+ (asymmetric)
Hashing:              bcrypt (passwords), SHA-256/SHA-3 (data integrity)
Key Derivation:       PBKDF2, Argon2id
Transport Security:   TLS 1.3+ (no TLS 1.2 or earlier)
Digital Signatures:   RSA-PSS, ECDSA (P-256, P-384)
Random Generation:    secrets module (not random)
```

**Implementation**:
```python
# Enterprise cryptography manager with key rotation
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa, padding
import bcrypt
import secrets

class EnterpriseCryptoManager:
    """Production-ready cryptography with key rotation."""

    def __init__(self, master_key: bytes):
        self.master_key = master_key
        self.cipher = AESGCM(master_key)

    # Data Encryption (A02: Cryptographic Failures)
    def encrypt_data(self, plaintext: str) -> dict:
        """Encrypt data with AES-256-GCM."""
        nonce = secrets.token_bytes(12)  # 96-bit nonce for GCM
        ciphertext = self.cipher.encrypt(
            nonce,
            plaintext.encode('utf-8'),
            associated_data=None
        )

        return {
            'ciphertext': ciphertext.hex(),
            'nonce': nonce.hex(),
            'algorithm': 'AES-256-GCM'
        }

    def decrypt_data(self, encrypted_data: dict) -> str:
        """Decrypt AES-256-GCM encrypted data."""
        nonce = bytes.fromhex(encrypted_data['nonce'])
        ciphertext = bytes.fromhex(encrypted_data['ciphertext'])

        plaintext = self.cipher.decrypt(nonce, ciphertext, associated_data=None)
        return plaintext.decode('utf-8')

    # Password Hashing (A02, A07)
    def hash_password(self, password: str, rounds: int = 12) -> str:
        """Hash password with bcrypt (2^12 = 4096 iterations)."""
        salt = bcrypt.gensalt(rounds=rounds)
        hashed = bcrypt.hashpw(password.encode('utf-8'), salt)
        return hashed.decode('utf-8')

    def verify_password(self, password: str, hashed: str) -> bool:
        """Verify password against bcrypt hash."""
        return bcrypt.checkpw(password.encode('utf-8'), hashed.encode('utf-8'))

    # Digital Signatures (A08: Software/Data Integrity)
    def generate_rsa_keypair(self, key_size: int = 2048) -> tuple:
        """Generate RSA key pair for digital signatures."""
        private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=key_size
        )
        public_key = private_key.public_key()

        return (private_key, public_key)

    def sign_data(self, data: bytes, private_key: rsa.RSAPrivateKey) -> bytes:
        """Sign data with RSA-PSS."""
        signature = private_key.sign(
            data,
            padding.PSS(
                mgf=padding.MGF1(hashes.SHA256()),
                salt_length=padding.PSS.MAX_LENGTH
            ),
            hashes.SHA256()
        )
        return signature

    def verify_signature(self, data: bytes, signature: bytes, public_key: rsa.RSAPublicKey) -> bool:
        """Verify RSA-PSS signature."""
        try:
            public_key.verify(
                signature,
                data,
                padding.PSS(
                    mgf=padding.MGF1(hashes.SHA256()),
                    salt_length=padding.PSS.MAX_LENGTH
                ),
                hashes.SHA256()
            )
            return True
        except Exception:
            return False
```

**Use Case**: Protecting sensitive data (PII, PHI, financial) with AES-256 encryption, securing passwords with bcrypt, and ensuring data integrity with digital signatures.

---

## 📖 Advanced Documentation

This Skill uses **Progressive Disclosure Architecture** for optimal learning. Core patterns above provide immediate value; detailed implementation strategies are modularized:

**Module Structure**:
- **[modules/owasp-compliance.md](modules/owasp-compliance.md)** - OWASP Top 10 2021 detailed patterns (A01-A10)
- **[modules/zero-trust-architecture.md](modules/zero-trust-architecture.md)** - Zero-trust implementation with adaptive authentication
- **[modules/threat-modeling.md](modules/threat-modeling.md)** - STRIDE, PASTA, LINDDUN methodologies with examples
- **[modules/devsecops-automation.md](modules/devsecops-automation.md)** - Security CI/CD integration patterns
- **[modules/cryptography-standards.md](modules/cryptography-standards.md)** - Encryption, hashing, key management
- **[modules/secure-coding-patterns.md](modules/secure-coding-patterns.md)** - Language-specific secure coding practices
- **[modules/access-control.md](modules/access-control.md)** - RBAC, ABAC, policy-based access control
- **[modules/reference.md](modules/reference.md)** - API reference, compliance checklists, tool guides

---

## 🎯 Security Implementation Workflow

**Standard Security Assessment Process**:

```
Step 1: Architecture Analysis
  ├─ Identify system components
  ├─ Map data flows
  └─ Document trust boundaries

Step 2: Threat Modeling
  ├─ Apply STRIDE methodology
  ├─ Identify attack vectors
  └─ Calculate risk scores

Step 3: Security Control Design
  ├─ Map threats to OWASP Top 10
  ├─ Design defense-in-depth layers
  └─ Apply zero-trust principles

Step 4: Implementation
  ├─ Apply secure coding patterns
  ├─ Implement cryptography standards
  └─ Build authentication/authorization

Step 5: Testing & Validation
  ├─ SAST, DAST, IAST scanning
  ├─ Penetration testing
  └─ Compliance validation

Step 6: Continuous Monitoring
  ├─ Deploy security monitoring
  ├─ Configure SIEM integration
  └─ Establish incident response
```

---

## 🔗 Context7 MCP Integration

**Latest Security Patterns** (2025):

This skill leverages Context7 MCP for real-time access to latest security standards and patterns:

```python
# Fetch latest OWASP patterns
owasp_patterns = await context7.get_library_docs(
    context7_library_id="/owasp/top-ten",
    topic="OWASP Top 10 2021 mitigation patterns vulnerability protection",
    tokens=5000
)

# Fetch latest zero-trust architectures
zerotrust_patterns = await context7.get_library_docs(
    context7_library_id="/nist/zero-trust",
    topic="zero-trust architecture adaptive authentication",
    tokens=4000
)

# Fetch cryptography best practices
crypto_patterns = await context7.get_library_docs(
    context7_library_id="/cryptography/hazmat",
    topic="AES-256 RSA bcrypt TLS 1.3 encryption standards",
    tokens=3000
)
```

**Relevant Libraries**:
| Library | Context7 ID | Use Case |
|---------|-------------|----------|
| OWASP Top 10 | `/owasp/top-ten` | Vulnerability patterns and mitigations |
| NIST Zero Trust | `/nist/zero-trust` | Zero-trust architecture guidance |
| Cryptography | `/cryptography/hazmat` | Python cryptography library patterns |
| OWASP ZAP | `/owasp/zap` | Dynamic security testing |
| Bandit | `/pycqa/bandit` | Python security linting |

---

## 🔗 Integration with Other Skills

**Security Ecosystem**:
- `moai-security-owasp` - OWASP compliance validation and testing
- `moai-security-identity` - Identity and access management (IAM)
- `moai-security-api` - API security patterns and best practices
- `moai-security-zero-trust` - Zero-trust architecture deep dive
- `moai-security-threat` - Advanced threat modeling techniques
- `moai-domain-cloud` - Cloud security patterns (AWS, GCP, Azure)
- `moai-domain-devops` - DevOps infrastructure security
- `moai-domain-backend` - Backend security patterns

---

## 📈 Best Practices

### ✅ DO
- Apply defense-in-depth with multiple security layers
- Use Context7 for latest vulnerability patterns (2025 standards)
- Implement zero-trust architecture for sensitive applications
- Conduct threat modeling before implementation
- Automate security testing in CI/CD pipelines
- Use modern cryptography (AES-256, TLS 1.3, bcrypt)
- Enforce least privilege access control (RBAC/ABAC)
- Monitor and log security events with SIEM integration
- Maintain security compliance (SOC 2, ISO 27001, GDPR)
- Regular security training and awareness programs

### ❌ DON'T
- Use deprecated cryptography (MD5, SHA1, DES, 3DES)
- Store secrets in code or version control
- Skip input validation and sanitization
- Ignore security scanning results
- Use outdated dependencies with known CVEs
- Apply security as an afterthought
- Trust internal network traffic (apply zero-trust)
- Expose detailed error messages to users
- Use weak password policies
- Neglect security logging and monitoring

---

## 📊 Success Metrics

**Security KPIs** (Enterprise Benchmarks):
- **Vulnerability Detection Rate**: 95%+ critical/high vulnerabilities detected
- **Mean Time to Remediate (MTTR)**: <7 days for critical, <30 days for high
- **Security Test Coverage**: ≥85% code covered by SAST/DAST
- **Incident Response Time**: <1 hour for critical incidents
- **Compliance Score**: 100% for mandatory controls (SOC 2, GDPR)
- **False Positive Rate**: <10% in automated security scans
- **Security Training Completion**: 100% of engineering team annually

---

## 📈 Version History

**2.0.0** (2025-11-24)
- 🔄 Complete restructuring with Progressive Disclosure architecture
- ✨ Enhanced metadata with auto-trigger keywords
- ✨ Context7 MCP integration for latest security patterns
- ✨ 5 comprehensive core patterns (OWASP, zero-trust, threat modeling, DevSecOps, crypto)
- ✨ CommonMark compatibility and TRUST 5 validation
- ✨ Consolidated redundant content across modules
- ✨ Updated cross-references and module organization
- ✨ Added success metrics and best practices

**1.1.0** (2025-11-23)
- 🔄 Refactored with Progressive Disclosure
- ✨ 5 Core Patterns highlighted
- ✨ Modularized advanced content

**1.0.0** (2025-11-12)
- ✨ Initial release
- ✨ OWASP Top 10 compliance
- ✨ Zero-trust architecture
- ✨ Threat modeling (STRIDE)

---

**Maintained by**: MoAI-ADK Security Team
**Domain**: Enterprise Security Architecture
**Status**: Production Ready (Enterprise)
**Generated with**: MoAI-ADK Skill Factory
**Enhanced with**: Context7 MCP Integration

---

**End of Core Skill** | See `modules/` for advanced patterns | Status: ✅ Optimized 2025-11-24
