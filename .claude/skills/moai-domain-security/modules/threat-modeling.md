# Threat Modeling Methodologies

STRIDE and PASTA frameworks for systematic threat identification.

## STRIDE Framework

### S: Spoofing of User Identity

**Definition**: Attacker impersonates legitimate user

**Attack Vectors**:
- Credentials theft through phishing
- Session hijacking via XSS
- Token forging and replay attacks
- Man-in-the-middle impersonation
- Social engineering attacks

**Real-World Example (Web Application)**:
```python
# Vulnerable: Weak authentication
def login(username, password):
    user = db.query("SELECT * FROM users WHERE username = ?", username)
    if user and user.password == password:  # ❌ Plain text comparison
        return create_session(user.id)

# Secure: Strong authentication with MFA
import bcrypt
import pyotp

def secure_login(username, password, totp_token):
    user = db.query("SELECT * FROM users WHERE username = ?", username)

    # Verify password hash
    if not user or not bcrypt.checkpw(password.encode(), user.password_hash):
        raise AuthenticationError("Invalid credentials")

    # Verify TOTP token (MFA)
    totp = pyotp.TOTP(user.totp_secret)
    if not totp.verify(totp_token):
        raise AuthenticationError("Invalid MFA token")

    return create_session(user.id)
```

**Mitigations**:
```python
# Strong authentication implementation
class SecureAuthenticationService:
    def __init__(self):
        self.bcrypt_rounds = 12
        self.session_timeout = timedelta(hours=1)

    def hash_password(self, password: str) -> str:
        salt = bcrypt.gensalt(rounds=self.bcrypt_rounds)
        return bcrypt.hashpw(password.encode(), salt).decode()

    def verify_password(self, password: str, hash: str) -> bool:
        return bcrypt.checkpw(password.encode(), hash.encode())

    def setup_mfa(self, user_id: str) -> dict:
        secret = pyotp.random_base32()
        totp = pyotp.TOTP(secret)

        return {
            'secret': secret,
            'qr_uri': totp.provisioning_uri(
                name=f'user_{user_id}',
                issuer_name='SecureApp'
            )
        }
```

### T: Tampering with Data

**Definition**: Attacker modifies data in transit or at rest

**Attack Vectors**:
- SQL injection attacks
- Man-in-the-middle data modification
- Form tampering (hidden field modification)
- Session data manipulation
- Database record alteration

**Real-World Example (API Endpoint)**:
```python
# Vulnerable: SQL injection
def get_user_orders(user_id: str):
    query = f"SELECT * FROM orders WHERE user_id = '{user_id}'"  # ❌ Concatenation
    return db.execute(query)
    # Attack: user_id = "1' OR '1'='1" exposes all orders

# Secure: Parameterized queries + HMAC
from cryptography.hazmat.primitives import hmac, hashes

def secure_get_user_orders(user_id: str, request_signature: str):
    # Verify request signature
    expected_signature = self._generate_signature(user_id)
    if not hmac.compare_digest(request_signature, expected_signature):
        raise TamperingError("Invalid request signature")

    # Parameterized query
    query = "SELECT * FROM orders WHERE user_id = ?"
    return db.execute(query, (user_id,))

def _generate_signature(self, data: str) -> str:
    h = hmac.HMAC(self.secret_key, hashes.SHA256())
    h.update(data.encode())
    return h.finalize().hex()
```

**Mitigations**:
```python
# Data integrity protection
- HTTPS/TLS 1.3 for all communications
- Parameterized SQL queries (never concatenate)
- Input validation and sanitization
- Digital signatures for critical data
- Checksum verification for file uploads
```

### R: Repudiation

**Definition**: User denies performing an action

**Attack Scenario**: User performs malicious transaction, then denies responsibility

