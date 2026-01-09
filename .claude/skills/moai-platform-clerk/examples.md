# Clerk Authentication - Working Examples

Complete working examples for Clerk authentication in Next.js 15 with React 19 and TypeScript.

---

## 1. Basic Setup

### Next.js App Router with Clerk Provider

Complete project initialization with environment setup:

```bash
# Install dependencies
npm install @clerk/nextjs

# Create environment file
cat > .env.local << EOF
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_your_key_here
CLERK_SECRET_KEY=sk_test_your_secret_key_here
NEXT_PUBLIC_CLERK_SIGN_IN_URL=/sign-in
NEXT_PUBLIC_CLERK_SIGN_UP_URL=/sign-up
NEXT_PUBLIC_CLERK_AFTER_SIGN_IN_URL=/dashboard
NEXT_PUBLIC_CLERK_AFTER_SIGN_UP_URL=/dashboard
EOF
```

Root layout with ClerkProvider:

```tsx
// app/layout.tsx
import type { Metadata } from "next";
import { ClerkProvider } from "@clerk/nextjs";
import "./globals.css";

export const metadata: Metadata = {
  title: "My App",
  description: "Application with Clerk authentication",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <ClerkProvider>
      <html lang="en">
        <body>{children}</body>
      </html>
    </ClerkProvider>
  );
}
```

---

## 2. Authentication Flow

### Sign-Up Page with Pre-built Component

```tsx
// app/sign-up/[[...sign-up]]/page.tsx
import { SignUp } from "@clerk/nextjs";

export default function SignUpPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="w-full max-w-md">
        <SignUp
          appearance={{
            elements: {
              rootBox: "mx-auto",
              card: "shadow-lg",
            },
          }}
        />
      </div>
    </div>
  );
}
```

### Sign-In Page with Pre-built Component

```tsx
// app/sign-in/[[...sign-in]]/page.tsx
import { SignIn } from "@clerk/nextjs";

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="w-full max-w-md">
        <SignIn
          appearance={{
            elements: {
              rootBox: "mx-auto",
              card: "shadow-lg",
            },
          }}
          routing="path"
          path="/sign-in"
        />
      </div>
    </div>
  );
}
```

### Navigation Header with Auth Controls

```tsx
// components/nav-header.tsx
import {
  SignInButton,
  SignUpButton,
  SignedIn,
  SignedOut,
  UserButton,
} from "@clerk/nextjs";
import Link from "next/link";

export function NavHeader() {
  return (
    <header className="border-b">
      <div className="container mx-auto flex h-16 items-center justify-between px-4">
        <Link href="/" className="text-xl font-bold">
          My App
        </Link>

        <nav className="flex items-center gap-4">
          <SignedOut>
            <SignInButton mode="modal">
              <button className="rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700">
                Sign In
              </button>
            </SignInButton>
            <SignUpButton mode="modal">
              <button className="rounded-md border border-gray-300 px-4 py-2 hover:bg-gray-50">
                Sign Up
              </button>
            </SignUpButton>
          </SignedOut>

          <SignedIn>
            <Link href="/dashboard" className="hover:text-blue-600">
              Dashboard
            </Link>
            <UserButton
              appearance={{
                elements: {
                  avatarBox: "w-10 h-10",
                },
              }}
            />
          </SignedIn>
        </nav>
      </div>
    </header>
  );
}
```

### Sign-Out Button with Redirect

```tsx
// components/sign-out-button.tsx
"use client";

import { useClerk } from "@clerk/nextjs";
import { useRouter } from "next/navigation";

export function SignOutButton() {
  const { signOut } = useClerk();
  const router = useRouter();

  const handleSignOut = async () => {
    await signOut();
    router.push("/");
  };

  return (
    <button
      onClick={handleSignOut}
      className="rounded-md bg-red-600 px-4 py-2 text-white hover:bg-red-700"
    >
      Sign Out
    </button>
  );
}
```

---

## 3. WebAuthn and Passkeys

### Enable WebAuthn in Clerk Dashboard

Enable passkeys in Clerk Dashboard:

1. Navigate to User & Authentication → Email, Phone, Username
2. Enable "Passkeys" authentication method
3. Configure passkey settings

### Client Component with Passkey Support

