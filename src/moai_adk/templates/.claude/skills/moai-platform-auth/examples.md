# Authentication Platform Implementation Examples

## Auth0 Production Examples

### Enterprise SSO with SAML

Next.js Integration with Auth0 SDK:

```typescript
// lib/auth0.ts
import { Auth0Client } from '@auth0/auth0-spa-js';

const auth0Config = {
  domain: process.env.NEXT_PUBLIC_AUTH0_DOMAIN!,
  clientId: process.env.NEXT_PUBLIC_AUTH0_CLIENT_ID!,
  authorizationParams: {
    redirect_uri: typeof window !== 'undefined' ? window.location.origin : '',
    audience: process.env.NEXT_PUBLIC_AUTH0_AUDIENCE,
    scope: 'openid profile email',
  },
};

let auth0Client: Auth0Client | null = null;

export async function getAuth0Client(): Promise<Auth0Client> {
  if (!auth0Client) {
    auth0Client = new Auth0Client(auth0Config);
  }
  return auth0Client;
}

export async function loginWithSSO(connection: string, organization?: string) {
  const client = await getAuth0Client();
  await client.loginWithRedirect({
    authorizationParams: {
      connection,
      organization,
    },
  });
}

export async function handleCallback() {
  const client = await getAuth0Client();
  await client.handleRedirectCallback();
  return client.getUser();
}
```

### Auth0 Organizations with RBAC

Organization Management API:

```typescript
// lib/auth0-management.ts
import { ManagementClient } from 'auth0';

const management = new ManagementClient({
  domain: process.env.AUTH0_DOMAIN!,
  clientId: process.env.AUTH0_M2M_CLIENT_ID!,
  clientSecret: process.env.AUTH0_M2M_CLIENT_SECRET!,
});

export class Auth0OrganizationManager {
  async createOrganization(data: {
    name: string;
    displayName: string;
    branding?: {
      logoUrl?: string;
      colors?: { primary: string; page_background: string };
    };
    metadata?: Record<string, string>;
  }) {
    return management.organizations.create({
      name: data.name.toLowerCase().replace(/\s+/g, '-'),
      display_name: data.displayName,
      branding: data.branding,
      metadata: data.metadata,
    });
  }

  async addMember(organizationId: string, userId: string, roles?: string[]) {
    await management.organizations.addMembers(
      { id: organizationId },
      { members: [userId] }
    );

    if (roles && roles.length > 0) {
      await management.organizations.addMemberRoles(
        { id: organizationId, user_id: userId },
        { roles }
      );
    }
  }

  async getMemberRoles(organizationId: string, userId: string) {
    return management.organizations.getMemberRoles({
      id: organizationId,
      user_id: userId,
    });
  }

  async inviteMember(
    organizationId: string,
    email: string,
    roles: string[],
    inviterName: string
  ) {
    return management.organizations.createInvitation(
      { id: organizationId },
      {
        invitee: { email },
        inviter: { name: inviterName },
        roles,
        client_id: process.env.AUTH0_CLIENT_ID!,
        connection_id: process.env.AUTH0_CONNECTION_ID,
        send_invitation_email: true,
      }
    );
  }

  async listOrganizationMembers(organizationId: string, page = 0, perPage = 50) {
    return management.organizations.getMembers({
      id: organizationId,
      page,
      per_page: perPage,
    });
  }
}
```

### Auth0 Python SDK Integration

FastAPI Backend with Auth0:

```python
# auth/auth0.py
from functools import wraps
from typing import Optional, List
import jwt
from jwt import PyJWKClient
from fastapi import HTTPException, Depends, Security
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel
import httpx
import os

class Auth0Config:
    DOMAIN = os.getenv("AUTH0_DOMAIN")
    API_AUDIENCE = os.getenv("AUTH0_API_AUDIENCE")
    ALGORITHMS = ["RS256"]
    ISSUER = f"https://{DOMAIN}/"

class TokenPayload(BaseModel):
    sub: str
    email: Optional[str] = None
    permissions: List[str] = []
    org_id: Optional[str] = None
    org_name: Optional[str] = None

security = HTTPBearer()
jwks_client = PyJWKClient(f"https://{Auth0Config.DOMAIN}/.well-known/jwks.json")

async def verify_token(
    credentials: HTTPAuthorizationCredentials = Security(security)
) -> TokenPayload:
    """Verify Auth0 JWT token and extract payload."""
    token = credentials.credentials

    try:
        signing_key = jwks_client.get_signing_key_from_jwt(token)

        payload = jwt.decode(
            token,
            signing_key.key,
            algorithms=Auth0Config.ALGORITHMS,
            audience=Auth0Config.API_AUDIENCE,
            issuer=Auth0Config.ISSUER,
        )

        return TokenPayload(
            sub=payload["sub"],
            email=payload.get("email"),
            permissions=payload.get("permissions", []),
            org_id=payload.get("org_id"),
            org_name=payload.get("org_name"),
        )

    except jwt.ExpiredSignatureError:
        raise HTTPException(status_code=401, detail="Token has expired")
    except jwt.InvalidTokenError as e:
        raise HTTPException(status_code=401, detail=f"Invalid token: {str(e)}")

def require_permissions(*required_permissions: str):
    """Decorator to require specific permissions."""
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, token: TokenPayload = Depends(verify_token), **kwargs):
            for permission in required_permissions:
                if permission not in token.permissions:
                    raise HTTPException(
                        status_code=403,
                        detail=f"Missing required permission: {permission}"
                    )
            return await func(*args, token=token, **kwargs)
        return wrapper
    return decorator

def require_organization():
    """Ensure user belongs to an organization."""
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, token: TokenPayload = Depends(verify_token), **kwargs):
            if not token.org_id:
                raise HTTPException(
                    status_code=403,
                    detail="Organization membership required"
                )
            return await func(*args, token=token, **kwargs)
        return wrapper
    return decorator
```

