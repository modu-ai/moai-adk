---
name: moai-platform-firebase-auth
description: Firebase Authentication specialist covering Google ecosystem, social auth, phone auth, and mobile-first patterns. Use when building Firebase-backed or Google ecosystem apps.
version: 1.0.0
category: platform
updated: 2025-12-07
status: active
tags: firebase, google, social-auth, mobile, authentication
context7-libraries: /firebase/firebase-docs
related-skills: moai-platform-firestore, moai-lang-flutter
allowed-tools: Read, Write, Bash, Grep, Glob
---

# Firebase Authentication Specialist

Comprehensive Firebase Authentication implementation covering Google ecosystem integration, social authentication providers, phone authentication, anonymous auth, custom claims, and Security Rules integration.

## Quick Reference (30 seconds)

Firebase Auth Core Features:

- Google Sign-In: Native Google ecosystem integration with Cloud Identity
- Social Auth: Facebook, Twitter/X, GitHub, Apple, Microsoft, Yahoo
- Phone Auth: SMS-based verification with international support
- Anonymous Auth: Guest access with account linking
- Custom Claims: Role-based access and admin privileges
- Security Rules: Firestore, Storage, and Realtime Database integration

Context7 Library Access:

- Firebase Documentation: /firebase/firebase-docs
- Use resolve-library-id with "firebase" then get-library-docs for latest API

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

## Implementation Guide

### Google Sign-In Integration

Google Sign-In provides seamless authentication within the Google ecosystem with access to Google APIs and services.

Web Implementation:

Step 1: Enable Google Sign-In provider in Firebase Console under Authentication then Sign-in method
Step 2: Configure OAuth consent screen in Google Cloud Console
Step 3: Import and configure Firebase Auth SDK

```typescript
import { getAuth, signInWithPopup, GoogleAuthProvider } from 'firebase/auth';
const auth = getAuth();
const provider = new GoogleAuthProvider();
provider.addScope('https://www.googleapis.com/auth/calendar.readonly');
const result = await signInWithPopup(auth, provider);
const credential = GoogleAuthProvider.credentialFromResult(result);
```

Flutter Implementation:

```dart
Future<UserCredential> signInWithGoogle() async {
  final googleUser = await GoogleSignIn().signIn();
  final googleAuth = await googleUser?.authentication;
  final credential = GoogleAuthProvider.credential(
    accessToken: googleAuth?.accessToken, idToken: googleAuth?.idToken);
  return await FirebaseAuth.instance.signInWithCredential(credential);
}
```

Mobile Configuration Requirements:

- iOS: Configure URL schemes in Info.plist with reversed client ID
- Android: Add SHA-1 fingerprint to Firebase project settings
- Web: Configure authorized domains in Firebase Console

### Social Authentication Providers

Firebase Auth supports major social identity providers with unified API.

Facebook Login:

```typescript
import { FacebookAuthProvider, signInWithPopup } from 'firebase/auth';
const provider = new FacebookAuthProvider();
provider.addScope('email');
provider.addScope('public_profile');
const result = await signInWithPopup(auth, provider);
```

Apple Sign-In (Required for iOS apps with third-party login):

```swift
let provider = OAuthProvider(providerID: "apple.com")
provider.scopes = ["email", "fullName"]
provider.getCredentialWith(nil) { credential, error in
    if let credential = credential {
        Auth.auth().signIn(with: credential) { result, error in }
    }
}
```

Twitter/X and GitHub Authentication:

```typescript
// Twitter
const twitterProvider = new TwitterAuthProvider();
await signInWithPopup(auth, twitterProvider);

// GitHub
const githubProvider = new GithubAuthProvider();
githubProvider.addScope('repo');
await signInWithPopup(auth, githubProvider);
```

### Phone Number Authentication

SMS-based phone authentication with international support and reCAPTCHA verification.

Web Implementation:

```typescript
import { RecaptchaVerifier, signInWithPhoneNumber } from 'firebase/auth';
const recaptchaVerifier = new RecaptchaVerifier(auth, 'recaptcha-container', {
  size: 'normal', callback: () => { /* reCAPTCHA solved */ }
});
const confirmationResult = await signInWithPhoneNumber(auth, '+1234567890', recaptchaVerifier);
const credential = await confirmationResult.confirm(verificationCode);
```

Flutter Implementation:

```dart
await FirebaseAuth.instance.verifyPhoneNumber(
  phoneNumber: '+1234567890',
  verificationCompleted: (credential) async {
    await FirebaseAuth.instance.signInWithCredential(credential);
  },
  verificationFailed: (e) => print('Failed: ${e.message}'),
  codeSent: (verificationId, resendToken) { /* Store verificationId */ },
  codeAutoRetrievalTimeout: (verificationId) {},
);
```

Phone Auth Best Practices:

- Use E.164 format for phone numbers
- Handle auto-verification on Android
- Provide manual code entry fallback
- Consider SMS costs and rate limits

### Anonymous Authentication

Anonymous auth enables guest access with account upgrade path.

```typescript
// Anonymous sign-in
const result = await signInAnonymously(auth);
console.log('Anonymous UID:', result.user.uid);

// Account linking (upgrade to permanent)
import { linkWithCredential, EmailAuthProvider } from 'firebase/auth';
const credential = EmailAuthProvider.credential(email, password);
const linked = await linkWithCredential(auth.currentUser, credential);

// Link with social provider
const googleProvider = new GoogleAuthProvider();
await linkWithPopup(auth.currentUser, googleProvider);
```

