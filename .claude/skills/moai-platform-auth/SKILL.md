---
name: moai-platform-auth
description: Authentication platform specialist for Auth0, Clerk, and Firebase Auth. Use when implementing SSO, WebAuthn, organizations, multi-tenant auth, or security best practices.
version: 1.0.0
category: platform
updated: 2025-12-07
status: active
allowed-tools: Read, Write, Bash, Grep, Glob
---

# Authentication Platform Specialist

Enterprise authentication integration hub covering Auth0, Clerk, and Firebase Auth with SSO, WebAuthn, organizations, and security best practices.

## Quick Reference (30 seconds)

Authentication Provider Selection Guide:

- Auth0: Enterprise SSO with 50+ connections, B2B SaaS, SAML/OIDC/ADFS
- Clerk: Modern WebAuthn, passkeys, passwordless, beautiful UI components
- Firebase Auth: Google ecosystem, mobile-first, social auth integration

Context7 Library Mappings:

- Auth0: /auth0/auth0-docs
- Clerk: /clerk/clerk-docs
- Firebase: /firebase/firebase-docs

Quick Decision Tree:

- Need Enterprise SSO with 50+ connections? Use Auth0
- Need modern WebAuthn and passkeys? Use Clerk
- Need Google ecosystem integration? Use Firebase Auth
- Need B2B SaaS with organizations? Use Auth0 or Clerk
- Need mobile-first authentication? Use Firebase Auth

---

## Implementation Guide

### Auth0: Enterprise SSO Specialist

Enterprise SSO Configuration (SAML, OIDC, ADFS):

Auth0 excels at enterprise identity federation with support for 50+ pre-built enterprise connections including Okta, Azure AD, Google Workspace, and custom SAML providers.

SAML Connection Setup:

Step 1: Navigate to Auth0 Dashboard and select Authentication then Enterprise
Step 2: Select SAML and create new connection with IdP metadata URL
Step 3: Configure attribute mappings for user profile synchronization
Step 4: Enable connection for target applications

OIDC Connection Setup:

Step 1: Select OpenID Connect in enterprise connections
Step 2: Provide discovery URL from identity provider
Step 3: Configure client credentials and scopes
Step 4: Map claims to Auth0 user profile attributes

ADFS Integration:

Step 1: Configure ADFS as SAML identity provider
Step 2: Export ADFS federation metadata
Step 3: Import metadata into Auth0 SAML connection
Step 4: Configure relying party trust in ADFS

B2B SaaS Organizations with RBAC:

Auth0 Organizations enable multi-tenant B2B SaaS applications with isolated authentication contexts per organization.

Organization Features:

- Isolated user pools per organization
- Organization-specific identity providers
- Role-based access control per organization
- Invitation and membership management
- Custom branding per organization

Organization RBAC Configuration:

Step 1: Enable Organizations feature in Auth0 tenant settings
Step 2: Define roles at organization level (admin, member, viewer)
Step 3: Assign permissions to roles
Step 4: Configure organization login experience
Step 5: Implement role checks in application

Custom Database Connections:

Auth0 supports custom database connections for migrating users from legacy systems without requiring immediate password resets.

Migration Strategy:

- Lazy migration imports users on first login
- Bulk import migrates all users with password hashes
- Custom scripts validate credentials against legacy database

Security Features:

- Anomaly detection for suspicious login attempts
- Brute-force protection with configurable thresholds
- Breached password detection
- Multi-factor authentication with multiple options
- Adaptive MFA based on risk assessment

### Clerk: Modern Authentication Specialist

WebAuthn and Passkey Implementation:

Clerk provides first-class WebAuthn support enabling passwordless authentication with biometrics and hardware security keys.

WebAuthn Configuration:

Step 1: Enable WebAuthn in Clerk Dashboard under User and Authentication
Step 2: Configure passkey requirements (required, optional, or disabled)
Step 3: Set verification requirements for passkey registration
Step 4: Implement passkey UI using Clerk components or custom flow

Passkey User Experience:

- Registration flow prompts for biometric or security key
- Login flow automatically detects available passkeys
- Fallback to password if passkeys unavailable
- Cross-device passkey support with FIDO Alliance standards

Passwordless Authentication Options:

- Email magic links with customizable templates
- SMS one-time passwords
- Email one-time passwords
- Passkeys with WebAuthn

Organization Management:

Clerk Organizations provide team and workspace management with invitations, roles, and permissions.

Organization Features:

- Create and manage organizations programmatically
- Invite users via email with customizable invitations
- Role-based permissions (admin, member, custom roles)
- Organization switching for users with multiple memberships
- Domain verification for automatic organization membership

Multi-Platform SDK Support:

Clerk provides SDKs for multiple platforms ensuring consistent authentication across web and mobile.

Supported Platforms:

- React: @clerk/clerk-react
- Next.js: @clerk/nextjs with middleware support
- Vue: @clerk/vue (community maintained)
- React Native: @clerk/clerk-expo
- Node.js: @clerk/clerk-sdk-node

Pre-built UI Components:

Clerk provides beautiful, customizable authentication components reducing development time.

Available Components:

- SignIn: Complete sign-in form with social and email options
- SignUp: Registration form with verification
- UserButton: User avatar dropdown with profile management
- OrganizationSwitcher: Organization selection dropdown
- UserProfile: Full user profile management
- CreateOrganization: Organization creation flow

Backend API Features:

- User management API for CRUD operations
- Session management with token verification
- Webhooks for user lifecycle events
- JWT customization for claims and templates

### Firebase Auth: Google Ecosystem Integration