Auth0 Management API Python Client:

```python
# auth/management.py
from auth0.management import Auth0
import os

class Auth0ManagementClient:
    def __init__(self):
        self.client = Auth0(
            domain=os.getenv("AUTH0_DOMAIN"),
            token=os.getenv("AUTH0_MANAGEMENT_TOKEN"),
        )

    async def create_organization(
        self,
        name: str,
        display_name: str,
        metadata: dict = None
    ):
        """Create a new organization."""
        return self.client.organizations.create({
            "name": name.lower().replace(" ", "-"),
            "display_name": display_name,
            "metadata": metadata or {},
        })

    async def add_organization_member(
        self,
        org_id: str,
        user_id: str,
        roles: list = None
    ):
        """Add user to organization with optional roles."""
        self.client.organizations.create_organization_members(
            org_id,
            {"members": [user_id]}
        )

        if roles:
            self.client.organizations.create_organization_member_roles(
                org_id,
                user_id,
                {"roles": roles}
            )

    async def get_user_organizations(self, user_id: str):
        """Get all organizations a user belongs to."""
        return self.client.users.list_organizations(user_id)

    async def create_user(
        self,
        email: str,
        password: str = None,
        connection: str = "Username-Password-Authentication",
        verify_email: bool = True
    ):
        """Create a new user."""
        user_data = {
            "email": email,
            "connection": connection,
            "email_verified": not verify_email,
        }

        if password:
            user_data["password"] = password

        return self.client.users.create(user_data)
```

---

## Clerk Production Examples

### Next.js with Clerk Middleware

Complete Middleware Setup:

```typescript
// middleware.ts
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server';
import { NextResponse } from 'next/server';

const isPublicRoute = createRouteMatcher([
  '/',
  '/sign-in(.*)',
  '/sign-up(.*)',
  '/api/webhook(.*)',
  '/api/public(.*)',
]);

const isAdminRoute = createRouteMatcher([
  '/admin(.*)',
  '/api/admin(.*)',
]);

const isOrganizationRoute = createRouteMatcher([
  '/organization(.*)',
  '/team(.*)',
  '/api/organization(.*)',
]);

export default clerkMiddleware(async (auth, req) => {
  const { userId, orgId, orgRole, sessionClaims } = await auth();

  // Allow public routes
  if (isPublicRoute(req)) {
    return NextResponse.next();
  }

  // Require authentication for all other routes
  if (!userId) {
    const signInUrl = new URL('/sign-in', req.url);
    signInUrl.searchParams.set('redirect_url', req.url);
    return NextResponse.redirect(signInUrl);
  }

  // Admin routes require admin role
  if (isAdminRoute(req)) {
    const isAdmin = sessionClaims?.metadata?.role === 'admin';
    if (!isAdmin) {
      return NextResponse.redirect(new URL('/unauthorized', req.url));
    }
  }

  // Organization routes require org membership
  if (isOrganizationRoute(req)) {
    if (!orgId) {
      return NextResponse.redirect(new URL('/select-organization', req.url));
    }
  }

  return NextResponse.next();
});

export const config = {
  matcher: [
    '/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)',
    '/(api|trpc)(.*)',
  ],
};
```

### Clerk Organization Components

Organization Dashboard:

```typescript
// app/organization/page.tsx
import { OrganizationProfile, OrganizationSwitcher } from '@clerk/nextjs';
import { auth, currentUser } from '@clerk/nextjs/server';
import { redirect } from 'next/navigation';

export default async function OrganizationPage() {
  const { orgId, orgRole } = await auth();
  const user = await currentUser();

  if (!user) {
    redirect('/sign-in');
  }

  if (!orgId) {
    redirect('/select-organization');
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold">Organization Settings</h1>
        <OrganizationSwitcher
          afterSelectOrganizationUrl="/organization"
          afterLeaveOrganizationUrl="/"
        />
      </div>

      <div className="grid gap-6">
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Current Role: {orgRole}</h2>
          {orgRole === 'org:admin' && (
            <p className="text-green-600">You have admin privileges</p>
          )}
        </div>

        <OrganizationProfile
          appearance={{
            elements: {
              rootBox: 'w-full',
              card: 'shadow-none',
            },
          }}
        />
      </div>
    </div>
  );
}
```