**Mitigations**:
```python
import logging
import hashlib
from datetime import datetime

class AuditLogger:
    def __init__(self):
        self.logger = logging.getLogger('audit')
        self.logger.setLevel(logging.INFO)

    def log_action(self, user_id: str, action: str, resource: str,
                   ip_address: str, user_agent: str, data: dict = None):
        """
        Comprehensive audit logging with non-repudiation
        """
        timestamp = datetime.utcnow().isoformat()

        audit_entry = {
            'timestamp': timestamp,
            'user_id': user_id,
            'action': action,
            'resource': resource,
            'ip_address': ip_address,
            'user_agent': user_agent,
            'data_hash': self._hash_data(data) if data else None
        }

        # Generate entry signature
        audit_entry['signature'] = self._sign_entry(audit_entry)

        self.logger.info(json.dumps(audit_entry))

        # Store in tamper-proof log
        self.db.insert('audit_log', audit_entry)

    def _hash_data(self, data: dict) -> str:
        return hashlib.sha256(json.dumps(data, sort_keys=True).encode()).hexdigest()

    def _sign_entry(self, entry: dict) -> str:
        # Sign with server private key
        signature_data = f"{entry['timestamp']}:{entry['user_id']}:{entry['action']}"
        return self.signer.sign(signature_data.encode()).hex()
```

### I: Information Disclosure

**Definition**: Unauthorized data access or exposure

**Attack Vectors**:
- SQL injection exposing sensitive data
- Directory traversal attacks
- Verbose error messages revealing internals
- Debug information left in production
- Unencrypted sensitive data

**Real-World Example (API Response)**:
```python
# Vulnerable: Detailed error exposure
@app.route('/api/user/<user_id>')
def get_user(user_id):
    try:
        user = db.get_user(user_id)
        return jsonify(user)
    except Exception as e:
        return jsonify({'error': str(e)}), 500  # ❌ Exposes stack trace

# Secure: Minimal error disclosure
@app.route('/api/user/<user_id>')
def secure_get_user(user_id):
    try:
        # Validate user_id format
        if not re.match(r'^[a-zA-Z0-9]{8,}$', user_id):
            return jsonify({'error': 'Invalid user ID'}), 400

        user = db.get_user(user_id)

        # Filter sensitive fields
        safe_user = {
            'id': user.id,
            'username': user.username,
            'created_at': user.created_at
        }

        return jsonify(safe_user), 200

    except UserNotFoundError:
        return jsonify({'error': 'User not found'}), 404
    except Exception as e:
        # Log full error internally
        logger.error(f"Error retrieving user: {e}", exc_info=True)
        # Return generic message to client
        return jsonify({'error': 'Internal server error'}), 500
```

**Mitigations**:
```python
# Secure error handling
- Generic error messages to clients
- Encryption at rest (AES-256)
- Encryption in transit (TLS 1.3+)
- Strict access controls (RBAC)
- Secure logging without sensitive data
```

### D: Denial of Service

**Definition**: Service becomes unavailable to legitimate users

**Attack Vectors**:
- DDoS attacks (volumetric flooding)
- Resource exhaustion (CPU, memory, disk)
- Rate limit bypass
- Algorithmic complexity attacks
- Database connection pool exhaustion

**Real-World Example (Rate Limiting)**:
```python
from functools import wraps
from flask import request, abort
import redis

# Rate limiting implementation
class RateLimiter:
    def __init__(self, redis_client):
        self.redis = redis_client

    def limit(self, max_requests: int = 100, window_seconds: int = 60):
        def decorator(f):
            @wraps(f)
            def wrapped(*args, **kwargs):
                # Get client identifier
                client_id = self._get_client_id(request)
                key = f"rate_limit:{client_id}"

                # Increment request count
                current = self.redis.incr(key)

                # Set expiry on first request
                if current == 1:
                    self.redis.expire(key, window_seconds)

                # Check limit
                if current > max_requests:
                    abort(429, "Too many requests")

                return f(*args, **kwargs)
            return wrapped
        return decorator

    def _get_client_id(self, request) -> str:
        # Use API key if present, else IP address
        return request.headers.get('X-API-Key', request.remote_addr)

# Usage
@app.route('/api/data')
@rate_limiter.limit(max_requests=100, window_seconds=60)
def get_data():
    return jsonify({'data': 'sensitive information'})
```

**Mitigations**:
```python
# DoS protection strategies
- Rate limiting per user/IP
- DDoS protection (Cloudflare, AWS Shield)
- Resource monitoring and auto-scaling
- Graceful degradation
- Request size limits
- Timeout enforcement
```

