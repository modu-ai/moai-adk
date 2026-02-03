# JWT Fundamentals

JSON Web Tokens (JWTs) are an open standard (RFC 7519) providing a compact, self-contained method for securely transmitting information between parties as a JSON object.

## Structure

JWTs consist of three parts separated by dots:
- Header
- Payload
- Signature

Format: header.payload.signature

### Header

Contains metadata about the token:
- typ: Token type (JWT)
- alg: Signing algorithm

Common Algorithms:
- HS256: HMAC with SHA-256 (symmetric)
- RS256: RSA with SHA-256 (asymmetric)
- ES256: ECDSA with SHA-256 (asymmetric)

### Payload

Contains claims (statements about the entity and additional data).

Registered Claims (standard):
- iss (issuer): Token issuer
- sub (subject): Subject identifier
- aud (audience): Intended recipients
- exp (expiration): Expiration time
- nbf (not before): Token valid after this time
- iat (issued at): Token issue time
- jti (JWT ID): Unique identifier

Public Claims:
- Registered in IANA JWT Registry
- Common across implementations

Private Claims:
- Custom claims for specific use
- Require namespace in Auth0

### Signature

Ensures token integrity and authenticity.

Signature Creation:
- Encode header and payload
- Apply signing algorithm
- Use secret (HS256) or private key (RS256)

## Signing Algorithms

### HS256 (Symmetric)

Characteristics:
- Single shared secret
- Secret known by issuer and verifier
- Simpler key management
- Must protect secret carefully

Use Cases:
- Single application scenarios
- Trusted environments
- Simple implementations

### RS256 (Asymmetric)

Characteristics:
- Public/private key pair
- Private key signs, public key verifies
- Public key can be shared freely
- No secret sharing required

Advantages:
- Multiple verifiers without secret sharing
- Key rotation without application changes
- Better for distributed systems

Auth0 Recommendation: Use RS256 for most scenarios.

### Algorithm Comparison

HS256:
- Faster signing/verification
- Simpler setup
- Secret must be protected everywhere

RS256:
- Slower but more flexible
- Only issuer has private key
- Public key verification
- Better for microservices

## Validation

### Signature Verification

Steps:
1. Decode header (base64url)
2. Identify algorithm
3. Get verification key (secret or public key)
4. Verify signature mathematically

### Claim Validation

Required Checks:
- exp: Token not expired
- iss: Issuer matches expected value
- aud: Audience includes your application

Optional Checks:
- nbf: Current time is after not-before
- iat: Issued at time is reasonable
- Custom claims as needed

### Key Management

For RS256:
- Retrieve JWKS from issuer
- Match kid (key ID) in header
- Cache keys with appropriate TTL
- Handle key rotation gracefully

JWKS Endpoint: {your-domain}/.well-known/jwks.json

## Security Considerations

### Signed vs Encrypted

Auth0 JWTs are Signed (JWS):
- Signature verifies integrity
- Content is NOT encrypted
- Anyone can read payload
- Only issuer can create valid signature

JWE (Encrypted):
- Content is encrypted
- Requires decryption key
- Used for sensitive data
- More complex implementation

### Security Best Practices

Never Trust Unverified Tokens:
- Always verify signature
- Always check claims
- Use established libraries

Sensitive Data:
- Never store sensitive data in JWT payload
- Payload is only base64 encoded, not encrypted
- Assume payload contents are public

Transmission:
- Always use HTTPS
- Never include in URL parameters
- Use Authorization header

Algorithm Attacks:
- Never accept "none" algorithm
- Specify expected algorithm in verification
- Use library that enforces algorithm

## Common Vulnerabilities

### Algorithm Confusion

Attack: Changing RS256 to HS256 and using public key as secret.

Prevention:
- Explicitly specify expected algorithm
- Use libraries that require algorithm specification

### Token Injection

Attack: Modifying payload and re-signing with weak key.

Prevention:
- Use strong keys
- Validate all claims
- Check issuer strictly

### Replay Attacks

Attack: Reusing captured tokens.

Prevention:
- Short expiration times
- Use jti claim for uniqueness
- Implement token binding (DPoP)

## Implementation

### Using Libraries

Recommended:
- Use official Auth0 SDKs
- Use well-maintained JWT libraries
- Avoid custom implementations

Popular Libraries:
- Node.js: jsonwebtoken, jose
- Python: PyJWT, python-jose
- Java: java-jwt, nimbus-jose-jwt
- .NET: System.IdentityModel.Tokens.Jwt

### Caching

JWKS Caching:
- Cache public keys
- Set appropriate TTL
- Invalidate on verification failure
- Handle rotation gracefully

Token Caching:
- Cache validation results
- Consider token lifetime
- Invalidate on expiration

## Auth0 Specifics

### Token Signing

Auth0 uses:
- RS256 by default for ID tokens
- Configurable per API for access tokens
- Tenant-specific signing keys

### JWKS Location

Endpoint: https://{your-tenant}.auth0.com/.well-known/jwks.json

Contains:
- Public keys for verification
- Key IDs for matching
- Algorithm information

### Key Rotation

Auth0 rotates signing keys:
- New key added to JWKS
- Old key remains temporarily
- Applications should handle multiple keys
- Cache should handle rotation gracefully
