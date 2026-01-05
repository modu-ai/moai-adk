# Custom Claims and Role Management Module

Comprehensive guide for implementing custom claims, role-based access control (RBAC), and admin privileges with Firebase Authentication.

---

## Overview

Custom claims are key-value pairs attached to Firebase ID tokens that enable role-based access control. Claims are set server-side using the Admin SDK and propagate to clients through token refresh.

Key Characteristics:
- Claims are embedded in ID tokens
- Maximum 1000 bytes total claim size
- Require token refresh to propagate changes
- Used in Security Rules for access control

---

## Setting Custom Claims (Admin SDK)

### Node.js Admin SDK

```typescript
import { getAuth } from 'firebase-admin/auth';

// Set single claim
await getAuth().setCustomUserClaims(uid, { admin: true });

// Set multiple claims
await getAuth().setCustomUserClaims(uid, {
  role: 'editor',
  organizationId: 'org_123',
  permissions: ['read', 'write', 'delete'],
  tier: 'premium'
});

// Add claims without overwriting
const user = await getAuth().getUser(uid);
const currentClaims = user.customClaims || {};
await getAuth().setCustomUserClaims(uid, {
  ...currentClaims,
  newClaim: 'value'
});

// Clear all claims
await getAuth().setCustomUserClaims(uid, null);
```

### Python Admin SDK

```python
from firebase_admin import auth

auth.set_custom_user_claims(uid, {'admin': True})

auth.set_custom_user_claims(uid, {
    'role': 'editor',
    'organizationId': 'org_123',
    'permissions': ['read', 'write']
})
```

---

## Cloud Functions for Claim Management

### HTTP Callable Function

```typescript
import { onCall, HttpsError } from 'firebase-functions/v2/https';
import { getAuth } from 'firebase-admin/auth';

export const setUserRole = onCall(async (request) => {
  if (!request.auth?.token.admin) {
    throw new HttpsError('permission-denied', 'Only admins can set roles');
  }

  const { targetUid, role } = request.data;

  const validRoles = ['viewer', 'editor', 'admin'];
  if (!validRoles.includes(role)) {
    throw new HttpsError('invalid-argument', 'Invalid role');
  }

  await getAuth().setCustomUserClaims(targetUid, { role });
  return { success: true };
});
```

### Auth Trigger: Set Default Claims

```typescript
import { auth } from 'firebase-functions';
import { getAuth } from 'firebase-admin/auth';

export const onUserCreate = auth.user().onCreate(async (user) => {
  const defaultClaims = {
    role: 'member',
    tier: 'free',
    createdAt: Date.now()
  };

  await getAuth().setCustomUserClaims(user.uid, defaultClaims);
});
```

### Blocking Function: Validate Before Creation

```typescript
import { beforeUserCreated } from 'firebase-functions/v2/identity';
import { HttpsError } from 'firebase-functions/v2/https';

export const validateNewUser = beforeUserCreated((event) => {
  const user = event.data;

  if (!user.email?.endsWith('@company.com')) {
    throw new HttpsError('permission-denied', 'Unauthorized email domain');
  }

  return {
    customClaims: { role: 'employee', organizationId: 'company_123' }
  };
});
```

---

## Reading Claims on Client

### Web/JavaScript

```typescript
import { getAuth, onIdTokenChanged } from 'firebase/auth';

const auth = getAuth();

const getClaimsFromToken = async () => {
  const user = auth.currentUser;
  if (!user) return null;

  const idTokenResult = await user.getIdTokenResult();
  return idTokenResult.claims;
};

const isAdmin = async (): Promise<boolean> => {
  const claims = await getClaimsFromToken();
  return claims?.admin === true;
};

// Listen for token changes
onIdTokenChanged(auth, async (user) => {
  if (user) {
    const idTokenResult = await user.getIdTokenResult();
    updateUIForRole(idTokenResult.claims.role);
  }
});

// Force token refresh after claim update
const refreshClaims = async () => {
  await auth.currentUser?.getIdToken(true);
  return await auth.currentUser?.getIdTokenResult();
};
```

### Flutter

```dart
import 'package:firebase_auth/firebase_auth.dart';

class ClaimsService {
  final FirebaseAuth _auth = FirebaseAuth.instance;

  Future<Map<String, dynamic>?> getClaims() async {
    final user = _auth.currentUser;
    if (user == null) return null;

    final idTokenResult = await user.getIdTokenResult();
    return idTokenResult.claims;
  }

  Future<bool> isAdmin() async {
    final claims = await getClaims();
    return claims?['admin'] == true;
  }

  Future<void> forceRefreshClaims() async {
    await _auth.currentUser?.getIdToken(true);
  }
}
```

