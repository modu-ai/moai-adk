# Clerk Platform Reference

Modern authentication platform with beautiful pre-built UI components, WebAuthn support, and seamless Next.js integration.

## SDK Versions

Current versions as of February 2026:

- @clerk/nextjs: Version 6.x (Core 2)
- @clerk/clerk-react: Version 5.x (Core 2)
- @clerk/express: Version 1.x
- Minimum Requirements: Next.js 13.0.4+, React 18+, Node.js 18.17.0+

**Core 2 Release:**
Major version with breaking changes released in 2024. See Core 2 Migration section for upgrade guidance.

## Environment Variables

Required environment variables for Next.js applications:

**.env.local:**
```
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...
NEXT_PUBLIC_CLERK_SIGN_IN_URL=/sign-in
NEXT_PUBLIC_CLERK_SIGN_UP_URL=/sign-up
NEXT_PUBLIC_CLERK_AFTER_SIGN_IN_URL=/dashboard
NEXT_PUBLIC_CLERK_AFTER_SIGN_UP_URL=/onboarding
```

**Variable Descriptions:**

NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY:
- Public key starting with pk_test (development) or pk_live (production)
- Safe to expose in client-side code
- Used for client-side Clerk initialization

CLERK_SECRET_KEY:
- Secret key starting with sk_test (development) or sk_live (production)
- NEVER expose in client-side code
- Used for server-side API calls

NEXT_PUBLIC_CLERK_SIGN_IN_URL:
- Route path for sign-in page (default: /sign-in)
- Must match your sign-in route
- Used for redirect after sign-out

NEXT_PUBLIC_CLERK_SIGN_UP_URL:
- Route path for sign-up page (default: /sign-up)
- Must match your sign-up route
- Used for account creation flows

NEXT_PUBLIC_CLERK_AFTER_SIGN_IN_URL:
- Redirect destination after successful sign-in
- Can be dynamic based on user role

NEXT_PUBLIC_CLERK_AFTER_SIGN_UP_URL:
- Redirect destination after successful sign-up
- Useful for onboarding flows

## ClerkProvider Setup

Root layout configuration for Next.js App Router.

**app/layout.tsx:**

Import ClerkProvider from @clerk/nextjs. Create RootLayout component accepting children prop. Wrap html tag with ClerkProvider. Nest body tag inside with children rendered.

**Key Points:**
- ClerkProvider must wrap entire app
- No configuration props needed (uses environment variables)
- Provides authentication context to all components
- Supports dynamic theme customization
- Handles session management automatically

**With Custom Theme:**

Import dark theme from @clerk/themes. Pass appearance prop to ClerkProvider with baseTheme set to dark. Customize variables for colors and typography.

## Authentication Components

Pre-built UI components for common authentication flows.

### User Authentication State

**SignedIn Component:**
Wrapper component that renders children only when user is authenticated. Use for protected content in layouts or pages.

**SignedOut Component:**
Wrapper component that renders children only when user is not authenticated. Use for login prompts or public content.

**SignInButton Component:**
Pre-styled button that opens sign-in modal. Accepts mode prop for modal or redirect. Supports custom styling via className.

**SignUpButton Component:**
Pre-styled button that opens sign-up modal. Accepts mode prop for modal or redirect. Supports custom styling via className.

**UserButton Component:**
Dropdown menu with user avatar showing profile, organization switcher, and sign-out. Automatically displays user name and image. Supports appearance customization.

**Example Layout with Auth Components:**

Import all authentication components. Create header with conditional rendering. Show SignInButton and SignUpButton when SignedOut. Show UserButton when SignedIn. Position with flexbox or Tailwind utilities.

### Dedicated Authentication Pages

**Sign-In Page:**

Create app/sign-in/[[...sign-in]]/page.tsx with catch-all route. Import SignIn component from @clerk/nextjs. Render SignIn with centered layout using flexbox. Minimum height full screen with items-center and justify-center.

**Sign-Up Page:**

Create app/sign-up/[[...sign-up]]/page.tsx with catch-all route. Import SignUp component from @clerk/nextjs. Render SignUp with centered layout matching sign-in page. Consistent styling for better UX.

