# Encryption Reference

NIST standards, algorithm specifications, and security best practices.

---

## Encryption Algorithms

### AES-GCM (Advanced Encryption Standard - Galois/Counter Mode)

**Specification**: NIST SP 800-38D

**Parameters**:

- Key size: 256 bits (32 bytes)
- IV/Nonce: 96 bits (12 bytes) - MUST be unique per encryption
- Tag size: 128 bits (16 bytes) - authentication tag

**Security Properties**:

- Confidentiality: ✅ (symmetric encryption)
- Integrity: ✅ (authenticated encryption)
- Forward secrecy: ❌ (requires key rotation)

**Performance**: ~1-2 GB/s (AES-NI hardware acceleration)

**Use Cases**: Database encryption, file encryption, session data

**Compliance**: FIPS 140-2, PCI DSS, HIPAA

---

### RSA (Rivest-Shamir-Adleman)

**Specification**: NIST FIPS 186-4

**Key Sizes** (2025 recommendations):

- 2048 bits: Legacy (not recommended for new systems)
- 3072 bits: Minimum
- 4096 bits: **Recommended**

**Padding Schemes**:

- **OAEP** (Optimal Asymmetric Encryption Padding) - ✅ Recommended
- PKCS#1 v1.5 - ❌ Deprecated (vulnerable to padding oracle attacks)

**Security Properties**:

- Confidentiality: ✅ (asymmetric encryption)
- Digital signatures: ✅
- Key exchange: ✅

**Performance**: ~1000x slower than AES (use for key exchange, not bulk data)

**Use Cases**: TLS/SSL, digital signatures, key wrapping

**Compliance**: FIPS 140-2, Common Criteria

---

### bcrypt

**Specification**: Based on Blowfish cipher

**Parameters**:

- Cost factor (rounds): 10-14 (2025: **12 recommended**)
- Salt: 128 bits (automatically generated)
- Output: 192 bits (24 bytes)

**Security Properties**:

- Adaptive: Cost can increase over time
- Salt: Unique per password
- Slow: Designed to be computationally expensive (~100-500ms)

**Formula**: `2^cost` iterations

**Use Cases**: Password storage (NOT for encryption)

**Resistance**:

- Brute force: ✅ (high cost factor)
- Rainbow tables: ✅ (unique salts)
- GPU attacks: ⚠️ (partially resistant)

---

### Argon2id

**Specification**: RFC 9106 (2021)

**Parameters**:

- Time cost (iterations): 3
- Memory cost: 64 MB (65536 KB)
- Parallelism: 4 threads
- Salt: 128 bits
- Output: 256 bits

**Variants**:

- Argon2d: Resistant to GPU attacks, vulnerable to side-channel
- Argon2i: Resistant to side-channel, weaker against GPU
- **Argon2id**: ✅ Hybrid (recommended)

**Security Properties**:

- Memory-hard: Requires significant RAM
- Side-channel resistant: Constant-time operations
- GPU/ASIC resistant: ✅ (best among password hashing algorithms)

**Use Cases**: Password storage (modern replacement for bcrypt)

**Compliance**: PHC (Password Hashing Competition) winner

---

## Key Management

### Key Derivation (PBKDF2)

**Specification**: NIST SP 800-132

**Parameters**:

- Iterations: 100,000+ (2025 recommendation)
- Hash function: SHA-256 or SHA-512
- Salt: 128 bits minimum
- Output length: 256 bits

**Use Case**: Derive encryption key from password

```python
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2
from cryptography.hazmat.primitives import hashes

kdf = PBKDF2(
    algorithm=hashes.SHA256(),
    length=32,  # 256 bits
    salt=salt,
    iterations=600000,  # OWASP 2023 recommendation
)
key = kdf.derive(password.encode())
```

---

### Key Rotation

**Frequency Recommendations**:

- Symmetric keys (AES): Annually or after 2^32 encryptions
- Asymmetric keys (RSA): Every 2-3 years
- Password hashes: On credential change or breach

**Rotation Process**:

1. Generate new key (KEK₂)
2. Decrypt DEK with old KEK₁
3. Re-encrypt DEK with new KEK₂
4. Update key reference in database
5. Securely delete KEK₁

**Graceful Migration**:

```python
def migrate_encryption(old_kek, new_kek):
    # Decrypt with old key
    plaintext_dek = decrypt(encrypted_dek, old_kek)

    # Re-encrypt with new key
    new_encrypted_dek = encrypt(plaintext_dek, new_kek)

    # Update database
    db.update(encrypted_dek=new_encrypted_dek, key_version=2)
```