```tsx
// components/passkey-setup.tsx
"use client";

import { useUser } from "@clerk/nextjs";
import { useState } from "react";

export function PasskeySetup() {
  const { user } = useUser();
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const createPasskey = async () => {
    if (!user) return;

    setIsCreating(true);
    setError(null);

    try {
      await user.createPasskey();
      alert("Passkey created successfully!");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create passkey");
    } finally {
      setIsCreating(false);
    }
  };

  const passkeys = user?.passkeys || [];

  return (
    <div className="rounded-lg border p-6">
      <h2 className="mb-4 text-xl font-semibold">Passkeys</h2>

      {passkeys.length > 0 ? (
        <div className="mb-4 space-y-2">
          {passkeys.map((passkey) => (
            <div
              key={passkey.id}
              className="flex items-center justify-between rounded border p-3"
            >
              <div>
                <p className="font-medium">
                  {passkey.name || "Unnamed Passkey"}
                </p>
                <p className="text-sm text-gray-600">
                  Created: {new Date(passkey.createdAt).toLocaleDateString()}
                </p>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <p className="mb-4 text-gray-600">No passkeys configured</p>
      )}

      {error && (
        <div className="mb-4 rounded bg-red-50 p-3 text-red-700">{error}</div>
      )}

      <button
        onClick={createPasskey}
        disabled={isCreating}
        className="rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:bg-gray-400"
      >
        {isCreating ? "Creating..." : "Add Passkey"}
      </button>
    </div>
  );
}
```

### Sign In with Passkey Flow

```tsx
// components/passkey-signin.tsx
"use client";

import { useSignIn } from "@clerk/nextjs";
import { useState } from "react";
import { useRouter } from "next/navigation";

export function PasskeySignIn() {
  const { signIn, setActive } = useSignIn();
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();

  const handlePasskeySignIn = async () => {
    if (!signIn) return;

    setIsLoading(true);

    try {
      const signInAttempt = await signIn.authenticateWithPasskey();

      if (signInAttempt.status === "complete") {
        await setActive({ session: signInAttempt.createdSessionId });
        router.push("/dashboard");
      }
    } catch (err) {
      console.error("Passkey authentication failed:", err);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <button
      onClick={handlePasskeySignIn}
      disabled={isLoading}
      className="w-full rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:bg-gray-400"
    >
      {isLoading ? "Authenticating..." : "Sign In with Passkey"}
    </button>
  );
}
```

---

## 4. Session Management

### Client-Side Session Access

```tsx
// components/session-info.tsx
"use client";

import { useAuth, useSession } from "@clerk/nextjs";

export function SessionInfo() {
  const { userId, sessionId, isLoaded, isSignedIn } = useAuth();
  const { session } = useSession();

  if (!isLoaded) {
    return <div>Loading session...</div>;
  }

  if (!isSignedIn) {
    return <div>No active session</div>;
  }

  return (
    <div className="rounded-lg border p-6">
      <h2 className="mb-4 text-xl font-semibold">Session Information</h2>

      <dl className="space-y-2">
        <div>
          <dt className="font-medium">User ID:</dt>
          <dd className="font-mono text-sm text-gray-600">{userId}</dd>
        </div>

        <div>
          <dt className="font-medium">Session ID:</dt>
          <dd className="font-mono text-sm text-gray-600">{sessionId}</dd>
        </div>

        <div>
          <dt className="font-medium">Last Active:</dt>
          <dd className="text-sm text-gray-600">
            {session?.lastActiveAt
              ? new Date(session.lastActiveAt).toLocaleString()
              : "N/A"}
          </dd>
        </div>

        <div>
          <dt className="font-medium">Expires At:</dt>
          <dd className="text-sm text-gray-600">
            {session?.expireAt
              ? new Date(session.expireAt).toLocaleString()
              : "N/A"}
          </dd>
        </div>
      </dl>
    </div>
  );
}
```

### Server-Side Session Access

