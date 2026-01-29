---
name: moai-platform-firestore
description: >
  Firebase Firestore specialist covering NoSQL patterns, real-time sync, offline
  caching, and Security Rules. Use when building mobile-first apps with offline
  support, implementing real-time listeners, or configuring Firestore security.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Bash Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.1.0"
  category: "platform"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "firestore, firebase, nosql, realtime, offline, mobile"
  context7-libraries: "/firebase/firebase-docs"
  related-skills: "moai-platform-firebase-auth, moai-lang-flutter, moai-lang-typescript"

# MoAI Extension: Triggers
triggers:
  keywords: ["firestore", "nosql", "real-time", "offline", "mobile database", "security rules", "sync"]
---

# moai-platform-firestore: Firebase Firestore Specialist

## Quick Reference

Firebase Firestore Expertise: NoSQL document database with real-time synchronization, offline-first architecture, Security Rules, Cloud Functions triggers, and mobile-optimized SDKs.

### Core Capabilities

Real-time Sync provides automatic synchronization across all connected clients.

Offline Caching uses IndexedDB persistence with automatic sync when online.

Security Rules offer declarative field-level access control.

Cloud Functions enable document triggers for server-side processing.

Composite Indexes support complex query optimization.

### When to Use Firestore

Use Firestore for mobile-first applications with offline support, real-time collaborative features, cross-platform apps across iOS, Android, Web, and Flutter, projects requiring Google Cloud integration, and apps with flexible evolving data structures.

### Context7 Documentation Access

Step 1: Resolve the Firebase library ID using mcp__context7__resolve-library-id with libraryName "firebase" to get the Context7-compatible library ID.

Step 2: Fetch Firestore documentation using mcp__context7__get-library-docs with the resolved ID. Set topic to specific areas such as "firestore security-rules", "firestore offline", or "firestore real-time". Allocate 6000 to 8000 tokens for comprehensive coverage.

### Module Index

This skill is organized into specialized modules for deep implementation guidance:

Security Rules Module at modules/security-rules.md covers basic rule structure and syntax, role-based access control patterns, custom claims integration, field-level validation, and testing and debugging rules.

Offline and Caching Module at modules/offline-cache.md covers persistent cache configuration, multi-tab manager setup, cache size optimization, sync status handling, and network state management.

Real-time Listeners Module at modules/realtime-listeners.md covers snapshot listener patterns, metadata handling for pending writes, query subscription optimization, listener lifecycle management, and error handling strategies.

Transactions Module at modules/transactions.md covers atomic batch operations, transaction patterns and constraints, distributed counter implementation, conflict resolution strategies, and performance considerations.

---

## Implementation Guide

### Firestore Initialization with Offline Persistence

Import initializeApp from firebase/app and initializeFirestore, persistentLocalCache, persistentMultipleTabManager, and CACHE_SIZE_UNLIMITED from firebase/firestore. Create the app with initializeApp passing the config object containing apiKey, authDomain, and projectId from environment variables. Export the db instance created with initializeFirestore, passing the app and a localCache configuration using persistentLocalCache with tabManager set to persistentMultipleTabManager() and cacheSizeBytes set to CACHE_SIZE_UNLIMITED.

### Basic Security Rules Structure

Define Security Rules using rules_version 2 and service cloud.firestore. In the databases match block, create a users collection match with userId wildcard allowing read and write if request.auth.uid equals userId. Create a documents collection match with docId wildcard. Allow read if the document isPublic field is true, or if request.auth.uid equals the ownerId field, or if request.auth.uid is in the collaborators array. Allow create if request.auth is not null and the new document ownerId equals request.auth.uid. Allow update and delete if request.auth.uid equals the document ownerId.

### Real-time Listener with Metadata

Import collection, query, where, orderBy, and onSnapshot from firebase/firestore. Create a subscribeToDocuments function taking userId and a callback. Build a query on the documents collection filtering where collaborators array-contains userId and ordering by createdAt descending. Call onSnapshot with the query, includeMetadataChanges set to true, and a callback that maps snapshot.docs to objects containing id, document data spread, _pending set to doc.metadata.hasPendingWrites, and _fromCache set to doc.metadata.fromCache. Return the unsubscribe function.

### Composite Indexes Configuration

Define indexes in firestore.indexes.json with an indexes array. Each index object specifies collectionGroup, queryScope as COLLECTION, and a fields array. Fields specify fieldPath and either order as ASCENDING or DESCENDING, or arrayConfig as CONTAINS for array fields. Create indexes for common query patterns such as organizationId with createdAt descending, or tags array with createdAt descending.

---

## Performance and Pricing

### Performance Characteristics

Read Latency ranges from 50 to 200 milliseconds depending on region.

Write Latency ranges from 100 to 300 milliseconds.

Real-time Propagation takes 100 to 500 milliseconds.

Offline Sync occurs automatically on reconnection.

### Free Tier (2024)

Storage: 1GB. Daily Reads: 50,000. Daily Writes: 20,000. Daily Deletes: 20,000.

---

## Works Well With

- moai-platform-firebase-auth for Firebase Authentication integration
- moai-lang-flutter for Flutter SDK patterns
- moai-lang-typescript for TypeScript client patterns
- moai-domain-mobile for mobile architecture patterns
- moai-quality-security for Security Rules best practices

---

## Additional Resources

- reference.md provides extended documentation and best practices
- examples.md provides complete working code examples
- modules/security-rules.md provides Security Rules deep dive
- modules/offline-cache.md covers offline persistence patterns
- modules/realtime-listeners.md covers real-time subscription patterns
- modules/transactions.md covers batch and transaction operations

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2026-01-11
Platform: Firebase Firestore
