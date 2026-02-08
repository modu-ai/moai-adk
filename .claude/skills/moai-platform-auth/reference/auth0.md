# Auth0 Platform Reference

Enterprise-grade authentication with advanced security features, compliance certifications, and attack protection.

## Dashboard Navigation

Access security features from the Auth0 Dashboard:

- Attack Protection: Dashboard > Security > Attack Protection
- Multi-Factor Auth: Dashboard > Security > Multi-factor Auth
- Security Center: Dashboard > Security > Security Center
- Applications: Dashboard > Applications > Applications
- APIs: Dashboard > Applications > APIs
- Actions: Dashboard > Actions > Library
- Organizations: Dashboard > Organizations

## Attack Protection

### Bot Detection

Protects against automated attacks using machine learning and behavioral analysis.

**Configuration:**
Navigate to Dashboard > Security > Attack Protection > Bot Detection.

**Sensitivity Levels:**
- Low: Minimal false positives, may miss sophisticated bots
- Medium: Balanced detection (recommended for most applications)
- High: Maximum protection, may increase false positives

**Response Types:**
- Auth Challenge (recommended): Multi-step verification for suspicious requests
- Simple CAPTCHA: Standard CAPTCHA challenge
- Third-party: Integration with reCAPTCHA or hCaptcha

**IP AllowList:**
Supports up to 100 IP addresses or CIDR ranges that bypass bot detection.

**Supported Flows:**
- Universal Login
- Classic Login
- Lock.js v12.4.0 or higher
- Native mobile applications

**Not Supported:**
- Enterprise connections (SAML, WS-Fed, LDAP)
- Social login providers
- Cross-origin authentication

### Breached Password Detection

Prevents credential stuffing by detecting compromised passwords from known breaches.

**Detection Modes:**

Standard Detection:
- Detection window: 7-13 months after breach
- Database: Public breach databases
- Free for all tiers

Credential Guard (Enterprise):
- Detection window: 12-36 hours after breach
- Database: Commercial threat intelligence
- Real-time monitoring of dark web

**Response Actions:**
- Block compromised credentials
- Send notification to user
- Send notification to admin
- Trigger webhook for custom response

**Testing:**
Use passwords starting with `AUTH0-TEST-` prefix to simulate breached credentials without blocking real users.

**Configuration:**
Enable for signup, login, or both flows. Configure notification templates and response actions.

### Brute Force Protection

Protects against password guessing attacks by rate limiting failed login attempts.

**Default Configuration:**
- Threshold: 10 failed attempts (configurable 1-100)
- Lockout duration: Until manual intervention
- Scope: Per user account and per IP address

**Protection Mechanisms:**

Account Lockout:
- Blocks account after threshold reached
- Requires password reset or admin unblock
- Sends notification email to user

IP Blocking:
- Blocks IP address after threshold reached
- Persists for 30 days
- Can be manually removed from dashboard

**Unlock Methods:**
- User clicks unblock link in email
- User changes password
- Admin removes block from dashboard
- Block expires after 30 days (IP blocks only)

**Configuration:**
Set failed attempt threshold, configure notification templates, manage blocked IPs and accounts from dashboard.

### Suspicious IP Throttling

Detects and blocks high-velocity attacks from single IP addresses using rate limiting.

**Configuration:**

Login Attempts:
- Threshold: Maximum attempts per day per IP
- Default: 100 attempts/day
- Response: HTTP 429 Too Many Requests

Signup Attempts:
- Threshold: Maximum signups per minute per IP
- Default: 10 signups/minute
- Response: HTTP 429 Too Many Requests

**Detection Logic:**
Monitors request patterns for:
- Velocity anomalies (sudden spikes)
- Distributed attacks (coordinated IPs)
- Geographic anomalies (impossible travel)

**Response Actions:**
- Rate limit with exponential backoff
- Temporary IP suspension
- Security Center alert

## Multi-Factor Authentication