```tsx
// app/dashboard/session/page.tsx
import { auth, clerkClient } from "@clerk/nextjs/server";
import { redirect } from "next/navigation";

export default async function SessionPage() {
  const { userId, sessionId } = await auth();

  if (!userId || !sessionId) {
    redirect("/sign-in");
  }

  const session = await clerkClient().sessions.getSession(sessionId);

  return (
    <div className="container mx-auto max-w-2xl py-8">
      <h1 className="mb-6 text-3xl font-bold">Session Details</h1>

      <div className="rounded-lg border p-6">
        <h2 className="mb-4 text-xl font-semibold">Server-Side Session Data</h2>

        <dl className="space-y-3">
          <div>
            <dt className="font-medium">Session ID:</dt>
            <dd className="font-mono text-sm text-gray-600">{session.id}</dd>
          </div>

          <div>
            <dt className="font-medium">User ID:</dt>
            <dd className="font-mono text-sm text-gray-600">
              {session.userId}
            </dd>
          </div>

          <div>
            <dt className="font-medium">Status:</dt>
            <dd className="text-sm">
              <span
                className={`rounded px-2 py-1 ${
                  session.status === "active"
                    ? "bg-green-100 text-green-800"
                    : "bg-gray-100 text-gray-800"
                }`}
              >
                {session.status}
              </span>
            </dd>
          </div>

          <div>
            <dt className="font-medium">Created:</dt>
            <dd className="text-sm text-gray-600">
              {new Date(session.createdAt).toLocaleString()}
            </dd>
          </div>

          <div>
            <dt className="font-medium">Last Active:</dt>
            <dd className="text-sm text-gray-600">
              {new Date(session.lastActiveAt).toLocaleString()}
            </dd>
          </div>
        </dl>
      </div>
    </div>
  );
}
```

---

## 5. Protected Routes

### Middleware with Route Matching

```typescript
// middleware.ts
import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server";

// Define public routes (no authentication required)
const isPublicRoute = createRouteMatcher([
  "/",
  "/sign-in(.*)",
  "/sign-up(.*)",
  "/about",
  "/api/public(.*)",
]);

// Define admin routes (require specific role)
const isAdminRoute = createRouteMatcher(["/admin(.*)"]);

export default clerkMiddleware(async (auth, req) => {
  // Allow public routes
  if (isPublicRoute(req)) {
    return;
  }

  // Protect admin routes with role check
  if (isAdminRoute(req)) {
    await auth.protect((has) => {
      return has({ role: "admin" });
    });
    return;
  }

  // Protect all other routes
  await auth.protect();
});

export const config = {
  matcher: [
    // Skip Next.js internals and static files
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    // Always run for API routes
    "/(api|trpc)(.*)",
  ],
};
```

### Server Component Route Protection

```tsx
// app/dashboard/page.tsx
import { auth, currentUser } from "@clerk/nextjs/server";
import { redirect } from "next/navigation";

export default async function DashboardPage() {
  const { userId } = await auth();

  if (!userId) {
    redirect("/sign-in");
  }

  const user = await currentUser();

  return (
    <div className="container mx-auto py-8">
      <h1 className="mb-6 text-3xl font-bold">
        Welcome, {user?.firstName || "User"}!
      </h1>

      <div className="grid gap-6 md:grid-cols-2">
        <div className="rounded-lg border p-6">
          <h2 className="mb-2 text-xl font-semibold">Profile</h2>
          <p className="text-gray-600">Manage your account settings</p>
        </div>

        <div className="rounded-lg border p-6">
          <h2 className="mb-2 text-xl font-semibold">Activity</h2>
          <p className="text-gray-600">View your recent activity</p>
        </div>
      </div>
    </div>
  );
}
```

### API Route Protection

```typescript
// app/api/protected/route.ts
import { auth } from "@clerk/nextjs/server";
import { NextResponse } from "next/server";

export async function GET() {
  const { userId } = await auth();

  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  // Fetch user-specific data
  const data = {
    userId,
    message: "This is protected data",
    timestamp: new Date().toISOString(),
  };

  return NextResponse.json(data);
}

export async function POST(req: Request) {
  const { userId, sessionClaims } = await auth();

  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const body = await req.json();

  // Process authenticated request
  return NextResponse.json({
    success: true,
    userId,
    data: body,
  });
}
```

---

## 6. User Profile Management

### User Profile Display Component