### E: Elevation of Privilege

**Definition**: Attacker gains higher privilege level than authorized

**Attack Vectors**:
- Authentication bypass
- Authorization bypass (broken access control)
- Privilege escalation through code vulnerabilities
- SQL injection leading to admin access
- Insecure direct object references

**Real-World Example (Authorization Bypass)**:
```python
# Vulnerable: No authorization check
@app.route('/api/admin/users')
def list_all_users():
    users = db.query("SELECT * FROM users")  # ❌ Anyone can access
    return jsonify(users)

# Secure: Role-based access control (RBAC)
from functools import wraps
from flask import g, abort

def require_role(role: str):
    def decorator(f):
        @wraps(f)
        def wrapped(*args, **kwargs):
            if not hasattr(g, 'user'):
                abort(401, "Authentication required")

            if role not in g.user.roles:
                abort(403, "Insufficient permissions")

            return f(*args, **kwargs)
        return wrapped
    return decorator

@app.route('/api/admin/users')
@require_role('admin')
def secure_list_all_users():
    # Only accessible to admin role
    users = db.query("SELECT * FROM users")
    return jsonify(users)

# RBAC implementation
class RBACManager:
    def __init__(self):
        self.permissions = {
            'admin': ['read', 'write', 'delete', 'admin'],
            'editor': ['read', 'write'],
            'viewer': ['read']
        }

    def check_permission(self, user_roles: list, required_permission: str) -> bool:
        for role in user_roles:
            if required_permission in self.permissions.get(role, []):
                return True
        return False
```

---

## PASTA Framework

**Process**: 7-stage risk-centric threat modeling methodology

### Stage 1: Define Objectives

```python
class ThreatModelingObjectives:
    def define_scope(self):
        return {
            'application': 'E-commerce platform',
            'business_objectives': [
                'Protect customer payment information',
                'Maintain PCI DSS compliance',
                'Prevent fraud transactions',
                'Ensure 99.9% uptime'
            ],
            'components': ['API Gateway', 'Database', 'Frontend', 'Payment Processor'],
            'data_types': ['PII', 'Payment info', 'Session tokens', 'Order history'],
            'threat_categories': ['Financial fraud', 'Data breach', 'Service disruption'],
            'compliance_requirements': ['PCI DSS', 'GDPR', 'SOC 2']
        }
```

### Stage 2: Define Technical Scope

```python
class TechnicalScope:
    def enumerate_components(self):
        return {
            'frontend': {
                'technology': ['React 18', 'WebSocket'],
                'security_features': ['CSP headers', 'XSS protection']
            },
            'backend': {
                'technology': ['Node.js 20', 'Express', 'GraphQL'],
                'security_features': ['JWT authentication', 'Rate limiting']
            },
            'database': {
                'technology': ['PostgreSQL 16'],
                'security_features': ['Encrypted at rest', 'SSL connections']
            },
            'infrastructure': {
                'technology': ['AWS', 'CloudFront', 'WAF'],
                'security_features': ['DDoS protection', 'TLS 1.3']
            },
            'integrations': {
                'payment': ['Stripe API'],
                'logging': ['DataDog'],
                'monitoring': ['Sentry']
            }
        }
```

### Stage 3: Decompose Application

```python
class ApplicationDecomposition:
    def create_data_flow_diagram(self):
        """
        Data Flow Diagram (DFD) components
        """
        return {
            'external_entities': [
                {'name': 'Customer', 'trust_level': 'untrusted'},
                {'name': 'Admin', 'trust_level': 'trusted'},
                {'name': 'Payment Provider', 'trust_level': 'trusted'}
            ],
            'processes': [
                {
                    'name': 'Authentication Service',
                    'inputs': ['credentials', 'MFA token'],
                    'outputs': ['JWT token', 'session'],
                    'trust_boundary': 'internal'
                },
                {
                    'name': 'Order Processing',
                    'inputs': ['order data', 'payment info'],
                    'outputs': ['order confirmation', 'payment status'],
                    'trust_boundary': 'internal'
                }
            ],
            'data_stores': [
                {
                    'name': 'User Database',
                    'data_types': ['PII', 'credentials'],
                    'encryption': 'at rest (AES-256)'
                },
                {
                    'name': 'Order Database',
                    'data_types': ['order data', 'payment references'],
                    'encryption': 'at rest (AES-256)'
                }
            ],
            'data_flows': [
                {
                    'from': 'Customer',
                    'to': 'API Gateway',
                    'data': 'HTTP Request',
                    'protocol': 'HTTPS/TLS 1.3'
                },
                {
                    'from': 'API Gateway',
                    'to': 'Authentication Service',
                    'data': 'Validated Request',
                    'protocol': 'Internal RPC'
                },
                {
                    'from': 'Authentication Service',
                    'to': 'User Database',
                    'data': 'Query',
                    'protocol': 'PostgreSQL SSL'
                }
            ]
        }
```