### MFA Factors

**Independent Factors (at least one required):**

WebAuthn with FIDO Security Keys:
- Physical security keys (YubiKey, Titan)
- Platform authenticators (Touch ID, Face ID, Windows Hello)
- Phishing-resistant authentication
- Supports passwordless flows

One-Time Password (OTP):
- Time-based OTP (TOTP) with authenticator apps
- Compatible with Google Authenticator, Authy, 1Password
- QR code enrollment
- Backup codes for recovery

Push Notifications:
- Auth0 Guardian app for iOS and Android
- Biometric verification on mobile device
- Real-time approval notifications
- Works offline with TOTP fallback

Phone Message:
- SMS or voice call verification
- International number support
- Rate limiting to prevent abuse
- Not recommended as sole factor (SIM swap vulnerability)

Cisco Duo Security:
- Enterprise MFA integration
- Supports push, SMS, phone call, hardware tokens
- Advanced policy controls
- Requires Duo account

**Dependent Factors (require independent factor):**

WebAuthn Biometrics:
- Platform biometrics (fingerprint, face recognition)
- Requires initial enrollment with independent factor

Email Verification:
- OTP sent to registered email
- Not recommended for high-security applications

Recovery Codes:
- One-time use codes for account recovery
- Generated during MFA enrollment
- Store securely offline

### MFA Policies

**Never:**
MFA is optional, users can skip enrollment. Use for low-risk applications or gradual rollout.

**Always:**
MFA is required for all users after initial login. Use for high-security applications.

**Adaptive MFA (Enterprise):**
MFA is required based on risk assessment. Evaluates multiple signals per authentication attempt.

### Adaptive MFA Risk Signals

**NewDevice:**
Triggers when device not seen in past 30 days. Considers browser fingerprint, user agent, screen resolution.

**ImpossibleTravel:**
Triggers when user location changes impossibly fast. Calculates distance and time between login locations.

**UntrustedIP:**
Triggers when IP has suspicious activity history. Monitors failed logins, velocity attacks, known proxies/VPNs.

**Risk Scoring:**
Combines multiple signals with configurable weights. High-risk transactions require MFA regardless of session state.

### Step-Up Authentication

Enhanced verification for sensitive operations within authenticated sessions.

**Implementation Patterns:**

API Approach:
- Add high-privilege scopes to sensitive endpoints
- Request step-up with `acr_values` parameter
- Verify `amr` claim includes `mfa` in access token

Web App Approach:
- Check ID token `amr` claim for recent MFA
- Redirect to MFA challenge if needed
- Verify `auth_time` claim for time-based policies

## Token Security

### JWT Fundamentals

**Structure:**

Header:
- Algorithm: RS256 (recommended) or HS256
- Type: JWT
- Key ID: kid for signature verification

Payload (Claims):
- iss: Issuer (Auth0 tenant URL)
- sub: Subject (user ID)
- aud: Audience (API identifier)
- exp: Expiration timestamp
- iat: Issued at timestamp
- scope: Authorized permissions

Signature:
- HMAC with HS256 (symmetric)
- RSA with RS256 (asymmetric, recommended)
- ECDSA with ES256 (elliptic curve)

**Security Rules:**
- Always validate signature before trusting token
- Verify issuer matches expected Auth0 tenant
- Check expiration and reject expired tokens
- Validate audience matches your API identifier
- Use HTTPS only for token transmission
- Never store sensitive data in token payload (publicly visible)

### Access Tokens

Authorize API access with specific scopes and permissions.

**Token Types:**

Opaque Tokens:
- Random string with no embedded information
- Require introspection endpoint for validation
- Cannot be validated client-side
- Smaller payload size

JWT Access Tokens:
- Self-contained with claims
- Can be validated without API call
- Include user ID, scopes, expiration
- Larger payload size

