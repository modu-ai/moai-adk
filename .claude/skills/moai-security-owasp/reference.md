# OWASP Top 10 2021: Complete Reference Guide

## OWASP Top 10 Vulnerabilities (Full Breakdown)

### A01:2021 - Broken Access Control

**Description**: Unauthorized access to resources or functions.

**Attack Scenarios**:
```python
# Vulnerable: Direct object reference
@app.route('/user/<user_id>')
def get_user(user_id):
    # No authorization check!
    user = db.query(f"SELECT * FROM users WHERE id = {user_id}")
    return user

# Secure: Proper authorization
@app.route('/user/<user_id>')
@require_authentication
def get_user(user_id):
    # Check authorization
    if not current_user.can_access_user(user_id):
        abort(403, "Access denied")

    user = db.query("SELECT * FROM users WHERE id = ?", [user_id])
    return user
```

**Prevention**:
- Implement role-based access control (RBAC)
- Deny by default (whitelist approach)
- Validate access for every request
- Log access control failures
- Rate limit API calls

**Checklist**:
- [ ] Authorization checks on all endpoints
- [ ] Proper session management
- [ ] CORS policy configured
- [ ] API rate limiting enabled
- [ ] Access logs monitored

---

### A02:2021 - Cryptographic Failures

**Description**: Weak or missing encryption of sensitive data.

**Attack Scenarios**:
```python
# Vulnerable: Storing passwords in plaintext
def register_user(username, password):
    db.execute("INSERT INTO users (username, password) VALUES (?, ?)",
               [username, password])  # BAD: Plaintext!

# Secure: Proper password hashing
import bcrypt

def register_user(username, password):
    # Hash password with bcrypt
    hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt(rounds=12))
    db.execute("INSERT INTO users (username, password_hash) VALUES (?, ?)",
               [username, hashed])
```

**Prevention**:
- Use strong encryption algorithms (AES-256)
- Hash passwords with bcrypt/argon2
- Use TLS 1.3+ for data in transit
- Encrypt sensitive data at rest
- Proper key management

**Checklist**:
- [ ] TLS 1.3+ enabled
- [ ] Strong cipher suites configured
- [ ] Passwords hashed with bcrypt (12+ rounds)
- [ ] Sensitive data encrypted at rest
- [ ] Keys stored in secure vaults

---

### A03:2021 - Injection

**Description**: Untrusted data sent to interpreter as part of command.

**Attack Scenarios**:
```python
# Vulnerable: SQL Injection
def get_user(username):
    query = f"SELECT * FROM users WHERE username = '{username}'"
    # Attacker input: ' OR '1'='1
    return db.execute(query)

# Secure: Parameterized query
def get_user(username):
    query = "SELECT * FROM users WHERE username = ?"
    return db.execute(query, [username])
```

**NoSQL Injection Example**:
```python
# Vulnerable: MongoDB injection
def find_user(username):
    return mongo.users.find_one({"username": username})
    # Attacker: {"$ne": null} bypasses authentication

# Secure: Input validation
def find_user(username):
    if not isinstance(username, str):
        raise ValueError("Invalid username type")

    # Sanitize input
    username = username.replace("$", "").replace("{", "").replace("}", "")

    return mongo.users.find_one({"username": username})
```

**Prevention**:
- Use parameterized queries
- Validate and sanitize all inputs
- Use ORM frameworks properly
- Principle of least privilege for DB users
- Input validation with whitelists

**Checklist**:
- [ ] All SQL queries parameterized
- [ ] Input validation on all fields
- [ ] ORM security best practices followed
- [ ] Database user has minimum permissions
- [ ] NoSQL injection protections in place

---

### A04:2021 - Insecure Design

**Description**: Missing or ineffective security controls in design.

**Attack Scenarios**:
```python
# Vulnerable: No rate limiting on password reset
@app.route('/reset-password', methods=['POST'])
def reset_password():
    email = request.json['email']
    send_reset_email(email)
    return {"message": "Reset email sent"}
    # Attacker can enumerate valid emails!

# Secure: Rate limiting + generic response
from flask_limiter import Limiter

limiter = Limiter(app, key_func=lambda: request.remote_addr)

@app.route('/reset-password', methods=['POST'])
@limiter.limit("5 per hour")
def reset_password():
    email = request.json['email']

    # Generic response (don't reveal if email exists)
    if is_valid_email(email):
        send_reset_email(email)

    return {"message": "If account exists, reset email was sent"}
```

