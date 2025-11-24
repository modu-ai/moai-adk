---
name: moai-quality-security
description: Security consolidating OWASP, auth, API security, encryption, compliance, threat modeling, and dependency scanning
version: 1.0.0
modularized: true
last_updated: 2025-11-24
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - Read
  - Write
  - Edit
compliance_score: 88
modules:
  - owasp-validation
  - authentication-authorization
  - api-security
  - encryption-secrets
  - compliance-standards
dependencies:
  - moai-foundation-trust
  - moai-lang-python
  - moai-lang-typescript
deprecated: false
successor: null
category_tier: 3
auto_trigger_keywords:
  - security
  - owasp
  - vulnerability
  - authentication
  - authorization
  - encryption
  - ssl
  - tls
  - secrets
  - compliance
  - privacy
  - injection
  - xss
  - csrf
agent_coverage:
  - security-expert
  - quality-gate
context7_references:
  - owasp
  - cwe
  - cvss
invocation_api_version: "1.0"
---

## Quick Reference (30 seconds)

**Enterprise Security Consolidated**

Unified security framework consolidating 12 security skills (OWASP, auth/authz, API security, encryption, secrets, compliance, threat modeling, injection prevention, XSS/CSRF protection, data protection, and dependency scanning) into single comprehensive skill.

**Core Capabilities**:
- ✅ OWASP Top 10 validation (injection, auth, sensitive data, XXE, broken access, CSRF, XSS, deserialization, components, logging)
- ✅ Authentication & Authorization (OAuth2, JWT, RBAC, ABAC)
- ✅ API Security (rate limiting, API keys, token validation, CORS, HTTPS)
- ✅ Encryption & Secrets (AES, RSA, TLS 1.3, secrets management)
- ✅ Dependency Scanning (CVE detection, supply chain security)
- ✅ Compliance (GDPR, HIPAA, SOC2, PCI-DSS, WCAG)
- ✅ Threat Modeling (STRIDE, attack trees)

**When to Use**:
- Implementing authentication and authorization
- Validating API endpoints for security
- Protecting sensitive data
- Scanning dependencies for vulnerabilities
- Ensuring GDPR/HIPAA compliance
- Threat modeling new features
- Security code review

**Core Framework**: OWASP + STRIDE + COMPLIANCE
```
1. Threat Identification (STRIDE)
   ↓
2. Vulnerability Assessment (OWASP Top 10)
   ↓
3. Control Implementation (Auth, Encryption)
   ↓
4. Compliance Validation (GDPR, HIPAA)
   ↓
5. Continuous Scanning (Dependencies, Code)
```

---

## Core Patterns (5-10 minutes each)

### Pattern 1: OWASP Top 10 Injection Prevention

**Concept**: Prevent SQL injection, command injection, and NoSQL injection through parameterized queries and input validation.

```python
# ❌ VULNERABLE - String concatenation
def find_user(user_id):
    query = f"SELECT * FROM users WHERE id = {user_id}"  # VULNERABLE!
    return db.execute(query)

# ✅ SECURE - Parameterized query
from sqlalchemy import text

def find_user(user_id: int):
    query = text("SELECT * FROM users WHERE id = :id")
    return db.execute(query, {"id": user_id})

# ✅ SECURE - ORM (SQLAlchemy)
from sqlalchemy import select

def find_user(user_id: int):
    stmt = select(User).where(User.id == user_id)
    return db.execute(stmt).scalar_one_or_none()
```

**Use Case**: Prevent SQL injection attacks (OWASP-A03:2021).

---

### Pattern 2: Authentication with JWT and OAuth2

**Concept**: Implement stateless authentication using JWT tokens with OAuth2 flow.

```python
from fastapi import FastAPI, Depends, HTTPException
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm
from jose import JWTError, jwt
from datetime import timedelta, datetime

SECRET_KEY = os.getenv("SECRET_KEY")  # Never hardcode!
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 30

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")

def create_access_token(data: dict, expires_delta: timedelta = None):
    """Create JWT token with expiration."""
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=15)
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    return encoded_jwt

async def get_current_user(token: str = Depends(oauth2_scheme)):
    """Verify JWT token and return user."""
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username: str = payload.get("sub")
        if username is None:
            raise HTTPException(status_code=401, detail="Invalid token")
    except JWTError:
        raise HTTPException(status_code=401, detail="Invalid token")
    user = get_user(username)
    if user is None:
        raise HTTPException(status_code=401, detail="User not found")
    return user

@app.post("/token")
async def login(form_data: OAuth2PasswordRequestForm = Depends()):
    """OAuth2 login endpoint."""
    user = authenticate_user(form_data.username, form_data.password)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid credentials")
    access_token = create_access_token(
        data={"sub": user.username},
        expires_delta=timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    )
    return {"access_token": access_token, "token_type": "bearer"}

@app.get("/users/me")
async def read_users_me(current_user = Depends(get_current_user)):
    """Protected endpoint requiring authentication."""
    return current_user
```

