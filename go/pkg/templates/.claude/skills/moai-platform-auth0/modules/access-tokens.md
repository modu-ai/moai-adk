# Access Tokens

Access tokens enable applications to call APIs after users authenticate and authorize access. They serve as credentials proving the bearer has permission to perform specific actions.

## Token Types

### Opaque Access Tokens

Characteristics:
- Proprietary format
- No user information visible
- Require server-side validation
- Smaller size

Validation:
- Call token introspection endpoint
- Auth0 validates and returns token info
- Suitable when token details needed server-side

### JWT Access Tokens

Characteristics:
- Self-contained (claims included)
- Standard format
- Local validation possible
- Larger size

Validation:
- Verify signature locally
- Check claims (expiration, audience)
- No Auth0 call required
- Faster validation

When Issued:
- For custom APIs
- For Management API
- When JWT format explicitly configured

## Token Structure

### Standard Claims

iss (Issuer):
- Auth0 tenant URL
- Identifies token source

sub (Subject):
- User identifier
- Format: auth0|user_id or provider|id

aud (Audience):
- API identifier(s)
- Must include your API identifier

azp (Authorized Party):
- Client ID of requesting application
- Identifies which app obtained token

scope:
- Granted permissions
- Space-separated list

exp (Expiration):
- Token expiration timestamp
- Unix epoch time

iat (Issued At):
- Token issue timestamp
- Unix epoch time

### Custom Claims

Adding custom claims:
- Use Actions post-login
- Namespace required
- Include in access token

Namespace Format: https://your-domain.com/claim-name

## Token Lifetime

### Default Lifetimes

Custom APIs:
- Default: 86400 seconds (24 hours)
- Configurable per API

Management API:
- 24 hours maximum
- Cannot exceed 24 hours

Userinfo Endpoint:
- Implicit flow: 7200 seconds (2 hours)
- Authorization Code/Hybrid: 86400 seconds (24 hours)

### Configuring Lifetime

API Settings:
1. Navigate to Dashboard > Applications > APIs
2. Select your API
3. Configure Token Expiration
4. Save changes

Considerations:
- Shorter = more secure, more token refreshes
- Longer = better performance, more risk
- Balance based on sensitivity

## Token Validation

### JWT Validation Steps

1. Parse token structure
2. Decode header (get algorithm, key ID)
3. Retrieve signing key from JWKS
4. Verify signature
5. Validate expiration (exp)
6. Validate issuer (iss)
7. Validate audience (aud)
8. Validate scopes as needed

### Opaque Token Validation

Call introspection endpoint:
- POST to /oauth/introspect
- Include token and credentials
- Receive token metadata

### Validation Libraries

Use official libraries:
- Handles signature verification
- Manages key caching
- Validates standard claims

## Scopes and Permissions

### OIDC Scopes

openid:
- Required for OIDC
- Enables ID token

profile:
- User profile claims
- name, nickname, picture, etc.

email:
- Email-related claims
- email, email_verified

### Custom API Scopes

Define in API settings:
- Specific permissions
- Granted to applications
- Included in access token

Example Scopes:
- read:users
- write:posts
- admin:settings

### Scope Validation

On API side:
- Extract scope claim
- Check required scope present
- Deny if missing

## Token Usage

### API Requests

Include in Authorization header:
- Format: Authorization: Bearer {access_token}
- Standard OAuth 2.0 approach

### Token Refresh

When token expires:
- Use refresh token to get new access token
- Silent authentication for SPAs
- Automatic in SDKs

### Multiple APIs

Single token for multiple APIs:
- Configure same audience
- Include all required scopes
- Or use token exchange

## Security Considerations

### Treat as Opaque

Even for JWTs:
- Don't decode on client for decisions
- Only validate on resource server
- Token format may change

### Secure Transmission

- Always use HTTPS
- Never in URL parameters
- Proper header usage

### Token Storage

Browser:
- Memory preferred
- Avoid localStorage

Server:
- Secure server-side storage
- Encrypted if persisted

### Minimal Scope

Request only needed scopes:
- Principle of least privilege
- Reduces impact if compromised
- Better user consent experience

## Best Practices

### Lifetime

- Short lifetimes for sensitive APIs
- Use refresh tokens for session continuity
- Balance security and performance

### Validation

- Always validate before trusting
- Use established libraries
- Cache validation results appropriately

### Scope Design

- Granular scopes
- Clear naming convention
- Document scope requirements

### Error Handling

Handle token errors:
- 401 Unauthorized: Token invalid/expired
- 403 Forbidden: Insufficient scope
- Implement refresh logic
