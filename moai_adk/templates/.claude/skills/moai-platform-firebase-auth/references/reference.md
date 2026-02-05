# Firebase Authentication Reference

Extended reference documentation for Firebase Authentication covering advanced patterns, configuration guides, and security best practices.

---

## Context7 Documentation Access

For the latest Firebase Authentication documentation, use Context7 MCP tools:

Step 1: Resolve library ID using resolve-library-id with query "firebase"

Step 2: Fetch documentation using get-library-docs with the resolved Context7 ID and topic "authentication"

This provides access to the most current API documentation, breaking changes, and best practices.

---

## Firebase Console Configuration

### Authentication Provider Setup

Navigate to Firebase Console at console.firebase.google.com, select your project, then go to Authentication section.

Sign-in Method Configuration:
- Enable desired providers under Sign-in method tab
- Configure OAuth credentials for social providers
- Set up authorized domains for web applications
- Configure email templates for email-based authentication

### OAuth Provider Configuration

Google Sign-In:
- Automatically configured with Firebase project
- Configure OAuth consent screen in Google Cloud Console
- Add authorized domains and redirect URIs

Facebook Login:
- Create Facebook App at developers.facebook.com
- Copy App ID and App Secret to Firebase Console
- Configure OAuth redirect URI from Firebase Console

Apple Sign-In:
- Enable Sign In with Apple in Apple Developer Console
- Configure Service ID for web authentication
- Generate private key for server-to-server communication
- Add team ID, key ID, and private key to Firebase Console

Twitter/X:
- Create Twitter App at developer.twitter.com
- Copy API Key and API Secret to Firebase Console

GitHub:
- Create GitHub OAuth App at github.com/settings/developers
- Copy Client ID and Client Secret to Firebase Console

Microsoft:
- Register application in Azure Portal
- Configure as multi-tenant or single-tenant
- Copy Application ID and Client Secret to Firebase Console

---

## Admin SDK Setup

### Node.js Initialization

```typescript
import { initializeApp, cert } from 'firebase-admin/app';
import { getAuth } from 'firebase-admin/auth';

// Initialize with service account
initializeApp({
  credential: cert('./service-account-key.json')
});

// Or use default credentials (Cloud Functions, Cloud Run)
initializeApp();

const auth = getAuth();
```

### Python Initialization

```python
import firebase_admin
from firebase_admin import credentials, auth

cred = credentials.Certificate('./service-account-key.json')
firebase_admin.initialize_app(cred)
```

### Go Initialization

```go
import (
    "context"
    firebase "firebase.google.com/go/v4"
    "google.golang.org/api/option"
)

opt := option.WithCredentialsFile("./service-account-key.json")
app, err := firebase.NewApp(context.Background(), nil, opt)
client, err := app.Auth(context.Background())
```

---

## Multi-Factor Authentication

### Enrollment Flow

```typescript
import { multiFactor, PhoneMultiFactorGenerator, PhoneAuthProvider } from 'firebase/auth';

const enrollMFA = async (phoneNumber: string) => {
  const user = auth.currentUser;
  if (!user) throw new Error('Not signed in');

  const mfUser = multiFactor(user);
  const session = await mfUser.getSession();

  const phoneAuthProvider = new PhoneAuthProvider(auth);
  const verificationId = await phoneAuthProvider.verifyPhoneNumber(
    { phoneNumber, session },
    recaptchaVerifier
  );

  // User enters verification code
  const code = await promptUserForCode();

  const credential = PhoneAuthProvider.credential(verificationId, code);
  const assertion = PhoneMultiFactorGenerator.assertion(credential);

  await mfUser.enroll(assertion, 'Phone');
};
```

### Sign-In with MFA Challenge

```typescript
import { getMultiFactorResolver } from 'firebase/auth';

try {
  await signInWithEmailAndPassword(auth, email, password);
} catch (error: any) {
  if (error.code === 'auth/multi-factor-auth-required') {
    const resolver = getMultiFactorResolver(auth, error);

    // Show available second factors
    const hints = resolver.hints;

    // Verify with selected factor
    const phoneAuthProvider = new PhoneAuthProvider(auth);
    const verificationId = await phoneAuthProvider.verifyPhoneNumber(
      { multiFactorHint: hints[0], session: resolver.session },
      recaptchaVerifier
    );

    const code = await promptUserForCode();
    const credential = PhoneAuthProvider.credential(verificationId, code);
    const assertion = PhoneMultiFactorGenerator.assertion(credential);

    await resolver.resolveSignIn(assertion);
  }
}
```

---

## Cloud Functions Auth Triggers

### User Lifecycle Events