**Use Case**: Implement secure authentication (OWASP-A07:2021).

---

### Pattern 3: CORS and API Rate Limiting

**Concept**: Control API access with CORS, rate limiting, and request validation.

```python
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from slowapi import Limiter
from slowapi.util import get_remote_address

app = FastAPI()

# CORS configuration
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "https://example.com",
        "https://app.example.com"
    ],
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["Content-Type", "Authorization"],
)

# Rate limiting
limiter = Limiter(key_func=get_remote_address)

@app.get("/api/data")
@limiter.limit("100/minute")
async def get_data(request: Request):
    """Limit to 100 requests per minute."""
    return {"data": "protected"}

@app.post("/api/login")
@limiter.limit("5/minute")
async def login(request: Request, credentials: LoginRequest):
    """Strict rate limiting for login (brute force protection)."""
    return authenticate(credentials)
```

**Use Case**: Prevent CSRF (OWASP-A01:2021) and brute force attacks.

---

### Pattern 4: Secrets Management & Encryption

**Concept**: Manage secrets securely using environment variables and encryption libraries.

```python
import os
from cryptography.fernet import Fernet
from dotenv import load_dotenv

# Load secrets from .env (NEVER commit .env!)
load_dotenv()
DATABASE_URL = os.getenv("DATABASE_URL")
SECRET_KEY = os.getenv("SECRET_KEY")
ENCRYPTION_KEY = os.getenv("ENCRYPTION_KEY")

# Encrypt sensitive data
cipher = Fernet(ENCRYPTION_KEY)

def encrypt_ssn(ssn: str) -> str:
    """Encrypt sensitive data like SSN."""
    return cipher.encrypt(ssn.encode()).decode()

def decrypt_ssn(encrypted_ssn: str) -> str:
    """Decrypt sensitive data."""
    return cipher.decrypt(encrypted_ssn.encode()).decode()

# Example: Store encrypted PII
class User(Base):
    __tablename__ = "users"
    id = Column(Integer, primary_key=True)
    email = Column(String, unique=True)
    encrypted_ssn = Column(String)  # Stored encrypted

# Use TLS 1.3 for transport security
SQLALCHEMY_DATABASE_URI = "postgresql://...?sslmode=require"
```

**Use Case**: Protect sensitive data (OWASP-A02:2021).

---

### Pattern 5: Dependency Scanning & Vulnerability Management

**Concept**: Automatically scan dependencies for known vulnerabilities.

```bash
# pip: Check for vulnerabilities
pip install safety
safety check

# npm: Check for vulnerabilities
npm audit
npm audit fix

# Python: Security linter
pip install bandit
bandit -r src/

# GitHub: Enable dependency scanning
# Settings → Security & analysis → Enable Dependabot
```

```yaml
# .github/dependabot.yml - Automated dependency updates
version: 2
updates:
  - package-ecosystem: "pip"
    directory: "/"
    schedule:
      interval: "weekly"
    security-updates-only: true

  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
```

**Use Case**: Prevent vulnerable dependencies (OWASP-A06:2021).

---

### Pattern 6: GDPR Compliance & Data Protection

**Concept**: Implement GDPR-compliant data handling (consent, deletion, portability).

```python
# GDPR: Right to be forgotten
class GDPRMixin:
    """Mixin for GDPR-compliant deletion."""

    def delete_user_data(user_id: int):
        """Delete all personal data (right to be forgotten)."""
        # 1. Delete direct data
        user = User.query.get(user_id)
        db.session.delete(user)

        # 2. Delete related data
        db.session.query(Order).filter(Order.user_id == user_id).delete()
        db.session.query(LoginHistory).filter(LoginHistory.user_id == user_id).delete()

        # 3. Anonymize audit logs
        db.session.query(AuditLog).filter(AuditLog.user_id == user_id).update({
            "user_id": None,
            "user_email": "[DELETED]"
        })

        db.session.commit()

    def export_user_data(user_id: int) -> dict:
        """Export all user data (data portability)."""
        user = User.query.get(user_id)
        orders = db.session.query(Order).filter(Order.user_id == user_id).all()
        return {
            "user": user.to_dict(),
            "orders": [o.to_dict() for o in orders],
            "preferences": get_user_preferences(user_id)
        }
```