**Prevention**:
- Threat modeling in design phase
- Secure design patterns
- Defense in depth
- Principle of least privilege
- Secure by default

**Checklist**:
- [ ] Threat model created
- [ ] Security requirements defined
- [ ] Rate limiting implemented
- [ ] Secure defaults configured
- [ ] Security review completed

---

### A05:2021 - Security Misconfiguration

**Description**: Missing security hardening or misconfigured settings.

**Attack Scenarios**:
```yaml
# Vulnerable: Debug mode enabled in production
DEBUG=true
SECRET_KEY=default_secret
ALLOWED_HOSTS=*

# Secure: Production configuration
DEBUG=false
SECRET_KEY=<random_256_bit_key>
ALLOWED_HOSTS=example.com,www.example.com
```

**Prevention**:
- Remove default accounts
- Disable directory listing
- Custom error pages
- Security headers configured
- Regular security updates

**Checklist**:
- [ ] Debug mode disabled
- [ ] Default credentials changed
- [ ] Security headers configured
- [ ] Unnecessary services disabled
- [ ] Error messages don't leak info

---

### A06:2021 - Vulnerable and Outdated Components

**Description**: Using components with known vulnerabilities.

**Attack Scenarios**:
```json
// Vulnerable: Outdated dependencies
{
  "dependencies": {
    "express": "4.16.0",  // Known vulnerability
    "lodash": "4.17.15",  // Known vulnerability
    "jquery": "3.3.0"     // Known vulnerability
  }
}

// Secure: Latest versions
{
  "dependencies": {
    "express": "^4.18.2",
    "lodash": "^4.17.21",
    "jquery": "^3.7.1"
  }
}
```

**Prevention**:
- Automated dependency scanning
- Regular updates
- Remove unused dependencies
- Only use trusted sources
- Monitor security advisories

**Checklist**:
- [ ] Automated vulnerability scanning
- [ ] Dependencies up to date
- [ ] No known CVEs in dependencies
- [ ] Unused dependencies removed
- [ ] Software Bill of Materials (SBOM) maintained

---

### A07:2021 - Identification and Authentication Failures

**Description**: Broken authentication and session management.

**Attack Scenarios**:
```python
# Vulnerable: Weak password policy
def register_user(password):
    if len(password) >= 6:  # Too weak!
        save_user(password)

# Secure: Strong password policy
import re

def register_user(password):
    if len(password) < 12:
        raise ValueError("Password must be at least 12 characters")

    if not re.search(r'[A-Z]', password):
        raise ValueError("Password must contain uppercase letter")

    if not re.search(r'[a-z]', password):
        raise ValueError("Password must contain lowercase letter")

    if not re.search(r'[0-9]', password):
        raise ValueError("Password must contain number")

    if not re.search(r'[!@#$%^&*]', password):
        raise ValueError("Password must contain special character")

    # Check against common passwords
    if password in COMMON_PASSWORDS:
        raise ValueError("Password too common")

    save_user(password)
```

**Prevention**:
- Multi-factor authentication (MFA)
- Strong password policies
- Secure session management
- Account lockout after failed attempts
- Password reset with proper verification

**Checklist**:
- [ ] MFA enabled for privileged accounts
- [ ] Strong password requirements
- [ ] Session timeout configured
- [ ] Account lockout after 5 failed attempts
- [ ] Secure password reset flow

---

### A08:2021 - Software and Data Integrity Failures

**Description**: Insufficient verification of code/data integrity.

**Attack Scenarios**:
```python
# Vulnerable: Deserializing untrusted data
import pickle

def load_user_data(user_input):
    # DANGEROUS: Pickle can execute arbitrary code
    data = pickle.loads(user_input)
    return data

# Secure: Use safe serialization
import json

def load_user_data(user_input):
    # JSON is safe (no code execution)
    data = json.loads(user_input)

    # Validate structure
    if not isinstance(data, dict):
        raise ValueError("Invalid data format")

    # Validate required fields
    required_fields = ['username', 'email']
    if not all(field in data for field in required_fields):
        raise ValueError("Missing required fields")

    return data
```

**Prevention**:
- Verify software signatures
- Use trusted repositories
- CI/CD pipeline security
- Code signing
- Integrity checks

