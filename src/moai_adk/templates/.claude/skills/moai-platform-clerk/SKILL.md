---
name: "moai-platform-clerk"
description: "Clerk modern authentication specialist covering WebAuthn, passkeys, passwordless, and beautiful UI components. Use when implementing modern auth with great UX."
version: 2.1.0
category: "platform"
modularized: true
user-invocable: false
tags: ['clerk', 'webauthn', 'passkeys', 'passwordless', 'authentication']
context7-libraries: "/clerk/clerk-docs"
related-skills: "moai-platform-auth0, moai-lang-typescript"
updated: 2026-01-11
status: "active"
allowed-tools:
  - Read
  - Write
  - Bash
  - Grep
  - Glob
---

# Clerk Modern Authentication Specialist

Modern authentication platform with WebAuthn, passkeys, passwordless flows, beautiful pre-built UI components, and multi-tenant organization support.

SDK Versions as of December 2025:

- @clerk/nextjs version 6.x with Core 2 requires Next.js 13.0.4 or higher and React 18 or higher
- @clerk/clerk-react version 5.x with Core 2 requires React 18 or higher
- @clerk/express version 1.x
- Node.js 18.17.0 or higher required

## Quick Reference

Environment Variables:

The .env.local file requires NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY set to your publishable key starting with pk_test, CLERK_SECRET_KEY set to your secret key starting with sk_test, NEXT_PUBLIC_CLERK_SIGN_IN_URL set to the sign-in route path, and NEXT_PUBLIC_CLERK_SIGN_UP_URL set to the sign-up route path.

ClerkProvider Setup:

In the app/layout.tsx file, import ClerkProvider from @clerk/nextjs. The RootLayout component wraps the html and body elements with ClerkProvider. The children prop is passed through the body element.

Basic Middleware:

In the middleware.ts file, import clerkMiddleware from @clerk/nextjs/server. Export the default clerkMiddleware function call. Export a config object with a matcher array containing patterns for all routes except static files and Next.js internals, plus patterns for api and trpc routes.

Context7 Access:

- Library: /clerk/clerk-docs
- Resolution: Use resolve-library-id with "clerk" then get-library-docs

---

## Implementation Guide

### ClerkProvider with Authentication Components

The full layout with sign-in and sign-out controls imports ClerkProvider, SignInButton, SignUpButton, SignedIn, SignedOut, and UserButton from @clerk/nextjs. The RootLayout component wraps everything in ClerkProvider. The header element contains SignedOut wrapper showing SignInButton and SignUpButton when user is not authenticated, and SignedIn wrapper showing UserButton when user is authenticated. Children are rendered in the body below the header.

### Protecting Routes with Middleware

Route protection with createRouteMatcher involves importing clerkMiddleware and createRouteMatcher from @clerk/nextjs/server. Define isProtectedRoute using createRouteMatcher with an array of protected route patterns such as /dashboard, /forum, and /api/private paths with wildcard suffixes. The middleware function checks if the request matches a protected route and calls auth.protect() if so. Export the config object with the matcher array.

To protect all routes except public ones, define isPublicRoute with createRouteMatcher containing public paths like sign-in, sign-up, root, and about. The middleware calls auth.protect() for any request that is not a public route.

### useAuth Hook

The useAuth hook provides access to authentication state and tokens. Import useAuth from @clerk/nextjs in a client component. Destructure userId, sessionId, getToken, isLoaded, and isSignedIn from the hook. The getToken function retrieves the session token for API calls with Authorization Bearer header. Check isLoaded before rendering and isSignedIn before showing authenticated content.

### useUser Hook

The useUser hook provides access to user profile data. Import useUser from @clerk/nextjs in a client component. Destructure isSignedIn, user, and isLoaded from the hook. Access user properties including firstName, primaryEmailAddress.emailAddress, and imageUrl for profile display.

### SignIn and SignUp Pages

For dedicated authentication pages, create app/sign-in/[[...sign-in]]/page.tsx and import SignIn from @clerk/nextjs. The SignInPage component renders the SignIn component centered on the page using flexbox with min-h-screen and items-center justify-center classes.

Similarly, create app/sign-up/[[...sign-up]]/page.tsx and import SignUp from @clerk/nextjs. The SignUpPage component renders the SignUp component with the same centered layout.

### Server-Side Authentication

For App Router server components, import auth and currentUser from @clerk/nextjs/server and redirect from next/navigation. In an async DashboardPage component, await auth() to get userId. If userId is null, redirect to sign-in. Await currentUser() to get the full user object for display.

For Route Handler authentication, import auth from @clerk/nextjs/server and NextResponse from next/server. In the GET handler, await auth() to get userId. If userId is null, return a 401 Unauthorized response. Otherwise return the userId in the response.

### Organization Management

The OrganizationSwitcher component is imported from @clerk/nextjs and rendered in dashboard layouts. It provides a dropdown for users to switch between organizations they belong to.

For custom organization switching, use the useOrganizationList hook with userMemberships configuration set to infinite mode. Check isLoaded before rendering. Map over userMemberships.data to display organization names with select buttons that call setActive with the organization ID. Include a load more button when hasNextPage is true.

The useOrganization hook provides access to the current organization. Destructure organization and isLoaded from the hook. Display organization name, imageUrl, and membersCount when an organization is selected.

---

## Advanced Patterns

For advanced implementation patterns including Core 2 migration, role-based access control, webhooks, custom auth flows, and external service integration, see the modules/advanced-patterns.md file.

Key Core 2 Migration Changes:

- Environment: CLERK_FRONTEND_API changed to NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY
- Middleware: authMiddleware() changed to clerkMiddleware()
- Imports: @clerk/nextjs changed to @clerk/nextjs/server for server-side
- Session: setSession() changed to setActive()
- Images: profileImageUrl and logoUrl changed to imageUrl

Migration Tool:

Run the Clerk upgrade command with npx @clerk/upgrade specifying --from=core-1 to migrate from Core 1 to Core 2.

---

## Resources

Official Documentation:

- Quickstart at clerk.com/docs/quickstarts/nextjs
- SDK Reference at clerk.com/docs/reference/nextjs/overview
- Core 2 Migration at clerk.com/docs/guides/development/upgrading/upgrade-guides/core-2/nextjs
- Webhooks at clerk.com/docs/integrations/webhooks

Works Well With:

- moai-platform-auth0 for alternative enterprise SSO solution
- moai-platform-supabase for Supabase authentication integration
- moai-platform-vercel for Vercel deployment with Clerk
- moai-lang-typescript for TypeScript development patterns
- moai-domain-frontend for React and Next.js integration

---

Status: Production Ready
Version: 2.1.0
Last Updated: 2026-01-11
SDK Version: @clerk/nextjs 6.x (Core 2)
