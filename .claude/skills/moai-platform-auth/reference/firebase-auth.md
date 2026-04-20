# Firebase Auth Platform Reference

Mobile-first authentication with seamless Google ecosystem integration and Firebase services.

## Platform Overview

Firebase Authentication provides backend services, SDKs, and UI libraries for authenticating users across multiple platforms. Integrates natively with other Firebase services (Firestore, Storage, Cloud Functions) and supports Google Sign-In, social providers, phone authentication, and anonymous authentication.

## Project Setup

### Create Firebase Project

Navigate to Firebase Console at console.firebase.google.com. Click Add Project. Enter project name and ID. Accept terms and create project. Wait for project initialization.

### Add Application

Select platform:
- Web: Register web app with app nickname
- iOS: Register iOS bundle ID
- Android: Register Android package name
- Flutter: Follow Flutter setup guide

Download configuration files:
- Web: Copy firebaseConfig object
- iOS: Download GoogleService-Info.plist
- Android: Download google-services.json
- Flutter: Use FlutterFire CLI

### Enable Authentication Providers

Navigate to Console > Authentication > Sign-in method. Click Add new provider. Select provider (Email/Password, Google, Facebook, etc.). Configure provider settings. Enable provider. Save changes.

## SDK Initialization

### Web SDK

**Installation:**
```bash
npm install firebase
```

**Initialization:**

Import initializeApp from firebase/app module. Import getAuth and onAuthStateChanged from firebase/auth module. Create firebaseConfig object with apiKey, authDomain, projectId, storageBucket, messagingSenderId, and appId. Call initializeApp with config to get app instance. Call getAuth with app to get auth instance. Set up auth state listener with onAuthStateChanged. Listener callback receives user object when signed in or null when signed out.

**Configuration Object:**
- apiKey: Web API key from Firebase Console
- authDomain: Auth domain (project-id.firebaseapp.com)
- projectId: Firebase project ID
- storageBucket: Storage bucket URL (optional)
- messagingSenderId: Cloud messaging sender ID (optional)
- appId: Firebase app ID

### Flutter SDK

**Installation:**

Add firebase_core and firebase_auth to pubspec.yaml dependencies. Run flutter pub get to install.

**FlutterFire CLI Setup:**

Install FlutterFire CLI globally with npm or dart pub global. Run flutterfire configure to select Firebase project. CLI generates firebase_options.dart file automatically.

**Initialization:**

Import firebase_core package. Import firebase_auth package. In main function, ensure Flutter bindings are initialized with WidgetsFlutterBinding.ensureInitialized(). Await Firebase.initializeApp(). Run app after initialization.

**Auth State Listener:**

Use FirebaseAuth.instance.authStateChanges() stream. Listen to stream with listen method. Callback receives User object when signed in or null when signed out. Use in provider or state management.

### iOS SDK

**Installation:**

Add Firebase to iOS project using CocoaPods or Swift Package Manager. For CocoaPods, add pod Firebase/Auth to Podfile. Run pod install.

**Configuration:**

Download GoogleService-Info.plist from Firebase Console. Add plist file to Xcode project. Import FirebaseCore and FirebaseAuth in AppDelegate. Call FirebaseApp.configure() in application didFinishLaunchingWithOptions.

**Auth State Listener:**

Use Auth.auth().addStateDidChangeListener method. Listener receives auth and user parameters. User is nil when signed out or User object when signed in. Store listener handle to remove later.

### Android SDK

**Installation:**

Add Firebase to Android project using Gradle. Add google-services plugin to project-level build.gradle. Add Firebase Auth dependency to app-level build.gradle. Sync project.

**Configuration:**

Download google-services.json from Firebase Console. Place file in app/ directory. Firebase automatically initialized on app start.

**Auth State Listener:**

Use FirebaseAuth.getInstance().addAuthStateListener(). Listener receives FirebaseAuth instance. Call getCurrentUser() to get current user or null.

### React Native SDK

**Installation:**

```bash
npm install @react-native-firebase/app @react-native-firebase/auth
```

**iOS Setup:**

Run pod install in ios/ directory. Add GoogleService-Info.plist to Xcode project.

**Android Setup:**

Add google-services.json to android/app/ directory. Plugin automatically configured.

**Usage:**

Import auth from @react-native-firebase/auth. Call auth() to get auth instance. Use onAuthStateChanged listener. Works identically to web SDK.