**Key Claims:**
- iss: Auth0 tenant URL
- sub: User identifier (auth0|user_id)
- aud: API identifier (configured in Auth0 APIs)
- scope: Space-separated permission list
- exp: Expiration timestamp (default 86400 seconds / 24 hours)
- azp: Authorized party (application client ID)

**Scope Management:**
Define scopes in Auth0 APIs dashboard. Request specific scopes during authorization. Validate scopes in API middleware.

**Lifetime Configuration:**
Navigate to Dashboard > Applications > APIs > Settings. Configure token expiration (300 to 2592000 seconds). Consider security vs. user experience tradeoff.

### Refresh Tokens

Enable long-lived sessions without storing passwords.

**Configuration:**

Rotation:
- Enabled by default for public clients
- Issues new refresh token on each use
- Invalidates predecessor token
- Prevents token replay attacks

Expiration:
- Absolute: Maximum lifetime regardless of activity
- Idle: Expires after inactivity period
- Configurable per application

Limits:
- Maximum 200 active refresh tokens per user per application
- Oldest tokens automatically revoked when limit reached

**Security Features:**

Automatic Reuse Detection:
- Detects compromised tokens
- Invalidates token family on reuse
- Requires user re-authentication

Revocation:
- Programmatic via Management API
- User-initiated from dashboard
- Admin-initiated from Security Center

**Best Practices:**
- Store securely (httpOnly cookie or secure storage)
- Use rotation for public clients
- Set appropriate expiration policies
- Monitor usage patterns for anomalies
- Implement revocation on logout

### Token Best Practices

**Signing Keys:**
- Treat as critical credentials
- Rotate regularly (recommended: 90 days)
- Use RS256 over HS256 for public clients
- Store private keys securely (HSM, KMS)
- Monitor key usage from Security Center

**Validation:**
- Verify signature using Auth0 public keys (JWKS endpoint)
- Check all standard claims (iss, aud, exp)
- Validate custom claims for authorization
- Reject tokens with unknown claims
- Implement token validation middleware

**Storage:**
- Server-side: Use in-memory or encrypted database
- Web apps: httpOnly secure cookies
- SPAs: Memory only with refresh token rotation
- Mobile apps: Keychain (iOS) or Keystore (Android)
- Never: localStorage or sessionStorage (XSS risk)

**Performance:**
- Cache tokens until expiration
- Reuse tokens across requests
- Implement token refresh before expiration
- Use background refresh for silent renewal

## Sender Constraining

Binds tokens to specific clients to prevent token theft attacks.

### DPoP (Demonstrating Proof-of-Possession)

Application-layer token binding using asymmetric cryptography.

**Implementation Steps:**

Step 1 - Generate Key Pair:
- Algorithm: ES256 (ECDSA P-256) recommended for mobile
- Algorithm: RS256 (RSA 2048) for web
- Store private key securely (never transmit)
- Generate JSON Web Key (JWK) from public key

Step 2 - Create DPoP Proof:
- Generate JWT with specific structure
- Sign with private key
- Include in DPoP header (not Authorization header)

Step 3 - Token Request:
- Send DPoP proof JWT in DPoP header
- Auth0 binds token to JWK thumbprint
- Token includes cnf claim with JWK thumbprint

Step 4 - API Requests:
- Create new DPoP proof for each request
- Include ath claim (hash of access token)
- Send in DPoP header alongside Authorization header
- Resource server validates DPoP proof

**DPoP Proof Structure:**

Header:
- typ: dpop+jwt
- alg: ES256 or RS256
- jwk: Public key in JWK format

Payload:
- jti: Unique identifier (UUID)
- htm: HTTP method (GET, POST, etc.)
- htu: HTTP URI (API endpoint URL)
- iat: Issued at timestamp
- ath: Access token hash (SHA-256, base64url-encoded)

**Error Handling:**

use_dpop_nonce Error:
- Server requires nonce for replay protection
- Extract nonce from DPoP-Nonce header
- Include in next DPoP proof as nonce claim

