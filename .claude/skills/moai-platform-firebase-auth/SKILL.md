---
name: "moai-platform-firebase-auth"
description: "Firebase Authentication specialist covering Google ecosystem, social auth, phone auth, and mobile-first patterns. Use when building Firebase-backed or Google ecosystem apps, implementing social login, or adding phone verification."
version: 2.1.0
category: "platform"
modularized: true
user-invocable: false
tags: ['firebase, google, social-auth, mobile, authentication']
updated: 2026-01-11
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

- Google Sign-In provides native Google ecosystem integration with Cloud Identity
- Social Auth supports Facebook, Twitter/X, GitHub, Apple, Microsoft, and Yahoo providers
- Phone Auth provides SMS-based verification with international support
- Anonymous Auth offers guest access with account linking upgrade path
- Custom Claims enable role-based access and admin privileges
- Security Rules provide integration with Firestore, Storage, and Realtime Database

Context7 Documentation Access:

- Use resolve-library-id with "firebase" to get Context7 library ID
- Use get-library-docs with resolved ID and topic "authentication" for latest API

Platform SDK Support:

- Web uses firebase/auth with modular SDK version 9 and higher
- iOS uses FirebaseAuth with Swift and SwiftUI
- Android uses firebase-auth with Kotlin
- Flutter uses the firebase_auth package
- React Native uses @react-native-firebase/auth

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

Key Topics include Google Sign-In with OAuth scopes and custom parameters, Facebook Login with Graph API access, Apple Sign-In which is required for iOS apps with third-party login, account linking between providers, and error handling for popup and redirect flows.

### Phone Authentication

File: modules/phone-auth.md

Covers SMS-based phone number authentication with international support, reCAPTCHA verification, and platform-specific implementations.

Key Topics include web implementation with RecaptchaVerifier, Flutter verifyPhoneNumber flow with auto-verification, iOS and Android native implementations, E.164 phone number formatting, and error handling and rate limiting.

### Custom Claims and Role Management

File: modules/custom-claims.md

Covers custom claims, role-based access control (RBAC), and admin privileges with Security Rules integration.

Key Topics include setting claims with Admin SDK for Node.js, Python, and Go, Cloud Functions for claim management, reading claims on client for Web, Flutter, and iOS, Security Rules integration for Firestore and Storage, and token refresh strategies.

---

## Implementation Guide

### Firebase Project Setup

Step 1: Create Firebase project at console.firebase.google.com. Step 2: Add your app for Web, iOS, Android, or Flutter. Step 3: Enable desired authentication providers. Step 4: Download and configure SDK credentials.

### Web SDK Initialization

Import initializeApp from firebase/app and getAuth with onAuthStateChanged from firebase/auth. Create a firebaseConfig object with apiKey, authDomain, and projectId values. Call initializeApp with the config to get the app instance. Call getAuth with the app to get the auth instance. Use onAuthStateChanged to listen for authentication state changes, logging signed in with user.uid when user exists or signed out when null.

### Flutter SDK Initialization

Import firebase_core and firebase_auth packages. In the main function, ensure Flutter bindings are initialized, await Firebase.initializeApp(), then run the app. Set up an auth state listener using FirebaseAuth.instance.authStateChanges().listen() to handle User objects, logging signed in with user.uid when user is not null.

### Anonymous Authentication

Import signInAnonymously, linkWithCredential, and EmailAuthProvider from firebase/auth. Call signInAnonymously with the auth instance to create an anonymous user, logging the anonymous UID from result.user.uid. To upgrade to a permanent account, create a credential using EmailAuthProvider.credential with email and password, then call linkWithCredential with the current user and credential.

### Session Persistence

Import setPersistence, browserLocalPersistence, and browserSessionPersistence from firebase/auth. Call setPersistence with auth and browserLocalPersistence to persist across browser sessions as the default behavior. Call setPersistence with auth and browserSessionPersistence to clear the session on tab close.

---

## Security Rules Integration

### Firestore Rules

For Firestore security rules, use rules_version 2 and service cloud.firestore. In the databases match block, create a users collection match with userId wildcard. Allow read and write if request.auth is not null and request.auth.uid equals userId. Create an admin collection match with document wildcard. Allow read and write if request.auth.token.admin equals true.

### Storage Rules

For Firebase Storage security rules, use rules_version 2 and service firebase.storage. In the bucket match block, create a users path match with userId wildcard and allPaths wildcard. Allow read and write if request.auth is not null and request.auth.uid equals userId.

---

## Advanced Patterns

For advanced implementation patterns, see reference.md covering MFA, session cookies, Cloud Functions triggers, and Admin SDK setup. See examples.md for complete authentication services, React hooks, and Flutter providers.

---

## Resources

Extended Documentation:

- reference.md provides advanced patterns and configuration guides
- examples.md provides working code examples across platforms

Module Files:

- modules/social-auth.md covers social identity providers
- modules/phone-auth.md covers phone number authentication
- modules/custom-claims.md covers role-based access control

Firebase Official Resources:

- Firebase Console at console.firebase.google.com
- Authentication Documentation at firebase.google.com/docs/auth
- Security Rules Reference at firebase.google.com/docs/rules

Works Well With:

- moai-platform-firestore for Firestore database with auth-based security
- moai-lang-flutter for Flutter SDK for mobile Firebase Auth
- moai-lang-typescript for TypeScript patterns for Firebase SDK
- moai-domain-backend for backend architecture with Firebase Admin SDK

---

Status: Production Ready
Version: 2.1.0 (Modular Architecture)
Last Updated: 2026-01-11
Provider Coverage: Firebase Authentication Only
