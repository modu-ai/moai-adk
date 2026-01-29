# Firestore Security Rules Module

## Overview

Firestore Security Rules provide declarative, server-side access control for your database. Rules are evaluated on every read and write operation and determine whether the operation should be allowed or denied.

---

## Basic Rule Structure

### Rules Version and Service Declaration

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Rules go here
  }
}
```

Always use `rules_version = '2'` for access to all current features including recursive wildcards.

### Match Statements

Single Document Match:
```javascript
match /users/{userId} {
  allow read, write: if request.auth.uid == userId;
}
```

Subcollection Match:
```javascript
match /organizations/{orgId}/projects/{projectId} {
  allow read: if isOrgMember(orgId);
}
```

Recursive Wildcard (match all subcollections):
```javascript
match /organizations/{orgId}/{document=**} {
  allow read: if isOrgMember(orgId);
}
```

---

## Authentication Checks

### Basic Authentication

```javascript
// Require any authenticated user
allow read: if request.auth != null;

// Require specific user
allow write: if request.auth.uid == userId;

// Check email verification
allow write: if request.auth.token.email_verified == true;
```

### Email Domain Restriction

```javascript
function isCompanyEmail() {
  return request.auth.token.email.matches('.*@company\\.com$');
}

match /internal/{document=**} {
  allow read, write: if isCompanyEmail();
}
```

---

## Custom Claims for Role-Based Access

### Setting Custom Claims (Admin SDK)

```typescript
import { getAuth } from 'firebase-admin/auth'

await getAuth().setCustomUserClaims(userId, {
  admin: true,
  organizationId: 'org-123',
  role: 'editor'
})
```

### Reading Custom Claims in Rules

```javascript
function isAdmin() {
  return request.auth.token.admin == true;
}

function hasRole(role) {
  return request.auth.token.role == role;
}

function belongsToOrg(orgId) {
  return request.auth.token.organizationId == orgId;
}

match /admin/{document=**} {
  allow read, write: if isAdmin();
}

match /organizations/{orgId}/{document=**} {
  allow read: if belongsToOrg(orgId);
  allow write: if belongsToOrg(orgId) && hasRole('admin');
}
```

---

## Document-Based Access Control

### Owner-Based Access

```javascript
match /documents/{docId} {
  allow read, write: if request.auth.uid == resource.data.ownerId;
  allow read: if resource.data.isPublic == true;
  allow read: if request.auth.uid in resource.data.collaborators;
}
```

### Cross-Document Validation with get()

```javascript
function getMemberRole(orgId) {
  return get(/databases/$(database)/documents/organizations/$(orgId)/members/$(request.auth.uid)).data.role;
}

function isMember(orgId) {
  return exists(/databases/$(database)/documents/organizations/$(orgId)/members/$(request.auth.uid));
}

match /organizations/{orgId}/projects/{projectId} {
  allow read: if isMember(orgId);
  allow create: if isMember(orgId) && getMemberRole(orgId) in ['admin', 'editor'];
  allow update, delete: if getMemberRole(orgId) == 'admin';
}
```

Each `get()` or `exists()` call counts as one document read toward the 10-call limit per rule evaluation.

---

## Data Validation

### Field Type Validation

```javascript
function isValidDocument() {
  let data = request.resource.data;
  return data.title is string
    && data.title.size() > 0
    && data.title.size() <= 200
    && data.createdAt is timestamp
    && data.ownerId is string
    && (data.tags is list || !('tags' in data));
}

match /documents/{docId} {
  allow create: if request.auth != null && isValidDocument();
}
```

### Required Fields Validation

```javascript
function hasRequiredFields() {
  let required = ['title', 'ownerId', 'createdAt'];
  return request.resource.data.keys().hasAll(required);
}

function noExtraFields() {
  let allowed = ['title', 'description', 'ownerId', 'createdAt', 'updatedAt', 'tags'];
  return request.resource.data.keys().hasOnly(allowed);
}

