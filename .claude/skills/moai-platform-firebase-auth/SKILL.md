---
name: "moai-platform-firebase-auth"
description: "Firebase Authentication specialist covering Google ecosystem, social auth, phone auth, and mobile-first patterns. Use when building Firebase-backed or Google ecosystem apps, implementing social login, or adding phone verification."
version: 2.0.0
category: "platform"
modularized: true
user-invocable: false
tags: ['firebase, google, social-auth, mobile, authentication']
updated: 2026-01-08
status: "active"
context7-libraries: "/firebase/firebase-docs"
related-skills: "moai-platform-firestore, moai-lang-flutter, moai-lang-typescript"
allowed-tools:
  - Read
  - Write
  - Bash
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Firebase Authentication Specialist

Comprehensive Firebase Authentication implementation covering Google ecosystem integration, social authentication providers, phone authentication, anonymous auth, custom claims, and Security Rules integration.

---

## Quick Reference

Firebase Auth Core Features:
- Google Sign-In: Native Google ecosystem integration with Cloud Identity
- Social Auth: Facebook, Twitter/X, GitHub, Apple, Microsoft, Yahoo
- Phone Auth: SMS-based verification with international support
- Anonymous Auth: Guest access with account linking upgrade path
- Custom Claims: Role-based access and admin privileges
- Security Rules: Firestore, Storage, and Realtime Database integration

Context7 Documentation Access:
- Use resolve-library-id with "firebase" to get Context7 library ID
- Use get-library-docs with resolved ID and topic "authentication" for latest API

Platform SDK Support:
- Web: firebase/auth with modular SDK (v9+)
- iOS: FirebaseAuth with Swift and SwiftUI
- Android: firebase-auth with Kotlin
- Flutter: firebase_auth package
- React Native: @react-native-firebase/auth

Quick Decision Tree:
- Need Google ecosystem integration? Use Firebase Auth
- Building mobile-first application? Use Firebase Auth
- Need serverless Cloud Functions? Use Firebase Auth
- Need anonymous guest access? Use Firebase Auth
- Existing Firebase infrastructure? Use Firebase Auth

---

## Module Index

This skill uses progressive disclosure with specialized modules for detailed implementation guidance.

### Social Authentication
File: modules/social-auth.md

Covers Google Sign-In, Facebook Login, Apple Sign-In, Twitter/X, GitHub, and Microsoft authentication. Includes web, Flutter, iOS, and Android implementations with configuration requirements and error handling.

Key Topics:
- Google Sign-In with OAuth scopes and custom parameters
- Facebook Login with Graph API access
- Apple Sign-In (required for iOS apps with third-party login)
- Account linking between providers
- Error handling for popup and redirect flows

### Phone Authentication
File: modules/phone-auth.md

Covers SMS-based phone number authentication with international support, reCAPTCHA verification, and platform-specific implementations.

Key Topics:
- Web implementation with RecaptchaVerifier
- Flutter verifyPhoneNumber flow with auto-verification
- iOS and Android native implementations
- E.164 phone number formatting
- Error handling and rate limiting

### Custom Claims and Role Management
File: modules/custom-claims.md

Covers custom claims, role-based access control (RBAC), and admin privileges with Security Rules integration.

Key Topics:
- Setting claims with Admin SDK (Node.js, Python, Go)
- Cloud Functions for claim management
- Reading claims on client (Web, Flutter, iOS)
- Security Rules integration for Firestore and Storage
- Token refresh strategies

---

## Implementation Guide

### Firebase Project Setup

Step 1: Create Firebase project at console.firebase.google.com
Step 2: Add your app (Web, iOS, Android, or Flutter)
Step 3: Enable desired authentication providers
Step 4: Download and configure SDK credentials

### Web SDK Initialization

```typescript
import { initializeApp } from 'firebase/app';
import { getAuth, onAuthStateChanged } from 'firebase/auth';

const firebaseConfig = {
  apiKey: "your-api-key",
  authDomain: "your-project.firebaseapp.com",
  projectId: "your-project-id"
};

const app = initializeApp(firebaseConfig);
const auth = getAuth(app);

onAuthStateChanged(auth, (user) => {
  if (user) {
    console.log('Signed in:', user.uid);
  } else {
    console.log('Signed out');
  }
});
```

### Flutter SDK Initialization

```dart
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_auth/firebase_auth.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Firebase.initializeApp();
  runApp(MyApp());
}

// Auth state listener
FirebaseAuth.instance.authStateChanges().listen((User? user) {
  if (user != null) {
    print('Signed in: ${user.uid}');
  }
});
```

### Anonymous Authentication

```typescript
import { signInAnonymously, linkWithCredential, EmailAuthProvider } from 'firebase/auth';

// Sign in anonymously
const result = await signInAnonymously(auth);
console.log('Anonymous UID:', result.user.uid);

// Upgrade to permanent account
const credential = EmailAuthProvider.credential(email, password);
await linkWithCredential(auth.currentUser, credential);
```

### Session Persistence

```typescript
import { setPersistence, browserLocalPersistence, browserSessionPersistence } from 'firebase/auth';

// Persist across browser sessions (default)
await setPersistence(auth, browserLocalPersistence);

// Clear on tab close
await setPersistence(auth, browserSessionPersistence);
```

---

## Security Rules Integration

### Firestore Rules

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    match /users/{userId} {
      allow read, write: if request.auth != null && request.auth.uid == userId;
    }
    match /admin/{document=**} {
      allow read, write: if request.auth.token.admin == true;
    }
  }
}
```

### Storage Rules

```javascript
rules_version = '2';
service firebase.storage {
  match /b/{bucket}/o {
    match /users/{userId}/{allPaths=**} {
      allow read, write: if request.auth != null && request.auth.uid == userId;
    }
  }
}
```

---

## Advanced Patterns

For advanced implementation patterns, see:
- reference.md: MFA, session cookies, Cloud Functions triggers, Admin SDK setup
- examples.md: Complete authentication services, React hooks, Flutter providers

---

## Resources

Extended Documentation:
- reference.md: Advanced patterns and configuration guides
- examples.md: Working code examples across platforms

Module Files:
- modules/social-auth.md: Social identity providers
- modules/phone-auth.md: Phone number authentication
- modules/custom-claims.md: Role-based access control

Firebase Official Resources:
- Firebase Console: console.firebase.google.com
- Authentication Documentation: firebase.google.com/docs/auth
- Security Rules Reference: firebase.google.com/docs/rules

Works Well With:
- moai-platform-firestore: Firestore database with auth-based security
- moai-lang-flutter: Flutter SDK for mobile Firebase Auth
- moai-lang-typescript: TypeScript patterns for Firebase SDK
- moai-domain-backend: Backend architecture with Firebase Admin SDK

---

Status: Production Ready
Version: 2.0.0 (Modular Architecture)
Last Updated: 2025-12-07
Provider Coverage: Firebase Authentication Only
