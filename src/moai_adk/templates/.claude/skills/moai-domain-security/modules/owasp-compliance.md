# OWASP Top 10 2021 Compliance

Detailed patterns for protecting against the top 10 critical web vulnerabilities.

## A01: Broken Access Control

### Vulnerability Overview
- Unauthorized users access restricted resources
- Vertical privilege escalation (user → admin)
- Horizontal privilege escalation (user A → user B's data)

### Protection Pattern

```python
class AccessControlMiddleware:
    def verify_access_control(self, request, resource):
        # Verify user authentication
        user = self.get_current_user(request)
        if not user:
            raise UnauthorizedError("Authentication required")

        # Verify user authorization
        required_role = resource.get('required_role')
        if not user.has_role(required_role):
            raise ForbiddenError("Insufficient permissions")

        # Verify resource ownership for data operations
        if resource.is_user_data() and resource.owner_id != user.id:
            raise ForbiddenError("Cannot access other user's data")

        return True

# Usage
@app.route('/api/users/<int:user_id>/data')
def get_user_data(user_id):
    resource = {'owner_id': user_id, 'required_role': 'user'}
    access_control.verify_access_control(request, resource)
    return jsonify(fetch_user_data(user_id))
```

## A02: Cryptographic Failures

### Vulnerability Overview
- Unencrypted sensitive data transmission
- Weak encryption algorithms (DES, MD5)
- Missing encryption at rest

### Protection Pattern

```python
from cryptography.fernet import Fernet
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2
import ssl

class CryptographyManager:
    def __init__(self):
        self.cipher_suite = Fernet(Fernet.generate_key())

    def encrypt_at_rest(self, plaintext: str) -> str:
        encrypted = self.cipher_suite.encrypt(plaintext.encode())
        return encrypted.decode()

    def encrypt_in_transit(self, key: bytes, salt: bytes) -> bytes:
        # Use TLS 1.3+ for all data in transit
        kdf = PBKDF2(
            algorithm=hashes.SHA256(),
            length=32,
            salt=salt,
            iterations=100000,
        )
        return kdf.derive(key)

    def hash_password(self, password: str) -> str:
        import bcrypt
        salt = bcrypt.gensalt(rounds=12)
        return bcrypt.hashpw(password.encode(), salt).decode()

# Configuration
app.config['SESSION_COOKIE_SECURE'] = True
app.config['SESSION_COOKIE_HTTPONLY'] = True
app.config['SESSION_COOKIE_SAMESITE'] = 'Strict'
```

## A03: Injection

### Vulnerability Overview
- SQL injection through unsanitized input
- Command injection via system calls
- LDAP/XML injection attacks

### Protection Pattern

```python
import paramiko

class InjectionPrevention:
    def __init__(self, db):
        self.db = db

    def query_database_safe(self, query: str, params: tuple):
        # Parameterized queries prevent SQL injection
        cursor = self.db.connection.cursor()
        cursor.execute(query, params)  # Separate query and data
        return cursor.fetchall()

    def safe_command_execution(self, user_input: str) -> str:
        import subprocess
        import shlex

        # Avoid shell=True, use list of arguments
        safe_input = shlex.quote(user_input)
        result = subprocess.run(
            ['echo', safe_input],
            capture_output=True,
            text=True
        )
        return result.stdout

    def validate_input(self, user_input: str, pattern: str) -> bool:
        import re
        return bool(re.match(pattern, user_input))

# Usage
@app.route('/search')
def search():
    query = request.args.get('q')

    # Validate input against whitelist
    if not injection_prevention.validate_input(query, r'^[a-zA-Z0-9\s]+$'):
        return jsonify({'error': 'Invalid search query'}), 400

    # Use parameterized query
    results = injection_prevention.query_database_safe(
        'SELECT * FROM products WHERE name LIKE %s',
        ('%' + query + '%',)
    )
    return jsonify(results)
```

## A04: Insecure Design

### Vulnerability Overview
- Missing security controls in architecture
- Lack of threat modeling
- No secure design patterns

### Protection Pattern

```python
class SecureDesignPattern:
    def threat_model_analysis(self, components: dict) -> dict:
        # STRIDE threat modeling
        threats = {
            'Spoofing': [],
            'Tampering': [],
            'Repudiation': [],
            'Information Disclosure': [],
            'Denial of Service': [],
            'Elevation of Privilege': []
        }

        for component_name, config in components.items():
            # Analyze each component for threats
            if config['type'] == 'api_endpoint':
                threats['Spoofing'].append({
                    'component': component_name,
                    'mitigation': 'Implement authentication (OAuth 2.0)'
                })

        return threats

    def defense_in_depth(self) -> dict:
        return {
            'network': ['WAF', 'DDoS protection', 'Network segmentation'],
            'application': ['Input validation', 'Output encoding', 'CSRF tokens'],
            'data': ['Encryption at rest', 'Encryption in transit', 'Access controls'],
            'monitoring': ['Log aggregation', 'Anomaly detection', 'Alerting']
        }
```

## A05: Broken Authentication

### Vulnerability Overview
- Weak password policies
- Session fixation attacks
- Brute force attacks

### Protection Pattern

```python
from flask_limiter import Limiter
import secrets

class AuthenticationSecurity:
    def __init__(self, limiter):
        self.limiter = limiter

    @limiter.limit("5 per minute")  # Brute force protection
    def login(self, username: str, password: str):
        user = self.db.find_user(username)

        if not user:
            # Never reveal if user exists
            raise AuthenticationError("Invalid credentials")

        # Secure password comparison
        if not bcrypt.checkpw(password.encode(), user.password):
            raise AuthenticationError("Invalid credentials")

        # Generate secure session
        session_token = secrets.token_urlsafe(32)
        session = Session(
            user_id=user.id,
            token=session_token,
            created_at=datetime.utcnow(),
            expires_at=datetime.utcnow() + timedelta(hours=1)
        )
        self.db.save(session)

        return {'token': session_token, 'expires_in': 3600}

    def password_policy_validation(self, password: str) -> bool:
        import re

        patterns = [
            (r'[A-Z]', 'at least one uppercase letter'),
            (r'[a-z]', 'at least one lowercase letter'),
            (r'[0-9]', 'at least one digit'),
            (r'[!@#$%^&*]', 'at least one special character'),
        ]

        for pattern, requirement in patterns:
            if not re.search(pattern, password):
                return False

        return len(password) >= 12  # Minimum 12 characters
```

## A07: Cross-Site Scripting (XSS)

### Vulnerability Overview
- Unescaped user input in HTML
- JavaScript execution in templates
- DOM-based XSS

### Protection Pattern

```python
from markupsafe import escape
import bleach

class XSSPrevention:
    def output_encoding(self, user_input: str) -> str:
        # HTML encode user input
        return escape(user_input)

    def sanitize_html(self, html_content: str) -> str:
        # Allow only safe HTML tags
        allowed_tags = ['p', 'br', 'strong', 'em', 'a']
        allowed_attrs = {'a': ['href']}

        return bleach.clean(
            html_content,
            tags=allowed_tags,
            attributes=allowed_attrs
        )

    def csp_headers(self, response):
        # Content Security Policy
        response.headers['Content-Security-Policy'] = (
            "default-src 'self'; "
            "script-src 'self'; "
            "style-src 'self' 'unsafe-inline'; "
            "img-src 'self' data:; "
            "font-src 'self'"
        )
        return response
```

## Best Practices

### ✅ DO
- Use parameterized queries
- Validate all user input against whitelist
- Encode output based on context (HTML, JavaScript, URL)
- Implement rate limiting
- Use security headers (CSP, HSTS, X-Frame-Options)
- Enable HTTPS everywhere
- Implement session timeout
- Use secure password hashing (bcrypt, scrypt)

### ❌ DON'T
- Trust client-side validation
- Use weak encryption
- Hardcode secrets
- Log sensitive data
- Expose detailed error messages
- Use MD5 or SHA-1
- Trust user input
- Share session tokens

---

**Tools**: OWASP ZAP, Burp Suite, SonarQube
