# moai-security-auth: Reference & Official Documentation

## Official Standards & Specifications

### Authentication Standards
- OAuth 2.1 Specification: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1
- OpenID Connect Core 1.0: https://openid.net/specs/openid-connect-core-1_0.html
- RFC 7519 (JWT): https://tools.ietf.org/html/rfc7519
- RFC 4226 (HOTP): https://tools.ietf.org/html/rfc4226
- RFC 6238 (TOTP): https://tools.ietf.org/html/rfc6238

### FIDO2 & WebAuthn
- FIDO2 Specifications: https://fidoalliance.org/fido2/fido2-web-authentication/
- WebAuthn Level 2 (W3C): https://www.w3.org/TR/webauthn-2/
- WebAuthn Level 3 (Draft): https://www.w3.org/TR/webauthn-3/
- FIDO Alliance Certified Products: https://fidoalliance.org/certification/certified-products/

### Password & Session Security
- OWASP Authentication Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- OWASP Session Management Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
- OWASP Password Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- NIST SP 800-63B (Authentication): https://pages.nist.gov/800-63-3/sp800-63b.html

## Framework Documentation

### NextAuth.js
- Official Documentation: https://next-auth.js.org/
- NextAuth.js v5 Migration Guide: https://next-auth.js.org/getting-started/upgrade-v5
- Providers Documentation: https://next-auth.js.org/providers/
- Callbacks Reference: https://next-auth.js.org/configuration/callbacks
- Database Adapters: https://next-auth.js.org/adapters/overview

### Passport.js
- Official Documentation: http://www.passportjs.org/
- Strategies: http://www.passportjs.org/packages/
- API Reference: http://www.passportjs.org/docs/
- Express Integration: http://www.passportjs.org/howtos/

### SimpleWebAuthn
- Official Documentation: https://simplewebauthn.dev/
- Server Library Docs: https://simplewebauthn.dev/docs/packages/server
- Browser Library Docs: https://simplewebauthn.dev/docs/packages/browser
- Examples & Tutorials: https://simplewebauthn.dev/docs/guide/

## Libraries & Tools (November 2025)

### Authentication Libraries
- **next-auth**: (5.0.x): https://github.com/nextauthjs/next-auth
- **passport**: (0.7.x): http://www.passportjs.org/
- **@simplewebauthn/server**: (10.0.x): https://github.com/MasterKale/SimpleWebAuthn
- **jsonwebtoken**: (9.x): https://github.com/auth0/node-jsonwebtoken

### Multi-Factor Authentication
- **otplib**: (12.x): https://github.com/yeojz/otplib
- **speakeasy**: (2.0.x): https://github.com/speakeasyjs/speakeasy
- **qrcode**: (1.5.x): https://github.com/davidshimjs/qrcode
- **totp-generator**: (0.0.x): https://github.com/chrisveness/hotp-totp

### Password Hashing
- **bcryptjs**: (2.4.x): https://github.com/dcodeIO/bcrypt.js
- **argon2**: (0.31.x): https://github.com/ranisalt/node-argon2
- **scryptsy**: (2.1.x): https://github.com/ricmoo/scryptsy

### Session Management
- **express-session**: (1.17.x): https://github.com/expressjs/session
- **redis**: (5.0.x): https://github.com/redis/node-redis
- **ioredis**: (5.3.x): https://github.com/luin/ioredis

## Common Vulnerabilities & CWE References

### Broken Authentication
- CWE-287: https://cwe.mitre.org/data/definitions/287.html
- OWASP A02:2021: https://owasp.org/Top10/A02_2021-Cryptographic_Failures/
- Mitigation: Use OAuth 2.1, strong password hashing

### Session Fixation
- CWE-384: https://cwe.mitre.org/data/definitions/384.html
- OWASP A01:2021: https://owasp.org/Top10/A01_2021-Broken_Access_Control/
- Mitigation: Regenerate session ID on login

### Brute Force Attack
- CWE-307: https://cwe.mitre.org/data/definitions/307.html
- OWASP A07:2021: https://owasp.org/Top10/A07_2021-Identification_and_Authentication_Failures/
- Mitigation: Rate limiting, account lockout, CAPTCHA

### Credential Stuffing
- CWE-640: https://cwe.mitre.org/data/definitions/640.html
- OWASP A07:2021: https://owasp.org/Top10/A07_2021-Identification_and_Authentication_Failures/
- Mitigation: MFA, breach detection, password change on compromise

### Weak Password Storage
- CWE-256: https://cwe.mitre.org/data/definitions/256.html
- OWASP A02:2021: https://owasp.org/Top10/A02_2021-Cryptographic_Failures/
- Mitigation: Use Argon2id, proper salting, key stretching

## Testing & Validation Tools

### WebAuthn Testing
- SimpleWebAuthn Testing: https://webauthn.io/
- FIDO2 Test Suite: https://github.com/duo-labs/py_webauthn
- WebAuthn Playground: https://github.com/MasterKale/SimpleWebAuthn/tree/master/examples

### API Testing
- Postman: https://www.postman.com/
- Insomnia: https://insomnia.rest/
- Thunder Client: https://www.thunderclient.com/

### Security Testing
- OWASP ZAP: https://www.zaproxy.org/
- Burp Suite: https://portswigger.net/burp
- Hydra (Password Brute Force): https://github.com/vanhauser-thc/thc-hydra

### Load Testing
- k6: https://k6.io/
- Apache JMeter: https://jmeter.apache.org/
- Locust: https://locust.io/

## Industry Standards & Best Practices

### NIST Cybersecurity
- NIST SP 800-63-3 (Digital Identity Guidelines): https://pages.nist.gov/800-63-3/
- NIST SP 800-52 Rev 2 (TLS Guidelines): https://csrc.nist.gov/publications/detail/sp/800-52/rev-2/final
- NIST SP 800-63B (Authentication): https://pages.nist.gov/800-63-3/sp800-63b.html

### FIDO Alliance
- FIDO2 Specifications: https://fidoalliance.org/fido2/fido2-web-authentication/
- Certified Products: https://fidoalliance.org/certification/certified-products/
- Developer Resources: https://fidoalliance.org/developer-resources/

### CIS Controls
- CIS Controls v8: https://www.cisecurity.org/cis-controls/v8
- CIS Microsoft Azure Foundations Benchmark: https://www.cisecurity.org/benchmark/microsoft_azure

## Conference Talks & Articles

### OWASP Events
- AppSec Global: https://www.globalappsec.org/
- OWASP Top 10: https://owasp.org/Top10/

### WebAuthn & Passkeys
- "Passkeys: The Future of Authentication": https://www.webauthn.io/
- "FIDO2 & WebAuthn: Complete Guide": https://simplewebauthn.dev/docs/guide/
- "Passwordless Authentication Best Practices": https://pages.nist.gov/800-63-3/

## Related Skills

- **moai-security-api**: OAuth 2.1, JWT implementation, rate limiting
- **moai-security-encryption**: Password hashing, session encryption
- **moai-security-owasp**: Brute force prevention, session security
- **moai-domain-web-app**: User authentication in web applications