Anonymous Auth Use Cases:

- E-commerce cart before checkout
- Content preview before signup
- Game progress before account creation

### Custom Claims and Tokens

Custom claims enable role-based access control and admin privileges.

Server Side (Admin SDK):

```typescript
import { getAuth } from 'firebase-admin/auth';
await getAuth().setCustomUserClaims(uid, { admin: true });
await getAuth().setCustomUserClaims(uid, {
  role: 'editor', organizationId: 'org_123', permissions: ['read', 'write']
});
```

Client Side:

```typescript
const idTokenResult = await auth.currentUser.getIdTokenResult();
if (idTokenResult.claims.admin === true) { /* Show admin UI */ }

// Force token refresh after claim update
await auth.currentUser.getIdToken(true);
```

Custom Claims Best Practices:

- Keep claims small (under 1000 bytes total)
- Use claims for access control, not user data
- Token refresh required after claim changes

### Security Rules Integration

Firebase Security Rules use authentication state for access control.

Firestore Security Rules:

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
    match /organizations/{orgId}/documents/{docId} {
      allow read, write: if request.auth.token.organizationId == orgId;
    }
  }
}
```

Cloud Storage Security Rules:

```javascript
rules_version = '2';
service firebase.storage {
  match /b/{bucket}/o {
    match /users/{userId}/{allPaths=**} {
      allow read, write: if request.auth != null && request.auth.uid == userId;
    }
    match /public/{allPaths=**} {
      allow read: if true;
      allow write: if request.auth != null;
    }
  }
}
```

---

## Advanced Patterns

### Cloud Functions Auth Triggers

Firebase Cloud Functions respond to authentication lifecycle events.

```typescript
import { auth } from 'firebase-functions';
import { getFirestore } from 'firebase-admin/firestore';

export const onUserCreate = auth.user().onCreate(async (user) => {
  await getFirestore().collection('users').doc(user.uid).set({
    email: user.email, displayName: user.displayName,
    createdAt: FieldValue.serverTimestamp(), role: 'member'
  });
});
```

Blocking Functions:

```typescript
import { beforeUserCreated } from 'firebase-functions/v2/identity';
export const validateUserCreate = beforeUserCreated((event) => {
  if (!event.data.email?.endsWith('@company.com')) {
    throw new HttpsError('invalid-argument', 'Unauthorized email domain');
  }
  return { customClaims: { role: 'employee' } };
});
```

### Multi-Factor Authentication

Firebase Auth supports SMS-based second factor.

```typescript
import { multiFactor, PhoneMultiFactorGenerator, PhoneAuthProvider } from 'firebase/auth';

// Enroll MFA
const mfUser = multiFactor(auth.currentUser);
const session = await mfUser.getSession();
const verificationId = await new PhoneAuthProvider(auth)
  .verifyPhoneNumber({ phoneNumber: '+1234567890', session }, recaptchaVerifier);
const credential = PhoneAuthProvider.credential(verificationId, code);
await mfUser.enroll(PhoneMultiFactorGenerator.assertion(credential), 'Phone');

// Sign-in with MFA challenge
try { await signInWithEmailAndPassword(auth, email, password); }
catch (error) {
  if (error.code === 'auth/multi-factor-auth-required') {
    const resolver = getMultiFactorResolver(auth, error);
    // Complete MFA verification with resolver.hints
  }
}
```

### Session Management

```typescript
import { setPersistence, browserLocalPersistence, browserSessionPersistence,
         inMemoryPersistence, onAuthStateChanged } from 'firebase/auth';

// Persistence options
await setPersistence(auth, browserLocalPersistence);  // Default - persist across sessions
await setPersistence(auth, browserSessionPersistence); // Clear on tab close
await setPersistence(auth, inMemoryPersistence);       // Memory only

// Auth state listener
const unsubscribe = onAuthStateChanged(auth, (user) => {
  if (user) { console.log('Signed in:', user.uid); }
  else { console.log('Signed out'); }
});
```

### Firebase Auth Emulator

```typescript
import { connectAuthEmulator } from 'firebase/auth';
if (process.env.NODE_ENV === 'development') {
  connectAuthEmulator(auth, 'http://localhost:9099');
}
```

Emulator Features:

- Test phone auth without real SMS
- Create test users programmatically
- Reset auth state between tests
- Debug token generation

---

## Resources

Context7 Documentation Access:

Use resolve-library-id with "firebase" then get-library-docs with topic "authentication" for comprehensive API documentation.

Firebase Official Resources:

- Firebase Console: console.firebase.google.com
- Authentication Documentation: firebase.google.com/docs/auth
- Security Rules Reference: firebase.google.com/docs/rules

Works Well With:

- moai-platform-firestore: Firestore database integration with auth-based security
- moai-lang-flutter: Flutter SDK for mobile Firebase Auth
- moai-lang-typescript: TypeScript patterns for Firebase SDK
- moai-domain-backend: Backend architecture with Firebase Admin SDK
- moai-quality-security: Security best practices for authentication

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v1.0
Last Updated: 2025-12-07
Provider Coverage: Firebase Authentication Only
