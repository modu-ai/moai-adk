# Cryptographic Failures (A02)

## Overview

Cryptographic failures involve weak or missing encryption of sensitive data, leading to data breaches and unauthorized access.

## Critical Vulnerabilities

### Weak Cryptography
- Using outdated algorithms (MD5, SHA1, DES)
- Insufficient key lengths
- Poor random number generation
- Hardcoded encryption keys

### TLS/SSL Issues
- TLS 1.0/1.1 (deprecated)
- Weak cipher suites
- Missing certificate validation
- Missing HSTS headers

### Password Storage
- Plaintext passwords
- Weak hashing algorithms
- Insufficient salt/rounds
- Reversible encryption

## Remediation Patterns

### Secure Password Hashing

**Python (bcrypt)**:
```python
import bcrypt

def hash_password(password: str) -> str:
    """Hash password with bcrypt (12 rounds)."""
    salt = bcrypt.gensalt(rounds=12)
    return bcrypt.hashpw(password.encode(), salt).decode()

def verify_password(password: str, hashed: str) -> bool:
    """Verify password against hash."""
    return bcrypt.checkpw(password.encode(), hashed.encode())
```

**Node.js (bcrypt)**:
```javascript
const bcrypt = require('bcrypt');

async function hashPassword(password) {
  const saltRounds = 12;
  return await bcrypt.hash(password, saltRounds);
}

async function verifyPassword(password, hash) {
  return await bcrypt.compare(password, hash);
}
```

### TLS Configuration

**Node.js (Express with TLS 1.3)**:
```javascript
const https = require('https');
const fs = require('fs');

const options = {
  key: fs.readFileSync('private-key.pem'),
  cert: fs.readFileSync('certificate.pem'),
  minVersion: 'TLSv1.3',
  ciphers: [
    'TLS_AES_256_GCM_SHA384',
    'TLS_CHACHA20_POLY1305_SHA256',
    'TLS_AES_128_GCM_SHA256'
  ].join(':')
};

https.createServer(options, app).listen(443);
```

**Python (SSL context)**:
```python
import ssl

context = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
context.minimum_version = ssl.TLSVersion.TLSv1_3
context.set_ciphers('ECDHE+AESGCM:ECDHE+CHACHA20:DHE+AESGCM')
```

### Data Encryption at Rest

**AES-256-GCM Encryption**:
```python
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
import os

def encrypt_data(plaintext: bytes, key: bytes) -> tuple:
    """Encrypt data with AES-256-GCM."""
    nonce = os.urandom(12)
    aesgcm = AESGCM(key)
    ciphertext = aesgcm.encrypt(nonce, plaintext, None)
    return nonce, ciphertext

def decrypt_data(nonce: bytes, ciphertext: bytes, key: bytes) -> bytes:
    """Decrypt data with AES-256-GCM."""
    aesgcm = AESGCM(key)
    return aesgcm.decrypt(nonce, ciphertext, None)
```

### Secure Key Management

**AWS KMS Integration**:
```python
import boto3

def get_data_key():
    """Get data encryption key from AWS KMS."""
    kms = boto3.client('kms')
    response = kms.generate_data_key(
        KeyId='alias/my-key',
        KeySpec='AES_256'
    )
    return response['Plaintext'], response['CiphertextBlob']

def decrypt_key(encrypted_key: bytes) -> bytes:
    """Decrypt data key using KMS."""
    kms = boto3.client('kms')
    response = kms.decrypt(CiphertextBlob=encrypted_key)
    return response['Plaintext']
```

## Validation Checklist

- [ ] TLS 1.3+ enabled
- [ ] Strong cipher suites configured
- [ ] Passwords hashed with bcrypt (12+ rounds)
- [ ] Sensitive data encrypted at rest (AES-256)
- [ ] Keys stored in secure vaults (KMS/HSM)
- [ ] No hardcoded credentials
- [ ] Certificate validation enabled
- [ ] HSTS headers configured

## Common Mistakes

**DON'T**:
- ❌ Use MD5/SHA1 for passwords
- ❌ Store passwords in plaintext
- ❌ Hardcode encryption keys
- ❌ Use weak cipher suites
- ❌ Disable certificate validation
- ❌ Use predictable IVs/nonces

**DO**:
- ✅ Use bcrypt/argon2 for passwords
- ✅ Use AES-256-GCM for data encryption
- ✅ Store keys in KMS/HSM
- ✅ Enable TLS 1.3+
- ✅ Generate random IVs/nonces
- ✅ Implement key rotation

## Testing

```python
def test_password_hashing():
    """Verify password hashing security."""
    password = "TestPassword123!"

    # Hash should be different each time (random salt)
    hash1 = hash_password(password)
    hash2 = hash_password(password)
    assert hash1 != hash2

    # Verification should work
    assert verify_password(password, hash1)
    assert verify_password(password, hash2)

    # Wrong password should fail
    assert not verify_password("WrongPassword", hash1)
```

---

**Last Updated**: 2025-11-24
**OWASP Category**: A02:2021
**CWE**: CWE-327 (Use of Broken or Risky Cryptographic Algorithm)