### Stage 4: Analyze Threats

```python
class ThreatAnalysis:
    def identify_threats(self, component: str) -> list:
        """
        Threat identification using STRIDE per component
        """
        threats = []

        if component == 'web_frontend':
            threats.extend([
                {
                    'type': 'XSS (Cross-Site Scripting)',
                    'severity': 'High',
                    'stride_category': 'Tampering, Information Disclosure',
                    'attack_vector': 'Injecting malicious JavaScript into user input',
                    'impact': 'Session hijacking, credential theft',
                    'likelihood': 'Medium'
                },
                {
                    'type': 'CSRF (Cross-Site Request Forgery)',
                    'severity': 'Medium',
                    'stride_category': 'Elevation of Privilege',
                    'attack_vector': 'Forging requests from authenticated user',
                    'impact': 'Unauthorized actions on behalf of user',
                    'likelihood': 'Low'
                },
                {
                    'type': 'Clickjacking',
                    'severity': 'Medium',
                    'stride_category': 'Spoofing',
                    'attack_vector': 'Transparent iframe overlay',
                    'impact': 'Unintended user actions',
                    'likelihood': 'Low'
                }
            ])

        elif component == 'api_backend':
            threats.extend([
                {
                    'type': 'SQL Injection',
                    'severity': 'Critical',
                    'stride_category': 'Tampering, Information Disclosure',
                    'attack_vector': 'Malicious SQL in user input',
                    'impact': 'Full database compromise',
                    'likelihood': 'Medium'
                },
                {
                    'type': 'Broken Authentication',
                    'severity': 'Critical',
                    'stride_category': 'Spoofing, Elevation of Privilege',
                    'attack_vector': 'Weak password policy, session fixation',
                    'impact': 'Account takeover',
                    'likelihood': 'High'
                }
            ])

        return threats
```

### Stage 5: Vulnerability Analysis

```python
class VulnerabilityAnalysis:
    def assess_vulnerabilities(self, threat: dict) -> dict:
        """
        Analyze exploitability and impact
        """
        return {
            'threat': threat['type'],
            'vulnerable_components': self._identify_vulnerable_components(threat),
            'exploitability': self._calculate_exploitability(threat),
            'impact': self._calculate_impact(threat),
            'risk_score': self._calculate_risk_score(threat),
            'cvss_score': self._calculate_cvss(threat)
        }

    def _calculate_risk_score(self, threat: dict) -> int:
        """
        Risk = Likelihood × Impact (1-10 scale)
        """
        likelihood_map = {'Low': 3, 'Medium': 6, 'High': 9}
        impact_map = {'Low': 3, 'Medium': 6, 'High': 9, 'Critical': 10}

        likelihood = likelihood_map.get(threat.get('likelihood', 'Medium'), 6)
        impact = impact_map.get(threat.get('severity', 'Medium'), 6)

        return likelihood * impact  # 9-90 scale

    def _calculate_cvss(self, threat: dict) -> float:
        """
        CVSS v3.1 base score calculation
        """
        # Simplified CVSS calculation
        # Real implementation would use full CVSS metrics
        severity_to_cvss = {
            'Low': 3.5,
            'Medium': 6.5,
            'High': 8.5,
            'Critical': 9.8
        }
        return severity_to_cvss.get(threat.get('severity', 'Medium'), 6.5)
```

