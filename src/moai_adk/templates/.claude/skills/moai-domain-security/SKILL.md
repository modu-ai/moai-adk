---
name: moai-domain-security
version: 4.0.0
created: '2025-10-28'
updated: '2025-11-18'
status: stable
tier: domain
description: Enterprise-grade security expertise with production-ready patterns for
  OWASP Top 10 2021, zero-trust architecture, threat modeling (STRIDE, PASTA), secure
  SDLC, DevSecOps automation, cloud security, cryptography, identity & access management,
  and compliance frameworks (SOC 2, ISO 27001, GDPR, CCPA).
allowed-tools: Read, Bash, WebSearch, WebFetch, mcp__context7__resolve-library-id,
  mcp__context7__get-library-docs
primary-agent: security-expert
secondary-agents:
- qa-validator
- alfred
- doc-syncer
keywords:
- security
- owasp
- zero-trust
- devsecops
- threat-modeling
- cryptography
- compliance
- soc2
- iso27001
- gdpr
stability: stable
---


# moai-domain-security — Enterprise Security Architecture

**Enterprise Security Expertise & Implementation**

> **Primary Agent**: security-expert
> **Secondary Agents**: qa-validator, alfred, doc-syncer
> **Version**: 4.0.0

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference

Enterprise security expertise with **OWASP Top 10 2021** compliance and **zero-trust architecture**.

**Core Capabilities**:
- **Threat Modeling**: STRIDE, PASTA methodologies
- **Secure SDLC**: Security-by-design in development lifecycle
- **DevSecOps**: Security automation and CI/CD integration
- **Cloud Security**: AWS, Azure, GCP security patterns
- **Cryptography**: Encryption, hashing, digital signatures
- **Identity & Access Management**: OAuth, JWT, RBAC implementation

**When to Use**:
- ✅ Application security assessments and penetration testing
- ✅ Secure architecture design and threat modeling
- ✅ DevSecOps pipeline implementation
- ✅ Compliance frameworks (SOC 2, ISO 27001, GDPR)
- ✅ Cloud security hardening and monitoring

---

### Level 2: Practical Implementation

#### Pattern 1: OWASP Top 10 2021 Protection

**Objective**: Protect against the OWASP Top 10 2021 vulnerabilities.

```python
# Security middleware for Flask/Django applications
import re
from functools import wraps

class SecurityMiddleware:
    """OWASP Top 10 protection middleware."""
    
    def __init__(self, app=None):
        self.app = app
        if app:
            self.init_app(app)
    
    def init_app(self, app):
        app.before_request(self.before_request_handler)
        app.after_request(self.after_request_handler)
    
    def before_request_handler(self, request):
        # A01: Broken Access Control
        self._verify_access_control(request)
        # A03: Injection prevention
        self._prevent_injection_attacks(request)
        return None
    
    def after_request_handler(self, response):
        response.headers['X-Content-Type-Options'] = 'nosniff'
        response.headers['X-Frame-Options'] = 'DENY'
        response.headers['X-XSS-Protection'] = '1; mode=block'
        response.headers['Strict-Transport-Security'] = 'max-age=31536000'
        return response
    
    def _prevent_injection_attacks(self, request):
        if hasattr(request, 'form'):
            for key, value in request.form.items():
                if self._detect_sql_injection(value):
                    raise SecurityError("SQL injection attempt detected")
    
    def _detect_sql_injection(self, input_str: str) -> bool:
        patterns = [
            r"(\b(union|select|insert|update|delete|drop)\b)",
            r"([';]|--|/\*|\*/|xp_|sp_)",
            r"(or\s+1\s*=\s*1|and\s+1\s*=\s*1)",
        ]
        return any(re.search(pattern, input_str, re.IGNORECASE) for pattern in patterns)

# Example usage
from flask import Flask, request, jsonify

app = Flask(__name__)
security = SecurityMiddleware(app)

@app.route('/api/users/<int:user_id>')
def get_user(user_id):
    # Parameterized query prevents SQL injection
    query = "SELECT id, username, email FROM users WHERE id = %s"
    user = db.execute(query, (user_id,)).fetchone()
    
    if not user:
        return jsonify({'error': 'User not found'}), 404
    
    return jsonify({
        'id': user['id'],
        'username': user['username'],
        'email': user['email']
    })
```

---

#### Pattern 2: Zero-Trust Architecture

**Objective**: Implement zero-trust security principles.

