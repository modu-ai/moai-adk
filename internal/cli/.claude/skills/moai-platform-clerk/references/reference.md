# Clerk Modern Authentication - Reference Documentation

## Official Resources

### Core Documentation
- **Clerk Documentation**: https://clerk.com/docs
  - Complete documentation hub
  - Getting started guides
  - API references

- **Quickstart Guides**: https://clerk.com/docs/quickstarts
  - Next.js quickstart
  - React quickstart
  - Express quickstart
  - Remix quickstart

- **SDK Reference**: https://clerk.com/docs/reference
  - JavaScript SDK
  - React hooks
  - Server SDKs
  - Backend APIs

### Framework-Specific Documentation

#### Next.js Integration
- **Next.js Documentation**: https://clerk.com/docs/quickstarts/nextjs
  - App Router setup
  - Middleware configuration
  - Server components
  - Route handlers

- **Core 2 Migration Guide**: https://clerk.com/docs/guides/development/upgrading/upgrade-guides/core-2/nextjs
  - Breaking changes
  - Migration tool: `npx @clerk/upgrade --from=core-1`
  - Environment variable changes
  - API updates

#### React Integration
- **Clerk React SDK**: https://clerk.com/docs/reference/clerk-react
  - useAuth hook
  - useUser hook
  - Components reference
  - TypeScript types

- **React Reference**: https://clerk.com/docs/reference/js
  - Clerk object
  - Authentication methods
  - Session management

### Authentication Features

#### WebAuthn & Passkeys
- **WebAuthn Guide**: https://clerk.com/docs/reference/javascript/webauthn
  - Passkey configuration
  - Biometric authentication
  - Security key setup

- **Passwordless Authentication**: https://clerk.com/docs/reference/javascript/passwordless
  - Magic links
  - Email codes
  - Phone verification

#### Multi-Factor Authentication
- **MFA Guide**: https://clerk.com/docs/authentication/mfa
  - TOTP setup
  - Backup codes
  - MFA policies

#### Organization Management
- **Organizations**: https://clerk.com/docs/organizations
  - Organization creation
  - Member management
  - Role-based access
  - Organization switching

### API Reference

#### JavaScript SDK
- **Clerk JavaScript**: https://clerk.com/docs/reference/javascript
  - Client initialization
  - Authentication methods
  - Session handling
  - User management

#### React Hooks
- **useAuth**: https://clerk.com/docs/reference/clerk-react/use-auth
  - Authentication state
  - Token access
  - Session management

- **useUser**: https://clerk.com/docs/reference/clerk-react/use-user
  - User profile data
  - User metadata
  - Profile updates

- **useOrganization**: https://clerk.com/docs/reference/clerk-react/use-organization
  - Organization context
  - Membership data
  - Organization switching

#### Server SDKs
- **Next.js Server**: https://clerk.com/docs/reference/nextjs
  - auth() function
  - currentUser()
  - Middleware helpers

- **Express Server**: https://clerk.com/docs/reference/express
  - ClerkExpressRequireAuth
  - requireMiddleware
  - Session handling

#### Backend API
- **Backend API**: https://clerk.com/docs/reference/backend-api
  - User management
  - Organization management
  - Token verification
  - Webhook handling

### Components Reference

#### Authentication Components
- **SignIn**: https://clerk.com/docs/components/sign-in
  - Sign-in page configuration
  - Flow customization
  - Redirect handling

- **SignUp**: https://clerk.com/docs/components/sign-up
  - Sign-up page configuration
  - Field customization
  - Verification flows

- **UserButton**: https://clerk.com/docs/components/user-button
  - User menu component
  - Account management
  - Sign-out functionality

- **OrganizationSwitcher**: https://clerk.com/docs/components/organization-switcher
  - Organization selection
  - Organization creation
  - Custom styling

### Configuration & Settings

#### Environment Variables
- **Configuration Reference**: https://clerk.com/docs/guides/configuration
  - NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY
  - CLERK_SECRET_KEY
  - NEXT_PUBLIC_CLERK_SIGN_IN_URL
  - NEXT_PUBLIC_CLERK_SIGN_UP_URL

#### Middleware
- **Middleware Guide**: https://clerk.com/docs/nextjs/middleware
  - Route protection
  - createRouteMatcher
  - Clerk middleware setup

### Webhooks & Events

- **Webhooks**: https://clerk.com/docs/integrations/webhooks
  - Event types
  - Webhook setup
  - Signature verification
  - Svix integration

### Security & Compliance

- **Security Best Practices**: https://clerk.com/docs/guides/security
  - Token security
  - Session management
  - CORS configuration
  - Rate limiting

- **Compliance**: https://clerk.com/docs/compliance
  - GDPR compliance
  - SOC 2 certification
  - Data residency
  - Privacy policy

### Context7 Integration

- **Library ID**: /clerk/clerk-docs
- **Resolution**: Use `mcp__context7__resolve-library-id` with query "clerk"
- **Documentation**: Use `mcp__context7__get-library-docs` for latest docs

### Module Organization

This skill contains 1 module:

- `modules/advanced-patterns.md` - Advanced implementation patterns including:
  - Core 2 migration details
  - Role-based access control (RBAC)
  - Webhook integration
  - Custom authentication flows
  - External service integration
  - Performance optimization
  - Testing strategies

### Community & Support

- **Clerk Discord**: https://discord.gg/v5qHfQPW2g
  - Community chat
  - Real-time support
  - Feature discussions

- **GitHub Discussions**: https://github.com/clerk/javascript/discussions
  - Q&A threads
  - Feature requests
  - Community solutions

- **Support Portal**: https://clerk.com/support
  - Email support
  - Priority support (Pro/Enterprise)
  - Status page

### SDK Packages

#### Current Versions (December 2025)
- **@clerk/nextjs**: 6.x (Core 2)
  - Requires Next.js 13.0.4+
  - Requires React 18+
  - Node.js 18.17.0+

- **@clerk/clerk-react**: 5.x (Core 2)
  - Requires React 18+
  - TypeScript support

- **@clerk/express**: 1.x
  - Express.js integration
  - Middleware support

- **@clerk/backend**: 1.x
  - Backend API utilities
  - JWT verification
  - Webhook verification

### Examples & Templates

- **Next.js Examples**: https://github.com/clerk/nextjs-examples
  - App Router examples
  - Pages Router examples
  - Middleware examples
  - Organization examples

- **React Examples**: https://github.com/clerk/react-examples
  - CRA templates
  - Vite examples
  - Remix examples

### Migration Tools

- **Core 2 Upgrade Tool**: https://clerk.com/docs/guides/development/upgrading/upgrade-guides/core-2
  ```bash
  npx @clerk/upgrade --from=core-1
  ```
  - Automated code migration
  - Environment variable updates
  - Breaking change detection

### Related Skills

- **moai-platform-auth0**: Enterprise SSO alternative
- **moai-platform-firebase-auth**: Firebase authentication
- **moai-lang-typescript**: TypeScript development patterns
- **moai-domain-frontend**: React and Next.js integration
- **moai-platform-supabase**: Supabase authentication
- **moai-platform-vercel**: Vercel deployment

---

**Last Updated**: 2025-12-30
**Skill Version**: 2.0.0
**Total Modules**: 1
**SDK Version**: @clerk/nextjs 6.x (Core 2)