match /documents/{docId} {
  allow create: if hasRequiredFields() && noExtraFields();
}
```

### Immutable Fields

```javascript
function immutableFieldsUnchanged() {
  return request.resource.data.ownerId == resource.data.ownerId
    && request.resource.data.createdAt == resource.data.createdAt;
}

match /documents/{docId} {
  allow update: if request.auth.uid == resource.data.ownerId
    && immutableFieldsUnchanged();
}
```

---

## Time-Based Rules

### Rate Limiting

```javascript
function notRateLimited() {
  return request.time > resource.data.lastWriteAt + duration.value(1, 's');
}

match /rateLimited/{docId} {
  allow update: if notRateLimited();
}
```

### Scheduled Content

```javascript
function isPublished() {
  return resource.data.publishAt <= request.time;
}

match /articles/{articleId} {
  allow read: if isPublished() || request.auth.uid == resource.data.authorId;
}
```

---

## Advanced Patterns

### Hierarchical Permissions

```javascript
function getOrgRole(orgId) {
  return get(/databases/$(database)/documents/organizations/$(orgId)/members/$(request.auth.uid)).data.role;
}

function canRead(orgId) {
  return getOrgRole(orgId) in ['owner', 'admin', 'editor', 'viewer'];
}

function canWrite(orgId) {
  return getOrgRole(orgId) in ['owner', 'admin', 'editor'];
}

function canAdmin(orgId) {
  return getOrgRole(orgId) in ['owner', 'admin'];
}

match /organizations/{orgId} {
  allow read: if canRead(orgId);
  allow update: if canAdmin(orgId);

  match /members/{memberId} {
    allow read: if canRead(orgId);
    allow create, update: if canAdmin(orgId);
  }

  match /projects/{projectId} {
    allow read: if canRead(orgId);
    allow create: if canWrite(orgId);
    allow update: if canWrite(orgId) || request.auth.uid == resource.data.createdBy;
    allow delete: if canAdmin(orgId);
  }
}
```

---

## Testing Security Rules

### Firebase Emulator Testing

```typescript
import { initializeTestEnvironment, assertSucceeds, assertFails } from '@firebase/rules-unit-testing'
import { doc, getDoc, setDoc } from 'firebase/firestore'

describe('Firestore Security Rules', () => {
  let testEnv

  beforeAll(async () => {
    testEnv = await initializeTestEnvironment({
      projectId: 'test-project',
      firestore: {
        rules: fs.readFileSync('firestore.rules', 'utf8')
      }
    })
  })

  test('users can read their own document', async () => {
    const userId = 'user-123'
    const context = testEnv.authenticatedContext(userId)
    const db = context.firestore()

    await testEnv.withSecurityRulesDisabled(async (adminContext) => {
      await setDoc(doc(adminContext.firestore(), 'users', userId), { name: 'Test' })
    })

    await assertSucceeds(getDoc(doc(db, 'users', userId)))
  })

  test('users cannot read other user documents', async () => {
    const context = testEnv.authenticatedContext('user-123')
    const db = context.firestore()

    await assertFails(getDoc(doc(db, 'users', 'other-user')))
  })
})
```

---

## Common Pitfalls

### Avoid Broad Wildcards Without Conditions

```javascript
// BAD: Too permissive
match /{document=**} {
  allow read, write: if request.auth != null;
}

// GOOD: Specific rules per collection
match /users/{userId} {
  allow read, write: if request.auth.uid == userId;
}
```

### Remember get() and exists() Limits

Each rule evaluation allows maximum 10 get() or exists() calls. Consolidate checks where possible.

### Validate Both Create and Update

```javascript
// GOOD: Validate both
match /documents/{docId} {
  allow create: if isValidDocument() && validOwner();
  allow update: if isValidDocument() && immutableFieldsUnchanged();
}
```

---

Status: Module Documentation
Last Updated: 2026-01-06