```python
# Zero-Trust authentication and authorization
import jwt
import secrets
from datetime import datetime, timedelta

class ZeroTrustAuth:
    """Zero-trust authentication and authorization system."""
    
    def __init__(self, secret_key: str, token_expiry: int = 3600):
        self.secret_key = secret_key
        self.token_expiry = token_expiry
        self.active_sessions = {}
    
    def authenticate_user(self, credentials: dict, context: dict) -> dict:
        user = self._verify_credentials(credentials)
        if not user:
            raise AuthenticationError("Invalid credentials")
        
        risk_score = self._calculate_risk_score(user, context)
        trust_level = self._determine_trust_level(risk_score)
        
        token_claims = {
            'user_id': user['id'],
            'username': user['username'],
            'roles': user['roles'],
            'trust_level': trust_level,
            'session_id': secrets.token_urlsafe(32),
            'device_fingerprint': context.get('device_fingerprint'),
            'ip_address': context.get('ip_address'),
            'risk_score': risk_score,
            'exp': datetime.utcnow() + timedelta(seconds=self.token_expiry)
        }
        
        token = jwt.encode(token_claims, self.secret_key, algorithm='HS256')
        
        self.active_sessions[token_claims['session_id']] = {
            'user_id': user['id'],
            'created_at': datetime.utcnow(),
            'context': context,
            'risk_score': risk_score
        }
        
        return {
            'token': token,
            'session_id': token_claims['session_id'],
            'trust_level': trust_level,
            'expires_in': self.token_expiry
        }
    
    def _calculate_risk_score(self, user: dict, context: dict) -> int:
        risk_score = 0
        
        if self._is_unusual_location(user['id'], context.get('ip_address')):
            risk_score += 20
        
        if self._is_new_device(user['id'], context.get('device_fingerprint')):
            risk_score += 15
        
        if self._is_unusual_time(user['id']):
            risk_score += 10
        
        if self._is_unusual_behavior(user['id'], context):
            risk_score += 25
        
        return min(risk_score, 100)
    
    def _determine_trust_level(self, risk_score: int) -> str:
        if risk_score < 20:
            return 'high'
        elif risk_score < 50:
            return 'medium'
        else:
            return 'low'

# Decorator for zero-trust verification
def zero_trust_required(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        token = request.headers.get('Authorization')
        if not token:
            return jsonify({'error': 'Authorization required'}), 401
        
        try:
            auth_result = zero_trust.verify_request(
                token=token,
                request_context={
                    'ip_address': request.remote_addr,
                    'device_fingerprint': request.headers.get('User-Agent'),
                    'endpoint': request.endpoint,
                    'method': request.method
                }
            )
            
            request.auth_result = auth_result
            return f(*args, **kwargs)
            
        except AuthorizationError as e:
            return jsonify({'error': str(e)}), 401
    
    return decorated_function

@app.route('/api/sensitive-data')
@zero_trust_required
def get_sensitive_data():
    if 'read_sensitive_data' not in request.auth_result['permissions']:
        return jsonify({'error': 'Insufficient permissions'}), 403
    
    sensitive_data = get_data_for_user(request.auth_result['user_id'])
    
    audit_log.info(
        "Sensitive data accessed",
        user_id=request.auth_result['user_id'],
        trust_level=request.auth_result['trust_level'],
        risk_score=request.auth_result['risk_score']
    )
    
    return jsonify(sensitive_data)
```

---

#### Pattern 3: Threat Modeling (STRIDE)

**Objective**: Implement STRIDE threat modeling for system security.