```tsx
// components/user-profile-display.tsx
"use client";

import { useUser } from "@clerk/nextjs";
import Image from "next/image";

export function UserProfileDisplay() {
  const { isLoaded, isSignedIn, user } = useUser();

  if (!isLoaded) {
    return <div>Loading profile...</div>;
  }

  if (!isSignedIn || !user) {
    return <div>Please sign in to view your profile</div>;
  }

  return (
    <div className="rounded-lg border p-6">
      <div className="mb-4 flex items-center gap-4">
        {user.imageUrl && (
          <Image
            src={user.imageUrl}
            alt={`${user.firstName || "User"} profile`}
            width={80}
            height={80}
            className="rounded-full"
          />
        )}

        <div>
          <h2 className="text-2xl font-bold">
            {user.firstName} {user.lastName}
          </h2>
          {user.username && <p className="text-gray-600">@{user.username}</p>}
        </div>
      </div>

      <dl className="space-y-2">
        <div>
          <dt className="font-medium">Email:</dt>
          <dd className="text-gray-600">
            {user.primaryEmailAddress?.emailAddress}
          </dd>
        </div>

        {user.primaryPhoneNumber && (
          <div>
            <dt className="font-medium">Phone:</dt>
            <dd className="text-gray-600">
              {user.primaryPhoneNumber.phoneNumber}
            </dd>
          </div>
        )}

        <div>
          <dt className="font-medium">Member Since:</dt>
          <dd className="text-gray-600">
            {new Date(user.createdAt).toLocaleDateString()}
          </dd>
        </div>
      </dl>
    </div>
  );
}
```

### Update User Profile

```tsx
// components/update-profile-form.tsx
"use client";

import { useUser } from "@clerk/nextjs";
import { useState } from "react";

export function UpdateProfileForm() {
  const { user } = useUser();
  const [firstName, setFirstName] = useState(user?.firstName || "");
  const [lastName, setLastName] = useState(user?.lastName || "");
  const [isUpdating, setIsUpdating] = useState(false);
  const [message, setMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!user) return;

    setIsUpdating(true);
    setMessage(null);

    try {
      await user.update({
        firstName,
        lastName,
      });

      setMessage({ type: "success", text: "Profile updated successfully!" });
    } catch (err) {
      setMessage({
        type: "error",
        text: err instanceof Error ? err.message : "Failed to update profile",
      });
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label htmlFor="firstName" className="block text-sm font-medium">
          First Name
        </label>
        <input
          id="firstName"
          type="text"
          value={firstName}
          onChange={(e) => setFirstName(e.target.value)}
          className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
          required
        />
      </div>

      <div>
        <label htmlFor="lastName" className="block text-sm font-medium">
          Last Name
        </label>
        <input
          id="lastName"
          type="text"
          value={lastName}
          onChange={(e) => setLastName(e.target.value)}
          className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
          required
        />
      </div>

      {message && (
        <div
          className={`rounded p-3 ${
            message.type === "success"
              ? "bg-green-50 text-green-700"
              : "bg-red-50 text-red-700"
          }`}
        >
          {message.text}
        </div>
      )}

      <button
        type="submit"
        disabled={isUpdating}
        className="w-full rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:bg-gray-400"
      >
        {isUpdating ? "Updating..." : "Update Profile"}
      </button>
    </form>
  );
}
```

### Upload Profile Image

```tsx
// components/profile-image-upload.tsx
"use client";

import { useUser } from "@clerk/nextjs";
import { useState } from "react";
import Image from "next/image";

export function ProfileImageUpload() {
  const { user } = useUser();
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !user) return;

    setIsUploading(true);
    setError(null);

    try {
      await user.setProfileImage({ file });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to upload image");
    } finally {
      setIsUploading(false);
    }
  };

  if (!user) return null;

  return (
    <div className="flex items-center gap-4">
      {user.imageUrl && (
        <Image
          src={user.imageUrl}
          alt="Profile"
          width={100}
          height={100}
          className="rounded-full"
        />
      )}

      <div>
        <label
          htmlFor="profile-image"
          className="cursor-pointer rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
        >
          {isUploading ? "Uploading..." : "Upload Image"}
        </label>
        <input
          id="profile-image"
          type="file"
          accept="image/*"
          onChange={handleImageUpload}
          disabled={isUploading}
          className="hidden"
        />

        {error && <p className="mt-2 text-sm text-red-600">{error}</p>}
      </div>
    </div>
  );
}
```

---

## 7. Organizations

### Organization Switcher Component