### Clerk Backend API Usage

Server-Side User Management:

```typescript
// lib/clerk-server.ts
import { clerkClient } from '@clerk/nextjs/server';

export class ClerkUserManager {
  async createUser(data: {
    email: string;
    firstName?: string;
    lastName?: string;
    password?: string;
  }) {
    const client = await clerkClient();

    return client.users.createUser({
      emailAddress: [data.email],
      firstName: data.firstName,
      lastName: data.lastName,
      password: data.password,
    });
  }

  async updateUserMetadata(userId: string, metadata: {
    public?: Record<string, unknown>;
    private?: Record<string, unknown>;
  }) {
    const client = await clerkClient();

    return client.users.updateUserMetadata(userId, {
      publicMetadata: metadata.public,
      privateMetadata: metadata.private,
    });
  }

  async createOrganization(data: {
    name: string;
    createdBy: string;
    publicMetadata?: Record<string, unknown>;
  }) {
    const client = await clerkClient();

    return client.organizations.createOrganization({
      name: data.name,
      createdBy: data.createdBy,
      publicMetadata: data.publicMetadata,
    });
  }

  async inviteToOrganization(
    orgId: string,
    email: string,
    role: 'org:admin' | 'org:member' = 'org:member'
  ) {
    const client = await clerkClient();

    return client.organizations.createOrganizationInvitation({
      organizationId: orgId,
      emailAddress: email,
      role,
      redirectUrl: `${process.env.NEXT_PUBLIC_APP_URL}/accept-invitation`,
    });
  }

  async getOrganizationMembers(orgId: string) {
    const client = await clerkClient();

    return client.organizations.getOrganizationMembershipList({
      organizationId: orgId,
    });
  }
}
```

### Clerk Webhooks

Webhook Handler for User Sync:

```typescript
// app/api/webhook/clerk/route.ts
import { Webhook } from 'svix';
import { headers } from 'next/headers';
import { WebhookEvent } from '@clerk/nextjs/server';
import { db } from '@/lib/database';

export async function POST(req: Request) {
  const WEBHOOK_SECRET = process.env.CLERK_WEBHOOK_SECRET;

  if (!WEBHOOK_SECRET) {
    throw new Error('Missing CLERK_WEBHOOK_SECRET');
  }

  const headerPayload = await headers();
  const svixId = headerPayload.get('svix-id');
  const svixTimestamp = headerPayload.get('svix-timestamp');
  const svixSignature = headerPayload.get('svix-signature');

  if (!svixId || !svixTimestamp || !svixSignature) {
    return new Response('Missing svix headers', { status: 400 });
  }

  const payload = await req.json();
  const body = JSON.stringify(payload);

  const wh = new Webhook(WEBHOOK_SECRET);
  let event: WebhookEvent;

  try {
    event = wh.verify(body, {
      'svix-id': svixId,
      'svix-timestamp': svixTimestamp,
      'svix-signature': svixSignature,
    }) as WebhookEvent;
  } catch (err) {
    console.error('Webhook verification failed:', err);
    return new Response('Invalid signature', { status: 400 });
  }

  switch (event.type) {
    case 'user.created':
      await handleUserCreated(event.data);
      break;

    case 'user.updated':
      await handleUserUpdated(event.data);
      break;

    case 'user.deleted':
      await handleUserDeleted(event.data);
      break;

    case 'organization.created':
      await handleOrganizationCreated(event.data);
      break;

    case 'organizationMembership.created':
      await handleMembershipCreated(event.data);
      break;

    case 'organizationMembership.deleted':
      await handleMembershipDeleted(event.data);
      break;
  }

  return new Response('Webhook processed', { status: 200 });
}

async function handleUserCreated(data: any) {
  await db.user.create({
    data: {
      clerkId: data.id,
      email: data.email_addresses[0]?.email_address,
      firstName: data.first_name,
      lastName: data.last_name,
      imageUrl: data.image_url,
    },
  });
}

async function handleUserUpdated(data: any) {
  await db.user.update({
    where: { clerkId: data.id },
    data: {
      email: data.email_addresses[0]?.email_address,
      firstName: data.first_name,
      lastName: data.last_name,
      imageUrl: data.image_url,
    },
  });
}

async function handleUserDeleted(data: any) {
  await db.user.delete({
    where: { clerkId: data.id },
  });
}

async function handleOrganizationCreated(data: any) {
  await db.organization.create({
    data: {
      clerkId: data.id,
      name: data.name,
      slug: data.slug,
      imageUrl: data.image_url,
    },
  });
}

async function handleMembershipCreated(data: any) {
  await db.membership.create({
    data: {
      clerkUserId: data.public_user_data.user_id,
      clerkOrgId: data.organization.id,
      role: data.role,
    },
  });
}

async function handleMembershipDeleted(data: any) {
  await db.membership.delete({
    where: {
      clerkUserId_clerkOrgId: {
        clerkUserId: data.public_user_data.user_id,
        clerkOrgId: data.organization.id,
      },
    },
  });
}
```