Firebase Ecosystem Integration:

Firebase Auth integrates seamlessly with Firebase services including Firestore, Cloud Functions, Cloud Storage, and Analytics.

Firebase Services Integration:

- Firestore: Security rules using auth.uid
- Cloud Functions: Authentication triggers
- Cloud Storage: User-scoped file access
- Analytics: User identification and events

Social Authentication Providers:

Firebase Auth supports major social providers with simple configuration.

Supported Providers:

- Google Sign-In with Google Cloud integration
- Facebook Login with customizable permissions
- Apple Sign-In for iOS applications
- Twitter/X authentication
- GitHub authentication
- Microsoft authentication
- Yahoo authentication

Cloud Functions Auth Triggers:

Firebase Cloud Functions can respond to authentication events for custom logic.

Available Triggers:

- onCreate: Triggered when new user created
- onDelete: Triggered when user deleted
- beforeCreate: Blocking function before user creation
- beforeSignIn: Blocking function before sign-in

Use Cases:

- Send welcome email on user creation
- Initialize user profile in Firestore
- Validate custom claims before sign-in
- Block suspicious sign-in attempts
- Sync user data to external systems

Mobile SDK Implementation:

Firebase Auth provides native SDKs for mobile platforms with offline support.

Supported Mobile Platforms:

- iOS: Firebase SDK with Swift and Objective-C
- Android: Firebase SDK with Kotlin and Java
- Flutter: firebase_auth package
- React Native: @react-native-firebase/auth

Mobile Features:

- Offline authentication state persistence
- Automatic token refresh
- Anonymous authentication for onboarding
- Phone number authentication with SMS
- Multi-factor authentication support

---

## Advanced Patterns

### Provider Selection Decision Framework

Enterprise Authentication Requirements:

If SSO with SAML, OIDC, or ADFS required: Select Auth0
If 50+ enterprise connections needed: Select Auth0
If B2B SaaS with organization isolation: Select Auth0 or Clerk
If compliance requirements (SOC2, HIPAA): Select Auth0 Enterprise

Modern Authentication Requirements:

If WebAuthn and passkeys primary method: Select Clerk
If passwordless authentication priority: Select Clerk
If beautiful pre-built UI needed: Select Clerk
If multi-platform consistency required: Select Clerk

Mobile and Google Ecosystem:

If Google services integration needed: Select Firebase Auth
If mobile-first application: Select Firebase Auth
If existing Firebase infrastructure: Select Firebase Auth
If serverless Cloud Functions: Select Firebase Auth

### Cross-Provider Migration Strategies

Auth0 to Clerk Migration:

Phase 1: Export Auth0 users via Management API
Phase 2: Transform user data to Clerk format
Phase 3: Import users via Clerk Backend API
Phase 4: Migrate organization memberships
Phase 5: Update application authentication flows
Phase 6: Parallel run with gradual traffic shift

Clerk to Auth0 Migration:

Phase 1: Export Clerk users via Backend API
Phase 2: Transform to Auth0 user format
Phase 3: Import via Auth0 Management API
Phase 4: Configure enterprise connections
Phase 5: Migrate organization structures
Phase 6: Update application integration

Firebase to Auth0 Migration:

Phase 1: Export Firebase users with password hashes
Phase 2: Create Auth0 custom database connection
Phase 3: Configure lazy migration scripts
Phase 4: Import users with password hash verification
Phase 5: Update security rules and triggers

### Security Best Practices

Token Security:

- Use short-lived access tokens (15 minutes default)
- Implement refresh token rotation
- Store tokens securely (httpOnly cookies preferred)
- Validate token signatures on backend

Session Management:

- Implement absolute session timeout
- Enable sliding session expiration
- Track active sessions per user
- Provide session revocation capability

Multi-Factor Authentication:

- Require MFA for sensitive operations
- Support multiple MFA methods
- Implement backup codes for recovery
- Consider adaptive MFA based on risk

Rate Limiting:

- Implement rate limits on authentication endpoints
- Configure progressive delays on failed attempts
- Monitor and alert on suspicious patterns
- Implement account lockout policies

### Integration Patterns

Authentication with Database Providers:

Auth0 with Supabase RLS:

- Configure Auth0 JWT in Supabase settings
- Map Auth0 claims to RLS policies
- Use organization ID for multi-tenant isolation

Clerk with Convex:

- Use Clerk JWT verification in Convex
- Sync user data via webhooks
- Implement organization-based access control

Firebase Auth with Firestore:

- Security rules reference request.auth
- User-specific document access
- Admin claims for elevated permissions

Authentication with Deployment Platforms:

Vercel Edge Functions:

- Verify tokens in Edge Middleware
- Use Clerk Next.js middleware for automatic protection
- Implement Auth0 session management

Railway Deployments:

- Configure environment variables for auth secrets
- Implement health checks with auth bypass
- Use service-to-service authentication

---

## Resources

Detailed API documentation and migration scripts: reference.md

Production-ready code examples and integration patterns: examples.md

Context7 Documentation Access:

Auth0 Documentation: Use resolve-library-id with "auth0" then get-library-docs
Clerk Documentation: Use resolve-library-id with "clerk" then get-library-docs
Firebase Documentation: Use resolve-library-id with "firebase" then get-library-docs

Works Well With:

- moai-platform-baas: Unified BaaS integration including database and deployment
- moai-domain-backend: Backend architecture patterns
- moai-quality-security: OWASP security validation
- moai-domain-frontend: Frontend integration patterns

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v1.0
Last Updated: 2025-12-07
Providers Covered: Auth0, Clerk, Firebase Auth
