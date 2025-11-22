# moai-security-encryption: Reference & Standards (2024-2025)

## NIST PQC Standardization (Post-Quantum Cryptography)

### NIST Selected Algorithms (2022-2024)
- **ML-KEM (Kyber)**: https://csrc.nist.gov/pubs/fips/203/final
  - Key Encapsulation Mechanism (KEM)
  - Lattice-based cryptography
  - Security levels: ML-KEM-512, ML-KEM-768, ML-KEM-1024

- **ML-DSA (Dilithium)**: https://csrc.nist.gov/pubs/fips/204/final
  - Digital Signature Algorithm
  - Lattice-based cryptography
  - Security levels: ML-DSA-44, ML-DSA-65, ML-DSA-87

- **SLH-DSA (SPHINCS+)**: https://csrc.nist.gov/pubs/fips/205/final
  - Stateless Hash-Based Signature Scheme
  - Hash-based cryptography
  - Security levels: SLH-DSA-128f, SLH-DSA-128s, SLH-DSA-256f

### PQC Implementation Status (2025)
- **Standardization**: Complete (FIPS 203, 204, 205)
- **Library Support**: OpenSSL 3.2+, Bouncy Castle, liboqs
- **Production Ready**: Yes (hybrid mode recommended)
- **Migration Timeline**: 2025-2030 (NIST recommendation)

## Modern Symmetric Encryption (2024-2025)

### Approved Algorithms
- **AES-256-GCM**: NIST FIPS 197 + NIST SP 800-38D
  - Authenticated Encryption with Associated Data (AEAD)
  - Performance: ~1.5 GB/s (AES-NI)
  - Status: ✅ Approved for all applications

- **ChaCha20-Poly1305**: RFC 8439
  - AEAD cipher for mobile/embedded devices
  - Performance: ~1.8 GB/s (software implementation)
  - Status: ✅ Approved (mobile-optimized)

- **AES-256-CBC**: NIST FIPS 197 + NIST SP 800-38A
  - Legacy mode (requires separate MAC)
  - Status: ⚠️ Allowed (prefer GCM)

### Deprecated/Removed Algorithms
- ❌ **DES**: Removed 1999 (56-bit key)
- ❌ **3DES**: Deprecated 2023 (NIST SP 800-131A Rev. 2)
- ❌ **RC4**: Removed 2015 (stream cipher vulnerabilities)
- ❌ **Blowfish**: Not recommended (64-bit block size)

## Modern Asymmetric Encryption (2024-2025)

### Current Standards
- **RSA-4096**: NIST FIPS 186-5
  - Key size: 4096 bits minimum (3072 acceptable until 2030)
  - Status: ✅ Approved (hybrid PQC recommended for long-term)

- **ECC P-384**: NIST FIPS 186-5
  - Elliptic Curve Cryptography (ECDH, ECDSA)
  - Status: ✅ Approved

- **Ed25519**: RFC 8032
  - Edwards-curve Digital Signature Algorithm
  - Status: ✅ Approved (fast signing)

### Deprecated Algorithms
- ❌ **RSA-1024**: Deprecated 2010
- ❌ **RSA-2048**: Transitioning out (acceptable until 2030)
- ❌ **DSA**: Deprecated 2024 (NIST)

## TLS/SSL Standards (2024-2025)

### TLS 1.3 (RFC 8446)
- **Status**: ✅ Mandatory for all new implementations
- **Cipher Suites** (Recommended):
  - TLS_AES_256_GCM_SHA384
  - TLS_CHACHA20_POLY1305_SHA256
  - TLS_AES_128_GCM_SHA256

### Deprecated TLS Versions
- ❌ **SSLv2**: Removed 2011
- ❌ **SSLv3**: Removed 2015 (POODLE attack)
- ❌ **TLS 1.0**: Deprecated 2020
- ❌ **TLS 1.1**: Deprecated 2020
- ⚠️ **TLS 1.2**: Legacy support only (prefer TLS 1.3)

## Modern Hashing (2024-2025)

### Approved Algorithms
- **SHA-3 (Keccak)**: NIST FIPS 202
  - SHA3-256, SHA3-384, SHA3-512
  - Status: ✅ Approved (quantum-resistant)

- **BLAKE2/BLAKE3**: RFC 7693
  - Fast cryptographic hash
  - Status: ✅ Approved (performance-critical)

- **SHA-256/SHA-384/SHA-512**: NIST FIPS 180-4
  - SHA-2 family
  - Status: ✅ Approved

### Deprecated Hashing
- ❌ **MD5**: Removed 2004 (collision attacks)
- ❌ **SHA-1**: Deprecated 2017 (collision attacks)
- ⚠️ **SHA-224**: Not recommended (prefer SHA-256)

## Password Hashing (2024-2025)

### Approved Algorithms
- **Argon2id**: RFC 9106
  - Winner of Password Hashing Competition (2015)
  - Memory-hard function (GPU-resistant)
  - Status: ✅ Recommended for all new applications