**Use Case**: Ensure GDPR compliance and data portability.

---

## Advanced Documentation

For detailed security patterns and implementation strategies:

- **[modules/owasp-validation.md](modules/owasp-validation.md)** - OWASP Top 10 detailed patterns
- **[modules/authentication-authorization.md](modules/authentication-authorization.md)** - Auth/AuthZ patterns
- **[modules/api-security.md](modules/api-security.md)** - API security best practices
- **[modules/encryption-secrets.md](modules/encryption-secrets.md)** - Encryption and secrets management
- **[modules/compliance-standards.md](modules/compliance-standards.md)** - Compliance frameworks

---

## Best Practices

### ✅ DO
- Use parameterized queries for all database operations
- Encrypt sensitive data at rest and in transit
- Implement strong authentication (JWT, OAuth2)
- Use environment variables for secrets (NEVER hardcode)
- Validate and sanitize ALL user input
- Enable HTTPS/TLS 1.3 everywhere
- Implement rate limiting for APIs
- Scan dependencies weekly for vulnerabilities
- Use RBAC/ABAC for authorization
- Log security events (authentication, authorization, data access)

### ❌ DON'T
- Hardcode secrets or API keys
- Use MD5/SHA1 for password hashing (use bcrypt/Argon2)
- Accept user input without validation
- Skip HTTPS (always use TLS 1.3)
- Store passwords in plaintext
- Trust client-side validation alone
- Expose sensitive error messages
- Ignore security warnings
- Use default credentials
- Log sensitive data (passwords, tokens)

---

## Success Metrics

- **OWASP Coverage**: All Top 10 vulnerabilities addressed
- **Vulnerability Scan**: Zero critical/high vulnerabilities
- **Dependency Age**: Dependencies updated within 30 days of release
- **Coverage**: Security tests cover all auth/authz paths
- **Compliance**: 100% GDPR/HIPAA compliance
- **Encryption**: Sensitive data encrypted at rest and in transit
- **Audit Logging**: All security events logged

---

## Context7 Integration

### Related Libraries & Tools
- **[OWASP Top 10](/owasp/Top10)**: Most critical web application security risks
- **[PyJWT](/jpadilla/pyjwt)**: JWT authentication for Python
- **[cryptography](/pyca/cryptography)**: Modern encryption library
- **[bandit](/PyCQA/bandit)**: Security linter for Python
- **[npm audit](/npm/npm)**: Vulnerability scanning for Node.js

### Official Documentation
- [OWASP Security Guidelines](https://owasp.org/www-community/)
- [OWASP Top 10 2021](https://owasp.org/Top10/)
- [GDPR Compliance](https://www.gdpr.eu/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)

---

## Related Skills

- `moai-foundation-trust` (TRUST 5 security requirements)
- `moai-lang-python` (Python security libraries)
- `moai-lang-typescript` (JavaScript security patterns)
- `moai-quality-testing` (Security testing)

---

## Consolidation Source Skills

This skill consolidates 12 security skills:
- owasp-validation
- authentication-patterns
- authorization-patterns
- api-security
- encryption-patterns
- secrets-management
- compliance-standards
- threat-modeling
- injection-prevention
- xss-csrf-protection
- data-protection
- dependency-scanning

All functionality preserved in unified architecture.

---

## Workflow Integration

**Typical Security Review Workflow**:
```
1. Threat Modeling (STRIDE)
   ↓
2. OWASP Validation
   ↓
3. Dependency Scanning
   ↓
4. Code Review for Security
   ↓
5. Compliance Check (GDPR, etc.)
   ↓
6. Penetration Testing
```

---

## Changelog

- **v1.0.0** (2025-11-24): Consolidated 12 security skills into unified moai-quality-security

---

**Status**: Production Ready (Enterprise)
**Generated with**: MoAI-ADK Skill Factory
**Modular Architecture**: SKILL.md + 5 modules (owasp, auth, api-security, encryption, compliance)
