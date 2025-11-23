---
name: moai-baas-firebase-ext
description: Enterprise Firebase Platform with AI-powered Google Cloud integration
version: 1.0.1
modularized: true
---

## 📊 Skill Metadata

**Name**: moai-baas-firebase-ext
**Domain**: Backend-as-a-Service (Firebase + Google Cloud)
**Freedom Level**: high
**Target Users**: Backend engineers, full-stack developers, cloud architects
**Invocation**: Skill("moai-baas-firebase-ext")
**Progressive Disclosure**: SKILL.md (core) → modules/ (detailed implementations)
**Last Updated**: 2025-11-23
**Modularized**: true

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Design and implement enterprise Firebase architectures with Google Cloud integration.

**Firebase Services**:
1. **Firestore** - NoSQL document database with real-time sync
2. **Authentication** - Multi-provider auth (social, enterprise, custom)
3. **Cloud Functions** - Serverless backend functions
4. **Cloud Storage** - Object storage with CDN and metadata
5. **Firebase Hosting** - Global web hosting with SSL

**Google Cloud Integration**:
- BigQuery for analytics
- Cloud Run for scalable containers
- Cloud Monitoring for observability
- Cloud Build for CI/CD

---

## 📚 Core Patterns (5-10 minutes)

### Pattern 1: Firebase Initialization

**Key Concept**: Proper Firebase Admin SDK initialization with configuration management

**When to Use**:
- Setting up backend services
- Initializing in cloud functions
- Server-side authentication

**Core Steps**:
1. Initialize Admin SDK with service account credentials
2. Get references to Firestore, Auth, Functions, Storage
3. Handle multiple app instances (check with getApps())
4. Export service references for module use

### Pattern 2: Firestore CRUD Operations

**Key Concept**: Efficient document operations with batch processing

**Common Operations**:
- **Create**: `setDoc()` with new document ID
- **Read**: `getDoc()` for single, `getDocs()` for query
- **Update**: `updateDoc()` for partial updates, `merge: true` for set
- **Delete**: `deleteDoc()` or batch operations

**Batch Pattern**:
```
1. Create batch reference
2. Add multiple operations (set, update, delete)
3. Commit atomically
4. All-or-nothing transaction
```

### Pattern 3: Real-Time Subscriptions

**Key Concept**: Listen to live data changes with query filters

**Pattern Components**:
1. Define collection path
2. Apply filters (where, orderBy, limit)
3. Create onSnapshot listener
4. Unsubscribe when done

**Multi-Filter Queries**:
- Firestore requires index creation for complex queries
- Use composite indexes for multiple where/orderBy clauses
- Consider query cost (every subscription reads matching documents)

### Pattern 4: Authentication with Custom Claims

**Key Concept**: Extend user authentication with role-based access

**Flow**:
1. Create/authenticate user in Firebase Auth
2. Set custom claims (roles, permissions, metadata)
3. Claims appear in ID token
4. Frontend verifies via claims in token
5. Backend validates claims before operations

**Example Claims**:
```
{
  "role": "admin",
  "permissions": ["read", "write", "delete"],
  "organization": "acme-corp"
}
```

### Pattern 5: Cloud Functions Integration

**Key Concept**: Serverless backend functions triggered by events

**Trigger Types**:
1. **HTTPS** - Direct HTTP requests
2. **Firestore** - Document create, update, delete
3. **Authentication** - User created, deleted
4. **Storage** - File uploaded, deleted
5. **Scheduled** - Cron-based execution

**Typical Flow**:
1. Define function with trigger
2. Handle event payload
3. Perform business logic
4. Return response or update database
5. Firebase manages scaling automatically

---

## 📖 Advanced Documentation

This Skill uses Progressive Disclosure. For detailed implementations:

- **[modules/typescript-implementation.md](modules/typescript-implementation.md)** - Complete TypeScript code examples
- **[modules/python-implementation.md](modules/python-implementation.md)** - Python Cloud Functions examples
- **[modules/advanced-patterns.md](modules/advanced-patterns.md)** - Query optimization, transactions, pagination
- **[modules/reference.md](modules/reference.md)** - API reference, troubleshooting, performance tuning

---

## 🏗️ Architecture Patterns

### Pattern: Scalable Real-Time App

```
Client (Web/Mobile)
    ↓
Firebase Authentication
    ↓
Firestore (Document DB)
    ↓
Cloud Functions (Business Logic)
    ↓
Cloud Storage (File Upload)
    ↓
BigQuery (Analytics)
```

### Pattern: Event-Driven Backend

```
Event Source (Firestore, Storage, Auth)
    ↓
Cloud Function Trigger
    ↓
Process Event
    ↓
Write Results to Firestore
    ↓
Emit to Pub/Sub (Optional)
    ↓
Update Client via Realtime Listener
```

---

## 🔗 Integration with Other Skills

**Complementary Skills**:
- Skill("moai-domain-backend") - Backend architecture patterns
- Skill("moai-domain-cloud") - Google Cloud infrastructure
- Skill("moai-domain-monitoring") - Firebase monitoring setup
- Skill("moai-security-identity") - Authentication best practices

**Typical Workflow**:
1. Use this Skill to design Firebase architecture
2. Use moai-domain-backend for overall API design
3. Use moai-domain-monitoring to setup observability
4. Use moai-security-identity for auth patterns

---

## 📈 Version History

**1.0.1** (2025-11-23)
- 🔄 Refactored with Progressive Disclosure pattern
- 📚 Code examples moved to modules/ for clarity
- ✨ Core patterns highlighted in SKILL.md
- ✨ Added integration with other Skills

**1.0.0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Enterprise Firebase patterns
- ✨ TypeScript + Python examples
- ✨ Google Cloud integration guide

---

**Maintained by**: alfred
**Domain**: Backend-as-a-Service (Firebase)
**Generated with**: MoAI-ADK Skill Factory
