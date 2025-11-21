# moai-security-auth: Reference & Official Documentation (2024-2025)

## OAuth 2.1 & OpenID Connect (Latest Standards)

### OAuth 2.1 Specification (2024)
- **OAuth 2.1**: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1-11
- **Key Changes from 2.0**:
  - PKCE (RFC 7636) now mandatory for all clients
  - Implicit Flow removed (security vulnerability)
  - Resource Owner Password Credentials Grant removed
  - Refresh Token Rotation recommended
  - State parameter mandatory for CSRF protection

### PKCE (Proof Key for Code Exchange)
- **RFC 7636**: https://datatracker.ietf.org/doc/html/rfc7636
- **Implementation Guide**: https://oauth.net/2/pkce/
- **PKCE Code Challenge Methods**: S256 (SHA-256) recommended

### OpenID Connect
- **OpenID Connect Core 1.0**: https://openid.net/specs/openid-connect-core-1_0.html
- **OpenID Connect Discovery**: https://openid.net/specs/openid-connect-discovery-1_0.html
- **ID Token Validation**: https://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation

## JWT (JSON Web Tokens) - 2025 Standards

### Core Specifications
- **RFC 7519 (JWT)**: https://datatracker.ietf.org/doc/html/rfc7519
- **RFC 7515 (JWS)**: https://datatracker.ietf.org/doc/html/rfc7515
- **RFC 7516 (JWE)**: https://datatracker.ietf.org/doc/html/rfc7516

### JWT Security Best Practices (2025)
- **JWT Best Current Practices**: https://datatracker.ietf.org/doc/html/rfc8725
- **Algorithm Security**:
  - ✅ Use: HS256 (≥256 bits), RS256, ES256
  - ❌ Avoid: HS256 (<256 bits), RS256 (<2048 bits), none
- **Token Storage**: httpOnly cookies (not localStorage)
- **Token Expiration**: Access tokens ≤15 minutes, Refresh tokens ≤7 days

## FIDO2 & WebAuthn (W3C Standards)

### WebAuthn Specifications
- **WebAuthn Level 2 (Stable)**: https://www.w3.org/TR/webauthn-2/
- **WebAuthn Level 3 (Draft)**: https://www.w3.org/TR/webauthn-3/
- **FIDO2 Project**: https://fidoalliance.org/fido2/

### Passkeys (2024-2025)
- **Passkeys Introduction**: https://passkeys.dev/
- **Apple Passkeys**: https://developer.apple.com/passkeys/
- **Google Passkeys**: https://developers.google.com/identity/passkeys
- **Microsoft Passkeys**: https://learn.microsoft.com/en-us/windows/security/identity-protection/passkeys

### FIDO Alliance Resources
- **Certified Authenticators**: https://fidoalliance.org/certification/certified-products/
- **FIDO2 Specifications**: https://fidoalliance.org/specifications/
- **Developer Resources**: https://fidoalliance.org/developer-resources/

## Multi-Factor Authentication (MFA)

### TOTP & HOTP Standards
- **RFC 6238 (TOTP)**: https://datatracker.ietf.org/doc/html/rfc6238
- **RFC 4226 (HOTP)**: https://datatracker.ietf.org/doc/html/rfc4226
- **Time-based One-Time Password**: 30-second window, 6-digit codes

### SMS Authentication (Deprecated 2024)
- **NIST Guidance**: SMS 2FA no longer recommended (SP 800-63B)
- **Preferred**: TOTP apps, hardware tokens, biometric authentication
- **Fallback Only**: Use SMS only as last resort backup

## Framework Documentation (November 2025)

### NextAuth.js 5.x
- **Official Documentation**: https://next-auth.js.org/
- **v5 Migration Guide**: https://next-auth.js.org/getting-started/upgrade-v5
- **Providers**: https://next-auth.js.org/providers/
- **Callbacks**: https://next-auth.js.org/configuration/callbacks
- **Database Adapters**: https://next-auth.js.org/adapters/overview

### Passport.js 0.7.x
- **Official Documentation**: http://www.passportjs.org/
- **Strategies**: http://www.passportjs.org/packages/
- **API Reference**: http://www.passportjs.org/docs/
- **Express Integration**: http://www.passportjs.org/howtos/

### SimpleWebAuthn 10.x
- **Official Documentation**: https://simplewebauthn.dev/
- **Server Library**: https://simplewebauthn.dev/docs/packages/server
- **Browser Library**: https://simplewebauthn.dev/docs/packages/browser
- **Examples**: https://simplewebauthn.dev/docs/guide/

## Libraries & Tools (November 2025 Versions)

### Authentication Libraries
- **next-auth** (5.0.x): https://github.com/nextauthjs/next-auth
- **passport** (0.7.x): http://www.passportjs.org/
- **@simplewebauthn/server** (10.0.x): https://github.com/MasterKale/SimpleWebAuthn
- **jsonwebtoken** (9.x): https://github.com/auth0/node-jsonwebtoken

