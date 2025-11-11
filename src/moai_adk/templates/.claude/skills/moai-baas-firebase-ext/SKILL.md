---
name: moai-baas-firebase-ext
version: 2.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: Firebase Firestore database integration, real-time data synchronization, and Firebase Authentication. Use when building mobile/web apps with real-time features, implementing Firebase auth, or managing NoSQL databases.
keywords: ['firebase', 'firestore', 'realtime', 'nosql', 'google-cloud']
allowed-tools:
  - Read
  - Bash
  - WebFetch
---

# Firebase Extension Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-firebase-ext |
| **Version** | 2.0.0 (2025-11-11) |
| **Allowed tools** | Read, Bash, WebFetch |
| **Auto-load** | On demand when Firebase integration detected |
| **Tier** | BaaS Extension |

---

## What It Does

Firebase Firestore database integration, real-time data synchronization, and Firebase Authentication.

**Key capabilities**:
- ✅ Firestore NoSQL database
- ✅ Real-time data sync
- ✅ Firebase Authentication
- ✅ Cloud Functions
- ✅ Firebase hosting

---

## When to Use

- ✅ Building mobile/web apps with real-time features
- ✅ Implementing Firebase authentication
- ✅ Managing NoSQL databases
- ✅ Creating serverless applications

---

## Core Firebase Patterns

### Database Architecture
1. **Firestore**: NoSQL document database
2. **Real-time Listeners**: Live data updates
3. **Collection Structure**: Hierarchical data organization
4. **Indexing**: Query optimization
5. **Security Rules**: Data access control

### Authentication Integration
- **Firebase Auth**: Email/password, social providers
- **Custom Claims**: User role management
- **Multi-tenant**: Separate user data
- **Anonymous Auth**: Guest users
- **Phone Auth**: SMS verification

---

## Dependencies

- Firebase project and configuration
- Firebase SDK for your platform
- Google Cloud account
- Firestore database rules

---

## Works Well With

- `moai-baas-foundation` (BaaS patterns)
- `moai-domain-database` (Database patterns)
- `moai-domain-frontend` (Frontend integration)

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, Firestore patterns
- **v1.0.0** (2025-10-22): Initial Firebase integration

---

**End of Skill** | Updated 2025-11-11
