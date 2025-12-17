---
name: moai-platform-clerk
description: Clerk modern authentication specialist covering WebAuthn, passkeys, passwordless, and beautiful UI components. Use when implementing modern auth with great UX.
version: 1.0.0
category: platform
tags: [clerk, webauthn, passkeys, passwordless, authentication]
context7-libraries: [/clerk/clerk-docs]
related-skills: [moai-platform-auth0, moai-lang-typescript]
updated: 2025-12-07
status: active
allowed-tools: Read, Write, Bash, Grep, Glob
---

# Clerk Modern Authentication Specialist

Modern authentication platform with WebAuthn, passkeys, passwordless flows, beautiful pre-built UI components, and multi-tenant organization support.

## Quick Reference (30 seconds)

Clerk Core Capabilities:

- WebAuthn and Passkeys: First-class biometric and hardware key support
- Passwordless: Email magic links, SMS OTP, email OTP
- Pre-built UI: SignIn, SignUp, UserButton, OrganizationSwitcher components
- Organizations: Multi-tenant team management with RBAC
- Multi-Platform: React, Next.js, Vue, React Native, Node.js SDKs

Context7 Access:

- Library: /clerk/clerk-docs
- Resolution: Use resolve-library-id with "clerk" then get-library-docs

Quick Decision Criteria:

- Need WebAuthn and passkeys? Clerk is ideal
- Need beautiful pre-built auth UI? Clerk provides ready components
- Need passwordless authentication? Clerk supports all methods
- Need multi-tenant organizations? Clerk Organizations feature
- Need React/Next.js integration? Clerk has first-class support

---

## Implementation Guide

### WebAuthn and Passkey Implementation

WebAuthn Configuration:

Clerk provides first-class WebAuthn support enabling passwordless authentication with biometrics and hardware security keys.

Step 1: Enable WebAuthn in Clerk Dashboard under User and Authentication
Step 2: Configure passkey requirements (required, optional, or disabled)
Step 3: Set verification requirements for passkey registration
Step 4: Implement passkey UI using Clerk components or custom flow

Passkey User Experience:

- Registration flow prompts for biometric or security key
- Login flow automatically detects available passkeys
- Fallback to password if passkeys unavailable
- Cross-device passkey support with FIDO Alliance standards

Passkey Registration Flow:

User clicks "Add Passkey" button in account settings
Clerk prompts for device biometric or security key
Browser WebAuthn API handles credential creation
Passkey stored securely in Clerk backend
User can manage multiple passkeys per account

Passkey Login Flow:

User navigates to sign-in page
Clerk detects available passkeys for user
User authenticates with biometric or security key
Session created automatically upon successful verification

### Passwordless Authentication

Email Magic Links:

Clerk sends secure magic links for passwordless sign-in with customizable email templates.

Configuration Steps:

Step 1: Enable Email magic link in Clerk Dashboard
Step 2: Customize email template with branding
Step 3: Configure link expiration time
Step 4: Set redirect URL after successful authentication

Magic Link Features:

- Customizable email templates with branding
- Configurable expiration times
- Secure one-time use tokens
- Automatic session creation on click

SMS One-Time Passwords:

Step 1: Enable SMS authentication in Dashboard
Step 2: Configure phone number verification requirements
Step 3: Set OTP expiration and retry limits
Step 4: Customize SMS message template

Email One-Time Passwords:

Step 1: Enable Email OTP in authentication settings
Step 2: Configure code length (6 or 8 digits)
Step 3: Set code expiration time
Step 4: Customize email template

### Pre-built UI Components

Available Components:

SignIn Component: Complete sign-in form with social and email options
SignUp Component: Registration form with verification
UserButton Component: User avatar dropdown with profile management
OrganizationSwitcher Component: Organization selection dropdown
UserProfile Component: Full user profile management
CreateOrganization Component: Organization creation flow

React Integration:

Install @clerk/clerk-react package
Wrap application with ClerkProvider
Use components directly in JSX
Customize appearance via theme prop

Next.js Integration:

Install @clerk/nextjs package
Add Clerk middleware for route protection
Use components in pages and layouts
Configure environment variables for API keys

Component Customization:

- Theme customization via appearance prop
- Custom CSS with provided class names
- Override individual elements
- Dark mode support built-in

### Organization Management (Multi-Tenancy)

Organization Features:

- Create and manage organizations programmatically
- Invite users via email with customizable invitations
- Role-based permissions (admin, member, custom roles)
- Organization switching for users with multiple memberships
- Domain verification for automatic organization membership

Creating Organizations:

Step 1: Enable Organizations feature in Dashboard
Step 2: Configure default roles and permissions
Step 3: Set invitation email templates
Step 4: Implement CreateOrganization component

Invitation System:

Step 1: Admin initiates invitation via dashboard or API
Step 2: Invitee receives customizable email invitation
Step 3: Invitee clicks link and completes signup or signin
Step 4: Automatic organization membership upon completion

Role-Based Access Control:

Default Roles:
- org:admin: Full organization management
- org:member: Standard member access

