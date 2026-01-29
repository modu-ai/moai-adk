# Token Best Practices

Comprehensive security recommendations for token management in Auth0 implementations covering storage, validation, lifecycle, and algorithm selection.

## Secret Management

### Signing Keys

Treat as Critical Credentials:
- Never expose signing secrets
- Secure storage with access controls
- Regular rotation schedule
- Audit access to keys

Key Protection:
- Hardware security modules (HSM) when possible
- Encrypted at rest
- Limited access

### Token Secrets

For HS256 (symmetric):
- Use cryptographically strong secrets
- Never share across environments
- Rotate periodically
- Same protection as passwords

## Token Payload

### Sensitive Data

Never Store in Tokens:
- Passwords or credentials
- Personal Identifiable Information (PII) beyond necessary
- Financial data
- Health information
- Any data requiring encryption

Reason:
- JWTs are signed, not encrypted
- Base64 encoding is not encryption
- Anyone can decode and read payload

### Claim Minimization

Include Only Necessary Claims:
- Reduces token size
- Limits data exposure
- Better performance
- Smaller attack surface

Standard Claims:
- iss, sub, aud, exp, iat (required)
- Additional claims only as needed
- Namespace custom claims

## Token Lifetime

### Access Token Lifetime

Recommendations:
- Short lifetimes (minutes to hours)
- Balance security and UX
- Consider operation sensitivity

Sensitive APIs: 5-15 minutes
Standard APIs: 1-24 hours
Low-risk APIs: Up to 24 hours

### Refresh Token Expiration

Configure Both:
- Idle timeout (inactivity)
- Absolute timeout (maximum lifetime)

Typical Values:
- Idle: 7-14 days
- Absolute: 30-90 days

### ID Token Lifetime

Considerations:
- Short lifetime recommended
- Used for authentication moment
- Re-validate for sensitive operations

## Token Storage

### Server-Side Applications

Preferred Storage:
- Server-side session
- Encrypted database
- HTTP-only secure cookies

Security Measures:
- Encryption at rest
- Secure cookie flags
- Session binding

### Single-Page Applications

Challenges:
- No secure storage in browser
- XSS vulnerability concerns
- Cookie limitations

Recommendations:
- Memory storage (most secure)
- Refresh token rotation
- Consider Backend-for-Frontend pattern

### Mobile Applications

Use Platform Security:
- iOS Keychain
- Android Keystore
- Encrypted SharedPreferences

Never:
- Store in plain text files
- Use unencrypted storage
- Log tokens

## Token Transmission

### Always HTTPS

Requirement:
- All token transmission over TLS
- Never HTTP for tokens
- Verify certificate validity

### Header Usage

Authorization Header:
- Standard: Authorization: Bearer {token}
- DPoP: Authorization: DPoP {token}

Avoid:
- URL query parameters (logged, cached)
- Request body for GETs

### CORS Considerations

For SPAs:
- Proper CORS configuration
- Restrict allowed origins
- Credential handling

## Algorithm Selection

### RS256 vs HS256

Prefer RS256 (Asymmetric):
- Public key for verification
- No secret sharing required
- Key rotation without app changes
- Multi-audience support

HS256 (Symmetric):
- Faster performance
- Single shared secret
- Both parties need secret
- Secret compromise = complete breach

### Algorithm Security

Never Allow:
- "none" algorithm
- Algorithm confusion attacks
- Weak algorithms

Validation:
- Explicitly specify expected algorithm
- Reject unexpected algorithms

## Token Validation

### Library Usage

Always Use Libraries:
- Proven implementations
- Regular security updates
- Handle edge cases

Never:
- Roll your own crypto
- Skip validation steps
- Trust without verification

### Validation Checklist

For Every Token:
1. Verify signature
2. Check expiration (exp)
3. Validate issuer (iss)
4. Validate audience (aud)
5. Check additional claims as needed

### Caching

Signing Keys:
- Cache JWKS for performance
- Set appropriate TTL
- Invalidate on signature failures
- Handle key rotation

Validation Results:
- Cache carefully
- Consider token lifetime
- Invalidate on logout

## Refresh Token Management

### Rotation

Enable Rotation:
- New token on each use
- Old token invalidated
- Detects compromise

Reuse Detection:
- Alert on old token use
- Revoke all tokens
- Force re-authentication

### Limits

Be Aware:
- Maximum 200 active per user per app
- Oldest revoked when exceeded
- Clean up test tokens

Testing Strategy:
- Create test users
- Delete after testing
- Avoid token accumulation

## Token Reuse

### Minimize Requests

Cache and Reuse:
- Don't request new token for each call
- Use until near expiration
- Refresh proactively

Benefits:
- Better performance
- Reduced Auth0 load
- Lower latency

### Token Recycling

Pattern:
- Check token expiration before use
- Refresh when close to expiry
- Buffer time for clock skew

## Error Handling

### Token Errors

Handle Gracefully:
- 401: Token invalid or expired
- 403: Insufficient permissions
- Implement refresh logic
- Clear UI for users

### Refresh Errors

When Refresh Fails:
- Redirect to login
- Clear stored tokens
- Preserve user context if possible

## Monitoring

### Track Token Usage

Metrics:
- Token issuance rates
- Refresh patterns
- Revocation events
- Validation failures

Alerts:
- Unusual token activity
- Mass token operations
- Repeated failures

### Audit

Log:
- Token requests (not token values)
- Refresh operations
- Revocation actions
- Access patterns

Review:
- Regular security audits
- Token lifecycle analysis
- Access pattern review
