# Refresh Tokens

Refresh tokens enable applications to request new access or ID tokens without requiring users to re-authenticate. They provide session continuity while maintaining security.

## Purpose

Enable Long Sessions:
- Access tokens have short lifetimes
- Refresh tokens maintain session
- User stays logged in

Avoid Re-authentication:
- Silent token renewal
- Better user experience
- Background token refresh

## Token Constraints

### Maximum Active Tokens

Limit: 200 active refresh tokens per user per application

When Limit Exceeded:
- Oldest token automatically revoked
- New token issued normally
- No error to application

What Counts:
- Only active tokens
- Expired tokens don't count
- Revoked tokens don't count

### Token Scope

Refresh tokens can obtain:
- New access tokens
- New ID tokens (if openid scope included)
- Cannot expand scopes

## Configuration

### Enable Refresh Tokens

API Settings:
1. Navigate to Dashboard > Applications > APIs
2. Select your API
3. Enable "Allow Offline Access"
4. Save changes

Application Request:
- Include offline_access scope
- Token returned with authorization response

### Token Rotation

Enable rotation for enhanced security:
- New refresh token issued with each use
- Previous token invalidated
- Detects token reuse attacks

Configuration:
1. Dashboard > Applications > Applications
2. Select application
3. Scroll to Refresh Token Rotation
4. Enable rotation

Reuse Detection:
- If old token used, all tokens revoked
- Potential theft detected
- User must re-authenticate

### Token Expiration

Configure expiration behavior:

Absolute Expiration:
- Maximum token lifetime
- Token expires regardless of use
- Forces re-authentication

Idle Expiration:
- Expires after inactivity period
- Resets on each use
- Balances security and convenience

Configuration:
1. Dashboard > Applications > Applications
2. Select application
3. Configure expiration settings

## Token Exchange

### Getting New Access Token

Request:
- POST to /oauth/token
- grant_type: refresh_token
- refresh_token: {your_refresh_token}
- client_id: {your_client_id}
- client_secret: {if confidential client}

Response:
- New access token
- New ID token (if applicable)
- New refresh token (if rotation enabled)
- Expiration information

### Token Refresh Flow

1. Access token expires
2. Application uses refresh token
3. Auth0 validates refresh token
4. New tokens issued
5. Application continues operation

## Security Features

### Rotation Benefits

Token Compromise Detection:
- Stolen token can only be used once
- Legitimate use invalidates stolen copy
- Attack detected on reuse

Reduced Exposure:
- Token lifetime effectively limited
- New token with each use
- Less time for attack

### Expiration Benefits

Idle Timeout:
- Inactive sessions expire
- Reduces risk of abandoned sessions
- Encourages active session management

Absolute Timeout:
- Forces periodic re-authentication
- Limits maximum exposure
- Ensures credential freshness

### Revocation

Revoke via Management API:
- Endpoint: POST /api/v2/users/{id}/refresh-tokens/revoke
- Revoke all user tokens
- Revoke specific token

Revoke on Logout:
- Call /oauth/revoke endpoint
- Include refresh token
- Ensure session terminated

## SDK Support

### Web Applications

Supported in:
- Node.js SDK
- ASP.NET Core SDK
- PHP SDK
- Java SDK

Implementation:
- Store refresh token server-side
- Automatic refresh in middleware
- Secure token storage

### Single-Page Applications

Considerations:
- Browser storage challenges
- Refresh token rotation recommended
- Consider silent authentication alternative

Implementation:
- Use Auth0 SPA SDK
- SDK handles token management
- Automatic background refresh

### Mobile/Native Applications

Benefits:
- Significant UX improvement
- Reduces authentication friction
- Essential for mobile experience

Implementation:
- Store in secure platform storage
- iOS Keychain, Android Keystore
- SDK handles refresh automatically

## Supported Flows

Flows that issue refresh tokens:
- Authorization Code Flow
- Authorization Code Flow with PKCE
- Resource Owner Password Grant
- Device Authorization Flow

Flows that do NOT issue refresh tokens:
- Implicit Flow
- Client Credentials Flow

## Best Practices

### Storage

Server-Side:
- Encrypted database storage
- Server-side session association
- Never expose to client

Client-Side (when necessary):
- Platform-specific secure storage
- Never in localStorage
- Encrypted if possible

### Expiration Strategy

High Security:
- Shorter idle timeout (hours)
- Reasonable absolute timeout (days)
- Enable rotation

User Convenience:
- Longer idle timeout (days)
- Longer absolute timeout (weeks/months)
- Enable rotation for security

### Logout

Always revoke on logout:
- Call revocation endpoint
- Clear local storage
- Ensure complete session termination

### Monitoring

Track refresh token usage:
- Unusual refresh patterns
- Multiple location refreshes
- Reuse detection events

## Troubleshooting

Token Not Returned:
- Check offline_access scope requested
- Verify API allows offline access
- Confirm OIDC-conformant enabled

Refresh Failing:
- Token may be expired
- Token may be revoked
- Rotation reuse may be detected

Too Many Tokens:
- Limit is 200 per user per app
- Old tokens auto-revoked
- Create test users for testing
