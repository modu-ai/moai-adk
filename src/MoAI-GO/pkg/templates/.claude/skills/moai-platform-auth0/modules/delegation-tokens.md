# Delegation Tokens

Module: moai-platform-auth0/modules/delegation-tokens.md
Version: 1.0.0
Last Updated: 2025-12-24

---

## Overview

Delegation tokens enable applications to call APIs of Application Add-ons (such as Firebase or SAP) registered in the same Auth0 tenant. Given an existing token, the delegation endpoint generates a new token signed with the target application's secret, flowing user identity from the application to an API.

---

## Deprecation Notice

IMPORTANT: Delegation is deprecated functionality with limited support.

Status: Delegation is disabled for tenants without an add-on in use as of June 8, 2017.

Legacy Support: Tenants currently using add-ons requiring delegation may continue, with advance notice if functionality changes.

Library Support: Delegation is supported in Auth0.js version 7 but NOT supported in version 8.

Recommendation: For new implementations, consider alternative approaches such as machine-to-machine tokens or custom API authorization patterns.

---

## Purpose

Delegation tokens serve specific legacy use cases:

Add-on Integration: Enable calling third-party services configured as Application Add-ons in Auth0.

Identity Flow: Transfer user identity from the primary application to downstream services.

Token Exchange: Generate service-specific tokens from existing Auth0 tokens.

---

## How Delegation Works

### Token Generation Process

Step 1: Application has an existing valid Auth0 token.

Step 2: Application calls the delegation endpoint with the existing token.

Step 3: Auth0 generates a new token signed with the target add-on's secret.

Step 4: Application uses the new token to call the add-on API.

### Requirements

Target Must Be Add-on: The target must be an Application Add-on (non-SAML or WS-Fed).

Configured Secrets: The Add-on must be configured with secrets from the provider.

Token Signing: These secrets sign the delegation token so the Add-on API can validate it.

---

## Token Formats

The delegation token type varies by provider:

Azure Blob Storage: SAS (Shared Access Signature) format.

Firebase Add-on: JWT format signed with Firebase credentials.

Other Providers: Format depends on the specific add-on requirements.

---

## Configurable Parameters

When requesting delegation tokens, applications can set:

target: The client ID of the target application or add-on.

scope: The scopes requested for the delegation token.

api_type: The type of API being accessed (provider-specific).

Additional Parameters: Free-form parameters depending on the provider requirements.

---

## Public Applications Caveat

When using delegation with Public Applications, specific configuration is required:

Algorithm Configuration: Configure the JsonWebToken Signature Algorithm as RS256.

Why RS256: The token endpoint forces RS256 signing even if set to HS256.

Validation Issue: Algorithm mismatch causes delegation validation to fail.

---

## Alternative Approaches

For new implementations, consider these alternatives:

### Machine-to-Machine Tokens

Use client credentials flow for service-to-service communication.

Benefits: Modern approach, well-supported, no deprecation concerns.

### Custom API Authorization

Implement custom token exchange patterns using Auth0 Actions or Rules.

Benefits: Flexible, maintainable, uses current Auth0 features.

### Direct API Integration

Configure APIs directly without using the legacy Add-on framework.

Benefits: Simpler architecture, better control over authorization.

---

## Migration Guidance

If currently using delegation tokens, consider migration planning:

Assessment: Identify all current delegation token usage in your applications.

Alternatives: Evaluate modern alternatives for each use case.

Timeline: Plan migration before potential feature deprecation.

Testing: Thoroughly test alternative implementations before switching.

---

## Related Modules

- tokens-overview.md: Token types overview
- access-tokens.md: Access token usage
- application-credentials.md: Application authentication methods

---

## Resources

Auth0 Documentation: Delegation Tokens
Auth0 Documentation: Application Add-ons
