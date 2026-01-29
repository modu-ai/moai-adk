# Tokens Overview

Auth0 uses various token types for identity and authorization. Understanding each token's purpose, structure, and security considerations is essential for secure implementations.

## Token Types

### ID Tokens

Purpose: Authenticate users to applications.

Format: JSON Web Token (JWT) per OpenID Connect specification.

Characteristics:
- Contains user identity information
- Intended only for the requesting application
- Audience claim (aud) must match client ID
- Should not be used for API authorization

Content:
- User profile information (name, email, picture)
- Authentication context (time, method)
- Standard OIDC claims
- Custom claims if configured

Critical Rule: Never use ID tokens to access APIs. They are for application authentication only.

### Access Tokens

Purpose: Authorize access to APIs and resources.

Format: Can be JWT or opaque string.

Characteristics:
- Contains authorization information (scopes)
- Intended for resource servers (APIs)
- Minimal user information (user ID only)
- Should be treated as opaque by applications

Content:
- Authorization data (scopes, permissions)
- User identifier (sub claim)
- Issuer and audience
- Expiration timestamp

Types:
- Opaque tokens: Require server-side validation
- JWT tokens: Self-contained, validate locally

### Refresh Tokens

Purpose: Obtain new access/ID tokens without re-authentication.

Characteristics:
- Long-lived credentials
- Enable session continuity
- Must be stored securely
- Can be revoked

Security Features:
- Rotation (new token issued, old invalidated)
- Expiration (idle and absolute)
- Maximum 200 per user per application

### IDP Access Tokens

Purpose: Access third-party identity provider APIs.

Usage:
- Returned by social/enterprise connections
- Enable calling provider APIs
- Subject to provider token policies

### Management API Tokens

Purpose: Access Auth0 Management API.

Characteristics:
- Short-lived
- Specific scopes for endpoint access
- Used for administrative operations

## Token Lifecycle

### Issuance

Tokens issued during:
- Successful authentication
- Token exchange
- Refresh token usage
- Device authorization

### Validation

ID Token Validation:
- Verify signature
- Check issuer and audience
- Validate expiration
- Confirm nonce (if used)

Access Token Validation:
- Opaque: Call introspection endpoint
- JWT: Verify signature and claims

### Expiration

Token Lifetimes:
- Access tokens: Configurable (default 24 hours)
- ID tokens: Configurable
- Refresh tokens: Configurable (idle/absolute)

### Revocation

Revocable Tokens:
- Refresh tokens: Via Management API
- Access/ID tokens: Wait for expiration

## Token Configuration

### Access Token Settings

API Settings:
- Token lifetime (seconds)
- Token dialect (opaque/JWT)
- Signing algorithm

Application Settings:
- Requested scopes
- Audience configuration
- Token endpoint auth method

### Refresh Token Settings

Configure in API settings:
- Enable refresh tokens
- Rotation policy
- Idle timeout
- Absolute timeout

### ID Token Settings

Configure claims:
- Standard OIDC claims
- Custom claims via Actions
- Namespace for custom claims

## Token Storage

### Browser Applications

Storage Options:
- Memory only (most secure, lost on refresh)
- Session storage (cleared on tab close)
- Cookies (httpOnly, secure, SameSite)

Avoid:
- localStorage (vulnerable to XSS)
- Exposing tokens to JavaScript when possible

### Native Applications

Storage Options:
- Secure enclave (iOS Keychain, Android Keystore)
- Encrypted storage
- Platform-specific secure storage

### Server Applications

Storage Options:
- Encrypted database
- HTTP-only session cookies
- Server-side session storage

## Token Transmission

Security Requirements:
- Always use HTTPS
- Never include in URLs
- Use Authorization header
- Consider token binding

Header Format:
- Bearer tokens: Authorization: Bearer {token}
- DPoP tokens: Authorization: DPoP {token}

## Custom Claims

### Adding Claims

Use Actions to add custom claims:
- User attributes
- Application metadata
- External data

### Namespacing

Requirement: Custom claims must use namespace URL.

Format: https://your-domain/claim-name

Purpose:
- Avoid collision with standard claims
- Identify claim source
- Enable claim management

## Token Security

### General Principles

- Keep tokens confidential
- Minimize token lifetime
- Use appropriate storage
- Validate before trusting

### Access Token Security

- Treat as opaque
- Never decode for authorization decisions
- Validate on resource server
- Use scopes for authorization

### Refresh Token Security

- Store securely
- Enable rotation
- Set appropriate expiration
- Revoke on logout

### ID Token Security

- Validate before using claims
- Check audience matches
- Verify signature
- Handle expiration

## Best Practices

Lifetime Management:
- Short access token lifetimes
- Appropriate refresh token expiration
- Balance security with user experience

Storage:
- Use most secure option available
- Consider application type
- Regular security review

Validation:
- Always validate tokens
- Use official libraries
- Cache validation results appropriately

Transmission:
- HTTPS only
- Proper header usage
- Avoid URL parameters
