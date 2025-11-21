## API Reference

### Core Encryption Operations
- `encrypt(data, keyId, algorithm)` - Encrypt data with specified algorithm
- `decrypt(encryptedData)` - Decrypt data with validation
- `generate_key(algorithm, metadata)` - Generate encryption key
- `sign_data(data, privateKeyId)` - Create digital signature
- `verify_signature(signature, publicKeyId)` - Verify digital signature

### Context7 Integration
- `get_latest_cryptography_docs()` - Cryptography via Context7
- `analyze_encryption_patterns()` - Encryption patterns via Context7
- `optimize_key_management()` - Key management via Context7

## Best Practices (November 2025)

### DO
- Use industry-standard cryptographic algorithms (AES-256, RSA-4096)
- Implement comprehensive key management with rotation
- Use authenticated encryption (AES-GCM) for data protection
- Implement proper error handling and secure disposal
- Use hardware security modules for key protection
- Maintain comprehensive audit logging and monitoring
- Follow compliance requirements (GDPR, PCI DSS, HIPAA)
- Implement quantum-resistant encryption preparation

### DON'T
- Implement custom cryptographic algorithms
- Store encryption keys with encrypted data
- Use deprecated or weak cryptographic algorithms
- Skip key rotation and lifecycle management
- Ignore compliance and regulatory requirements
- Forget to implement proper error handling
- Skip security testing and vulnerability assessments
- Use hardcoded keys or initialization vectors

## Works Well With

- `moai-security-api` (API security implementation)
- `moai-foundation-trust` (Trust and compliance)
- `moai-cc-configuration` (Configuration security)
- `moai-security-secrets` (Secrets management)
- `moai-baas-foundation` (BaaS security patterns)
- `moai-domain-backend` (Backend security)
- `moai-security-owasp` (Security best practices)
- `moai-security-compliance` (Compliance management)

## Changelog

- ** .0** (2025-11-13): Complete Enterprise   rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, advanced cryptographic patterns, and enterprise key management
- **v2.0.0** (2025-11-11): Complete metadata structure, encryption patterns, key management
- **v1.0.0** (2025-11-11): Initial encryption security foundation

---

**End of Skill** | Updated 2025-11-13

## Cryptographic Security

### Algorithm Selection
- AES-256-GCM for symmetric encryption with authentication
- RSA-4096 for asymmetric encryption and digital signatures
- ECC P-384 for efficient key exchange
- SHA-384 for cryptographic hashing
- PBKDF2 with 100,000 iterations for key derivation

### Enterprise Features
- Hardware Security Module (HSM) integration
- Automated key rotation and lifecycle management
- Comprehensive audit logging and compliance reporting
- Quantum-resistant encryption preparation
- Zero-knowledge proof implementation support

---

**End of Enterprise Encryption Security Expert **