---

## Firebase Auth Production Examples

### Firebase Admin SDK Setup

Node.js Backend Integration:

```typescript
// lib/firebase-admin.ts
import { initializeApp, cert, getApps, App } from 'firebase-admin/app';
import { getAuth, Auth } from 'firebase-admin/auth';
import { getFirestore, Firestore } from 'firebase-admin/firestore';

let app: App;
let auth: Auth;
let db: Firestore;

function initializeFirebaseAdmin() {
  if (getApps().length === 0) {
    app = initializeApp({
      credential: cert({
        projectId: process.env.FIREBASE_PROJECT_ID,
        clientEmail: process.env.FIREBASE_CLIENT_EMAIL,
        privateKey: process.env.FIREBASE_PRIVATE_KEY?.replace(/\\n/g, '\n'),
      }),
    });
  } else {
    app = getApps()[0];
  }

  auth = getAuth(app);
  db = getFirestore(app);

  return { app, auth, db };
}

export const firebase = initializeFirebaseAdmin();

export class FirebaseAuthManager {
  async createUser(data: {
    email: string;
    password?: string;
    displayName?: string;
    phoneNumber?: string;
    emailVerified?: boolean;
  }) {
    return firebase.auth.createUser({
      email: data.email,
      password: data.password,
      displayName: data.displayName,
      phoneNumber: data.phoneNumber,
      emailVerified: data.emailVerified ?? false,
    });
  }

  async setCustomClaims(uid: string, claims: Record<string, any>) {
    await firebase.auth.setCustomUserClaims(uid, claims);
  }

  async verifyIdToken(idToken: string) {
    return firebase.auth.verifyIdToken(idToken);
  }

  async createCustomToken(uid: string, claims?: Record<string, any>) {
    return firebase.auth.createCustomToken(uid, claims);
  }

  async getUserByEmail(email: string) {
    return firebase.auth.getUserByEmail(email);
  }

  async revokeRefreshTokens(uid: string) {
    await firebase.auth.revokeRefreshTokens(uid);
  }

  async deleteUser(uid: string) {
    await firebase.auth.deleteUser(uid);
  }
}
```

### Firebase Cloud Functions Auth Triggers

Authentication Event Handlers:

```typescript
// functions/src/auth-triggers.ts
import * as functions from 'firebase-functions';
import * as admin from 'firebase-admin';

admin.initializeApp();
const db = admin.firestore();

export const onUserCreated = functions.auth.user().onCreate(async (user) => {
  const { uid, email, displayName, photoURL, emailVerified } = user;

  // Create user document in Firestore
  await db.collection('users').doc(uid).set({
    email,
    displayName: displayName || email?.split('@')[0],
    photoURL,
    emailVerified,
    createdAt: admin.firestore.FieldValue.serverTimestamp(),
    role: 'user',
    subscription: {
      tier: 'free',
      startedAt: admin.firestore.FieldValue.serverTimestamp(),
    },
  });

  // Set default custom claims
  await admin.auth().setCustomUserClaims(uid, {
    role: 'user',
    tier: 'free',
  });

  // Send welcome email
  await db.collection('mail').add({
    to: email,
    template: {
      name: 'welcome',
      data: {
        displayName: displayName || 'there',
      },
    },
  });

  functions.logger.info(`User created: ${uid}`);
});

export const onUserDeleted = functions.auth.user().onDelete(async (user) => {
  const { uid, email } = user;

  // Archive user data before deletion
  const userDoc = await db.collection('users').doc(uid).get();
  if (userDoc.exists) {
    await db.collection('archived_users').doc(uid).set({
      ...userDoc.data(),
      deletedAt: admin.firestore.FieldValue.serverTimestamp(),
    });
  }

  // Delete user data
  const batch = db.batch();
  batch.delete(db.collection('users').doc(uid));

  // Delete user's subcollections
  const collections = ['projects', 'settings', 'notifications'];
  for (const collection of collections) {
    const docs = await db.collection('users').doc(uid).collection(collection).listDocuments();
    for (const doc of docs) {
      batch.delete(doc);
    }
  }

  await batch.commit();

  functions.logger.info(`User deleted and archived: ${uid}`);
});

export const beforeUserCreated = functions.auth.user().beforeCreate(async (user) => {
  const { email } = user;

  // Block disposable email domains
  const disposableDomains = ['tempmail.com', 'throwaway.com', 'guerrillamail.com'];
  const domain = email?.split('@')[1];

  if (domain && disposableDomains.includes(domain)) {
    throw new functions.auth.HttpsError(
      'invalid-argument',
      'Disposable email addresses are not allowed'
    );
  }

  // Check invite-only mode
  const settings = await db.collection('settings').doc('registration').get();
  if (settings.exists && settings.data()?.inviteOnly) {
    const invitation = await db.collection('invitations')
      .where('email', '==', email)
      .where('status', '==', 'pending')
      .limit(1)
      .get();

    if (invitation.empty) {
      throw new functions.auth.HttpsError(
        'permission-denied',
        'Registration is by invitation only'
      );
    }
  }
});

export const beforeUserSignedIn = functions.auth.user().beforeSignIn(async (user) => {
  const { uid, email } = user;

  // Check if user is banned
  const userDoc = await db.collection('users').doc(uid).get();
  if (userDoc.exists && userDoc.data()?.banned) {
    throw new functions.auth.HttpsError(
      'permission-denied',
      'Your account has been suspended'
    );
  }

  // Update last sign-in timestamp
  await db.collection('users').doc(uid).update({
    lastSignInAt: admin.firestore.FieldValue.serverTimestamp(),
  });
});
```

