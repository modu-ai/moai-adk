---
name: moai-security-encryption
description: Enterprise encryption security with cryptographic architecture and key management
version: 1.0.0
modularized: true
last_updated: 2025-11-22
compliance_score: 70
auto_trigger_keywords:
  - encryption
  - security
category_tier: 1
---

## Quick Reference (30 seconds)

# Enterprise Encryption Security

**Comprehensive data protection** with AI-powered cryptographic architecture, key management, and compliance monitoring.

**Core Capabilities**:
- AES-256-GCM/RSA-4096 encryption
- Enterprise key management (Vault, KMS, HSM)
- FIPS 140-2/3, NIST SP 800-57 compliance
- End-to-end encryption lifecycle
- Automated key rotation and audit

---

## Modules

### 1. [Encryption Architecture & Algorithms](modules/encryption-architecture-algorithms.md)
Modern cryptographic algorithms, implementation patterns, and security configuration.

**Topics**:
- AES-256-GCM symmetric encryption
- RSA-4096 asymmetric encryption
- ECC P-384 elliptic curve
- HMAC-SHA256 authentication

### 2. [Key Management Systems](modules/key-management-systems.md)
Enterprise key management with HSM, Vault integration, and rotation strategies.

**Topics**:
- HashiCorp Vault integration
- AWS KMS/Azure Key Vault
- Key rotation automation
- Compliance auditing

### 3. [Secure Communication & Compliance](modules/secure-communication-compliance.md)
TLS/SSL configuration, certificate management, and regulatory compliance.

**Topics**:
- TLS 1.3 configuration
- Certificate lifecycle
- GDPR/HIPAA/PCI DSS compliance
- Security audit logging

---

## Security Standards

- FIPS 140-2/3: Cryptographic module validation
- NIST SP 800-57: Key management guidelines
- PCI DSS: Payment security
- GDPR/HIPAA: Data protection compliance

---

## Context7 Integration

### Related Libraries & Tools
- [OpenSSL](/openssl/openssl): Cryptographic library
- [libsodium](/jedisct1/libsodium): Modern cryptography library
- [crypto](/nodejs/node): Node.js built-in crypto module
- [cryptography](/pyca/cryptography): Python cryptography library
- [TweetNaCl.js](/dchest/tweetnacl-js): Cryptographic library for JavaScript

### Official Documentation
- [NIST SP 800-38D](https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf)
- [OpenSSL Documentation](https://www.openssl.org/docs/)
- [libsodium Documentation](https://doc.libsodium.org/)

### Version-Specific Guides
Latest stable versions: OpenSSL 3.x, libsodium, TLS 1.3
- [AES-GCM Implementation](https://csrc.nist.gov/publications/detail/sp/800-38d/final)
- [Elliptic Curve Cryptography](https://safecurves.cr.yp.to/)
- [Key Derivation Functions](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Key_Management_Cheat_Sheet.html)

---

## Version History

**v4.0.0** (2025-11-21)
- Modularized structure (3 focused modules)
- Enhanced key management patterns
- Compliance automation

---

**Last Updated**: 2025-11-21  
**Classification**: Enterprise Encryption Security