## Google Sign-In

Native integration with Google accounts and Google ecosystem.

### Web Implementation

**Installation:**
```bash
npm install firebase
```

**Implementation:**

Import GoogleAuthProvider and signInWithPopup from firebase/auth. Create provider instance with new GoogleAuthProvider(). Call signInWithPopup with auth and provider. Handle promise with user credential result. Extract user from result.user. Handle errors with catch block.

**Popup vs Redirect:**

signInWithPopup:
- Opens popup window for authentication
- Better user experience on desktop
- May be blocked by popup blockers
- Returns promise with user credential

signInWithRedirect:
- Redirects entire page to Google
- Better for mobile browsers
- No popup blocker issues
- Use getRedirectResult() after redirect back

**Custom Parameters:**

Add scopes with provider.addScope(). Request additional permissions like email or profile. Add custom parameters with provider.setCustomParameters(). Example: login_hint for pre-filled email.

### Flutter Implementation

Import google_sign_in package. Create GoogleSignIn instance. Call signIn() to trigger Google authentication flow. User selects account or signs in. Get GoogleSignInAuthentication with accessToken and idToken. Create credential with GoogleAuthProvider.credential(). Call FirebaseAuth.instance.signInWithCredential(). User now signed into Firebase.

**Error Handling:**

Handle PlatformException for user cancellation. Catch FirebaseAuthException for authentication errors. Check error code for specific handling.

### iOS Implementation

Import GoogleSignIn framework. Configure GIDSignIn with client ID from GoogleService-Info.plist. Call GIDSignIn.sharedInstance.signIn() with presenting view controller. Handle result in completion callback. Get ID token and access token. Create credential with GoogleAuthProvider.credential(). Call Auth.auth().signIn(with:credential:).

**iOS Configuration:**

Add URL scheme to Info.plist with reversed client ID. Configure Google Sign-In in AppDelegate. Handle sign-in callback URL.

### Android Implementation

Configure Google Sign-In in build.gradle with web client ID. Create GoogleSignInOptions with requestIdToken. Build GoogleSignInClient. Launch sign-in intent. Handle result in onActivityResult. Get ID token from sign-in account. Create credential with GoogleAuthProvider.getCredential(). Call FirebaseAuth.getInstance().signInWithCredential().

## Social Authentication

OAuth providers for third-party authentication.

### Facebook Login

**Firebase Configuration:**

Enable Facebook provider in Firebase Console. Enter App ID and App Secret from Facebook Developer portal. Configure OAuth redirect URI in Facebook app settings.

**Web Implementation:**

Import FacebookAuthProvider and signInWithPopup. Create provider instance. Call signInWithPopup with auth and provider. User redirects to Facebook for authentication. Returns with user credential. Extract user information.

**Requesting Permissions:**

Add scopes with provider.addScope(). Common scopes: email, public_profile, user_friends. Additional scopes require Facebook app review.

**Flutter Implementation:**

Install flutter_facebook_auth package. Call FacebookAuth.instance.login() with permissions list. Get AccessToken from result. Create credential with FacebookAuthProvider.credential(). Sign in with Firebase.

### Apple Sign-In

**Required for iOS Apps:**

Apple requires Apple Sign-In if app offers third-party authentication. Failure to implement results in App Store rejection.

**Firebase Configuration:**

Enable Apple provider in Firebase Console. Configure Service ID and Team ID. Register redirect URL in Apple Developer portal.

**Web Implementation:**

Import OAuthProvider and create instance with apple.com. Call signInWithPopup with auth and provider. User authenticates with Apple ID. Returns user credential with Apple user data.

**iOS Implementation:**

Use AuthenticationServices framework. Create ASAuthorizationAppleIDRequest. Request full name and email scopes. Present ASAuthorizationController. Handle result in delegate. Get ID token and authorization code. Create credential with OAuthProvider.credential(). Sign in with Firebase.

### Twitter/X Authentication

**Configuration:**

Enable Twitter provider in Firebase Console. Enter API Key and API Secret from Twitter Developer portal. Configure callback URL.

**Implementation:**

Import TwitterAuthProvider. Create provider instance. Call signInWithPopup or signInWithRedirect. User authenticates on Twitter. Returns with user credential including profile data.

### GitHub Authentication

**Configuration:**

Enable GitHub provider in Firebase Console. Create OAuth App in GitHub Settings. Enter Client ID and Client Secret. Configure authorization callback URL.