### Multi-Factor Authentication
- **otplib** (12.x): https://github.com/yeojz/otplib
- **speakeasy** (2.0.x): https://github.com/speakeasyjs/speakeasy
- **qrcode** (1.5.x): https://github.com/soldair/node-qrcode
- **totp-generator** (1.x): https://github.com/chrisveness/hotp-totp

### Password Hashing (2025 Standards)
- **argon2** (0.31.x): https://github.com/ranisalt/node-argon2 ✅ Recommended
- **bcryptjs** (2.4.x): https://github.com/dcodeIO/bcrypt.js ✅ Acceptable
- **scrypt** (Node.js built-in): https://nodejs.org/api/crypto.html#cryptoscryptsyncpassword-salt-keylen-options

### Session Management
- **express-session** (1.18.x): https://github.com/expressjs/session
- **redis** (5.0.x): https://github.com/redis/node-redis
- **ioredis** (5.4.x): https://github.com/luin/ioredis

## OWASP Security Guidelines (2024)

### OWASP Top 10 2024 (Relevant to Authentication)
- **A01:2024 - Broken Access Control**: https://owasp.org/Top10/A01_2021-Broken_Access_Control/
- **A02:2024 - Cryptographic Failures**: https://owasp.org/Top10/A02_2021-Cryptographic_Failures/
- **A07:2024 - Authentication Failures**: https://owasp.org/Top10/A07_2021-Identification_and_Authentication_Failures/

### OWASP Cheat Sheets
- **Authentication Cheat Sheet**: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- **Session Management Cheat Sheet**: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
- **Password Storage Cheat Sheet**: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- **Credential Stuffing Prevention**: https://cheatsheetseries.owasp.org/cheatsheets/Credential_Stuffing_Prevention_Cheat_Sheet.html

## NIST Cybersecurity Framework (2024)

### NIST Digital Identity Guidelines
- **NIST SP 800-63-3**: https://pages.nist.gov/800-63-3/
- **NIST SP 800-63A (Enrollment & Identity Proofing)**: https://pages.nist.gov/800-63-3/sp800-63a.html
- **NIST SP 800-63B (Authentication & Lifecycle Management)**: https://pages.nist.gov/800-63-3/sp800-63b.html
- **NIST SP 800-63C (Federation & Assertions)**: https://pages.nist.gov/800-63-3/sp800-63c.html

### Key NIST Recommendations (2024)
- **Password Complexity**: Minimum 8 characters (12-15 recommended)
- **Password Hashing**: Argon2id, scrypt, or bcrypt
- **MFA**: Required for privileged accounts
- **SMS 2FA**: No longer recommended (deprecated)

## Testing & Validation Tools

### WebAuthn Testing
- **WebAuthn.io Debugger**: https://webauthn.io/
- **FIDO2 Test Suite**: https://github.com/duo-labs/py_webauthn
- **SimpleWebAuthn Example App**: https://github.com/MasterKale/SimpleWebAuthn/tree/master/example

### API Testing
- **Postman**: https://www.postman.com/
- **Insomnia**: https://insomnia.rest/
- **Thunder Client**: https://www.thunderclient.com/
- **HTTPie**: https://httpie.io/

### Security Testing
- **OWASP ZAP**: https://www.zaproxy.org/
- **Burp Suite**: https://portswigger.net/burp
- **Hydra (Brute Force Testing)**: https://github.com/vanhauser-thc/thc-hydra
- **John the Ripper (Password Testing)**: https://www.openwall.com/john/

## Common Vulnerabilities (CWE References)

### Authentication Vulnerabilities
- **CWE-287 (Improper Authentication)**: https://cwe.mitre.org/data/definitions/287.html
- **CWE-798 (Hardcoded Credentials)**: https://cwe.mitre.org/data/definitions/798.html
- **CWE-306 (Missing Authentication)**: https://cwe.mitre.org/data/definitions/306.html
- **CWE-307 (Improper Restriction of Excessive Authentication Attempts)**: https://cwe.mitre.org/data/definitions/307.html

### Session Management Vulnerabilities
- **CWE-384 (Session Fixation)**: https://cwe.mitre.org/data/definitions/384.html
- **CWE-613 (Insufficient Session Expiration)**: https://cwe.mitre.org/data/definitions/613.html
- **CWE-311 (Missing Encryption of Sensitive Data)**: https://cwe.mitre.org/data/definitions/311.html

### Password Security Vulnerabilities
- **CWE-256 (Plaintext Storage of Password)**: https://cwe.mitre.org/data/definitions/256.html
- **CWE-521 (Weak Password Requirements)**: https://cwe.mitre.org/data/definitions/521.html
- **CWE-640 (Weak Password Recovery)**: https://cwe.mitre.org/data/definitions/640.html

## Related MoAI Skills

- **moai-security-api**: OAuth 2.1, API security, rate limiting
- **moai-security-encryption**: Password hashing, session encryption
- **moai-security-owasp**: OWASP Top 10 compliance
- **moai-domain-web-api**: RESTful API authentication

---

**Last Updated**: 2025-11-22  
**Compliance**: OAuth 2.1, FIDO2, OWASP Top 10 2024, NIST SP 800-63B