**Browser Considerations:**
- Private key storage in Web Crypto API
- Consider key rotation complexity
- Implement key migration strategy

### mTLS (Mutual TLS)

Transport-layer token binding using X.509 certificates.

**Requirements:**
- Confidential clients only (no SPAs or mobile apps)
- Enterprise Plan with Highly Regulated Identity (HRI) add-on
- PKI infrastructure with certificate authority
- Client certificate management system

**Implementation Process:**

Step 1 - Certificate Setup:
- Generate X.509 client certificate
- Register certificate with certificate authority
- Configure client to present certificate during TLS handshake

Step 2 - mTLS Connection:
- Client establishes TLS connection to Auth0
- Presents client certificate during handshake
- Auth0 validates certificate chain
- Calculates SHA-256 thumbprint of certificate

Step 3 - Token Binding:
- Auth0 embeds certificate thumbprint in token
- Token includes cnf claim with x5t#S256 value
- Token cannot be used without matching certificate

Step 4 - API Protection:
- Resource server requires mTLS connection
- Extracts client certificate from TLS connection
- Calculates certificate thumbprint
- Compares with cnf claim in access token
- Rejects if thumbprints do not match

**Certificate Management:**
- Implement certificate rotation before expiration
- Monitor certificate validity periods
- Use short-lived certificates (recommended: 90 days)
- Automate certificate renewal
- Plan for certificate revocation

**Comparison with DPoP:**

DPoP Advantages:
- Works with public clients (SPAs, mobile)
- No PKI infrastructure required
- Simpler certificate management
- Browser and mobile SDK support

mTLS Advantages:
- Transport-layer protection
- Stronger security guarantees
- Industry-standard for high-security environments
- Regulatory acceptance (FAPI, financial services)

## Compliance and Certifications

### Highly Regulated Identity (HRI)

Enterprise add-on for financial services and highly regulated industries.

**Features:**

Strong Customer Authentication:
- Minimum two independent authentication factors
- Dynamic linking of transaction details to authorization
- Supports FAPI 1 Advanced and FAPI 2 compliance

Pushed Authorization Requests (PAR):
- Authorization parameters posted to Auth0 before redirect
- Returns request_uri for authorization request
- Prevents parameter tampering and leakage

JWT-Secured Authorization Requests (JAR):
- Authorization parameters in signed JWT
- Prevents parameter modification
- Supports request and request_uri parameters

Access Token Encryption (JWE):
- Encrypts access token payload
- Prevents token inspection in transit
- Requires client to decrypt with private key

Private Key JWT Authentication:
- Client authentication using asymmetric keys
- Private key never transmitted
- Supports key rotation without downtime

mTLS Client Authentication:
- Certificate-based client authentication
- Strongest security for confidential clients
- Required for FAPI compliance in some jurisdictions

### FAPI Implementation

Financial-grade API security profile for open banking and financial services.

**FAPI 1 Advanced Requirements:**
- Private Key JWT or mTLS client authentication
- PAR (Pushed Authorization Requests)
- JAR (JWT-Secured Authorization Requests)
- Sender-constrained tokens (mTLS or DPoP)
- Refresh token rotation
- PKCE for all flows

**Auth0 FAPI Certification:**
- FAPI 1 Advanced OpenID Provider certified
- OpenID Foundation certified
- Supports Open Banking implementations

**Implementation Checklist:**
1. Enable HRI add-on on Enterprise plan
2. Configure PAR endpoint for application
3. Enable JAR with signing algorithm (RS256, PS256, ES256)
4. Set up mTLS or Private Key JWT client authentication
5. Enable token binding (mTLS preferred for FAPI)
6. Configure refresh token rotation with reuse detection
7. Test with FAPI conformance suite

### GDPR Compliance

General Data Protection Regulation compliance features.

**Data Roles:**
- Customer: Data Controller (determines purpose and means)
- Auth0: Data Processor (processes data on behalf of controller)
- Data Processing Addendum (DPA) available in Enterprise contracts