### Stage 6: Attack Modeling

```python
class AttackModeling:
    def create_attack_tree(self, target_asset: str) -> dict:
        """
        Build attack tree for target asset
        """
        return {
            'goal': f'Compromise {target_asset}',
            'attack_paths': [
                {
                    'path': 'SQL Injection',
                    'steps': [
                        'Identify input field',
                        'Test for SQL injection vulnerability',
                        'Extract database schema',
                        'Exfiltrate sensitive data'
                    ],
                    'effort': 'Low',
                    'skill_level': 'Intermediate',
                    'detection_probability': 'Medium'
                },
                {
                    'path': 'Credential Stuffing',
                    'steps': [
                        'Obtain leaked credentials',
                        'Automated login attempts',
                        'Bypass rate limiting',
                        'Access user accounts'
                    ],
                    'effort': 'Low',
                    'skill_level': 'Beginner',
                    'detection_probability': 'High'
                }
            ]
        }
```

### Stage 7: Risk Mitigation

```python
class RiskMitigationStrategy:
    def prioritize_risks(self, threats: list) -> list:
        """
        Prioritize threats by risk score
        """
        scored_threats = []

        for threat in threats:
            risk_score = self._calculate_risk_score(threat)
            threat['risk_score'] = risk_score
            threat['priority'] = self._assign_priority(risk_score)
            scored_threats.append(threat)

        return sorted(scored_threats, key=lambda x: x['risk_score'], reverse=True)

    def generate_mitigation_plan(self, threat: dict) -> dict:
        """
        Generate mitigation strategy
        """
        return {
            'threat': threat['type'],
            'priority': threat['priority'],
            'mitigation_strategies': [
                {
                    'type': 'Preventive',
                    'actions': self._get_preventive_actions(threat),
                    'cost': 'Low',
                    'effort': '2 weeks'
                },
                {
                    'type': 'Detective',
                    'actions': self._get_detective_controls(threat),
                    'cost': 'Medium',
                    'effort': '1 week'
                },
                {
                    'type': 'Corrective',
                    'actions': self._get_corrective_actions(threat),
                    'cost': 'Low',
                    'effort': '3 days'
                }
            ],
            'residual_risk': self._calculate_residual_risk(threat)
        }

    def _assign_priority(self, risk_score: int) -> str:
        if risk_score >= 60:
            return 'Critical'
        elif risk_score >= 40:
            return 'High'
        elif risk_score >= 20:
            return 'Medium'
        else:
            return 'Low'
```

---

## Threat Modeling Output

### Threat Report Template

```python
class ThreatReport:
    def generate_report(self, threats: list) -> dict:
        """
        Comprehensive threat modeling report
        """
        return {
            'executive_summary': {
                'total_threats': len(threats),
                'critical_threats': len([t for t in threats if t.get('severity') == 'Critical']),
                'high_threats': len([t for t in threats if t.get('severity') == 'High']),
                'medium_threats': len([t for t in threats if t.get('severity') == 'Medium']),
                'low_threats': len([t for t in threats if t.get('severity') == 'Low']),
                'overall_risk_level': self._calculate_overall_risk(threats)
            },
            'threats_by_component': self._group_by_component(threats),
            'threats_by_stride_category': self._group_by_stride(threats),
            'mitigations': self._generate_mitigations(threats),
            'risk_matrix': self._create_risk_matrix(threats),
            'recommendations': self._generate_recommendations(threats)
        }

    def _create_risk_matrix(self, threats: list) -> dict:
        """
        Likelihood vs Impact matrix
        """
        matrix = {
            'High_Critical': [],
            'High_High': [],
            'High_Medium': [],
            'Medium_Critical': [],
            'Medium_High': [],
            'Medium_Medium': [],
            'Low_Critical': [],
            'Low_High': [],
            'Low_Medium': []
        }

        for threat in threats:
            key = f"{threat.get('likelihood', 'Medium')}_{threat.get('severity', 'Medium')}"
            if key in matrix:
                matrix[key].append(threat)

        return matrix
```

---

**Tools**: Microsoft Threat Modeling Tool, OWASP Threat Dragon, IriusRisk, Trike, CAIRIS