```typescript
import { auth } from 'firebase-functions';
import { getFirestore, FieldValue } from 'firebase-admin/firestore';

// On user creation
export const onUserCreate = auth.user().onCreate(async (user) => {
  await getFirestore().collection('users').doc(user.uid).set({
    email: user.email,
    displayName: user.displayName,
    photoURL: user.photoURL,
    createdAt: FieldValue.serverTimestamp(),
    providers: user.providerData.map(p => p.providerId)
  });
});

// On user deletion
export const onUserDelete = auth.user().onDelete(async (user) => {
  // Clean up user data
  const batch = getFirestore().batch();

  batch.delete(getFirestore().collection('users').doc(user.uid));

  const userDocs = await getFirestore()
    .collection('userContent')
    .where('userId', '==', user.uid)
    .get();

  userDocs.forEach(doc => batch.delete(doc.ref));

  await batch.commit();
});
```

### Blocking Functions

```typescript
import { beforeUserCreated, beforeUserSignedIn } from 'firebase-functions/v2/identity';

// Validate before user creation
export const validateUserCreate = beforeUserCreated((event) => {
  const user = event.data;

  // Domain restriction
  if (!user.email?.endsWith('@company.com')) {
    throw new HttpsError('permission-denied', 'Unauthorized email domain');
  }

  // Set initial claims
  return {
    customClaims: { role: 'member' }
  };
});

// Validate before sign-in
export const validateSignIn = beforeUserSignedIn((event) => {
  const user = event.data;

  // Check if user is disabled in external system
  // Return undefined to allow, throw to block

  return;
});
```

---

## Session Management

### Persistence Options

```typescript
import {
  setPersistence,
  browserLocalPersistence,
  browserSessionPersistence,
  inMemoryPersistence
} from 'firebase/auth';

// Persist across browser sessions (default)
await setPersistence(auth, browserLocalPersistence);

// Clear on tab close
await setPersistence(auth, browserSessionPersistence);

// No persistence (memory only)
await setPersistence(auth, inMemoryPersistence);
```

### Session Cookies (Server-Side)

```typescript
// Server: Create session cookie
const expiresIn = 60 * 60 * 24 * 5 * 1000; // 5 days
const sessionCookie = await getAuth().createSessionCookie(idToken, { expiresIn });

// Set cookie in response
res.cookie('session', sessionCookie, {
  maxAge: expiresIn,
  httpOnly: true,
  secure: true,
  sameSite: 'strict'
});

// Verify session cookie
const decodedClaims = await getAuth().verifySessionCookie(sessionCookie, true);
```

---

## Firebase Auth Emulator

### Configuration

```typescript
import { connectAuthEmulator } from 'firebase/auth';

if (process.env.NODE_ENV === 'development') {
  connectAuthEmulator(auth, 'http://localhost:9099');
}
```

### Emulator Features

Test Phone Numbers:
- Configure test phone numbers in Firebase Console
- Use any 6-digit code for verification
- No actual SMS sent

Emulator UI:
- Access at localhost:4000 when running emulator suite
- View and manage test users
- Inspect authentication state

Programmatic User Creation:
- Create users via Emulator REST API
- Useful for test setup

---

## Security Best Practices

Token Validation:
- Always verify ID tokens server-side for sensitive operations
- Check token expiration and issuer claims
- Use Admin SDK verifyIdToken method

Rate Limiting:
- Firebase automatically rate limits authentication requests
- Implement additional rate limiting for sensitive operations
- Monitor failed authentication attempts

Secure Configuration:
- Never expose service account keys in client code
- Use environment variables for sensitive configuration
- Rotate API keys periodically

Custom Claims Security:
- Never trust client-provided claims
- Validate claims server-side before granting access
- Use Security Rules for claim-based access control

---

## Error Codes Reference

Common Authentication Errors:
- auth/user-not-found: No user with provided identifier
- auth/wrong-password: Invalid password
- auth/email-already-in-use: Email already registered
- auth/weak-password: Password does not meet requirements
- auth/invalid-email: Malformed email adddess
- auth/user-disabled: User account is disabled
- auth/too-many-requests: Rate limit exceeded
- auth/network-request-failed: Network connectivity issue

Social Provider Errors:
- auth/popup-closed-by-user: User closed popup before completion
- auth/popup-blocked: Browser blocked popup window
- auth/account-exists-with-different-credential: Email linked to different provider
- auth/cancelled-popup-request: Multiple popup requests

Phone Auth Errors:
- auth/invalid-phone-number: Malformed phone number
- auth/invalid-verification-code: Wrong SMS code
- auth/code-expired: Verification code expired
- auth/quota-exceeded: SMS quota exceeded

---

## Platform SDK Versions

Web SDK:
- Firebase JS SDK v9+ (modular)
- Legacy v8 (compat mode available)

Mobile SDKs:
- iOS: FirebaseAuth via Swift Package Manager or CocoaPods
- Android: firebase-auth via Gradle

Cross-Platform:
- Flutter: firebase_auth package
- React Native: @react-native-firebase/auth

---

Version: 1.0.0
Last Updated: 2025-12-07
Parent Skill: moai-platform-firebase-auth