### Firebase Security Rules

Firestore Rules with Authentication:

```javascript
// firestore.rules
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {

    // Helper functions
    function isAuthenticated() {
      return request.auth != null;
    }

    function isOwner(userId) {
      return request.auth.uid == userId;
    }

    function hasRole(role) {
      return isAuthenticated() &&
             request.auth.token.role == role;
    }

    function isAdmin() {
      return hasRole('admin');
    }

    function isMember(orgId) {
      return isAuthenticated() &&
             exists(/databases/$(database)/documents/organizations/$(orgId)/members/$(request.auth.uid));
    }

    function hasOrgRole(orgId, role) {
      return isMember(orgId) &&
             get(/databases/$(database)/documents/organizations/$(orgId)/members/$(request.auth.uid)).data.role == role;
    }

    // User profiles
    match /users/{userId} {
      allow read: if isAuthenticated() && (isOwner(userId) || isAdmin());
      allow create: if isOwner(userId);
      allow update: if isOwner(userId) || isAdmin();
      allow delete: if isAdmin();

      // User's private data
      match /private/{document=**} {
        allow read, write: if isOwner(userId);
      }

      // User's public profile
      match /public/{document=**} {
        allow read: if isAuthenticated();
        allow write: if isOwner(userId);
      }
    }

    // Organizations
    match /organizations/{orgId} {
      allow read: if isMember(orgId);
      allow create: if isAuthenticated();
      allow update: if hasOrgRole(orgId, 'admin');
      allow delete: if hasOrgRole(orgId, 'owner');

      // Organization members
      match /members/{memberId} {
        allow read: if isMember(orgId);
        allow create: if hasOrgRole(orgId, 'admin');
        allow update: if hasOrgRole(orgId, 'admin');
        allow delete: if hasOrgRole(orgId, 'admin') && memberId != request.auth.uid;
      }

      // Organization projects
      match /projects/{projectId} {
        allow read: if isMember(orgId);
        allow create: if isMember(orgId);
        allow update: if isMember(orgId);
        allow delete: if hasOrgRole(orgId, 'admin');
      }
    }

    // Public content
    match /public/{document=**} {
      allow read: if true;
      allow write: if isAdmin();
    }
  }
}
```

### Flutter Firebase Auth Integration

Complete Auth Service:

