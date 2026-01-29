# Social Authentication Module

Comprehensive guide for implementing social identity providers with Firebase Authentication including Google, Facebook, Apple, Twitter/X, GitHub, and Microsoft.

---

## Google Sign-In

Google Sign-In provides seamless authentication within the Google ecosystem with access to Google APIs and services.

### Web Implementation

Step 1: Enable Google Sign-In provider in Firebase Console under Authentication then Sign-in method.

Step 2: Configure OAuth consent screen in Google Cloud Console with appropriate scopes.

Step 3: Import and configure Firebase Auth SDK:

```typescript
import { getAuth, signInWithPopup, signInWithRedirect, GoogleAuthProvider } from 'firebase/auth';

const auth = getAuth();
const provider = new GoogleAuthProvider();

// Add OAuth scopes for Google APIs access
provider.addScope('https://www.googleapis.com/auth/calendar.readonly');
provider.addScope('https://www.googleapis.com/auth/contacts.readonly');

// Set custom parameters
provider.setCustomParameters({
  login_hint: 'user@example.com',
  prompt: 'select_account'
});

// Popup flow (recommended for web)
const result = await signInWithPopup(auth, provider);
const credential = GoogleAuthProvider.credentialFromResult(result);
const accessToken = credential?.accessToken;
const user = result.user;

// Redirect flow (alternative for mobile browsers)
await signInWithRedirect(auth, provider);
const redirectResult = await getRedirectResult(auth);
```

### Flutter Implementation

```dart
import 'package:firebase_auth/firebase_auth.dart';
import 'package:google_sign_in/google_sign_in.dart';

Future<UserCredential> signInWithGoogle() async {
  final GoogleSignInAccount? googleUser = await GoogleSignIn().signIn();

  if (googleUser == null) {
    throw Exception('Google Sign-In was cancelled');
  }

  final GoogleSignInAuthentication googleAuth = await googleUser.authentication;

  final credential = GoogleAuthProvider.credential(
    accessToken: googleAuth.accessToken,
    idToken: googleAuth.idToken,
  );

  return await FirebaseAuth.instance.signInWithCredential(credential);
}
```

### iOS Native Implementation

```swift
import FirebaseAuth
import GoogleSignIn

func signInWithGoogle() {
    guard let clientID = FirebaseApp.app()?.options.clientID else { return }

    let config = GIDConfiguration(clientID: clientID)
    GIDSignIn.sharedInstance.configuration = config

    guard let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
          let rootViewController = windowScene.windows.first?.rootViewController else { return }

    GIDSignIn.sharedInstance.signIn(withPresenting: rootViewController) { result, error in
        if let error = error {
            print("Google Sign-In error: \(error.localizedDescription)")
            return
        }

        guard let user = result?.user,
              let idToken = user.idToken?.tokenString else { return }

        let credential = GoogleAuthProvider.credential(
            withIDToken: idToken,
            accessToken: user.accessToken.tokenString
        )

        Auth.auth().signIn(with: credential) { authResult, error in
            // Handle sign-in result
        }
    }
}
```

### Mobile Configuration Requirements

iOS Configuration:
- Add reversed client ID to URL schemes in Info.plist
- Enable Keychain Sharing capability
- Configure GoogleService-Info.plist with correct bundle ID

Android Configuration:
- Add SHA-1 and SHA-256 fingerprints to Firebase project settings
- Configure google-services.json with correct package name
- Ensure play-services-auth dependency is included

Web Configuration:
- Configure authorized domains in Firebase Console
- Set up OAuth consent screen in Google Cloud Console
- Add authorized JavaScript origins and redirect URIs

---

## Facebook Login

Facebook Login integration with Firebase Auth for social authentication.

### Web Implementation

```typescript
import { FacebookAuthProvider, signInWithPopup } from 'firebase/auth';

const provider = new FacebookAuthProvider();

provider.addScope('email');
provider.addScope('public_profile');

provider.setCustomParameters({
  display: 'popup',
  auth_type: 'rerequest'
});

const result = await signInWithPopup(auth, provider);
const credential = FacebookAuthProvider.credentialFromResult(result);
const accessToken = credential?.accessToken;
```

### Flutter Implementation

```dart
import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter_facebook_auth/flutter_facebook_auth.dart';

Future<UserCredential> signInWithFacebook() async {
  final LoginResult loginResult = await FacebookAuth.instance.login(
    permissions: ['email', 'public_profile'],
  );

  if (loginResult.status != LoginStatus.success) {
    throw Exception('Facebook login failed');
  }

  final OAuthCredential credential = FacebookAuthProvider.credential(
    loginResult.accessToken!.tokenString,
  );

  return await FirebaseAuth.instance.signInWithCredential(credential);
}
```

