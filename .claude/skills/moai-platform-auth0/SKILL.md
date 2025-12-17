---
name: moai-platform-auth0
description: Auth0 enterprise authentication specialist covering SSO, SAML, OIDC, organizations, and B2B multi-tenancy. Use when implementing enterprise identity federation or complex auth workflows.
version: 1.0.0
category: platform
tags: [auth0, sso, saml, oidc, enterprise, identity]
context7-libraries: [/auth0/auth0-docs]
related-skills: [moai-platform-clerk, moai-domain-backend]
updated: 2025-12-07
status: active
allowed-tools: Read, Write, Bash, Grep, Glob
---

# Auth0 Enterprise Authentication Specialist

Enterprise identity federation platform for B2B SaaS applications with SSO, SAML, OIDC, ADFS, Organizations, Actions, and Universal Login customization.

## Quick Reference (30 seconds)

Auth0 Core Capabilities:

- Enterprise SSO: SAML, OIDC, ADFS with 50+ pre-built connections
- Organizations: B2B multi-tenancy with isolated authentication contexts
- Actions: Serverless extensibility for custom auth logic
- Universal Login: Customizable branded login experience
- Management API: Comprehensive user and tenant management

When to Use Auth0:

- Enterprise SSO with SAML, OIDC, or ADFS required
- B2B SaaS with organization-level isolation
- 50+ enterprise identity provider integrations needed
- Complex authentication workflows with custom logic
- SOC2, HIPAA, or enterprise compliance requirements

Context7 Access:

Use resolve-library-id with "auth0" then get-library-docs for latest API documentation.

---

## Implementation Guide

### Enterprise SSO Configuration

SAML Identity Provider Integration:

Step 1: Navigate to Auth0 Dashboard, select Authentication, then Enterprise
Step 2: Select SAML and click Create Connection
Step 3: Provide connection name and IdP metadata URL or upload XML
Step 4: Configure attribute mappings for user profile synchronization
Step 5: Map SAML attributes to Auth0 user profile fields
Step 6: Enable connection for target applications

SAML Attribute Mapping Configuration:

Common attribute mappings include email from http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress, given_name from http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname, family_name from http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname, and groups from http://schemas.xmlsoap.org/claims/Group.

OIDC Connection Setup:

Step 1: Select OpenID Connect in enterprise connections
Step 2: Provide discovery URL from identity provider
Step 3: Configure client ID and client secret from IdP
Step 4: Define required scopes (openid, profile, email)
Step 5: Map OIDC claims to Auth0 user profile attributes
Step 6: Configure token validation settings

ADFS Integration:

Step 1: Configure ADFS as SAML identity provider in ADFS console
Step 2: Add Auth0 as relying party trust
Step 3: Export ADFS federation metadata XML
Step 4: Create SAML connection in Auth0 with ADFS metadata
Step 5: Configure claim rules in ADFS for required attributes
Step 6: Test connection with ADFS sign-in flow

### Organizations for B2B Multi-Tenancy

Organization Feature Overview:

Auth0 Organizations enable multi-tenant B2B SaaS applications with isolated authentication contexts per customer organization.

Organization Core Features:

- Isolated user pools per organization
- Organization-specific identity providers
- Role-based access control per organization
- Invitation and membership management
- Custom branding per organization
- Connection-level organization restrictions

Creating Organizations Programmatically:

Use Management API to create organizations with name, display_name, branding configuration (logo_url, colors), metadata for custom attributes, and enabled_connections for allowed identity providers.

Organization Membership Management:

Invite users via email with customizable invitation templates, assign roles during invitation, support multiple organization memberships per user, and implement domain-based auto-enrollment.

Organization RBAC Configuration:

Step 1: Enable Organizations in tenant settings
Step 2: Define organization roles (org_admin, org_member, org_viewer)
Step 3: Assign permissions to roles at organization level
Step 4: Configure organization login experience
Step 5: Implement role checks in application using organization claims

Organization-Specific Connections:

Enable different identity providers per organization, allowing Enterprise customers to use SAML SSO while standard customers use email/password.

### Actions and Rules

Actions Overview:

Auth0 Actions replace deprecated Rules and Hooks with a modern serverless extensibility system.

Action Triggers:

- post-login: Execute after successful authentication
- post-user-registration: Execute after user signs up
- pre-user-registration: Validate user before registration
- post-change-password: Execute after password change
- send-phone-message: Custom phone message providers

Post-Login Action Patterns:

Add custom claims to tokens based on user metadata, enforce organization membership requirements, implement progressive profiling, log authentication events to external systems, and block users based on custom conditions.

Post-Login Action Structure:

The exports.onExecutePostLogin function receives event and api parameters. Access user information via event.user, organization via event.organization, and modify tokens using api.idToken.setCustomClaim and api.accessToken.setCustomClaim methods.

Pre-User-Registration Actions:

Validate email domains before allowing registration, check against external systems for user approval, populate initial user metadata, and enforce custom registration requirements.

Action Secrets Management:

Store sensitive values like API keys in Action secrets, access via event.secrets object, rotate secrets without redeploying actions, and audit secret access in logs.

### Universal Login Customization

Universal Login Overview:

Auth0 Universal Login provides a centralized, secure authentication experience hosted on Auth0 infrastructure.

New Universal Login Features:

- Built-in customization without code
- Passwordless authentication support
- WebAuthn and passkeys integration
- Organization login picker
- Identifier-first authentication flow

Branding Configuration:

Configure logo, colors, and fonts in Dashboard under Branding. Set primary_color for buttons and links, page_background_color for login page, and upload logo images in recommended dimensions.

Custom Universal Login:

For advanced customization, use Auth0 Lock widget or custom HTML pages. Implement custom CSS, JavaScript logic, and integrate with design systems while maintaining security.

Page Templates:

Customize login, signup, password reset, and MFA pages. Support multiple languages with template variables. Implement A/B testing for conversion optimization.

### Management API

Management API Overview:

Auth0 Management API provides comprehensive programmatic access for user management, application configuration, and tenant administration.

Authentication for Management API:

Obtain Machine-to-Machine access tokens with appropriate scopes. Use client credentials flow with application client_id and client_secret targeting the Management API audience.

User Management Operations:

Create users with connection, email, password, and metadata. Search users with Lucene query syntax. Update user metadata (user_metadata for user-editable, app_metadata for application-controlled). Delete users and revoke sessions.

Application Management:

Create and configure applications programmatically. Manage allowed callbacks, logout URLs, and web origins. Configure JWT settings including token lifetime and signing algorithm.

Connection Management:

Create enterprise connections via API. Configure connection options and attribute mappings. Enable connections for specific applications. Manage connection-level settings.

Rate Limiting Considerations:

Management API enforces rate limits per endpoint. Implement exponential backoff for retry logic. Cache frequently accessed data. Use bulk operations where available.

---

## Advanced Patterns

### Enterprise Connection Patterns

Connection Selector for Multiple IdPs:

Implement Home Realm Discovery using email domain to route users to appropriate identity provider automatically.

Connection Configuration per Environment:

Maintain separate connections for development, staging, and production environments. Use environment-specific metadata for connection configuration.

Fallback Authentication Strategy:

Configure primary enterprise connection with database connection fallback. Allow password reset for users locked out of SSO.

### Token Customization

Custom Claims in Access Tokens:

Add organization_id, roles, and permissions as custom claims. Use namespaced claims following JWT best practices (e.g., https://myapp.com/claims/role).

Refresh Token Rotation:

Enable refresh token rotation for enhanced security. Configure absolute and inactivity expiration. Implement reuse detection for compromised tokens.

Token Lifetime Configuration:

Set appropriate lifetimes based on security requirements: access_token (15 minutes default), id_token (36000 seconds default), refresh_token (absolute and inactivity expiration).

### Migration Strategies

Lazy Migration from Legacy Database:

Step 1: Create custom database connection
Step 2: Implement Login script to validate against legacy DB
Step 3: Implement GetUser script for profile retrieval
Step 4: Auth0 creates user on successful legacy authentication
Step 5: Monitor migration progress via logs

Bulk User Import:

Export users from legacy system with password hashes. Format as Auth0 bulk import JSON with supported hash algorithms (bcrypt, argon2, pbkdf2). Submit import job via Management API. Monitor job status and handle errors.

Organization Migration:

Map legacy tenant structure to Auth0 Organizations. Migrate users with organization memberships. Configure organization-specific connections. Update application to use organization context.

### Security Best Practices

Anomaly Detection:

Enable brute-force protection with configurable thresholds. Configure breached password detection. Set up suspicious IP throttling. Monitor authentication anomalies in logs.

Adaptive MFA:

Configure risk-based MFA challenges. Require MFA for sensitive operations. Support multiple factors (TOTP, SMS, WebAuthn, push). Implement step-up authentication for high-risk actions.

Token Security:

Use httpOnly cookies for token storage when possible. Implement token binding for enhanced security. Configure audience restrictions on access tokens. Validate tokens server-side before granting access.

---

## Resources

Context7 Documentation Access:

Use resolve-library-id with "auth0" then get-library-docs for comprehensive API reference and implementation guides.

Works Well With:

- moai-platform-clerk: Alternative for WebAuthn-first authentication
- moai-platform-supabase: Supabase authentication integration
- moai-platform-firebase-auth: Firebase authentication comparison
- moai-platform-vercel: Vercel deployment with Auth0
- moai-domain-backend: API development and token validation
- moai-quality-security: OWASP compliance and security validation

Auth0 Deployment Models:

- Public Cloud: Multi-tenant SaaS deployment
- Private Cloud: Dedicated tenant with enhanced isolation
- Managed Private Cloud: Customer-controlled infrastructure

Compliance Certifications:

SOC 2 Type II, ISO 27001, ISO 27018, HIPAA BAA available, GDPR compliant, PCI DSS for applicable services.

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v1.0
Last Updated: 2025-12-07
Platform: Auth0 Enterprise Authentication