```tsx
// app/dashboard/layout.tsx
import { OrganizationSwitcher } from "@clerk/nextjs";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div>
      <nav className="border-b">
        <div className="container mx-auto flex h-16 items-center justify-between px-4">
          <h1 className="text-xl font-bold">Dashboard</h1>

          <OrganizationSwitcher
            appearance={{
              elements: {
                rootBox: "flex items-center",
                organizationSwitcherTrigger: "rounded-md border px-3 py-2",
              },
            }}
            hidePersonal={false}
            afterCreateOrganizationUrl="/dashboard/organization"
            afterSelectOrganizationUrl="/dashboard/organization"
          />
        </div>
      </nav>

      <main className="container mx-auto py-8 px-4">{children}</main>
    </div>
  );
}
```

### Custom Organization List

```tsx
// components/organization-list.tsx
"use client";

import { useOrganizationList } from "@clerk/nextjs";
import Image from "next/image";

export function OrganizationList() {
  const { isLoaded, setActive, userMemberships } = useOrganizationList({
    userMemberships: {
      infinite: true,
    },
  });

  if (!isLoaded) {
    return <div>Loading organizations...</div>;
  }

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Your Organizations</h2>

      {userMemberships.data && userMemberships.data.length > 0 ? (
        <ul className="space-y-2">
          {userMemberships.data.map((membership) => (
            <li
              key={membership.id}
              className="flex items-center justify-between rounded-lg border p-4"
            >
              <div className="flex items-center gap-3">
                {membership.organization.imageUrl && (
                  <Image
                    src={membership.organization.imageUrl}
                    alt={membership.organization.name}
                    width={40}
                    height={40}
                    className="rounded"
                  />
                )}

                <div>
                  <p className="font-medium">{membership.organization.name}</p>
                  <p className="text-sm text-gray-600">
                    Role: {membership.role}
                  </p>
                </div>
              </div>

              <button
                onClick={() =>
                  setActive({ organization: membership.organization.id })
                }
                className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
              >
                Switch
              </button>
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-gray-600">
          You are not a member of any organizations
        </p>
      )}

      {userMemberships.hasNextPage && (
        <button
          onClick={() => userMemberships.fetchNext()}
          className="w-full rounded-md border px-4 py-2 hover:bg-gray-50"
        >
          Load More
        </button>
      )}
    </div>
  );
}
```

### Organization Members Management

```tsx
// components/organization-members.tsx
"use client";

import { useOrganization } from "@clerk/nextjs";
import Image from "next/image";

export function OrganizationMembers() {
  const { organization, memberships, isLoaded } = useOrganization({
    memberships: {
      infinite: true,
    },
  });

  if (!isLoaded) {
    return <div>Loading members...</div>;
  }

  if (!organization) {
    return <div>No organization selected</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">{organization.name} Members</h2>
        <span className="text-sm text-gray-600">
          {organization.membersCount} total members
        </span>
      </div>

      {memberships.data && memberships.data.length > 0 ? (
        <ul className="space-y-2">
          {memberships.data.map((membership) => (
            <li
              key={membership.id}
              className="flex items-center justify-between rounded-lg border p-4"
            >
              <div className="flex items-center gap-3">
                {membership.publicUserData.imageUrl && (
                  <Image
                    src={membership.publicUserData.imageUrl}
                    alt={membership.publicUserData.firstName || "Member"}
                    width={40}
                    height={40}
                    className="rounded-full"
                  />
                )}

                <div>
                  <p className="font-medium">
                    {membership.publicUserData.firstName}{" "}
                    {membership.publicUserData.lastName}
                  </p>
                  <p className="text-sm text-gray-600">
                    {membership.publicUserData.identifier}
                  </p>
                </div>
              </div>

              <span className="rounded bg-blue-100 px-3 py-1 text-sm text-blue-800">
                {membership.role}
              </span>
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-gray-600">No members found</p>
      )}

      {memberships.hasNextPage && (
        <button
          onClick={() => memberships.fetchNext()}
          className="w-full rounded-md border px-4 py-2 hover:bg-gray-50"
        >
          Load More Members
        </button>
      )}
    </div>
  );
}
```

---

## 8. Webhooks

### Webhook Handler with Svix Verification

```bash
# Install Svix for webhook verification
npm install svix
```