---

## Security Standards

### NIST Recommendations (SP 800-57)

**Key Sizes** (Through 2030):
| Algorithm | Key Size | Security Level |
|-----------|----------|----------------|
| AES | 128 bits | 128-bit |
| AES | **256 bits** | **256-bit** ✅ |
| RSA | 2048 bits | 112-bit |
| RSA | 3072 bits | 128-bit |
| RSA | **4096 bits** | **152-bit** ✅ |
| ECC | 256 bits | 128-bit |
| ECC | 384 bits | 192-bit |

---

### OWASP Cryptographic Storage Cheat Sheet

**DO**:

- ✅ Use AES-256-GCM for symmetric encryption
- ✅ Use RSA-4096 or ECC (P-384) for asymmetric
- ✅ Use bcrypt (cost 12+) or Argon2id for passwords
- ✅ Generate cryptographically secure random IVs/salts
- ✅ Use authenticated encryption (GCM, not CBC)
- ✅ Implement key rotation policy
- ✅ Store keys separately from encrypted data

**DON'T**:

- ❌ Use MD5 or SHA-1 for hashing (broken)
- ❌ Use ECB mode (insecure)
- ❌ Reuse IVs/nonces
- ❌ Roll your own crypto
- ❌ Store keys with encrypted data
- ❌ Use weak random number generators (e.g., `Math.random()`)

---

### Compliance Mappings

| Standard       | Requirement             | Algorithm          |
| -------------- | ----------------------- | ------------------ |
| **PCI DSS**    | Data at rest encryption | AES-256            |
| **HIPAA**      | PHI encryption          | AES-256, RSA-2048+ |
| **GDPR**       | Pseudonymization        | AES-256-GCM        |
| **FIPS 140-2** | Approved algorithms     | AES, RSA, SHA-2    |
| **SOC 2**      | Encryption in transit   | TLS 1.3            |

---

## Attack Vectors

### Known Attacks

**Padding Oracle Attack** (CBC mode):

- **Vulnerable**: AES-CBC without authentication
- **Mitigation**: Use GCM mode (authenticated encryption)

**Timing Attacks** (RSA):

- **Vulnerable**: Unpadded RSA, constant-time comparison
- **Mitigation**: Use OAEP padding, constant-time operations

**Brute Force** (Weak passwords):

- **Vulnerable**: Low bcrypt cost (<10)
- **Mitigation**: Cost 12+, enforce strong passwords

**Rainbow Tables** (Unsalted hashes):

- **Vulnerable**: `SHA-256(password)` without salt
- **Mitigation**: Use bcrypt/Argon2 (automatic salting)

---

## Entropy Requirements

**Random Number Generation**:

- **Cryptographic**: `os.urandom()` (Python), `crypto.randomBytes()` (Node.js)
- **NOT cryptographic**: `random.random()` (Python), `Math.random()` (JavaScript)

**Entropy Sources**:

- `/dev/urandom` (Linux)
- `CryptGenRandom` (Windows)
- Hardware RNG (TPM)

**Minimum Entropy**:

- Symmetric key (AES-256): 256 bits
- Salt: 128 bits
- IV/Nonce: 96 bits (GCM)
- Session token: 128 bits

```python
import secrets

# Cryptographically secure random
key = secrets.token_bytes(32)  # 256 bits
token = secrets.token_urlsafe(32)  # URL-safe token
```

---

## Quantum Resistance

### Post-Quantum Cryptography

**NIST PQC Selected Algorithms** (2022):

- **CRYSTALS-Kyber**: Key encapsulation
- **CRYSTALS-Dilithium**: Digital signatures
- **SPHINCS+**: Stateless hash-based signatures

**Timeline**:

- 2025-2030: Transition period
- 2030+: Quantum computers may break RSA/ECC

**Recommendation**: Plan for cryptographic agility (ability to switch algorithms)

---

## Tools & Libraries

### Python

```bash
pip install cryptography  # Recommended
pip install pycryptodome  # Alternative
pip install argon2-cffi   # Passw hashing
```

### Node.js

```bash
npm install crypto         # Built-in
npm install bcrypt         # Password hashing
npm install argon2         # Modern hashing
```

### Verification

```bash
# Test AES-GCM support
openssl enc -aes-256-gcm -P

# Generate RSA key pair
openssl genrsa -out private.pem 4096
openssl rsa -in private.pem -pubout -out public.pem
```

---

**See also**: [SKILL.md](./SKILL.md) for overview and [examples.md](./examples.md) for implementations.