---

## Security Rules Integration

### Firestore Security Rules

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {

    function isAuthenticated() {
      return request.auth != null;
    }

    function isAdmin() {
      return isAuthenticated() && request.auth.token.admin == true;
    }

    function hasRole(role) {
      return isAuthenticated() && request.auth.token.role == role;
    }

    function belongsToOrg(orgId) {
      return isAuthenticated() && request.auth.token.organizationId == orgId;
    }

    match /users/{userId} {
      allow read: if isAuthenticated() &&
                    (request.auth.uid == userId || isAdmin());
      allow write: if isAuthenticated() && request.auth.uid == userId;
      allow delete: if isAdmin();
    }

    match /adminData/{document=**} {
      allow read, write: if isAdmin();
    }

    match /content/{docId} {
      allow read: if isAuthenticated();
      allow create, update: if hasRole('editor') || hasRole('admin');
      allow delete: if hasRole('admin');
    }

    match /organizations/{orgId}/documents/{docId} {
      allow read, write: if belongsToOrg(orgId);
    }
  }
}
```

### Cloud Storage Security Rules

```javascript
rules_version = '2';
service firebase.storage {
  match /b/{bucket}/o {

    function isAdmin() {
      return request.auth != null && request.auth.token.admin == true;
    }

    match /users/{userId}/{allPaths=**} {
      allow read, write: if request.auth != null && request.auth.uid == userId;
    }

    match /public/{allPaths=**} {
      allow read: if true;
      allow write: if isAdmin();
    }
  }
}
```

---

## Role Management Patterns

### Hierarchical Role System

```typescript
const roleHierarchy: Record<string, number> = {
  viewer: 1,
  member: 2,
  editor: 3,
  manager: 4,
  admin: 5
};

const hasMinimumRole = async (requiredRole: string): Promise<boolean> => {
  const claims = await getClaimsFromToken();
  const userRole = claims?.role as string;

  if (!userRole || !roleHierarchy[userRole]) return false;
  if (!roleHierarchy[requiredRole]) return false;

  return roleHierarchy[userRole] >= roleHierarchy[requiredRole];
};
```

### Permission-Based System

```typescript
export const updateUserPermissions = onCall(async (request) => {
  if (!request.auth?.token.admin) {
    throw new HttpsError('permission-denied', 'Admin required');
  }

  const { targetUid, permissions } = request.data;
  const validPermissions = ['read', 'write', 'delete', 'manage'];
  const invalid = permissions.filter((p: string) => !validPermissions.includes(p));

  if (invalid.length > 0) {
    throw new HttpsError('invalid-argument', `Invalid: ${invalid.join(', ')}`);
  }

  const user = await getAuth().getUser(targetUid);
  await getAuth().setCustomUserClaims(targetUid, {
    ...user.customClaims,
    permissions
  });

  return { success: true };
});
```

---

## Token Refresh Strategies

### Proactive Token Refresh

```typescript
// Send FCM to trigger client refresh
export const notifyClaimsUpdate = onCall(async (request) => {
  const { targetUid } = request.data;

  await getMessaging().send({
    token: await getUserFcmToken(targetUid),
    data: { type: 'claims_updated' }
  });

  return { success: true };
});

// Client: Listen for refresh notification
onMessage(getMessaging(), async (payload) => {
  if (payload.data?.type === 'claims_updated') {
    await auth.currentUser?.getIdToken(true);
  }
});
```

---

## Best Practices

Claim Size Limits:
- Total custom claims cannot exceed 1000 bytes
- Use short key names to save space
- Store detailed data in Firestore, use claims for access control

Security Guidelines:
- Never trust client-provided claims
- Use Security Rules to enforce claim-based access control
- Audit claim changes for security-sensitive roles
- Implement proper admin verification before setting claims

Performance Considerations:
- Claims are included in every ID token
- Minimize claim count and size for better performance
- Use Firestore for complex permission queries

Token Propagation:
- Claims only propagate on token refresh (max 1 hour delay)
- Force refresh after critical claim changes
- Use FCM to notify clients of updates

---

Version: 1.0.0
Last Updated: 2025-12-07
Parent Skill: moai-platform-firebase-auth
