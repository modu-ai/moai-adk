---
name: moai-security-encryption
version: 4.0.0
updated: 2025-11-20
status: stable
description: Encryption patterns - AES-GCM, RSA, password hashing, envelope encryption
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# Encryption Security Expert

**Secure Data Encryption Patterns**

> **Focus**: AES-256-GCM, RSA-4096, Password Hashing (bcrypt/Argon2)  
> **Standards**: NIST, OWASP, FIPS 140-2

---

## Overview

Production-grade encryption for data at rest and in transit.

### Core Encryption Types

1.  **Symmetric**: AES-256-GCM (same key for encrypt/decrypt)
2.  **Asymmetric**: RSA-4096 (public/private key pair)
3.  **Password Hashing**: bcrypt, Argon2id (one-way)
4.  **Envelope Encryption**: Data encrypted with DEK, DEK encrypted with KEK

---

## Quick Start

### 1. Symmetric Encryption (AES-256-GCM)

For encrypting data at rest (files, database fields).

**Use Cases**: Database columns, file storage, session data

**Key Points**:

- 256-bit key
- GCM mode (authenticated encryption)
- Unique IV per encryption

See: [examples.md](./examples.md#aes-gcm-encryption)

### 2. Asymmetric Encryption (RSA)

For key exchange and digital signatures.

**Use Cases**: HTTPS, JWT signing, key distribution

**Key Points**:

- 4096-bit keys (2025 standard)
- OAEP padding
- Public key for encryption, private for decryption

See: [examples.md](./examples.md#rsa-encryption)

### 3. Password Hashing

For storing user passwords securely.

**Use Cases**: User authentication, API keys

**Algorithms**: bcrypt (recommended), Argon2id (modern)

**Key Points**:

- Never reversible
- Salt automatically included
- Cost factor: 12+ (2025)

See: [examples.md](./examples.md#password-hashing)

### 4. Envelope Encryption

For securing large datasets with key rotation.

**Pattern**:

1. Generate Data Encryption Key (DEK)
2. Encrypt data with DEK
3. Encrypt DEK with Key Encryption Key (KEK)
4. Store encrypted data + encrypted DEK

See: [examples.md](./examples.md#envelope-encryption)

---

## Encryption Decision Tree

```
Need to encrypt?
├─ Password? → bcrypt or Argon2id
├─ Large file/data? → AES-256-GCM (symmetric)
├─ Key exchange? → RSA-4096 (asymmetric)
└─ Multi-key management? → Envelope Encryption
```

---

## Best Practices

1.  **Key Management**: Never hardcode keys, use environment variables or KMS
2.  **Random IVs**: Generate new IV for each encryption
3.  **Authenticated Encryption**: Use GCM mode (prevents tampering)
4.  **Key Rotation**: Rotate keys annually or after breach
5.  **Secure Storage**: Store keys separately from encrypted data

---

## Security Checklist

- [ ] **Algorithm**: AES-256-GCM or RSA-4096?
- [ ] **Keys**: Stored securely (not in code)?
- [ ] **IVs**: Unique per encryption?
- [ ] **Mode**: Authenticated (GCM, not ECB)?
- [ ] **Rotation**: Key rotation policy defined?

---

## Related Skills

- `moai-security-api`: API encryption (TLS)
- `moai-security-auth`: Token encryption
- `moai-cloud-aws-advanced`: AWS KMS integration

---

## Additional Resources

- [examples.md](./examples.md): Implementation code
- [reference.md](./reference.md): NIST standards, algorithms

---

**Last Updated**: 2025-11-20