Custom Roles:
- Define custom roles in Dashboard
- Assign permissions to roles
- Check permissions in application code

Domain Verification:

Organizations can claim domains for automatic membership
Users with verified email from claimed domain auto-join
Reduces friction for enterprise onboarding

### Session Management

Session Features:

- Automatic token refresh
- Multi-device session tracking
- Session revocation capability
- Configurable session duration

Session Configuration:

Step 1: Configure session lifetime in Dashboard
Step 2: Set multi-session or single-session mode
Step 3: Configure token refresh behavior
Step 4: Enable session activity tracking

Token Management:

- Access tokens for API authentication
- Session tokens for frontend state
- Automatic refresh before expiration
- Secure httpOnly cookie storage option

### Multi-Platform SDK Support

Supported Platforms:

React: @clerk/clerk-react
Next.js: @clerk/nextjs with middleware support
Vue: @clerk/vue (community maintained)
React Native: @clerk/clerk-expo
Node.js: @clerk/clerk-sdk-node
Express: @clerk/express
Fastify: @clerk/fastify

Next.js Middleware:

Clerk middleware protects routes at Edge
Configure public and protected route patterns
Automatic redirect to sign-in for unauthenticated users
Access user session in middleware for custom logic

Backend Verification:

Node.js SDK verifies session tokens
Extract user ID and organization from token
Implement authorization logic in API routes
Webhook signature verification for events

---

## Advanced Patterns

### Custom Authentication Flows

Building Custom Sign-In:

Use useSignIn hook for programmatic control
Implement multi-step verification flows
Handle errors with custom UI
Support social OAuth alongside email

Building Custom Sign-Up:

Use useSignUp hook for registration logic
Implement progressive profiling
Custom verification code entry UI
Handle optional vs required fields

Headless Mode:

Full control over UI while using Clerk backend
Access all functionality via hooks
Implement completely custom designs
Maintain security without pre-built components

### Webhook Integration

Available Webhook Events:

user.created: New user registration completed
user.updated: User profile changes
user.deleted: User account deleted
session.created: New session started
session.ended: Session terminated
organization.created: New organization created
organization.membership.created: User joined organization

Webhook Configuration:

Step 1: Add webhook endpoint URL in Dashboard
Step 2: Select events to subscribe
Step 3: Copy signing secret for verification
Step 4: Implement signature verification in endpoint

Webhook Security:

Verify webhook signatures using svix library
Check timestamp to prevent replay attacks
Return 200 status for successful processing
Implement idempotency for duplicate handling

### JWT Customization

Custom Claims:

Add custom claims to session tokens
Include organization metadata
Add user roles and permissions
Configure claim templates in Dashboard

JWT Templates:

Create multiple JWT templates for different services
Configure issuer and audience
Set expiration times
Add conditional claims based on user attributes

### Integration Patterns

Clerk with Database Providers:

Clerk with Convex:
- Use Clerk JWT verification in Convex functions
- Sync user data via webhooks
- Implement organization-based access control

Clerk with Supabase:
- Configure Clerk JWT in Supabase settings
- Map Clerk claims to RLS policies
- Use organization ID for multi-tenant isolation

Clerk with Prisma:
- Sync user ID from Clerk to database
- Store additional user data with Clerk user ID as foreign key
- Handle user lifecycle via webhooks

Deployment Platforms:

Vercel: Native integration with Edge Middleware
Railway: Environment variable configuration
Netlify: Serverless function integration
AWS Lambda: SDK support for serverless

### Security Best Practices

Token Security:

- Use short-lived access tokens
- Enable automatic token refresh
- Store tokens securely in httpOnly cookies
- Validate tokens on backend for all requests

Rate Limiting:

- Clerk implements built-in rate limiting
- Configure custom limits per organization
- Monitor authentication attempts
- Alert on suspicious patterns

Multi-Factor Authentication:

Enable MFA in Dashboard settings
Support authenticator apps (TOTP)
Backup codes for account recovery
SMS as secondary verification option

Account Protection:

- Enable device verification for new logins
- Configure suspicious activity detection
- Implement session activity monitoring
- Provide users with security notifications

---

## Resources

Context7 Documentation Access:

Library Resolution: Use resolve-library-id with "clerk"
Documentation Fetch: Use get-library-docs with resolved ID

API Documentation:

Backend API: https://clerk.com/docs/reference/backend-api
Frontend SDK: https://clerk.com/docs/references/react/overview
Next.js SDK: https://clerk.com/docs/references/nextjs/overview
Webhooks: https://clerk.com/docs/integrations/webhooks

Works Well With:

- moai-platform-auth0: Alternative enterprise SSO solution
- moai-platform-supabase: Supabase authentication integration
- moai-platform-vercel: Vercel deployment with Clerk
- moai-platform-firebase-auth: Firebase authentication comparison
- moai-lang-typescript: TypeScript development patterns
- moai-domain-frontend: React and Next.js integration
- moai-quality-security: Security validation and OWASP compliance

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v1.0
Last Updated: 2025-12-07
Provider Coverage: Clerk Authentication Platform
