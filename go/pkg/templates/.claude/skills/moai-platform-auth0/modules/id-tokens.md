# ID Tokens

Module: moai-platform-auth0/modules/id-tokens.md
Version: 1.0.0
Last Updated: 2025-12-24

---

## Overview

ID tokens are JSON Web Tokens (JWTs) used in token-based authentication to cache user profile information and deliver it to client applications. After successful user authentication, applications receive an ID token, extract user information from it, and use this data to personalize the user experience.

---

## Purpose and Function

ID tokens serve specific purposes in authentication flows:

User Information Delivery: ID tokens contain claims about the authenticated user including name, email, and profile information.

Application Personalization: Applications use ID token claims to generate personalized experiences such as welcome messages, user dashboards, and customized content.

Authentication Proof: ID tokens prove that authentication occurred and provide information about when and how the user authenticated.

---

## Critical Security Distinction

ID tokens should never be used to obtain direct access to APIs or to make authorization decisions. Use access tokens for API authorization instead.

Why This Matters:

- ID tokens are designed for the client application only
- They contain user identity information, not authorization scopes
- APIs should validate access tokens, not ID tokens
- Using ID tokens for API access creates security vulnerabilities

---

## Token Structure

ID tokens follow the standard JWT structure:

Header: Contains the token type (JWT) and signing algorithm (typically RS256).

Payload: Contains claims about the user and authentication event.

Signature: Cryptographic signature for validation.

### Standard Claims

iss (Issuer): The Auth0 tenant URL that issued the token.

sub (Subject): The unique identifier for the user.

aud (Audience): The application client ID the token was issued for.

exp (Expiration): Unix timestamp when the token expires.

iat (Issued At): Unix timestamp when the token was issued.

### User Profile Claims

name: The user's full name.

email: The user's email adddess.

email_verified: Boolean indicating email verification status.

picture: URL to the user's profile picture.

nickname: The user's nickname or username.

---

## Token Lifetime

Default Lifetime: ID tokens remain valid for 36,000 seconds (10 hours) by default.

Configuration: Token lifetime can be shortened for security-sensitive applications. Navigate to Dashboard then Applications then select your application then Settings then Advanced Settings then OAuth to configure.

Considerations:

- Shorter lifetimes increase security but require more frequent authentication
- Longer lifetimes improve user experience but increase exposure window
- Balance security requirements with user experience needs

---

## Validation Requirements

Before using information contained in an ID token, validation is required.

### Validation Steps

Step 1: Verify the token signature using the public key from Auth0's JWKS endpoint.

Step 2: Validate the issuer (iss) claim matches your Auth0 tenant URL.

Step 3: Validate the audience (aud) claim matches your application's client ID.

Step 4: Check the expiration (exp) claim to ensure the token is not expired.

Step 5: Verify the issued at (iat) claim is in the past.

### Validation Libraries

Use established JWT libraries for validation. Available libraries at jwt.io include implementations for:

- JavaScript and Node.js
- Python
- Java
- .NET
- Ruby
- Go
- PHP

---

## Security Best Practices

### XSS Protection

Applications making API calls must ensure tokens and sensitive data are not vulnerable to cross-site scripting (XSS) attacks and cannot be read by malicious JavaScript.

Protection Measures:

- Store tokens securely (httpOnly cookies when possible)
- Sanitize all user input
- Implement Content Security Policy headers
- Use secure token handling libraries

### Token Storage

Server-Side Applications: Store tokens in server-side sessions.

Single-Page Applications: Store tokens in memory or secure storage with XSS protections.

Mobile Applications: Use secure device storage mechanisms.

---

## Getting ID Tokens

ID tokens are returned from authentication flows that include the openid scope.

### Authorization Code Flow

Request the openid scope during authorization.

Exchange the authorization code for tokens.

Receive ID token along with access token.

### Implicit Flow (Legacy)

Request response_type=id_token during authorization.

Receive ID token directly in the redirect.

Note: Implicit flow is deprecated for most use cases.

---

## Common Use Cases

### User Profile Display

Extract name, email, and picture claims to display user information in the application interface.

### Personalization

Use user claims to customize content, greetings, and application behavior.

### Audit Logging

Log authentication events using token claims for audit trails.

### Session Management

Use token claims to manage user sessions and permissions within the application.

---

## Related Modules

- access-tokens.md: API authorization tokens
- jwt-fundamentals.md: JWT structure and validation
- token-best-practices.md: Token security guidelines
- tokens-overview.md: Token types overview

---

## Resources

Auth0 Documentation: ID Tokens
Auth0 Documentation: Validate ID Tokens
Auth0 Documentation: Get ID Tokens