**Implementation:**

Import GithubAuthProvider. Create provider instance. Call signInWithPopup. User authenticates with GitHub account. Returns credential with GitHub profile.

**Requesting Scopes:**

Add scopes for GitHub API access. Common scopes: user, repo, gist. User must approve requested permissions.

### Microsoft/Azure AD

**Enterprise Authentication:**

Enable Microsoft provider in Firebase Console. Supports personal Microsoft accounts and Azure AD. Configure tenant ID for organizational accounts.

**Implementation:**

Import OAuthProvider with microsoft.com. Create provider instance. Optionally set custom parameters for tenant. Call signInWithPopup. User authenticates with Microsoft account.

## Phone Authentication

SMS-based verification for phone number authentication.

### Web Implementation

**reCAPTCHA Setup:**

Import RecaptchaVerifier from firebase/auth. Create verifier instance with container ID. Specify size (invisible or normal). Set callback for success. Render verifier before sign-in attempt.

**Phone Sign-In Flow:**

Import signInWithPhoneNumber. Call with auth instance, phone number (E.164 format), and RecaptchaVerifier. Returns ConfirmationResult. User receives SMS with verification code. Call confirmationResult.confirm(code) with user-entered code. Returns user credential on success.

**E.164 Format:**

Phone number must include country code with plus prefix. Example: +1 (US), +82 (Korea), +44 (UK). No spaces or special characters in final string.

**Error Handling:**

auth/invalid-phone-number: Phone number format incorrect. auth/missing-phone-number: Phone number not provided. auth/quota-exceeded: SMS quota exhausted. auth/captcha-check-failed: reCAPTCHA verification failed.

### Flutter Implementation

**verifyPhoneNumber Flow:**

Call FirebaseAuth.instance.verifyPhoneNumber() with multiple callbacks. Provide phoneNumber in E.164 format. Set verificationCompleted callback for automatic verification (Android only). Set verificationFailed callback for errors. Set codeSent callback receiving verificationId. Set codeAutoRetrievalTimeout callback for timeout.

**Manual Verification:**

In codeSent callback, store verificationId. Show UI for code input. Create credential with PhoneAuthProvider.credential(verificationId, code). Call signInWithCredential.

**Automatic Verification (Android):**

On some Android devices, SMS automatically read and verified. VerificationCompleted callback receives PhoneAuthCredential. Sign in immediately without manual code entry.

### iOS Implementation

Import FirebaseAuth. Call PhoneAuthProvider.provider().verifyPhoneNumber(). Provide phone number and completion callback. Callback receives verificationID or error. User receives SMS with code. Create credential with PhoneAuthProvider.provider().credential(). Pass verificationID and verification code. Sign in with credential.

**iOS Configuration:**

Enable reCAPTCHA verification in Firebase Console. Add URL scheme to Info.plist. Configure silent push notifications for automatic verification.

### Android Implementation

Call FirebaseAuth.getInstance().signInWithPhoneNumber(). Provide Activity, phone number, and callbacks. Callbacks include OnVerificationStateChangedCallbacks. Handle onVerificationCompleted for automatic verification. Handle onCodeSent for manual verification. Create credential with PhoneAuthProvider.getCredential(). Sign in with credential.

**Android Configuration:**

Add SHA-1 fingerprint to Firebase Console. Enable SafetyNet in Google Cloud Console. Configure reCAPTCHA fallback.

### Rate Limiting

Firebase enforces rate limits on phone authentication:
- 10 SMS per phone number per hour
- 100 SMS per IP address per day
- Exceeded quota returns auth/quota-exceeded

**Best Practices:**
- Implement client-side throttling
- Show clear error messages
- Provide alternative authentication methods
- Monitor quota usage in Firebase Console

## Anonymous Authentication

Guest access with optional account upgrade.

### Implementation

**Web:**

Import signInAnonymously from firebase/auth. Call signInAnonymously(auth). Returns user credential with anonymous user. User automatically assigned unique ID. No email or password required.

**Flutter:**

Call FirebaseAuth.instance.signInAnonymously(). Returns UserCredential. Access user with userCredential.user. User persists across app restarts.

**iOS:**

Call Auth.auth().signInAnonymously(). Handle completion with user or error. Anonymous user created immediately.

**Android:**