```typescript
// app/api/webhooks/clerk/route.ts
import { Webhook } from "svix";
import { headers } from "next/headers";
import { WebhookEvent } from "@clerk/nextjs/server";

export async function POST(req: Request) {
  // Get webhook secret from environment
  const WEBHOOK_SECRET = process.env.CLERK_WEBHOOK_SECRET;

  if (!WEBHOOK_SECRET) {
    throw new Error("CLERK_WEBHOOK_SECRET is not defined");
  }

  // Get headers
  const headerPayload = await headers();
  const svix_id = headerPayload.get("svix-id");
  const svix_timestamp = headerPayload.get("svix-timestamp");
  const svix_signature = headerPayload.get("svix-signature");

  if (!svix_id || !svix_timestamp || !svix_signature) {
    return new Response("Missing svix headers", { status: 400 });
  }

  // Get body
  const payload = await req.json();
  const body = JSON.stringify(payload);

  // Create Svix instance
  const wh = new Webhook(WEBHOOK_SECRET);

  let evt: WebhookEvent;

  // Verify webhook
  try {
    evt = wh.verify(body, {
      "svix-id": svix_id,
      "svix-timestamp": svix_timestamp,
      "svix-signature": svix_signature,
    }) as WebhookEvent;
  } catch (err) {
    console.error("Webhook verification failed:", err);
    return new Response("Invalid signature", { status: 400 });
  }

  // Handle events
  const eventType = evt.type;

  switch (eventType) {
    case "user.created":
      const { id, email_addresses, first_name, last_name } = evt.data;
      console.log("User created:", {
        id,
        email_addresses,
        first_name,
        last_name,
      });
      // Add user to database
      break;

    case "user.updated":
      console.log("User updated:", evt.data.id);
      // Update user in database
      break;

    case "user.deleted":
      console.log("User deleted:", evt.data.id);
      // Delete user from database
      break;

    case "organization.created":
      console.log("Organization created:", evt.data.id);
      // Create organization in database
      break;

    case "organization.updated":
      console.log("Organization updated:", evt.data.id);
      // Update organization in database
      break;

    case "organizationMembership.created":
      console.log("Organization membership created:", evt.data);
      // Add organization member to database
      break;

    default:
      console.log(`Unhandled event type: ${eventType}`);
  }

  return new Response("Webhook processed", { status: 200 });
}
```

### User Sync Service

```typescript
// lib/user-sync.ts
import { clerkClient } from "@clerk/nextjs/server";

export async function syncUserToDatabase(userId: string) {
  const user = await clerkClient().users.getUser(userId);

  // Your database sync logic
  const userData = {
    id: user.id,
    email: user.emailAddresses[0]?.emailAddress,
    firstName: user.firstName,
    lastName: user.lastName,
    imageUrl: user.imageUrl,
    createdAt: new Date(user.createdAt),
    updatedAt: new Date(user.updatedAt),
  };

  console.log("Syncing user to database:", userData);

  // Example: await db.users.upsert(userData)

  return userData;
}

export async function deleteUserFromDatabase(userId: string) {
  console.log("Deleting user from database:", userId);

  // Example: await db.users.delete({ where: { id: userId } })
}
```

---

## 9. Testing

### Mock Clerk in Tests

```bash
# Install testing dependencies
npm install -D @clerk/testing vitest @testing-library/react @testing-library/jest-dom
```

```tsx
// __tests__/components/protected-component.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";

// Mock Clerk hooks
vi.mock("@clerk/nextjs", () => ({
  useAuth: vi.fn(),
  useUser: vi.fn(),
}));

import { useAuth, useUser } from "@clerk/nextjs";
import { ProtectedComponent } from "@/components/protected-component";

describe("ProtectedComponent", () => {
  it("shows loading state when not loaded", () => {
    vi.mocked(useAuth).mockReturnValue({
      isLoaded: false,
      isSignedIn: false,
      userId: null,
      sessionId: null,
      getToken: vi.fn(),
    } as any);

    render(<ProtectedComponent />);
    expect(screen.getByText("Loading...")).toBeInTheDocument();
  });

  it("shows sign-in prompt when not authenticated", () => {
    vi.mocked(useAuth).mockReturnValue({
      isLoaded: true,
      isSignedIn: false,
      userId: null,
      sessionId: null,
      getToken: vi.fn(),
    } as any);

    render(<ProtectedComponent />);
    expect(screen.getByText("Please sign in")).toBeInTheDocument();
  });

  it("shows content when authenticated", () => {
    vi.mocked(useAuth).mockReturnValue({
      isLoaded: true,
      isSignedIn: true,
      userId: "user_123",
      sessionId: "sess_456",
      getToken: vi.fn(),
    } as any);

    vi.mocked(useUser).mockReturnValue({
      isLoaded: true,
      isSignedIn: true,
      user: {
        id: "user_123",
        firstName: "John",
        lastName: "Doe",
        emailAddresses: [{ emailAddress: "john@example.com" }],
      },
    } as any);

    render(<ProtectedComponent />);
    expect(screen.getByText("Welcome, John!")).toBeInTheDocument();
  });
});
```

