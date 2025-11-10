# Advanced Security Guide

A comprehensive guide to strengthening the security of MoAI-ADK projects following industry best practices and OWASP standards.

## Table of Contents

- [Security Principles](#security-principles)
- [OWASP Top 10 Response](#owasp-top-10-response)
- [Security Checklists](#security-checklists)
- [Encryption & Secrets Management](#encryption-and-secrets-management)
- [Vulnerability Scanning](#vulnerability-scanning)
- [Security Policies](#security-policies)
- [Incident Response](#incident-response)

______________________________________________________________________

## Security Principles

### Principle 1: Secure by Default

**Philosophy**: All systems should be secure in their default state

**Implementation**:
```python
# ❌ Insecure default
class Config:
    DEBUG = True  # Dangerous in production
    SECRET_KEY = "default"  # Predictable
    ALLOWED_HOSTS = ["*"]  # Too permissive

# ✅ Secure default
class Config:
    DEBUG = os.getenv("DEBUG", "False").lower() == "true"
    SECRET_KEY = os.getenv("SECRET_KEY")  # Required
    ALLOWED_HOSTS = os.getenv("ALLOWED_HOSTS", "").split(",")
    
    def __init__(self):
        if not self.SECRET_KEY:
            raise ValueError("SECRET_KEY must be set")
```

### Principle 2: Input Validation

**Philosophy**: Never trust user input

**Implementation**:
```python
from pydantic import BaseModel, validator, constr

class UserInput(BaseModel):
    username: constr(min_length=3, max_length=20)
    email: str
    age: int
    
    @validator('email')
    def validate_email(cls, v):
        import re
        pattern = r'^[\w\.-]+@[\w\.-]+\.\w+$'
        if not re.match(pattern, v):
            raise ValueError('Invalid email format')
        return v
    
    @validator('age')
    def validate_age(cls, v):
        if not (0 <= v <= 150):
            raise ValueError('Age must be between 0 and 150')
        return v

# Usage
try:
    user = UserInput(
        username="john",
        email="john@example.com",
        age=25
    )
except ValueError as e:
    print(f"Validation error: {e}")
```

### Principle 3: Least Privilege

**Philosophy**: Grant minimum necessary permissions

**Implementation**:
```python
# Role-Based Access Control (RBAC)
class Permission:
    READ = "read"
    WRITE = "write"
    DELETE = "delete"
    ADMIN = "admin"

class Role:
    VIEWER = {Permission.READ}
    EDITOR = {Permission.READ, Permission.WRITE}
    ADMIN = {Permission.READ, Permission.WRITE, Permission.DELETE, Permission.ADMIN}

def require_permission(permission):
    def decorator(func):
        def wrapper(user, *args, **kwargs):
            if permission not in user.permissions:
                raise PermissionError(f"Requires {permission}")
            return func(user, *args, **kwargs)
        return wrapper
    return decorator

@require_permission(Permission.WRITE)
def edit_document(user, doc_id):
    # Only users with WRITE permission can execute
    pass
```

### Principle 4: Defense in Depth

**Philosophy**: Multiple layers of security

**Layers**:
```
Layer 1: Network Security
  ├─ Firewall rules
  ├─ VPN/Private network
  └─ DDoS protection

Layer 2: Application Security
  ├─ Input validation
  ├─ Authentication
  └─ Authorization

Layer 3: Data Security
  ├─ Encryption at rest
  ├─ Encryption in transit
  └─ Database security

Layer 4: Monitoring & Response
  ├─ Intrusion detection
  ├─ Audit logging
  └─ Incident response
```

______________________________________________________________________

## OWASP Top 10 Response

### 1. Broken Access Control

**Risk**: Users access resources they shouldn't

**Prevention**:
```python
# ✅ Server-side access control
from functools import wraps
from flask import session, abort

def require_ownership(func):
    @wraps(func)
    def wrapper(resource_id, *args, **kwargs):
        resource = Resource.get(resource_id)
        if resource.owner_id != session.get('user_id'):
            abort(403)  # Forbidden
        return func(resource_id, *args, **kwargs)
    return wrapper

@app.route('/api/documents/<int:doc_id>', methods=['DELETE'])
@require_ownership
def delete_document(doc_id):
    # Only owner can delete
    Document.delete(doc_id)
    return {'status': 'deleted'}
```

**Testing**:
```python
def test_access_control():
    # User A creates document
    doc = create_document(owner_id=1)
    
    # User B tries to delete (should fail)
    with pytest.raises(PermissionError):
        delete_document(doc.id, user_id=2)
    
    # User A deletes (should succeed)
    delete_document(doc.id, user_id=1)
```

### 2. Cryptographic Failures

**Risk**: Sensitive data exposed due to weak/missing encryption

**Prevention**:
```python
from cryptography.fernet import Fernet
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2
import base64

class SecureStorage:
    def __init__(self, password: bytes):
        # Derive key from password
        kdf = PBKDF2(
            algorithm=hashes.SHA256(),
            length=32,
            salt=b'secure_salt_16_b',  # Should be random in production
            iterations=100000,
        )
        key = base64.urlsafe_b64encode(kdf.derive(password))
        self.cipher = Fernet(key)
    
    def encrypt(self, data: str) -> str:
        return self.cipher.encrypt(data.encode()).decode()
    
    def decrypt(self, encrypted: str) -> str:
        return self.cipher.decrypt(encrypted.encode()).decode()

# Usage
storage = SecureStorage(b"strong_password")
encrypted = storage.encrypt("sensitive data")
# Store encrypted in database

# Later...
decrypted = storage.decrypt(encrypted)
```

**Password Hashing**:
```python
from argon2 import PasswordHasher

ph = PasswordHasher()

# Hash password
password_hash = ph.hash("user_password_123")
# Store hash in database

# Verify password
try:
    ph.verify(password_hash, "user_password_123")
    print("Password correct")
except:
    print("Invalid password")

# Update hash if needed (automatic rehashing)
if ph.check_needs_rehash(password_hash):
    new_hash = ph.hash("user_password_123")
```

### 3. Injection Attacks

**Risk**: SQL, Command, LDAP injection

**SQL Injection Prevention**:
```python
# ❌ Vulnerable
def get_user(email):
    query = f"SELECT * FROM users WHERE email = '{email}'"
    return db.execute(query)
    # Attacker input: ' OR '1'='1

# ✅ Safe (parameterized queries)
def get_user(email):
    query = "SELECT * FROM users WHERE email = ?"
    return db.execute(query, [email])

# ✅ Safe (ORM)
from sqlalchemy import select
def get_user(email):
    return session.execute(
        select(User).where(User.email == email)
    ).scalar_one_or_none()
```

**Command Injection Prevention**:
```python
import subprocess

# ❌ Vulnerable
def ping_host(host):
    os.system(f"ping -c 1 {host}")
    # Attacker input: "example.com; rm -rf /"

# ✅ Safe
def ping_host(host):
    # Validate input
    import re
    if not re.match(r'^[\w\.-]+$', host):
        raise ValueError("Invalid host format")
    
    # Use subprocess with list (no shell)
    subprocess.run(['ping', '-c', '1', host], check=True)
```

### 4. Insecure Design

**Risk**: Flawed architecture enabling attacks

**Example: Rate Limiting**:
```python
from flask_limiter import Limiter
from flask_limiter.util import get_remote_address

limiter = Limiter(
    app,
    key_func=get_remote_address,
    default_limits=["200 per day", "50 per hour"]
)

@app.route('/api/login', methods=['POST'])
@limiter.limit("5 per minute")  # Prevent brute force
def login():
    username = request.json['username']
    password = request.json['password']
    # Login logic
```

**Example: Security Headers**:
```python
from flask import Flask

app = Flask(__name__)

@app.after_request
def set_security_headers(response):
    response.headers['X-Content-Type-Options'] = 'nosniff'
    response.headers['X-Frame-Options'] = 'DENY'
    response.headers['X-XSS-Protection'] = '1; mode=block'
    response.headers['Strict-Transport-Security'] = 'max-age=31536000; includeSubDomains'
    response.headers['Content-Security-Policy'] = "default-src 'self'"
    return response
```

### 5. Security Misconfiguration

**Risk**: Default/insecure configurations

**Secure Configuration**:
```python
# production_config.py
import os

class ProductionConfig:
    # Security
    SECRET_KEY = os.getenv('SECRET_KEY')  # Must be set
    DEBUG = False  # Never True in production
    TESTING = False
    
    # HTTPS
    SESSION_COOKIE_SECURE = True
    SESSION_COOKIE_HTTPONLY = True
    SESSION_COOKIE_SAMESITE = 'Lax'
    
    # CORS
    CORS_ORIGINS = ['https://app.example.com']
    
    # Database
    SQLALCHEMY_DATABASE_URI = os.getenv('DATABASE_URL')
    SQLALCHEMY_ECHO = False  # Don't log SQL in production
    
    # Logging
    LOG_LEVEL = 'WARNING'  # Don't log sensitive info
    
    def validate(self):
        if not self.SECRET_KEY:
            raise ValueError("SECRET_KEY must be set")
        if len(self.SECRET_KEY) < 32:
            raise ValueError("SECRET_KEY must be at least 32 characters")
```

### 6. Vulnerable and Outdated Components

**Risk**: Using components with known vulnerabilities

**Prevention**:
```bash
# Scan dependencies
pip-audit

# Example output:
# Found 2 vulnerabilities
# package-a 1.0.0 → 1.0.1 (CVE-2024-12345)
# package-b 2.0.0 → 2.1.0 (CVE-2024-67890)

# Update
uv pip install --upgrade package-a package-b

# Automate with Dependabot
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "pip"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10
```

### 7. Identification and Authentication Failures

**Risk**: Weak authentication allowing account takeover

**Secure Authentication**:
```python
import jwt
from datetime import datetime, timedelta
from werkzeug.security import generate_password_hash, check_password_hash

class AuthService:
    SECRET_KEY = "your-secret-key"
    
    @staticmethod
    def hash_password(password):
        # Check password strength
        if len(password) < 12:
            raise ValueError("Password must be at least 12 characters")
        if not any(c.isupper() for c in password):
            raise ValueError("Password must contain uppercase")
        if not any(c.islower() for c in password):
            raise ValueError("Password must contain lowercase")
        if not any(c.isdigit() for c in password):
            raise ValueError("Password must contain digit")
        
        return generate_password_hash(password, method='pbkdf2:sha256')
    
    @staticmethod
    def verify_password(password_hash, password):
        return check_password_hash(password_hash, password)
    
    @staticmethod
    def generate_token(user_id):
        payload = {
            'user_id': user_id,
            'exp': datetime.utcnow() + timedelta(hours=1),
            'iat': datetime.utcnow()
        }
        return jwt.encode(payload, AuthService.SECRET_KEY, algorithm='HS256')
    
    @staticmethod
    def verify_token(token):
        try:
            payload = jwt.decode(token, AuthService.SECRET_KEY, algorithms=['HS256'])
            return payload['user_id']
        except jwt.ExpiredSignatureError:
            raise ValueError("Token expired")
        except jwt.InvalidTokenError:
            raise ValueError("Invalid token")
```

### 8. Software and Data Integrity Failures

**Risk**: Unsigned updates, insecure CI/CD

**Prevention**:
```python
# Verify package integrity
import hashlib

def verify_file_hash(filepath, expected_hash):
    sha256_hash = hashlib.sha256()
    with open(filepath, "rb") as f:
        for byte_block in iter(lambda: f.read(4096), b""):
            sha256_hash.update(byte_block)
    
    actual_hash = sha256_hash.hexdigest()
    if actual_hash != expected_hash:
        raise ValueError(f"Hash mismatch: {actual_hash} != {expected_hash}")

# Verify before loading
verify_file_hash('update.zip', 'abc123...')
```

**Secure CI/CD**:
```yaml
# .github/workflows/security.yml
name: Security Checks

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Dependency scan
        run: |
          pip install pip-audit
          pip-audit
      
      - name: SAST scan
        run: |
          pip install bandit
          bandit -r src/
      
      - name: Secret scan
        run: |
          pip install detect-secrets
          detect-secrets scan
```

### 9. Security Logging and Monitoring Failures

**Risk**: Attacks go undetected

**Secure Logging**:
```python
import logging
from datetime import datetime

class SecurityLogger:
    def __init__(self):
        self.logger = logging.getLogger('security')
        handler = logging.FileHandler('security.log')
        formatter = logging.Formatter(
            '%(asctime)s - %(levelname)s - %(message)s'
        )
        handler.setFormatter(formatter)
        self.logger.addHandler(handler)
        self.logger.setLevel(logging.INFO)
    
    def log_login_attempt(self, username, success, ip_address):
        self.logger.info(
            f"Login attempt: user={username}, success={success}, ip={ip_address}"
        )
    
    def log_access_denied(self, user_id, resource, action):
        self.logger.warning(
            f"Access denied: user={user_id}, resource={resource}, action={action}"
        )
    
    def log_security_event(self, event_type, details):
        self.logger.critical(
            f"Security event: type={event_type}, details={details}"
        )

# Usage
security_log = SecurityLogger()
security_log.log_login_attempt('admin', False, '192.168.1.100')
```

### 10. Server-Side Request Forgery (SSRF)

**Risk**: Attacker makes server request internal resources

**Prevention**:
```python
import requests
from urllib.parse import urlparse

class SafeHTTPClient:
    ALLOWED_SCHEMES = ['http', 'https']
    BLOCKED_IPS = [
        '127.0.0.1',
        'localhost',
        '0.0.0.0',
        # Add private IP ranges
        '10.0.0.0/8',
        '172.16.0.0/12',
        '192.168.0.0/16'
    ]
    
    @staticmethod
    def is_safe_url(url):
        parsed = urlparse(url)
        
        # Check scheme
        if parsed.scheme not in SafeHTTPClient.ALLOWED_SCHEMES:
            return False
        
        # Check hostname
        hostname = parsed.hostname
        if not hostname:
            return False
        
        # Block private IPs (simplified)
        if hostname in SafeHTTPClient.BLOCKED_IPS:
            return False
        
        return True
    
    @staticmethod
    def fetch(url, timeout=5):
        if not SafeHTTPClient.is_safe_url(url):
            raise ValueError(f"Unsafe URL: {url}")
        
        return requests.get(url, timeout=timeout)

# Usage
try:
    response = SafeHTTPClient.fetch('https://api.example.com/data')
except ValueError as e:
    print(f"Blocked: {e}")
```

______________________________________________________________________

## Security Checklists

### Development Phase Checklist

```
Development Security Review
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Input Validation
  [ ] All user inputs validated
  [ ] Type checking enforced
  [ ] Length limits applied
  [ ] Pattern matching for formats (email, phone, etc.)

Authentication & Authorization
  [ ] Password hashing (Argon2/bcrypt)
  [ ] Strong password policy enforced
  [ ] JWT tokens with expiration
  [ ] Role-based access control (RBAC)
  [ ] Session management secure

Data Protection
  [ ] Sensitive data encrypted at rest
  [ ] HTTPS enforced for transit
  [ ] Secrets not in code/config
  [ ] No sensitive data in logs

Error Handling
  [ ] Generic error messages to users
  [ ] Detailed errors logged securely
  [ ] Stack traces never exposed
  [ ] Error codes documented

Code Quality
  [ ] No hardcoded credentials
  [ ] Parameterized queries (no SQL injection)
  [ ] Command injection prevention
  [ ] XSS prevention (output encoding)
```

### Deployment Phase Checklist

```
Deployment Security Review
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Infrastructure
  [ ] HTTPS enabled (TLS 1.3)
  [ ] Security headers configured
  [ ] CORS policy set
  [ ] Rate limiting enabled
  [ ] DDoS protection active

Configuration
  [ ] DEBUG=False in production
  [ ] Secrets in environment variables
  [ ] Database credentials secured
  [ ] API keys rotated
  [ ] Firewall rules configured

Dependencies
  [ ] Vulnerability scan passed (pip-audit)
  [ ] All packages up to date
  [ ] Dependabot enabled
  [ ] License compliance checked

Monitoring
  [ ] Security logging enabled
  [ ] Intrusion detection configured
  [ ] Alert system active
  [ ] Backup system verified
```

### Operations Phase Checklist

```
Ongoing Security Maintenance
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Regular Tasks (Weekly)
  [ ] Review security logs
  [ ] Check failed login attempts
  [ ] Monitor resource usage
  [ ] Review access logs

Regular Tasks (Monthly)
  [ ] Dependency updates
  [ ] Security patch application
  [ ] Access control review
  [ ] Backup restoration test

Regular Tasks (Quarterly)
  [ ] Security audit
  [ ] Penetration testing
  [ ] Incident response drill
  [ ] Security training
```

______________________________________________________________________

## Encryption and Secrets Management

### Environment Variables

```bash
# .env (NEVER commit to git)
SECRET_KEY=abc123...
DATABASE_URL=postgresql://user:pass@localhost/db
API_KEY=xyz789...

# Load in application
from dotenv import load_dotenv
import os

load_dotenv()

SECRET_KEY = os.getenv('SECRET_KEY')
if not SECRET_KEY:
    raise ValueError("SECRET_KEY must be set")
```

### Secrets in Production

```python
# Use cloud secret management
from google.cloud import secretmanager

def get_secret(secret_id):
    client = secretmanager.SecretManagerServiceClient()
    name = f"projects/{PROJECT_ID}/secrets/{secret_id}/versions/latest"
    response = client.access_secret_version(request={"name": name})
    return response.payload.data.decode('UTF-8')

# Usage
DATABASE_URL = get_secret('database-url')
```

______________________________________________________________________

## Vulnerability Scanning

### Automated Scans

```bash
# Python dependency scan
pip install pip-audit
pip-audit

# SAST (Static Application Security Testing)
pip install bandit
bandit -r src/

# Secret scanning
pip install detect-secrets
detect-secrets scan

# License compliance
pip install pip-licenses
pip-licenses
```

### Continuous Monitoring

```yaml
# .github/workflows/security-scan.yml
name: Security Scan
on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly
  push:
    branches: [main, develop]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Security scan
        run: |
          pip install pip-audit bandit
          pip-audit || true
          bandit -r src/ -f json -o bandit-report.json
      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: security-reports
          path: |
            bandit-report.json
```

______________________________________________________________________

## Security Policies

### Password Policy

```
Password Requirements
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Length: Minimum 12 characters
Complexity:
  - At least 1 uppercase letter
  - At least 1 lowercase letter
  - At least 1 number
  - At least 1 special character

Expiration: 90 days
History: Cannot reuse last 5 passwords
Lockout: 5 failed attempts = 15-minute lockout
```

### Access Control Policy

```
RBAC Matrix
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Role        Read    Write   Delete  Admin
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Viewer      ✓       ✗       ✗       ✗
Editor      ✓       ✓       ✗       ✗
Manager     ✓       ✓       ✓       ✗
Admin       ✓       ✓       ✓       ✓
```

______________________________________________________________________

## Incident Response

### Incident Response Plan

```
Phase 1: Detection
  1. Monitor alerts
  2. Investigate anomalies
  3. Confirm incident

Phase 2: Containment
  1. Isolate affected systems
  2. Preserve evidence
  3. Prevent spread

Phase 3: Eradication
  1. Remove malware/attacker access
  2. Patch vulnerabilities
  3. Update credentials

Phase 4: Recovery
  1. Restore from backup
  2. Verify system integrity
  3. Resume operations

Phase 5: Post-Incident
  1. Document lessons learned
  2. Update security measures
  3. Train team
```

______________________________________________________________________

**Next Steps**:
- [Extension Guide](extensions.md) - Customize security
- [Performance Optimization](performance.md) - Secure and fast
- [Architecture Guide](architecture.md) - Security by design