Call FirebaseAuth.getInstance().signInAnonymously(). Returns Task with AuthResult. Extract FirebaseUser from result.

### Account Linking

Convert anonymous account to permanent account with email/password, phone, or social provider.

**Link Email/Password:**

Import EmailAuthProvider and linkWithCredential. Create credential with EmailAuthProvider.credential(email, password). Call currentUser.linkWithCredential(credential). Anonymous data migrates to permanent account. User can now sign in with email/password.

**Link Social Provider:**

Create provider (GoogleAuthProvider, FacebookAuthProvider, etc.). Call linkWithPopup(provider) or linkWithRedirect(provider). User authenticates with social provider. Anonymous data migrates to social account.

**Link Phone Number:**

Use same flow as phone sign-in. Instead of signInWithPhoneNumber, use linkWithPhoneNumber. Verify code and confirm. Anonymous data migrates to phone account.

**Error Handling:**

auth/credential-already-in-use: Credential linked to different account. auth/email-already-in-use: Email already registered. auth/provider-already-linked: Provider already linked to this user.

## Custom Claims

Role-based access control with JWT custom claims.

### Setting Custom Claims

**Admin SDK (Node.js):**

Import admin from firebase-admin. Initialize admin SDK with credentials. Call admin.auth().setCustomUserClaims(uid, claims). Claims object contains custom key-value pairs. Example: { admin: true, role: 'editor' }. Claims included in user's ID token.

**Admin SDK (Python):**

Import firebase_admin and auth. Initialize admin SDK. Call auth.set_custom_user_claims(uid, claims). Claims propagate to ID token on next refresh.

**Admin SDK (Go):**

Import firebase.google.com/go/v4/auth. Initialize admin SDK. Call client.SetCustomUserClaims(ctx, uid, claims). Claims available in ID token.

### Reading Custom Claims

**Client Side (Web):**

Call user.getIdTokenResult(). Access claims via result.claims object. Check for custom claims like claims.admin or claims.role. Tokens expire after 1 hour, claims update on refresh.

**Client Side (Flutter):**

Call FirebaseAuth.instance.currentUser.getIdTokenResult(). Access claims via tokenResult.claims. Cast values to appropriate types.

**Client Side (iOS):**

Call user.getIDTokenResult(). Access claims via result.claims dictionary. Check for custom claim keys.

**Server Side:**

Verify ID token with Admin SDK. Decoded token includes custom claims. Access claims from decoded token object. Use for authorization logic.

### Force Token Refresh

Custom claims only update in ID token after refresh. Force refresh immediately with user.getIdToken(true) passing forceRefresh true. New token includes updated claims.

### Limitations

**Size Limit:**

Custom claims payload limited to 1000 bytes. Exceeding limit throws error. Keep claims minimal.

**Update Propagation:**

Claims update on next token refresh (hourly). Force refresh for immediate effect. Consider client-side caching for frequently checked claims.

**Security:**

Custom claims are public (in JWT). Do not store sensitive data. Use for authorization roles only.

## Security Rules Integration

Integrate authentication with Firebase services security.

### Firestore Security Rules

Use rules_version 2. Define service cloud.firestore. In match databases block, define rules for collections. Access authentication with request.auth. Check user ID with request.auth.uid. Check custom claims with request.auth.token.

**User Data Protection:**

Match users collection with userId wildcard. Allow read and write if request.auth exists and request.auth.uid equals userId. Ensures users only access own data.

**Admin-Only Access:**

Match admin collection. Allow read and write if request.auth.token.admin equals true. Use custom claims for role-based access.

**Organization Data:**

Match organizations collection with orgId wildcard. Allow read if request.auth exists. Allow write if request.auth.token.orgId equals orgId. Implements multi-tenant security.

### Firebase Storage Rules

Use rules_version 2. Define service firebase.storage. Access authentication with request.auth. Use request.auth.uid for user identification. Check request.auth.token for custom claims.

**User File Storage:**

Match users path with userId wildcard. Allow read and write if request.auth exists and request.auth.uid equals userId. Users upload files to own directory.

**Public Read, Private Write:**

Match public path. Allow read for all. Allow write only if request.auth exists. Useful for public content.

**File Size Limits:**

Check request.resource.size for upload size. Example: request.resource.size less than 5 * 1024 * 1024 for 5MB limit. Prevents large uploads.

### Realtime Database Rules

