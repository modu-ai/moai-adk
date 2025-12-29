# OAuth 2.0 State Parameters

Module: moai-platform-auth0/modules/state-parameters.md
Version: 1.0.0
Last Updated: 2025-12-24

---

## Overview

The OAuth 2.0 state parameter is a critical security mechanism that protects against Cross-Site Request Forgery (CSRF) attacks and enables post-authentication redirect handling. Proper implementation of state parameters is essential for secure authentication flows.

---

## CSRF Protection

### Primary Purpose

The primary reason for using the state parameter is to mitigate CSRF attacks by using a unique and non-guessable value associated with each authentication request.

### How CSRF Attacks Work

Attacker Scenario: An attacker tricks a user's browser into making an authentication request without the user's knowledge.

Without State Parameter: The application cannot distinguish between legitimate and forged authentication responses.

With State Parameter: The application can verify that the response corresponds to a request it initiated.

### Implementation Process

Step 1 - Generate Random String: Before redirecting to the Identity Provider, generate a cryptographically secure random string (for example, xyzABC123).

Step 2 - Store Locally: Store the generated value in cookies, sessions, or local storage depending on application type.

Step 3 - Include in Request: Add the state parameter to the authorization request URL.

Step 4 - Validate on Return: When receiving the authentication response, compare the returned state value with the stored value.

### Validation Logic

If the returned state matches the stored value: The response is legitimate, proceed with authentication.

If the returned state does not match: You may be the target of an attack because this is either a response for an unsolicited request or someone trying to forge the response. Reject the authentication attempt.

---

## Redirect Users Post-Authentication

### Context Preservation

The state parameter can preserve application context across the authentication flow.

Use Case: User attempts to access a protected resource, gets redirected to authenticate, and should return to the original resource after authentication.

### Implementation

Encode Information: Include the intended destination URL alongside the nonce in the state parameter.

After Validation: Extract the destination URL from the state and redirect the user accordingly.

Example Structure: The state value might contain both a random nonce and an encoded destination path.

---

## Storage Recommendations

Storage method depends on application type:

Regular Web Applications: Use server-side session storage or signed cookies.

Single-Page Applications: Use browser local storage with appropriate security measures.

Native Applications: Use device memory or secure local storage.

---

## Security Requirements

### State Value Characteristics

Uniqueness: Each authentication request must have a unique state value.

Opacity: State values should not be predictable or guessable.

Sufficient Entropy: Use cryptographically secure random number generators.

### Cookie-Based Storage Security

Signed Cookies: When storing state in cookies, sign the cookie to prevent tampering.

HttpOnly: Consider HttpOnly flag to prevent JavaScript access.

Secure: Use Secure flag to ensure transmission only over HTTPS.

SameSite: Configure SameSite attribute appropriately.

### URL Encoding Security

Avoid Plaintext: Do not use plaintext or predictable encoding for stored URLs.

Encryption: Consider encrypting sensitive redirect URLs.

Length Limits: Be aware that excessively long state values may trigger 414 Request-URI Too Large errors.

---

## Implementation Examples

### Authorization Request

When constructing the authorization URL:

Include the state parameter with the generated random value.

Store the corresponding value locally before redirecting.

Ensure the state is URL-encoded if it contains special characters.

### Response Handling

When receiving the authorization response:

Extract the state parameter from the response.

Retrieve the stored state value.

Compare the values for exact match.

Only proceed if values match.

---

## Common Mistakes to Avoid

Reusing State Values: Each authentication request needs a fresh state value.

Weak Random Generation: Use cryptographically secure random generators, not Math.random().

Not Validating State: Always validate the returned state, never skip this step.

Storing State Insecurely: Protect stored state values from unauthorized access.

Predictable Patterns: Avoid using timestamps or sequential numbers as state values.

---

## Error Handling

### Missing State Parameter

If the authorization response lacks a state parameter but one was sent, treat as a potential attack.

### State Mismatch

Log the mismatch for security monitoring.

Do not complete the authentication.

Display an appropriate error message to the user.

Consider implementing rate limiting if mismatches occur frequently.

---

## Related Modules

- attack-protection-overview.md: Overall attack protection strategy
- tokens-overview.md: Token security
- application-credentials.md: Application security

---

## Resources

Auth0 Documentation: State Parameter
Auth0 Documentation: Prevent Attacks with State Parameters
OAuth 2.0 RFC 6749: State Parameter Specification