**Why Catch-All Routes:**
Clerk components handle multiple sub-routes for verification, password reset, and MFA flows. Catch-all route [[...sign-in]] captures all sub-paths under /sign-in/*.

## Route Protection

Middleware-based authentication for protecting routes.

**Basic Middleware (middleware.ts):**

Import clerkMiddleware from @clerk/nextjs/server. Export default clerkMiddleware function call. Export config object with matcher array. Include patterns for all routes except static files and Next.js internals. Add explicit patterns for api and trpc routes.

**Protected Routes with createRouteMatcher:**

Import clerkMiddleware and createRouteMatcher. Define isProtectedRoute using createRouteMatcher with array of protected route patterns. Patterns support wildcards for nested paths. Middleware function checks if request matches protected route and calls auth.protect() if true.

**Public Routes with createRouteMatcher:**

Define isPublicRoute using createRouteMatcher with array of public route patterns. Include authentication pages, landing pages, and marketing content. Middleware calls auth.protect() for any request not matching public routes.

**Multiple Matchers:**

Combine protected and admin routes. Define isAdminRoute separately with admin-specific patterns. Call auth.protect() with custom role check for admin routes. Enables fine-grained authorization.

**Middleware Execution Order:**
1. Clerk middleware processes authentication
2. Route matcher evaluates current path
3. auth.protect() enforces authentication requirement
4. Request continues to route handler or page

## Client-Side Authentication

React hooks for accessing authentication state and user data.

### useAuth Hook

Provides authentication state, session tokens, and sign-out functionality.

**Available Properties:**

userId (string | null):
- Unique user identifier when authenticated
- null when not authenticated
- Use for conditional rendering

sessionId (string | null):
- Current session identifier
- null when no active session
- Use for session tracking

isLoaded (boolean):
- True when authentication state has been determined
- False during initial load
- Always check before rendering

isSignedIn (boolean):
- True when user is authenticated
- False when not authenticated
- Use for auth-dependent UI

getToken (function):
- Async function returning session token
- Accepts template name for custom JWT templates
- Returns null if not authenticated

signOut (function):
- Async function to sign out user
- Clears session and redirects
- Accepts redirect URL option

**Example Usage:**

Import useAuth in client component. Destructure properties. Check isLoaded before rendering. Return loading state if not loaded. Conditionally render based on isSignedIn. Call getToken for API requests with Authorization Bearer header.

### useUser Hook

Provides user profile data and metadata.

**Available Properties:**

isSignedIn (boolean):
- True when user is authenticated
- Same as useAuth isSignedIn

isLoaded (boolean):
- True when user data has loaded
- False during initial fetch

user (User | null):
- Full user object when authenticated
- null when not authenticated
- Contains profile, email, phone data

**User Object Structure:**

id:
- Unique user identifier
- Same as userId from useAuth

firstName:
- User's first name
- null if not provided

lastName:
- User's last name
- null if not provided

fullName:
- Computed full name
- Combines firstName and lastName

primaryEmailAddress:
- Email object with emailAddress string property
- null if no email

primaryPhoneNumber:
- Phone object with phoneNumber string property
- null if no phone

imageUrl:
- User profile image URL
- Clerk-hosted image

publicMetadata:
- Custom data visible to client
- Set via Dashboard or API

unsafeMetadata:
- Custom data editable by client
- Not recommended for sensitive data

**Example Usage:**

Import useUser in client component. Destructure isSignedIn, user, and isLoaded. Check isLoaded before rendering. Conditionally display user information. Access nested properties with optional chaining.

### useClerk Hook

Provides access to Clerk instance for advanced operations.

**Available Methods:**

openSignIn():
- Opens sign-in modal programmatically
- Accepts custom redirect URL

openSignUp():
- Opens sign-up modal programmatically
- Accepts custom redirect URL

redirectToSignIn():
- Redirects to sign-in page
- Useful for protected actions

redirectToSignUp():
- Redirects to sign-up page
- Useful for marketing CTAs

**Example Usage:**

Import useClerk. Destructure needed methods. Call in event handlers or useEffect. Handle navigation after authentication.

## Server-Side Authentication

Accessing authentication state in Server Components and Route Handlers.

### App Router Server Components

Import auth and currentUser from @clerk/nextjs/server. Make component async. Call await auth() to get userId. Redirect to sign-in if userId is null. Call await currentUser() to get full user object. Display user information.

**auth() Function:**

Returns object with:
- userId: User ID or null
- sessionId: Session ID or null
- getToken(): Get session token
- protect(): Throw error if not authenticated

**currentUser() Function:**

Returns full User object or null. Same structure as useUser hook. Includes email, name, image, metadata.

### Route Handlers

Import auth from @clerk/nextjs/server. Create async GET, POST, etc. handler. Call await auth() to get authentication state. Return 401 Unauthorized if userId is null. Process authenticated request. Return JSON response.

**Token Validation:**

Call getToken() from auth() result. Validate token if needed. Include token in requests to external APIs. Handle token expiration errors.

### Middleware Authentication

Access authentication in middleware for dynamic routing.

Import auth from @clerk/nextjs/server. Call auth() in middleware function. Check userId for authentication. Redirect unauthenticated users to sign-in. Continue to route for authenticated users.

## Organization Management

Multi-tenant features for B2B applications.

### Organization Concepts

**Organization:**
- Group of users with shared resources
- One user can belong to multiple organizations
- Each organization has roles and permissions

**Membership:**
- User's relationship to organization
- Includes role assignment
- Can be invited or joined

**Roles:**
- Admin: Full organization access
- Member: Standard user access
- Custom: Defined in Dashboard

### OrganizationSwitcher Component

Pre-built dropdown for switching between user's organizations.

**Basic Usage:**

Import OrganizationSwitcher from @clerk/nextjs. Render in header or navigation. Automatically displays current organization. Shows list of user's organizations. Provides create organization option.

**Customization:**

Pass appearance prop for styling. Set hidePersonal to hide personal workspace. Configure afterCreateOrganizationUrl for redirect after creation.

### useOrganizationList Hook

Programmatic access to user's organizations.

**Configuration:**

Call useOrganizationList with userMemberships configuration. Set mode to infinite for pagination. Access userMemberships.data for organization list. Use hasNextPage and fetchNext for pagination.

**Organization Switching:**

Call setActive from useOrganization with organization ID. Updates current organization context. Triggers re-render of organization-dependent components.

**Example Implementation:**

Import useOrganizationList. Configure with userMemberships. Map over userMemberships.data to display organizations. Render buttons to switch organizations. Show load more button when hasNextPage is true.

### useOrganization Hook

Access current organization data and operations.

**Available Properties:**

organization (Organization | null):
- Current organization object
- null if no organization selected

isLoaded (boolean):
- True when organization data loaded
- False during fetch

membership (OrganizationMembership | null):
- Current user's membership in organization
- Includes role information

**Organization Object:**

id:
- Unique organization identifier

name:
- Organization display name

imageUrl:
- Organization logo URL

membersCount:
- Number of organization members

publicMetadata:
- Custom organization data

**Example Usage:**

Import useOrganization. Destructure organization and isLoaded. Check isLoaded before rendering. Display organization name and logo. Show member count. Conditionally render based on organization selection.

## WebAuthn and Passkeys

Passwordless authentication with security keys and biometrics.

### Configuration

Navigate to Clerk Dashboard > User & Authentication > Email, Phone, Username. Enable passkey authentication. Configure passkey options for sign-in and sign-up.

**Supported Authenticators:**

Platform Authenticators:
- Touch ID (macOS, iOS)
- Face ID (iOS)
- Windows Hello (Windows)
- Android biometrics

Roaming Authenticators:
- YubiKey
- Titan Security Key
- Other FIDO2-compliant keys

### Client-Side Implementation

**Passkey Enrollment:**

Use SignUp component with passkey enabled. User clicks passkey option during sign-up. Browser prompts for authenticator. User verifies with biometric or security key. Passkey registered to account.

**Passkey Sign-In:**

Use SignIn component with passkey enabled. User clicks passkey option. Browser prompts for authenticator. User verifies with biometric or security key. Signed in without password.

**Programmatic Passkey Management:**

Use user.createPasskey() to register new passkey. Use user.passkeys to list registered passkeys. Use user.deletePasskey(id) to remove passkey.

### Server-Side Verification

Passkeys are automatically verified by Clerk. No additional server-side code needed. Session created after successful passkey verification. Access user data with auth() or currentUser().

## Social Authentication

Pre-configured OAuth providers with minimal setup.

### Supported Providers

Major Providers:
- Google
- GitHub
- Microsoft
- Facebook
- Apple
- Twitter/X
- LinkedIn
- Discord
- Twitch
- TikTok

### Configuration

Navigate to Clerk Dashboard > User & Authentication > Social Connections. Toggle desired providers. Configure OAuth credentials if needed. Clerk provides test credentials for development.

**OAuth Application Setup:**

For production, create OAuth application with provider. Configure redirect URL to Clerk's callback. Copy client ID and client secret. Paste in Clerk Dashboard. Save and enable provider.

**Google OAuth Setup:**

Create project in Google Cloud Console. Enable Google Sign-In API. Create OAuth 2.0 credentials. Add authorized redirect URI: https://your-clerk-subdomain.clerk.accounts.dev/v1/oauth_callback. Copy client ID and secret to Clerk Dashboard.

**GitHub OAuth Setup:**

Navigate to GitHub Settings > Developer settings > OAuth Apps. Create new OAuth App. Set authorization callback URL to Clerk's callback. Copy client ID and generate client secret. Add to Clerk Dashboard.

### Client-Side Usage

**Pre-built Components:**

SignIn and SignUp components automatically show enabled social providers. Rendered as buttons with provider logos. No additional code needed.

**Custom Social Sign-In:**

Import useSignIn hook. Destructure signIn.authenticateWithRedirect. Call with strategy parameter. Specify social provider like oauth_google. Handle redirect after authentication.

**Example:**

Import useSignIn from @clerk/nextjs. Get signIn from hook. Create button click handler. Call signIn.authenticateWithRedirect with strategy oauth_google and redirectUrl. User redirected to Google for authentication.

## Custom Authentication Flows

Programmatic control over authentication for custom UX.

### Email/Password Sign-Up

Import useSignUp hook. Get signUp object. Create email input and password input. On submit, call signUp.create with emailAddress and password. Check for verification requirement. If needed, call signUp.prepareEmailAddressVerification. Show code input for user. Call signUp.attemptEmailAddressVerification with code. Set active session with setActive.

### Email/Password Sign-In

Import useSignIn hook. Get signIn and setActive. Create email and password inputs. On submit, call signIn.create with identifier and password. Check for status complete. Call setActive with session ID. Redirect to dashboard or home.

### Phone Number Authentication

Import useSignUp for registration. Create phone number input. Call signUp.create with phoneNumber. Call preparePhoneNumberVerification to send SMS code. Show code input. Call attemptPhoneNumberVerification with code. Set active session.

### Email Link Authentication

Import useSignIn. Call signIn.create with identifier and empty object. Call prepareFirstFactor with strategy email_link and emailAddressId. User receives magic link email. Redirect to callback page. Call attemptFirstFactor in callback to complete sign-in.

## Webhooks

Real-time event notifications for user actions.

### Webhook Events

User Events:
- user.created: New user registered
- user.updated: User profile changed
- user.deleted: User account deleted

Session Events:
- session.created: User signed in
- session.ended: User signed out
- session.removed: Session expired

Organization Events:
- organization.created: New organization created
- organization.updated: Organization data changed
- organization.deleted: Organization removed
- organizationMembership.created: User joined organization
- organizationMembership.updated: User role changed
- organizationMembership.deleted: User left organization

### Webhook Setup

Navigate to Clerk Dashboard > Webhooks. Click Add Endpoint. Enter endpoint URL (must be HTTPS). Select events to receive. Copy signing secret. Save endpoint.

**Endpoint Implementation:**

Create API route handler. Verify webhook signature with Clerk SDK. Parse webhook payload. Process event based on type. Return 200 OK response.

**Signature Verification:**

Import Webhook from svix. Create verifier with signing secret. Call verifier.verify with payload and headers. Throws error if signature invalid. Safe to process event if verification succeeds.

### Webhook Security

**Signature Verification Required:**
Always verify webhook signature. Prevents spoofed requests. Use Clerk's provided signing secret.

**HTTPS Required:**
Webhooks only sent to HTTPS endpoints. Local testing requires tunneling (ngrok, Clerk's svix).

**Idempotency:**
Handle duplicate webhook deliveries. Use event ID for deduplication. Store processed event IDs.

## Core 2 Migration

Breaking changes when upgrading from Core 1 to Core 2.

### Environment Variable Changes

Core 1:
- CLERK_FRONTEND_API

Core 2:
- NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY

Migration:
Replace CLERK_FRONTEND_API with NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY in environment files. Get new publishable key from Clerk Dashboard.

### API Changes

**Middleware:**

Core 1:
- authMiddleware() with publicRoutes config

Core 2:
- clerkMiddleware() with createRouteMatcher
- Remove publicRoutes from config
- Use route matcher functions

**Server Imports:**

Core 1:
- Import from @clerk/nextjs

Core 2:
- Import from @clerk/nextjs/server for server-side functions
- Separates client and server imports

**Session Management:**

Core 1:
- setSession(sessionId)

Core 2:
- setActive({ session: sessionId })
- Unified method for session and organization

**Image URLs:**

Core 1:
- user.profileImageUrl
- organization.logoUrl

Core 2:
- user.imageUrl
- organization.imageUrl
- Consistent naming

### Automated Migration

Run Clerk upgrade command:

```bash
npx @clerk/upgrade --from=core-1
```

**What It Does:**
- Updates package versions
- Modifies import statements
- Replaces deprecated APIs
- Updates middleware syntax
- Suggests manual changes

**Manual Steps After:**
- Update environment variables
- Test authentication flows
- Verify route protection
- Check organization code
- Test webhooks

### Breaking Changes Checklist

- [ ] Update @clerk/nextjs to version 6.x
- [ ] Replace CLERK_FRONTEND_API with NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY
- [ ] Update middleware from authMiddleware to clerkMiddleware
- [ ] Change server imports to @clerk/nextjs/server
- [ ] Replace setSession with setActive
- [ ] Update profileImageUrl to imageUrl
- [ ] Update logoUrl to imageUrl
- [ ] Test all authentication flows
- [ ] Verify protected routes work
- [ ] Test organization features

## Performance Optimization

### Client-Side Optimization

**Code Splitting:**
Clerk components are automatically code-split. Only loaded when used. Reduces initial bundle size.

**Preloading:**
Use prefetch on Link components to sign-in pages. Loads authentication UI in background. Faster sign-in experience.

**Caching:**
Clerk SDK caches user and session data. Reduces API calls. Automatic cache invalidation on updates.

### Server-Side Optimization

**Session Token Caching:**
Cache getToken() results for API calls. Reduces token generation. Implement with short TTL (5 minutes).

**User Data Caching:**
Cache currentUser() results. Reduces database queries. Use Next.js cache or Redis.

**Middleware Optimization:**
Minimize middleware logic. Check authentication only where needed. Use route matchers for efficiency.

## Troubleshooting

### Common Issues

**Clerk Components Not Rendering:**

Check that ClerkProvider wraps app in layout.tsx. Verify environment variables are set. Ensure public variables have NEXT_PUBLIC_ prefix.

**Middleware Not Protecting Routes:**

Verify middleware.ts is in project root. Check matcher patterns include protected routes. Confirm clerkMiddleware is exported as default.

**Sign-In Redirect Loop:**

Check NEXT_PUBLIC_CLERK_SIGN_IN_URL matches actual route. Ensure sign-in page is public in middleware. Verify after-sign-in URL exists.

**User Data Not Loading:**

Check isLoaded before rendering. Handle null user state. Verify useUser is in client component.

**Session Token Empty:**

Ensure user is authenticated. Check getToken() is awaited. Verify token template if using custom JWT.

### Debug Mode

Enable debug logging in development:

Create .env.local entry for CLERK_DEBUG set to true. Clerk SDK logs to console. Shows authentication state changes. Displays API requests and responses.

## Resources

**Official Documentation:**
- Clerk Docs: https://clerk.com/docs
- Next.js Quickstart: https://clerk.com/docs/quickstarts/nextjs
- API Reference: https://clerk.com/docs/reference/nextjs/overview
- Core 2 Migration: https://clerk.com/docs/guides/development/upgrading/upgrade-guides/core-2/nextjs

**Clerk SDKs:**
- @clerk/nextjs: Next.js integration
- @clerk/clerk-react: React integration
- @clerk/express: Express.js integration
- @clerk/backend: Backend SDK for Node.js

**Community:**
- Discord: clerk.com/discord
- GitHub: github.com/clerk
- Twitter: @clerk_dev

**Tools:**
- Clerk Dashboard: dashboard.clerk.com
- Webhook Tester: dashboard.clerk.com/webhooks
- API Explorer: clerk.com/docs/reference/backend-api