Use JSON format for rules. Access authentication with auth. Check user ID with auth.uid. Check custom claims with auth.token.

**User Data Node:**

Define users node with $userId wildcard. Set .read and .write rules to auth not null and auth.uid equals $userId. Protects user-specific data.

**Admin Access:**

Set .read and .write to auth.token.admin equals true. Uses custom claims for authorization.

## Cloud Functions Triggers

React to authentication events with Cloud Functions.

### onCreate Trigger

Triggered when new user created. Receives UserRecord object with user data. Use for sending welcome emails, creating user documents, initializing user data.

**Implementation (Node.js):**

Import functions from firebase-functions. Import admin for Firestore access. Define onCreate function with functions.auth.user().onCreate(). Async handler receives UserRecord. Access user.uid, user.email, user.displayName. Perform setup tasks. Return promise.

**Use Cases:**
- Create Firestore user document
- Send welcome email
- Initialize user profile
- Grant default permissions
- Add to mailing list

### onDelete Trigger

Triggered when user deleted. Receives UserRecord object. Use for cleanup tasks, deleting user data, revoking permissions.

**Implementation:**

Define onDelete function with functions.auth.user().onDelete(). Handler receives UserRecord. Delete user's Firestore documents. Remove user files from Storage. Clean up external integrations. Return promise.

**Use Cases:**
- Delete user data (GDPR compliance)
- Remove user files
- Revoke API keys
- Cancel subscriptions
- Clean up related data

### HTTP Triggers for Custom Claims

Create HTTP function for setting custom claims. Verify admin authorization. Call admin.auth().setCustomUserClaims(). Return success response.

**Security:**

Verify calling user is admin. Check authentication token. Validate requested claims. Audit claim changes.

## Admin SDK Setup

Server-side Firebase operations with elevated privileges.

### Node.js Setup

**Installation:**
```bash
npm install firebase-admin
```

**Initialization:**

Import firebase-admin. For development, use service account key. Call admin.initializeApp with credential. For production, use default credentials on Firebase hosting or Cloud Functions.

**Service Account Key:**

Download from Firebase Console > Project Settings > Service Accounts. Store securely (environment variable or secret manager). Never commit to version control. Initialize with admin.credential.cert(serviceAccount).

### Python Setup

**Installation:**
```bash
pip install firebase-admin
```

**Initialization:**

Import firebase_admin and credentials. Create certificate credential with credentials.Certificate(). Call firebase_admin.initialize_app(). Use default credentials in production environments.

### Go Setup

**Installation:**
```bash
go get firebase.google.com/go/v4
```

**Initialization:**

Import firebase package. Create context. Call firebase.NewApp() with config. Extract auth client with app.Auth(). Use for user management operations.

### Common Admin Operations

**Get User:**

Call admin.auth().getUser(uid). Returns UserRecord. Access user properties. Throws error if user not found.

**List Users:**

Call admin.auth().listUsers(). Returns paginated results. Iterate through users. Use for bulk operations.

**Create User:**

Call admin.auth().createUser() with properties. Set email, password, displayName, etc. Returns UserRecord with new user ID.

**Update User:**

Call admin.auth().updateUser(uid, properties). Modify user attributes. Email verification status cannot be set directly.

**Delete User:**

Call admin.auth().deleteUser(uid). Permanently removes user. Triggers onDelete Cloud Function. Cannot be undone.

**Verify ID Token:**

Call admin.auth().verifyIdToken(token). Returns decoded token with claims. Use for server-side authentication. Validates signature and expiration.

## Multi-Factor Authentication

Additional security layer with second factor verification.

### MFA Enrollment

**Phone MFA (Web):**

Import PhoneAuthProvider, PhoneMultiFactorGenerator, multiFactor. Call multiFactor(user).getSession(). Use session to initialize PhoneAuthProvider.verifyPhoneNumber(). User receives SMS code. Create credential with PhoneAuthProvider.credential(). Wrap in PhoneMultiFactorGenerator.assertion(). Call multiFactor(user).enroll() with assertion and display name.

**TOTP MFA (Enterprise):**

Request TOTP secret. Display QR code for authenticator app scanning. User scans QR code. User enters verification code. Enroll MFA factor with code.

### MFA Sign-In

**Detecting MFA Requirement:**

Sign-in attempt throws auth/multi-factor-auth-required error. Extract resolver from error. Resolver contains enrolled factors and session.