```dart
// lib/services/auth_service.dart
import 'package:firebase_auth/firebase_auth.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:cloud_firestore/cloud_firestore.dart';

class AuthService {
  final FirebaseAuth _auth = FirebaseAuth.instance;
  final GoogleSignIn _googleSignIn = GoogleSignIn();
  final FirebaseFirestore _firestore = FirebaseFirestore.instance;

  // Stream of auth state changes
  Stream<User?> get authStateChanges => _auth.authStateChanges();

  // Current user
  User? get currentUser => _auth.currentUser;

  // Email/Password Sign In
  Future<UserCredential> signInWithEmail(String email, String password) async {
    try {
      return await _auth.signInWithEmailAndPassword(
        email: email,
        password: password,
      );
    } on FirebaseAuthException catch (e) {
      throw _handleAuthException(e);
    }
  }

  // Email/Password Sign Up
  Future<UserCredential> signUpWithEmail(
    String email,
    String password,
    String displayName,
  ) async {
    try {
      final credential = await _auth.createUserWithEmailAndPassword(
        email: email,
        password: password,
      );

      // Update display name
      await credential.user?.updateDisplayName(displayName);

      // Create user document
      await _createUserDocument(credential.user!);

      return credential;
    } on FirebaseAuthException catch (e) {
      throw _handleAuthException(e);
    }
  }

  // Google Sign In
  Future<UserCredential> signInWithGoogle() async {
    try {
      final GoogleSignInAccount? googleUser = await _googleSignIn.signIn();
      if (googleUser == null) {
        throw Exception('Google sign in cancelled');
      }

      final GoogleSignInAuthentication googleAuth =
          await googleUser.authentication;

      final credential = GoogleAuthProvider.credential(
        accessToken: googleAuth.accessToken,
        idToken: googleAuth.idToken,
      );

      final userCredential = await _auth.signInWithCredential(credential);

      // Create or update user document
      await _createOrUpdateUserDocument(userCredential.user!);

      return userCredential;
    } catch (e) {
      throw Exception('Google sign in failed: $e');
    }
  }

  // Apple Sign In
  Future<UserCredential> signInWithApple() async {
    final appleProvider = AppleAuthProvider()
      ..addScope('email')
      ..addScope('name');

    try {
      final userCredential = await _auth.signInWithProvider(appleProvider);
      await _createOrUpdateUserDocument(userCredential.user!);
      return userCredential;
    } on FirebaseAuthException catch (e) {
      throw _handleAuthException(e);
    }
  }

  // Phone Authentication
  Future<void> verifyPhoneNumber({
    required String phoneNumber,
    required Function(PhoneAuthCredential) verificationCompleted,
    required Function(FirebaseAuthException) verificationFailed,
    required Function(String, int?) codeSent,
    required Function(String) codeAutoRetrievalTimeout,
  }) async {
    await _auth.verifyPhoneNumber(
      phoneNumber: phoneNumber,
      verificationCompleted: verificationCompleted,
      verificationFailed: verificationFailed,
      codeSent: codeSent,
      codeAutoRetrievalTimeout: codeAutoRetrievalTimeout,
    );
  }

  Future<UserCredential> signInWithPhoneCredential(
    String verificationId,
    String smsCode,
  ) async {
    final credential = PhoneAuthProvider.credential(
      verificationId: verificationId,
      smsCode: smsCode,
    );
    return await _auth.signInWithCredential(credential);
  }

  // Password Reset
  Future<void> sendPasswordResetEmail(String email) async {
    await _auth.sendPasswordResetEmail(email: email);
  }

  // Sign Out
  Future<void> signOut() async {
    await _googleSignIn.signOut();
    await _auth.signOut();
  }

  // Create user document
  Future<void> _createUserDocument(User user) async {
    await _firestore.collection('users').doc(user.uid).set({
      'email': user.email,
      'displayName': user.displayName,
      'photoURL': user.photoURL,
      'createdAt': FieldValue.serverTimestamp(),
      'lastSignInAt': FieldValue.serverTimestamp(),
    });
  }

  // Create or update user document
  Future<void> _createOrUpdateUserDocument(User user) async {
    final docRef = _firestore.collection('users').doc(user.uid);
    final doc = await docRef.get();

    if (doc.exists) {
      await docRef.update({
        'lastSignInAt': FieldValue.serverTimestamp(),
      });
    } else {
      await _createUserDocument(user);
    }
  }

  // Handle auth exceptions
  String _handleAuthException(FirebaseAuthException e) {
    switch (e.code) {
      case 'user-not-found':
        return 'No user found with this email.';
      case 'wrong-password':
        return 'Wrong password provided.';
      case 'email-already-in-use':
        return 'An account already exists with this email.';
      case 'weak-password':
        return 'The password is too weak.';
      case 'invalid-email':
        return 'Invalid email address.';
      case 'user-disabled':
        return 'This account has been disabled.';
      case 'too-many-requests':
        return 'Too many attempts. Please try again later.';
      default:
        return 'An error occurred: ${e.message}';
    }
  }
}
```

---

## Migration Examples

### Auth0 to Clerk User Migration

Complete Migration Script:

```python
# scripts/migrate_auth0_to_clerk.py
import asyncio
import json
from typing import List, Dict, Any
import httpx
from auth0.management import Auth0
from datetime import datetime

class Auth0ToClerkMigrator:
    def __init__(self):
        self.auth0 = Auth0(
            domain=os.getenv('AUTH0_DOMAIN'),
            token=os.getenv('AUTH0_MANAGEMENT_TOKEN')
        )
        self.clerk_api_url = 'https://api.clerk.com/v1'
        self.clerk_secret = os.getenv('CLERK_SECRET_KEY')
        self.migration_log = []

    async def export_auth0_users(self) -> List[Dict[str, Any]]:
        """Export all users from Auth0."""
        users = []
        page = 0
        per_page = 100

        while True:
            result = self.auth0.users.list(
                page=page,
                per_page=per_page,
                fields=['user_id', 'email', 'email_verified', 'name',
                        'given_name', 'family_name', 'picture',
                        'user_metadata', 'app_metadata', 'identities']
            )

            if not result['users']:
                break

            users.extend(result['users'])
            page += 1

            print(f"Exported {len(users)} users...")

        return users

    async def export_auth0_organizations(self) -> List[Dict[str, Any]]:
        """Export all organizations from Auth0."""
        organizations = []
        page = 0
        per_page = 100

        while True:
            result = self.auth0.organizations.all_organizations(
                page=page,
                per_page=per_page
            )

            if not result:
                break

            organizations.extend(result)
            page += 1

        # Get members for each organization
        for org in organizations:
            members = self.auth0.organizations.list_members(org['id'])
            org['members'] = members

        return organizations

    def transform_user(self, auth0_user: Dict[str, Any]) -> Dict[str, Any]:
        """Transform Auth0 user to Clerk format."""
        return {
            'email_address': [auth0_user['email']],
            'first_name': auth0_user.get('given_name', ''),
            'last_name': auth0_user.get('family_name', ''),
            'username': auth0_user.get('user_metadata', {}).get('username'),
            'profile_image_url': auth0_user.get('picture'),
            'public_metadata': auth0_user.get('user_metadata', {}),
            'private_metadata': auth0_user.get('app_metadata', {}),
            'external_id': auth0_user['user_id'],
            'skip_password_requirement': True,
        }

    async def import_user_to_clerk(
        self,
        user_data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Import a single user to Clerk."""
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f'{self.clerk_api_url}/users',
                headers={
                    'Authorization': f'Bearer {self.clerk_secret}',
                    'Content-Type': 'application/json'
                },
                json=user_data
            )

            if response.status_code != 200:
                raise Exception(f"Failed to create user: {response.text}")

            return response.json()

    async def create_clerk_organization(
        self,
        name: str,
        slug: str,
        created_by: str
    ) -> Dict[str, Any]:
        """Create organization in Clerk."""
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f'{self.clerk_api_url}/organizations',
                headers={
                    'Authorization': f'Bearer {self.clerk_secret}',
                    'Content-Type': 'application/json'
                },
                json={
                    'name': name,
                    'slug': slug,
                    'created_by': created_by
                }
            )

            if response.status_code != 200:
                raise Exception(f"Failed to create organization: {response.text}")

            return response.json()

    async def add_member_to_clerk_organization(
        self,
        org_id: str,
        user_id: str,
        role: str = 'org:member'
    ):
        """Add member to Clerk organization."""
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f'{self.clerk_api_url}/organizations/{org_id}/memberships',
                headers={
                    'Authorization': f'Bearer {self.clerk_secret}',
                    'Content-Type': 'application/json'
                },
                json={
                    'user_id': user_id,
                    'role': role
                }
            )

            if response.status_code != 200:
                raise Exception(f"Failed to add member: {response.text}")

    async def run_migration(self):
        """Execute complete migration."""
        print("Starting Auth0 to Clerk migration...")

        # Export data
        print("\n1. Exporting Auth0 users...")
        auth0_users = await self.export_auth0_users()
        print(f"   Found {len(auth0_users)} users")

        print("\n2. Exporting Auth0 organizations...")
        auth0_orgs = await self.export_auth0_organizations()
        print(f"   Found {len(auth0_orgs)} organizations")

        # User ID mapping (Auth0 -> Clerk)
        user_id_map = {}

        # Import users
        print("\n3. Importing users to Clerk...")
        for i, auth0_user in enumerate(auth0_users):
            try:
                clerk_user_data = self.transform_user(auth0_user)
                clerk_user = await self.import_user_to_clerk(clerk_user_data)
                user_id_map[auth0_user['user_id']] = clerk_user['id']

                self.migration_log.append({
                    'type': 'user',
                    'auth0_id': auth0_user['user_id'],
                    'clerk_id': clerk_user['id'],
                    'status': 'success'
                })

                print(f"   Imported user {i+1}/{len(auth0_users)}: {auth0_user['email']}")

            except Exception as e:
                self.migration_log.append({
                    'type': 'user',
                    'auth0_id': auth0_user['user_id'],
                    'error': str(e),
                    'status': 'failed'
                })
                print(f"   Failed to import user: {auth0_user['email']} - {e}")

        # Import organizations
        print("\n4. Importing organizations to Clerk...")
        org_id_map = {}

        for org in auth0_orgs:
            try:
                # Find an admin to be the creator
                admin_id = None
                for member in org.get('members', []):
                    if 'Admin' in member.get('roles', []):
                        admin_id = user_id_map.get(member['user_id'])
                        break

                if not admin_id and org.get('members'):
                    admin_id = user_id_map.get(org['members'][0]['user_id'])

                if not admin_id:
                    print(f"   Skipping org {org['name']}: no valid creator found")
                    continue

                clerk_org = await self.create_clerk_organization(
                    name=org['display_name'],
                    slug=org['name'],
                    created_by=admin_id
                )
                org_id_map[org['id']] = clerk_org['id']

                # Add members
                for member in org.get('members', []):
                    clerk_user_id = user_id_map.get(member['user_id'])
                    if clerk_user_id:
                        role = 'org:admin' if 'Admin' in member.get('roles', []) else 'org:member'
                        await self.add_member_to_clerk_organization(
                            clerk_org['id'],
                            clerk_user_id,
                            role
                        )

                self.migration_log.append({
                    'type': 'organization',
                    'auth0_id': org['id'],
                    'clerk_id': clerk_org['id'],
                    'status': 'success'
                })

                print(f"   Imported org: {org['display_name']}")

            except Exception as e:
                self.migration_log.append({
                    'type': 'organization',
                    'auth0_id': org['id'],
                    'error': str(e),
                    'status': 'failed'
                })
                print(f"   Failed to import org: {org['name']} - {e}")

        # Save migration log
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        with open(f'migration_log_{timestamp}.json', 'w') as f:
            json.dump(self.migration_log, f, indent=2)

        # Summary
        successful_users = sum(1 for log in self.migration_log
                               if log['type'] == 'user' and log['status'] == 'success')
        failed_users = sum(1 for log in self.migration_log
                          if log['type'] == 'user' and log['status'] == 'failed')
        successful_orgs = sum(1 for log in self.migration_log
                              if log['type'] == 'organization' and log['status'] == 'success')
        failed_orgs = sum(1 for log in self.migration_log
                         if log['type'] == 'organization' and log['status'] == 'failed')

        print(f"\n=== Migration Complete ===")
        print(f"Users: {successful_users} success, {failed_users} failed")
        print(f"Organizations: {successful_orgs} success, {failed_orgs} failed")
        print(f"Log saved to: migration_log_{timestamp}.json")

if __name__ == '__main__':
    import os
    from dotenv import load_dotenv
    load_dotenv()

    migrator = Auth0ToClerkMigrator()
    asyncio.run(migrator.run_migration())
```