```python
# STRIDE Threat Modeling Framework
from enum import Enum
from dataclasses import dataclass
from typing import List, Dict

class ThreatCategory(Enum):
    SPOOFING = "Spoofing"
    TAMPERING = "Tampering"
    REPUDIATION = "Repudiation"
    INFORMATION_DISCLOSURE = "Information Disclosure"
    DENIAL_OF_SERVICE = "Denial of Service"
    ELEVATION_OF_PRIVILEGE = "Elevation of Privilege"

@dataclass
class Threat:
    category: ThreatCategory
    description: str
    impact: str
    likelihood: str
    mitigation: List[str]
    affected_components: List[str]

class ThreatModelAnalyzer:
    """STRIDE threat modeling analyzer."""
    
    def analyze_system(self, system_architecture: Dict) -> List[Threat]:
        threats = []
        
        for component_name, component_config in system_architecture.items():
            component_threats = self._analyze_component(component_name, component_config)
            threats.extend(component_threats)
        
        return threats
    
    def _analyze_component(self, component_name: str, component_config: Dict) -> List[Threat]:
        threats = []
        component_type = component_config.get('type', '')
        
        if component_type == 'web_application':
            threats.extend([
                Threat(
                    category=ThreatCategory.SPOOFING,
                    description="Attacker impersonates legitimate user",
                    impact="High",
                    likelihood="Medium",
                    mitigation=[
                        "Implement strong authentication (MFA)",
                        "Use CSRF tokens",
                        "Implement proper session management"
                    ],
                    affected_components=[component_name]
                ),
                Threat(
                    category=ThreatCategory.INFORMATION_DISCLOSURE,
                    description="Sensitive data exposed through vulnerabilities",
                    impact="High",
                    likelihood="High",
                    mitigation=[
                        "Encrypt data at rest and in transit",
                        "Implement proper access controls",
                        "Use secure coding practices"
                    ],
                    affected_components=[component_name]
                )
            ])
        
        return threats
    
    def generate_threat_report(self, threats: List[Threat]) -> Dict:
        threats_by_category = {}
        for threat in threats:
            category = threat.category.value
            if category not in threats_by_category:
                threats_by_category[category] = []
            threats_by_category[category].append(threat)
        
        high_risk_threats = [
            threat for threat in threats
            if threat.impact == "High" and threat.likelihood in ["High", "Medium"]
        ]
        
        return {
            'total_threats': len(threats),
            'threats_by_category': threats_by_category,
            'high_risk_threats': len(high_risk_threats),
            'recommendations': self._generate_recommendations(threats)
        }
    
    def _generate_recommendations(self, threats: List[Threat]) -> List[str]:
        recommendations = []
        
        mitigations = set()
        for threat in threats:
            mitigations.update(threat.mitigation)
        
        priority_mitigations = [
            "Implement strong authentication (MFA)",
            "Encrypt data at rest and in transit",
            "Use parameterized queries",
            "Implement proper access controls",
            "Use secure coding practices"
        ]
        
        for mitigation in priority_mitigations:
            if mitigation in mitigations:
                recommendations.append(mitigation)
                mitigations.remove(mitigation)
        
        recommendations.extend(sorted(mitigations))
        return recommendations

# Example usage
system_architecture = {
    'web_application': {
        'type': 'web_application',
        'technologies': ['React', 'Node.js', 'Express'],
        'exposed': True
    },
    'api': {
        'type': 'api',
        'technologies': ['FastAPI', 'Python'],
        'exposed': True
    },
    'database': {
        'type': 'database',
        'technologies': ['PostgreSQL'],
        'exposed': False
    }
}

analyzer = ThreatModelAnalyzer()
threats = analyzer.analyze_system(system_architecture)
report = analyzer.generate_threat_report(threats)

print(f"Threat Analysis Report")
print(f"Total threats: {report['total_threats']}")
print(f"High-risk threats: {report['high_risk_threats']}")
print(f"Top recommendations: {report['recommendations'][:3]}")
```

---

### Level 3: Advanced Integration

#### DevSecOps Pipeline Integration

**Security automation in CI/CD pipeline**:

```yaml
# .github/workflows/security.yml
name: Security Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  security-scan:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Run security scan
      run: |
        # OWASP ZAP Baseline Scan
        docker run -t owasp/zap2docker-stable zap-baseline.py -t http://app-url
        
        # Dependency vulnerability scan
        pip install safety
        safety check --json --output safety-report.json
        
        # Static application security testing
        pip install bandit
        bandit -r src/ -f json -o bandit-report.json
    
    - name: Threat modeling
      run: |
        python threat_modeling.py --architecture architecture.json --output threat-model.json
    
    - name: Compliance check
      run: |
        python compliance_check.py --framework soc2
        python compliance_check.py --framework gdpr
```

---

## 🔗 Integration with Alfred Workflow

### Command Integration

**Security Assessment**:
- Use: Security expert agent for threat modeling
- Tools: STRIDE analysis, vulnerability scanning

**Compliance Validation**:
- Use: QA validation agent for compliance checks
- Tools: SOC 2, ISO 27001, GDPR validation

### Skill Dependencies

- `moai-domain-cloud`: Cloud security patterns
- `moai-core-dev-guide`: Secure development practices
- `moai-core-best-practices`: Security best practices

---

## 📚 Key Benefits

### For Development Teams

1. **Proactive Security**: Build security in from the start
2. **Compliance Ready**: Meet regulatory requirements automatically
3. **Threat Prevention**: Identify and mitigate threats early
4. **Continuous Monitoring**: Real-time security posture assessment

### For Organizations

1. **Risk Management**: Quantified risk assessment and mitigation
2. **Audit Trail**: Comprehensive security logging and monitoring
3. **Zero Trust**: Never trust, always verify security model
4. **Scalable Security**: Security that grows with your organization

---

## 📚 Research Attribution

**Security Research**: Based on OWASP Top 10 2021, NIST Cybersecurity Framework, and zero-trust architecture principles

**Compliance Frameworks**: SOC 2, ISO 27001, GDPR, CCPA implementation patterns

**Last Updated**: 2025-11-18

---

## 🔗 Related Resources

**Complete Security Patterns**: See `examples.md`
**Compliance Checklists**: See `reference.md`
**Threat Modeling Templates**: See examples.md