**Checklist**:
- [ ] Dependencies verified with checksums
- [ ] Code signing implemented
- [ ] CI/CD pipeline secured
- [ ] No unsafe deserialization
- [ ] Integrity monitoring in place

---

### A09:2021 - Security Logging and Monitoring Failures

**Description**: Insufficient logging and monitoring.

**Attack Scenarios**:
```python
# Vulnerable: No logging of security events
@app.route('/login', methods=['POST'])
def login():
    username = request.json['username']
    password = request.json['password']

    if authenticate(username, password):
        return {"token": create_token(username)}
    else:
        return {"error": "Invalid credentials"}, 401
    # No logging of failed attempts!

# Secure: Comprehensive security logging
import logging

security_logger = logging.getLogger('security')

@app.route('/login', methods=['POST'])
def login():
    username = request.json['username']
    password = request.json['password']
    ip_address = request.remote_addr

    if authenticate(username, password):
        security_logger.info(f"Successful login: {username} from {ip_address}")
        return {"token": create_token(username)}
    else:
        security_logger.warning(
            f"Failed login attempt: {username} from {ip_address}"
        )

        # Check for brute force
        if count_failed_attempts(username, ip_address) > 5:
            security_logger.critical(
                f"Possible brute force: {username} from {ip_address}"
            )
            # Trigger alert
            send_security_alert(f"Brute force detected: {username}")

        return {"error": "Invalid credentials"}, 401
```

**Prevention**:
- Log all security events
- Centralized logging
- Real-time monitoring
- Automated alerts
- Regular log reviews

**Checklist**:
- [ ] Security events logged
- [ ] Centralized log management
- [ ] Real-time alerting configured
- [ ] Log retention policy
- [ ] Regular log reviews

---

### A10:2021 - Server-Side Request Forgery (SSRF)

**Description**: Fetching remote resource without validating URL.

**Attack Scenarios**:
```python
# Vulnerable: No URL validation
@app.route('/fetch-image')
def fetch_image():
    url = request.args.get('url')
    # Attacker: http://localhost:9200 (internal service)
    response = requests.get(url)
    return response.content

# Secure: Whitelist-based URL validation
from urllib.parse import urlparse

ALLOWED_DOMAINS = ['cdn.example.com', 'images.example.com']

@app.route('/fetch-image')
def fetch_image():
    url = request.args.get('url')

    # Parse URL
    parsed = urlparse(url)

    # Validate scheme
    if parsed.scheme not in ['http', 'https']:
        return {"error": "Invalid URL scheme"}, 400

    # Validate domain
    if parsed.hostname not in ALLOWED_DOMAINS:
        return {"error": "Domain not allowed"}, 403

    # Block internal IPs
    if is_internal_ip(parsed.hostname):
        return {"error": "Internal URLs not allowed"}, 403

    # Fetch with timeout
    response = requests.get(url, timeout=5)
    return response.content
```

**Prevention**:
- Whitelist allowed domains
- Block internal IPs
- Network segmentation
- Validate URL format
- Use allow lists, not deny lists

**Checklist**:
- [ ] URL validation implemented
- [ ] Internal IPs blocked
- [ ] Network segmentation configured
- [ ] Timeout on external requests
- [ ] Whitelist of allowed domains

---

## OWASP Top 10 Quick Reference Matrix

| Rank | Vulnerability | CVSS Avg | Prevalence | Impact |
|------|--------------|----------|------------|--------|
| A01 | Broken Access Control | 5.8 | 94% | High |
| A02 | Cryptographic Failures | 6.9 | 46% | High |
| A03 | Injection | 7.3 | 19% | Critical |
| A04 | Insecure Design | 5.9 | N/A | Medium |
| A05 | Security Misconfiguration | 5.5 | 90% | Medium |
| A06 | Vulnerable Components | 6.5 | 27% | Medium |
| A07 | Authentication Failures | 7.2 | 14% | High |
| A08 | Data Integrity Failures | 5.8 | 10% | Medium |
| A09 | Logging Failures | 4.5 | 53% | Medium |
| A10 | SSRF | 6.8 | 2.7% | High |

---

**Last Updated**: 2025-11-23
**Status**: Production Ready
**Lines**: 320
**Complete Coverage**: All OWASP Top 10 2021 with code examples