### Integration Test Setup

```typescript
// __tests__/setup.ts
import { beforeAll, afterAll, afterEach } from "vitest";
import { setupClerkTestingToken } from "@clerk/testing/vitest";

beforeAll(() => {
  // Setup Clerk testing environment
  setupClerkTestingToken();
});

afterEach(() => {
  // Clear mocks after each test
  vi.clearAllMocks();
});

afterAll(() => {
  // Cleanup
});
```

### E2E Test with Playwright

```typescript
// e2e/auth.spec.ts
import { test, expect } from "@playwright/test";

test.describe("Authentication Flow", () => {
  test("should sign in successfully", async ({ page }) => {
    await page.goto("http://localhost:3000");

    // Click sign-in button
    await page.click("text=Sign In");

    // Fill in credentials
    await page.fill('input[name="identifier"]', "test@example.com");
    await page.click('button[type="submit"]');

    await page.fill('input[name="password"]', "testpassword123");
    await page.click('button[type="submit"]');

    // Wait for redirect to dashboard
    await page.waitForURL("**/dashboard");

    // Verify user is signed in
    await expect(page.locator("text=Welcome")).toBeVisible();
  });

  test("should sign out successfully", async ({ page }) => {
    // Assume already signed in
    await page.goto("http://localhost:3000/dashboard");

    // Click user button and sign out
    await page.click('[data-testid="user-button"]');
    await page.click("text=Sign Out");

    // Verify redirect to home
    await page.waitForURL("http://localhost:3000");
    await expect(page.locator("text=Sign In")).toBeVisible();
  });
});
```

---

## 10. Advanced Features

### Custom JWT Template with Backend API

```typescript
// app/api/external/route.ts
import { auth } from "@clerk/nextjs/server";
import { NextResponse } from "next/server";

export async function GET() {
  const { userId, getToken } = await auth();

  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  // Get token with custom template
  const token = await getToken({ template: "supabase" });

  // Call external API with token
  const response = await fetch("https://api.external.com/data", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  const data = await response.json();
  return NextResponse.json(data);
}
```

### Multi-Factor Authentication Setup

```tsx
// components/mfa-setup.tsx
"use client";

import { useUser } from "@clerk/nextjs";
import { useState } from "react";

export function MFASetup() {
  const { user } = useUser();
  const [isEnabling, setIsEnabling] = useState(false);
  const [qrCode, setQrCode] = useState<string | null>(null);
  const [backupCodes, setBackupCodes] = useState<string[]>([]);

  const enableTOTP = async () => {
    if (!user) return;

    setIsEnabling(true);

    try {
      const totp = await user.createTOTP();
      setQrCode(totp.qr_code);

      // Generate backup codes
      const codes = await user.createBackupCode();
      setBackupCodes(codes.codes);
    } catch (err) {
      console.error("Failed to enable TOTP:", err);
    } finally {
      setIsEnabling(false);
    }
  };

  return (
    <div className="rounded-lg border p-6">
      <h2 className="mb-4 text-xl font-semibold">Two-Factor Authentication</h2>

      {!qrCode ? (
        <button
          onClick={enableTOTP}
          disabled={isEnabling}
          className="rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
        >
          {isEnabling ? "Setting up..." : "Enable 2FA"}
        </button>
      ) : (
        <div className="space-y-4">
          <div>
            <p className="mb-2">
              Scan this QR code with your authenticator app:
            </p>
            <img src={qrCode} alt="QR Code" className="rounded border" />
          </div>

          {backupCodes.length > 0 && (
            <div>
              <p className="mb-2 font-medium">Backup Codes:</p>
              <ul className="space-y-1 font-mono text-sm">
                {backupCodes.map((code, idx) => (
                  <li key={idx} className="rounded bg-gray-100 p-2">
                    {code}
                  </li>
                ))}
              </ul>
              <p className="mt-2 text-sm text-gray-600">
                Save these codes in a secure location
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
```

---

**Last Updated**: 2026-01-10