**User Rights Implementation:**

Right of Access:
- User can view their profile data
- Management API provides user data export
- Includes profile, metadata, logs

Right to Portability:
- Export user data in JSON format
- Includes all user attributes and metadata
- API: GET /api/v2/users/{id}

Right to Erasure:
- Delete user account and all associated data
- API: DELETE /api/v2/users/{id}
- Anonymizes logs after deletion
- Grace period configurable (default: immediate)

Right to Rectification:
- Users can update profile information
- Verification required for sensitive changes
- Audit trail in user logs

**Consent Management:**
- Track consent with user metadata
- Prompt for consent during signup or login
- Store consent version and timestamp
- Allow consent withdrawal

**Security Measures:**
- Profile encryption at rest
- Encryption in transit (TLS 1.2+)
- Breached password detection
- Brute force protection
- Regular security audits
- SOC 2 Type 2 certified

**Data Residency:**
- Choose Auth0 region during tenant creation
- Regions: US, EU, Australia, Japan
- Data stored in chosen region only
- Cannot change region after tenant creation

### Other Certifications

**ISO 27001/27017/27018:**
- Information security management
- Cloud security controls
- Personal data protection in public clouds

**SOC 2 Type 2:**
- Annual audit of security controls
- Security, availability, confidentiality
- Report available to Enterprise customers

**CSA STAR:**
- Cloud Security Alliance certification
- Level 2: Attestation
- Published in CSA registry

**HIPAA:**
- Business Associate Agreement (BAA) available
- Enterprise plan required
- PHI handling capabilities
- Audit logging and controls

**PCI DSS:**
- Compliant infrastructure
- Tokenization of payment data
- Not recommended for direct payment processing
- Use dedicated payment processor (Stripe, Square)

## Security Center

Real-time security monitoring and threat detection dashboard.

**Access:**
Dashboard > Security > Security Center

**Threat Categories:**

Credential Stuffing:
- Automated credential testing from breaches
- High-velocity login attempts across accounts
- Bot detection correlation

Signup Attacks:
- Automated account creation
- Suspicious IP throttling correlation
- Pattern recognition for fake accounts

MFA Bypass Attempts:
- Multiple MFA failures per user
- MFA enrollment anomalies
- Social engineering indicators

**Filtering Options:**
- Time period (up to 14 days retention)
- Applications
- Connections
- Attack types
- IP addresses

**Metrics Tracked:**
- Bot detection challenges and blocks
- IP throttling events and blocks
- Brute force protection triggers
- Breached password detections
- MFA enrollment and verification rates
- MFA success and failure rates
- Anomaly detection alerts

**Automated Responses:**
Configure automatic actions when threats detected:
- Block IP addresses
- Suspend user accounts
- Require MFA enrollment
- Trigger webhooks for custom logic
- Send alerts to security team

**Integration:**
- Export logs to SIEM systems
- Webhook notifications for real-time alerts
- Management API for programmatic access
- Log streaming to external services

## Application Credentials

Client authentication methods for secure application identification.

**Client Secret (Default):**

Characteristics:
- Symmetric secret (same key for client and Auth0)
- Simple implementation
- Vulnerable if secret is compromised
- Must be stored securely server-side
- Cannot be rotated without downtime

Security Considerations:
- Treat as password (never commit to VCS)
- Store in environment variables or secret manager
- Rotate regularly (recommended: 90 days)
- Monitor for exposure in logs or error messages

**Private Key JWT (Enterprise):**

Characteristics:
- Asymmetric key pairs (public/private)
- Private key never transmitted to Auth0
- Short-lived JWT assertions (recommended: 5 minutes)
- Zero-downtime key rotation
- Phishing-resistant

Implementation:
- Generate RSA (RS256, RS384, PS256) or ECDSA (ES256) key pair
- Register public key in Auth0 application settings
- Create JWT assertion for each token request
- Sign assertion with private key
- Include assertion in client_assertion parameter