---

## Security Best Practices Examples

### Token Validation Middleware

Express.js with Multiple Providers:

```typescript
// middleware/auth.ts
import { Request, Response, NextFunction } from 'express';
import jwt from 'jsonwebtoken';
import jwksRsa from 'jwks-rsa';
import { promisify } from 'util';

interface AuthConfig {
  provider: 'auth0' | 'clerk' | 'firebase';
  domain?: string;
  audience?: string;
  issuer?: string;
}

const createJwksClient = (config: AuthConfig) => {
  switch (config.provider) {
    case 'auth0':
      return jwksRsa({
        jwksUri: `https://${config.domain}/.well-known/jwks.json`,
        cache: true,
        cacheMaxEntries: 5,
        cacheMaxAge: 600000,
      });
    case 'clerk':
      return jwksRsa({
        jwksUri: `https://${config.domain}/.well-known/jwks.json`,
        cache: true,
      });
    case 'firebase':
      return jwksRsa({
        jwksUri: 'https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com',
        cache: true,
      });
  }
};

export const createAuthMiddleware = (config: AuthConfig) => {
  const client = createJwksClient(config);
  const getKey = promisify(client.getSigningKey);

  return async (req: Request, res: Response, next: NextFunction) => {
    const authHeader = req.headers.authorization;

    if (!authHeader?.startsWith('Bearer ')) {
      return res.status(401).json({ error: 'Missing authorization header' });
    }

    const token = authHeader.split(' ')[1];

    try {
      const decoded = jwt.decode(token, { complete: true });
      if (!decoded || !decoded.header.kid) {
        return res.status(401).json({ error: 'Invalid token format' });
      }

      const key = await getKey(decoded.header.kid);
      const publicKey = key.getPublicKey();

      const verified = jwt.verify(token, publicKey, {
        algorithms: ['RS256'],
        audience: config.audience,
        issuer: config.issuer,
      });

      req.user = verified as any;
      next();
    } catch (error) {
      console.error('Token verification failed:', error);
      return res.status(401).json({ error: 'Invalid token' });
    }
  };
};

// Usage
const auth0Middleware = createAuthMiddleware({
  provider: 'auth0',
  domain: process.env.AUTH0_DOMAIN,
  audience: process.env.AUTH0_AUDIENCE,
  issuer: `https://${process.env.AUTH0_DOMAIN}/`,
});

const clerkMiddleware = createAuthMiddleware({
  provider: 'clerk',
  domain: process.env.CLERK_DOMAIN,
  issuer: process.env.CLERK_ISSUER,
});
```

---

Last Updated: 2025-12-07
Examples: Production-ready implementations for Auth0, Clerk, Firebase Auth
Status: Production Ready