**Completing MFA:**

User selects factor. For phone, verify phone number with resolver session. User enters SMS code. Create credential with PhoneAuthProvider.credential(). Wrap in PhoneMultiFactorGenerator.assertion(). Call resolver.resolveSignIn() with assertion. Returns user credential on success.

### Managing MFA

**List Enrolled Factors:**

Call multiFactor(user).enrolledFactors. Returns array of MultiFactorInfo. Each factor has uid, displayName, and enrollmentTime.

**Unenroll Factor:**

Call multiFactor(user).unenroll(factorUid). Requires recent authentication. User must re-authenticate if session expired.

## Session Management

Control session persistence and security.

### Session Persistence (Web)

**Persistence Types:**

browserLocalPersistence:
- Session persists across browser tabs and restarts
- Default behavior
- Use for most applications

browserSessionPersistence:
- Session persists only in current tab
- Cleared when tab closed
- Use for shared computers

inMemoryPersistence:
- Session cleared on page refresh
- Not persisted
- Use for maximum security

**Setting Persistence:**

Import setPersistence and persistence types. Call setPersistence(auth, persistence) before sign-in. Returns promise. Sign in after persistence set.

### Session Cookies (Server)

Create session cookie for server-side sessions. Call admin.auth().createSessionCookie(idToken, options). Set expiresIn duration (5 minutes to 2 weeks). Return cookie to client. Client includes cookie in requests. Server verifies cookie with admin.auth().verifySessionCookie().

**Use Cases:**
- Server-side rendering
- Traditional session-based apps
- Progressive web apps with offline support

### Token Refresh

ID tokens expire after 1 hour. SDK automatically refreshes tokens. Force refresh with user.getIdToken(true). Monitor token expiration with onIdTokenChanged listener.

## Testing and Development

### Firebase Auth Emulator

**Setup:**

Install Firebase CLI globally. Run firebase init emulators. Select Authentication emulator. Configure port (default 9099). Start emulator with firebase emulators:start.

**Connecting to Emulator:**

Import connectAuthEmulator. Call connectAuthEmulator(auth, 'http://localhost:9099'). Must be called before any auth operations. All auth requests now go to emulator.

**Benefits:**
- No SMS costs during development
- No rate limiting
- Fast iteration
- Offline development
- Predictable test data

### Test Users

Create test users in emulator UI. Set custom claims without Admin SDK. View user data and tokens. Delete test data easily. Export and import auth data.

## Error Handling

### Common Error Codes

**auth/invalid-email:**
Email address malformed. Validate email format before submission.

**auth/user-not-found:**
No user record for identifier. Check if user exists or prompt sign-up.

**auth/wrong-password:**
Password incorrect for email. Limit retry attempts to prevent brute force.

**auth/email-already-in-use:**
Email already registered. Prompt user to sign in or recover password.

**auth/weak-password:**
Password too weak. Enforce password requirements (8+ characters, complexity).

**auth/operation-not-allowed:**
Provider not enabled in Firebase Console. Enable provider in settings.

**auth/account-exists-with-different-credential:**
Account exists with same email but different provider. Offer account linking.

**auth/requires-recent-login:**
Sensitive operation requires recent authentication. Re-authenticate user before operation.

### Error Handling Pattern

Wrap auth calls in try-catch. Check error code with error.code. Provide user-friendly messages. Log errors for monitoring. Implement retry logic for transient errors.

## Resources

**Official Documentation:**
- Firebase Auth Docs: https://firebase.google.com/docs/auth
- Web Guide: https://firebase.google.com/docs/auth/web/start
- Flutter Guide: https://firebase.google.com/docs/auth/flutter/start
- iOS Guide: https://firebase.google.com/docs/auth/ios/start
- Android Guide: https://firebase.google.com/docs/auth/android/start

**Firebase SDKs:**
- Web: firebase/auth
- Flutter: firebase_auth
- iOS: FirebaseAuth (Swift)
- Android: firebase-auth (Kotlin)
- React Native: @react-native-firebase/auth

**Tools:**
- Firebase Console: console.firebase.google.com
- Firebase CLI: firebase.google.com/docs/cli
- Emulator Suite: firebase.google.com/docs/emulator-suite

**Community:**
- Stack Overflow: stackoverflow.com/questions/tagged/firebase-authentication
- Firebase Community: firebase.google.com/community
- GitHub: github.com/firebase