JWT Assertion Structure:
- Header: alg (RS256), kid (optional key ID)
- Payload: iss (client ID), sub (client ID), aud (Auth0 token endpoint), exp (5 minutes), iat (now), jti (UUID)

Key Management:
- Register up to two public keys per application
- Implement key rotation without downtime
- Monitor key usage from dashboard
- Revoke compromised keys immediately

**mTLS for OAuth (HRI):**

Characteristics:
- X.509 client certificates
- Strongest security
- Requires PKI infrastructure
- Certificate-based authentication
- Enterprise plan with HRI add-on required

Implementation:
- Generate client certificate from CA
- Register certificate in Auth0 application
- Configure client to present certificate during TLS handshake
- Auth0 validates certificate and binds to client

Certificate Management:
- Monitor certificate expiration
- Automate certificate renewal
- Plan for certificate revocation
- Test certificate chain validation

**Recommendation:**
- Public clients (SPAs, mobile): PKCE only, no client secret
- Confidential clients (servers): Private Key JWT (preferred) or client secret
- High-security environments: mTLS with HRI

## Continuous Session Protection

Monitor and manage user sessions for security anomalies.

**Implementation:**
Use Auth0 Actions in token refresh flow to evaluate session context.

**Session Context Available:**

IP Address Tracking:
- Compare current IP with previous logins
- Detect geographic anomalies
- Flag IP changes during active session

ASN (Autonomous System Number):
- Track network provider changes
- Detect proxy or VPN usage
- Flag unusual network transitions

Device Tracking:
- Browser fingerprinting
- Device ID from mobile SDKs
- Flag device changes mid-session

Session Expiration:
- Custom logic based on user attributes
- Organization policies
- Role-based session lifetimes

**Dynamic Session Management:**

Customize Token Lifetime:
- Vary by user role (admin: 1 hour, user: 8 hours)
- Adjust by risk score
- Reduce for sensitive operations

Force Re-authentication:
- Terminate suspicious sessions
- Require step-up authentication
- Challenge based on anomaly detection

**Example Scenarios:**

Scenario 1 - IP Change Detection:
- User authenticates from IP A
- Token refresh from IP B triggers action
- Action checks geographic distance
- If impossible travel, revoke session and require re-authentication

Scenario 2 - Session Timeout by Role:
- Admin users: 1-hour session timeout
- Regular users: 8-hour session timeout
- Implemented in refresh token action
- Customizable per organization

Scenario 3 - Device Change:
- User authenticates on device A
- Refresh token used from device B
- Action detects device mismatch
- Requires MFA verification before issuing new tokens

**Monitoring:**
- Track session events in logs
- Alert on anomalous patterns
- Dashboard for active session visualization
- Integration with Security Center

## Resources

**Official Documentation:**
- Auth0 Docs: https://auth0.com/docs
- Attack Protection: https://auth0.com/docs/secure/attack-protection
- Multi-Factor Auth: https://auth0.com/docs/secure/multi-factor-authentication
- Tokens: https://auth0.com/docs/secure/tokens
- Sender Constraining: https://auth0.com/docs/secure/sender-constraining
- Compliance: https://auth0.com/docs/secure/data-privacy-and-compliance

**Auth0 SDKs:**
- Next.js: @auth0/nextjs-auth0
- React: @auth0/auth0-react
- Node.js: auth0 (Management API), express-oauth2-jwt-bearer (API protection)
- iOS: Auth0.swift
- Android: Auth0.Android

**Community:**
- Auth0 Community: community.auth0.com
- GitHub: github.com/auth0
- Stack Overflow: stackoverflow.com/questions/tagged/auth0

**Tools:**
- JWT Debugger: jwt.io
- FAPI Conformance Suite: openid.net/certification/fapi_op_testing/
- Auth0 CLI: github.com/auth0/auth0-cli
