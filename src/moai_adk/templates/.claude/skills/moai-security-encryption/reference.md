# moai-security-encryption: Reference & Official Documentation

## Official Standards & Specifications

### Cryptographic Standards
- NIST FIPS 197 (AES): https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.197.pdf
- NIST SP 800-38D (GCM): https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf
- RFC 3394 (AES Key Wrap): https://tools.ietf.org/html/rfc3394
- RFC 5116 (CRYPTOGRAPHIC INTERFACE): https://tools.ietf.org/html/rfc5116

### Key Derivation Functions
- RFC 5869 (HKDF): https://tools.ietf.org/html/rfc5869
- RFC 2898 (PBKDF2): https://tools.ietf.org/html/rfc2898
- Argon2: https://github.com/P-H-C/phc-winner-argon2
- scrypt: https://tools.ietf.org/html/rfc7914

### Hash Functions
- SHA-2 Family: https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.180-4.pdf
- SHA-3 Family: https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.202.pdf
- BLAKE2: https://blake2.net/

### TLS & Transport Security
- TLS 1.3 (RFC 8446): https://tools.ietf.org/html/rfc8446
- TLS 1.2 (RFC 5246): https://tools.ietf.org/html/rfc5246
- HSTS (RFC 6797): https://tools.ietf.org/html/rfc6797
- OCSP Stapling (RFC 6066): https://tools.ietf.org/html/rfc6066

### Authenticated Encryption
- ChaCha20-Poly1305 (RFC 7539): https://tools.ietf.org/html/rfc7539
- AES-CCM (SP 800-38C): https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38c.pdf
- AES-GCM (SP 800-38D): https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf

## Framework & Library Documentation

### Node.js Crypto Module
- Built-in crypto: https://nodejs.org/api/crypto.html
- Crypto Subtle API: https://nodejs.org/api/webcrypto.html
- OpenSSL Integration: https://nodejs.org/api/crypto.html#crypto_crypto_createcipher_algorithm_password_options

### Libsodium (C Library)
- Official Site: https://doc.libsodium.org/
- Secret Box (Encryption): https://doc.libsodium.org/secret-key_cryptography/secretbox
- Sealed Box (Public Key): https://doc.libsodium.org/public-key_cryptography/sealed_boxes
- Password Hashing: https://doc.libsodium.org/password_hashing

### Cryptographic Libraries
- **tweetnacl.js**: https://tweetnacl.js.org/
- **node-jose**: https://github.com/cisco/node-jose
- **crypto-js**: https://cryptojs.gitbook.io/docs/
- **libsodium.js**: https://github.com/jedisct1/libsodium.js

## Certificate & PKI

### Certificate Management
- Let's Encrypt: https://letsencrypt.org/
- Certbot: https://certbot.eff.org/
- Digital Ocean Certificate: https://www.digitalocean.com/products/certificates/
- AWS Certificate Manager: https://aws.amazon.com/certificate-manager/

### Public Key Infrastructure
- X.509 Standard: https://tools.ietf.org/html/rfc5280
- PKIX: https://tools.ietf.org/html/rfc6234
- Self-Signed Certificates: https://www.digitalocean.com/community/tutorials/how-to-create-self-signed-ssl-certificates

## Security Best Practices

### OWASP Cryptographic Guidelines
- Cryptographic Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html
- Cryptographic Failures (A02:2021): https://owasp.org/Top10/A02_2021-Cryptographic_Failures/
- Secrets Management: https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html

### NIST Guidelines
- NIST SP 800-52 (TLS Guidelines): https://csrc.nist.gov/publications/detail/sp/800-52/rev-2/final
- NIST SP 800-63 (Digital Identity): https://pages.nist.gov/800-63-3/
- NIST SP 800-175B (Cryptography): https://csrc.nist.gov/publications/detail/sp/800-175/b/final

## Common Vulnerabilities & CWE References

### Weak Cryptography
- CWE-326: https://cwe.mitre.org/data/definitions/326.html
- CWE-327: https://cwe.mitre.org/data/definitions/327.html
- Mitigation: Use AES-256-GCM, TLS 1.3

### Hard-coded Cryptographic Keys
- CWE-798: https://cwe.mitre.org/data/definitions/798.html
- OWASP A02:2021: https://owasp.org/Top10/A02_2021-Cryptographic_Failures/
- Mitigation: Use key management system, environment variables

### Inadequate Encryption
- CWE-311: https://cwe.mitre.org/data/definitions/311.html
- CWE-312: https://cwe.mitre.org/data/definitions/312.html
- Mitigation: Encrypt all sensitive data

### Use of Insufficiently Random Values
- CWE-338: https://cwe.mitre.org/data/definitions/338.html
- OWASP A02:2021: https://owasp.org/Top10/A02_2021-Cryptographic_Failures/
- Mitigation: Use crypto.randomBytes(), never Math.random()

### Weak Password Hash
- CWE-916: https://cwe.mitre.org/data/definitions/916.html
- OWASP A02:2021: https://owasp.org/Top10/A02_2021-Cryptographic_Failures/
- Mitigation: Use Argon2id, bcrypt with proper salt

## Testing & Validation Tools

### Cryptographic Testing
- OpenSSL: https://www.openssl.org/
- GnuPG: https://gnupg.org/
- hashcat: https://hashcat.net/
- John the Ripper: https://www.openwall.com/john/

### TLS Testing
- ssllabs.com: https://www.ssllabs.com/
- Nessus: https://www.tenable.com/products/nessus
- testssl.sh: https://github.com/drwetter/testssl.sh

### Load Testing
- k6: https://k6.io/
- Apache JMeter: https://jmeter.apache.org/

## Industry Standards & Benchmarks

### CIS Controls
- CIS Controls v8: https://www.cisecurity.org/cis-controls/v8
- CIS benchmarks: https://www.cisecurity.org/cis-benchmarks/

### NIST Cybersecurity Framework
- NIST CSF Overview: https://www.nist.gov/cyberframework
- Protect Function: https://www.nist.gov/cyberframework/framework/v1-1/functions

## Related Skills

- **moai-security-auth**: Password hashing, session encryption
- **moai-security-api**: JWT signing, API encryption
- **moai-security-owasp**: OWASP A02 (Cryptographic Failures)
- **moai-domain-database**: Database encryption