---

## Apple Sign-In

Apple Sign-In is required for iOS apps that offer third-party sign-in options.

### Web Implementation

```typescript
import { OAuthProvider, signInWithPopup } from 'firebase/auth';

const provider = new OAuthProvider('apple.com');
provider.addScope('email');
provider.addScope('name');

const result = await signInWithPopup(auth, provider);
const credential = OAuthProvider.credentialFromResult(result);
```

### iOS Native Implementation

```swift
import FirebaseAuth
import AuthenticationServices
import CryptoKit

class AppleSignInManager: NSObject, ASAuthorizationControllerDelegate {
    private var currentNonce: String?

    func startSignInWithAppleFlow() {
        let nonce = randomNonceString()
        currentNonce = nonce

        let appleIDProvider = ASAuthorizationAppleIDProvider()
        let request = appleIDProvider.createRequest()
        request.requestedScopes = [.fullName, .email]
        request.nonce = sha256(nonce)

        let authorizationController = ASAuthorizationController(authorizationRequests: [request])
        authorizationController.delegate = self
        authorizationController.performRequests()
    }

    func authorizationController(controller: ASAuthorizationController,
                                  didCompleteWithAuthorization authorization: ASAuthorization) {
        guard let appleIDCredential = authorization.credential as? ASAuthorizationAppleIDCredential,
              let nonce = currentNonce,
              let appleIDToken = appleIDCredential.identityToken,
              let idTokenString = String(data: appleIDToken, encoding: .utf8) else { return }

        let credential = OAuthProvider.appleCredential(
            withIDToken: idTokenString,
            rawNonce: nonce,
            fullName: appleIDCredential.fullName
        )

        Auth.auth().signIn(with: credential) { authResult, error in }
    }
}
```

---

## Twitter/X and GitHub Authentication

### Twitter/X Implementation

```typescript
import { TwitterAuthProvider, signInWithPopup } from 'firebase/auth';

const provider = new TwitterAuthProvider();
const result = await signInWithPopup(auth, provider);
const credential = TwitterAuthProvider.credentialFromResult(result);
```

### GitHub Implementation

```typescript
import { GithubAuthProvider, signInWithPopup } from 'firebase/auth';

const provider = new GithubAuthProvider();
provider.addScope('repo');
provider.addScope('read:user');

const result = await signInWithPopup(auth, provider);
const credential = GithubAuthProvider.credentialFromResult(result);
const token = credential?.accessToken;
```

### Microsoft Authentication

```typescript
import { OAuthProvider, signInWithPopup } from 'firebase/auth';

const provider = new OAuthProvider('microsoft.com');
provider.addScope('mail.read');
provider.addScope('calendars.read');

const result = await signInWithPopup(auth, provider);
```

---

## Error Handling

```typescript
import { AuthErrorCodes } from 'firebase/auth';

try {
  const result = await signInWithPopup(auth, provider);
} catch (error: any) {
  switch (error.code) {
    case AuthErrorCodes.POPUP_CLOSED_BY_USER:
      console.log('Sign-in popup was closed');
      break;
    case AuthErrorCodes.POPUP_BLOCKED:
      await signInWithRedirect(auth, provider);
      break;
    case 'auth/account-exists-with-different-credential':
      const methods = await fetchSignInMethodsForEmail(auth, error.customData?.email);
      break;
    default:
      console.error('Authentication error:', error.message);
  }
}
```

---

## Account Linking

```typescript
import { linkWithPopup, linkWithCredential, unlink } from 'firebase/auth';

// Link with popup
await linkWithPopup(auth.currentUser, new GoogleAuthProvider());

// Unlink provider
await unlink(auth.currentUser, 'google.com');

// Get linked providers
const providers = auth.currentUser?.providerData.map(p => p.providerId);
```

---

## Best Practices

Security Considerations:
- Always validate tokens server-side for sensitive operations
- Use nonce for Apple Sign-In to prevent replay attacks
- Implement proper error handling for all authentication flows

User Experience Guidelines:
- Provide multiple sign-in options for user convenience
- Handle account linking gracefully when email already exists
- Show loading states during authentication flows

Platform-Specific Notes:
- iOS requires Apple Sign-In if offering third-party sign-in
- Android requires SHA fingerprints for Google Sign-In
- Web popup may be blocked; implement redirect fallback

---

Version: 1.0.0
Last Updated: 2025-12-07
Parent Skill: moai-platform-firebase-auth
