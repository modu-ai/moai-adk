# Token Revocation

Token revocation provides mechanisms to invalidate tokens before their natural expiration, essential for security incident response and proper session management.

## Revocable vs Non-Revocable Tokens

### Refresh Tokens

Revocable: Yes

Methods:
- Auth0 Management API
- OAuth revocation endpoint
- User logout action

Effect:
- Token immediately invalidated
- Future use rejected
- Associated sessions affected

### Access Tokens

Revocable: No (by design)

Reason:
- Self-contained validation
- No server check required
- Performance optimization

Mitigation:
- Short token lifetimes
- Refresh token revocation
- Accept some validity window

### ID Tokens

Revocable: No

Reason:
- Similar to access tokens
- Self-contained
- Validated locally

Mitigation:
- Short lifetimes
- Re-validate on sensitive operations
- Session management layer

## Revocation Methods

### OAuth Revocation Endpoint

Endpoint: POST /oauth/revoke

Purpose:
- Standard OAuth 2.0 revocation
- Refresh token revocation
- Client-initiated

Request Parameters:
- token: Refresh token to revoke
- client_id: Application client ID
- client_secret: For confidential clients

Response:
- 200 OK on success
- No content returned
- Revocation is idempotent

### Management API

Revoke User's Refresh Tokens:

Endpoint: DELETE /api/v2/users/{user_id}/refresh-tokens

Effect:
- All refresh tokens for user revoked
- Forces re-authentication
- Administrative action

Revoke Specific Token:

Endpoint: DELETE /api/v2/device-credentials/{id}

Effect:
- Specific token revoked
- Other tokens unaffected

### User Logout

Federated Logout:
- Revoke Auth0 session
- Optionally log out of IdP
- Clear application session

Implementation:
- Redirect to /v2/logout
- Include returnTo parameter
- Revoke refresh token separately

## Revocation Strategies

### On Logout

Always revoke:
- Application-stored refresh tokens
- Clear session storage
- Redirect through logout endpoint

Implementation Steps:
1. Call /oauth/revoke with refresh token
2. Clear local token storage
3. Redirect to /v2/logout
4. Clear application session

### On Security Event

Password Change:
- Consider revoking all refresh tokens
- Force re-authentication
- Notify user of action

Suspicious Activity:
- Immediately revoke tokens
- Alert user
- Require MFA on next login

Account Compromise:
- Revoke all user tokens
- Reset password
- Review account activity

### Session Management

Timeout-Based:
- Allow natural expiration
- Idle timeout for refresh tokens
- Absolute timeout for sessions

Active Revocation:
- User-initiated logout
- Administrative action
- Automated security response

## Implementation Patterns

### Single Logout

When user logs out:
1. Revoke refresh token (if stored)
2. Clear access token from memory
3. Clear ID token from memory
4. Redirect to Auth0 logout
5. Clear application session/cookies

### Global Logout

Revoke all user sessions:
1. List user's refresh tokens
2. Revoke all tokens
3. Send session termination to applications
4. Force re-authentication everywhere

### Selective Revocation

Revoke specific session:
1. Identify token/session to revoke
2. Call appropriate revocation endpoint
3. Update session registry
4. Optionally notify user

## Timing Considerations

### Access Token Window

After Refresh Token Revocation:
- Existing access tokens remain valid
- Until their expiration
- This is by design

Mitigation:
- Short access token lifetimes
- Accept risk window
- Additional checks for sensitive operations

### Immediate Effect

Refresh Token Revocation:
- Immediate effect
- No new tokens issued
- Existing tokens from this refresh continue

Session Logout:
- Auth0 session terminated
- New authentications required
- Active tokens still valid until expiration

## Best Practices

### Always Revoke on Logout

- Don't rely on expiration alone
- Explicit revocation is cleaner
- Reduces security exposure

### Short Access Token Lifetimes

- Reduces risk window after revocation
- Balance with refresh frequency
- Consider application type

### Monitor Revocation Events

Track:
- Logout frequency
- Revocation patterns
- Forced revocations

Alert on:
- Mass revocations
- Unusual patterns
- Failed revocation attempts

### Test Revocation Flow

Verify:
- Revoked tokens are rejected
- Application handles rejection gracefully
- Re-authentication flow works

## Error Handling

### Revocation Errors

Token Not Found:
- May already be revoked
- May be expired
- Treat as success

Permission Denied:
- Check client credentials
- Verify token ownership
- Review API permissions

Rate Limiting:
- Implement backoff
- Batch revocations if possible
- Monitor usage

### Post-Revocation Errors

Application Token Rejection:
- 401 Unauthorized responses
- Trigger re-authentication
- Clear stored tokens

User Experience:
- Clear message about session end
- Smooth re-login flow
- Preserve context if appropriate