- **scrypt**: RFC 7914
  - Memory-hard function
  - Status: ✅ Approved

- **bcrypt**: Blowfish-based KDF
  - Status: ✅ Acceptable (legacy systems)

### Deprecated Password Hashing
- ❌ **MD5 (plain)**: Never use
- ❌ **SHA-1/SHA-256 (plain)**: Not suitable for passwords
- ❌ **PBKDF2**: Transitioning out (prefer Argon2id)

## Key Management Standards

### NIST Guidelines
- **NIST SP 800-57 Part 1 Rev. 5**: https://csrc.nist.gov/publications/detail/sp/800-57-part-1/rev-5/final
  - Key management lifecycle
  - Key length recommendations
  - Cryptoperiod guidelines

- **NIST SP 800-133 Rev. 2**: https://csrc.nist.gov/publications/detail/sp/800-133/rev-2/final
  - Key generation recommendations
  - Random number generation

### Key Rotation Guidelines (2025)
- **Symmetric keys**: 1-2 years or after 2^32 encryptions
- **Asymmetric keys**: 2-5 years
- **TLS certificates**: 90 days (recommended), 1 year (maximum)
- **API keys**: 90 days (automated rotation)

## Compliance Frameworks

### FIPS 140-3 (2024)
- **Module Validation**: https://csrc.nist.gov/projects/cryptographic-module-validation-program
- **Security Levels**: 1-4 (Level 3 recommended for enterprise)
- **Approved Algorithms**: FIPS 197, 186-5, 202

### PCI DSS 4.0 (2024)
- **Requirement 3**: Protect stored cardholder data
- **Requirement 4**: Encrypt transmission of cardholder data
- **Encryption Standards**: AES-256, TLS 1.3

### GDPR (EU)
- **Article 32**: Security of processing
- **Article 34**: Communication of data breach
- **Encryption Requirement**: State-of-the-art encryption

### HIPAA (Healthcare)
- **§ 164.312(a)(2)(iv)**: Encryption and decryption
- **§ 164.312(e)(1)**: Transmission security
- **Approved Standards**: NIST FIPS 140-3

## Hardware Security Modules (HSM)

### HSM Vendors (2025)
- **Thales Luna HSM**: FIPS 140-3 Level 3
- **nCipher nShield**: FIPS 140-3 Level 3
- **Utimaco SecurityServer**: Common Criteria EAL4+
- **AWS CloudHSM**: FIPS 140-3 Level 3
- **Azure Dedicated HSM**: FIPS 140-3 Level 3

### Cloud KMS Services
- **AWS KMS**: https://aws.amazon.com/kms/
- **Azure Key Vault**: https://azure.microsoft.com/en-us/services/key-vault/
- **Google Cloud KMS**: https://cloud.google.com/kms
- **HashiCorp Vault**: https://www.vaultproject.io/

## Libraries & Tools (2024-2025)

### Cryptography Libraries
- **OpenSSL 3.2+**: https://www.openssl.org/ (PQC support)
- **Bouncy Castle**: https://www.bouncycastle.org/ (Java/C#)
- **libsodium**: https://doc.libsodium.org/ (NaCl/modern crypto)
- **liboqs**: https://github.com/open-quantum-safe/liboqs (PQC)
- **Tink**: https://github.com/google/tink (Google's crypto library)

### Python Cryptography
- **cryptography** (43.0.x): https://cryptography.io/
- **PyNaCl** (1.5.x): https://pynacl.readthedocs.io/
- **PyCryptodome** (3.20.x): https://www.pycryptodome.org/

### Node.js Cryptography
- **node:crypto** (built-in): https://nodejs.org/api/crypto.html
- **@noble/ciphers**: https://github.com/paulmillr/noble-ciphers
- **tweetnacl**: https://tweetnacl.js.org/

## Testing & Validation Tools

### Encryption Testing
- **NIST Cryptographic Algorithm Validation Program (CAVP)**: https://csrc.nist.gov/projects/cavp
- **Keyczar**: https://github.com/google/keyczar
- **CrypTool**: https://www.cryptool.org/

### Security Scanners
- **Nessus**: https://www.tenable.com/products/nessus
- **Qualys SSL Labs**: https://www.ssllabs.com/ssltest/
- **testssl.sh**: https://testssl.sh/

## Related MoAI Skills

- **moai-security-auth**: Authentication and session encryption
- **moai-security-api**: API encryption and transport security
- **moai-security-secrets**: Secrets management and key storage
- **moai-security-compliance**: Regulatory compliance frameworks
- **moai-foundation-trust**: Trust and security foundations

---

**Last Updated**: 2025-11-22  
**Compliance**: NIST FIPS 140-3, PCI DSS 4.0, GDPR, HIPAA, PQC (FIPS 203/204/205)
