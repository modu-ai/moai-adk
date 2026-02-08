---
name: moai-platform-database-cloud
description: >
  Cloud database platform specialist covering Neon (serverless PostgreSQL), Supabase (PostgreSQL 16 with real-time),
  and Firebase Firestore (NoSQL with offline sync). Use when choosing cloud databases, setting up serverless
  PostgreSQL, implementing real-time subscriptions, configuring offline-first apps, or evaluating database
  platforms. Supports branching (Neon), real-time (Supabase), and mobile-first (Firestore).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Bash(psql:*) Bash(npm:*) Bash(npx:*) Bash(neonctl:*) Bash(firebase:*) Bash(supabase:*) Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.0.0"
  category: "platform"
  status: "active"
  updated: "2026-02-09"
  modularized: "true"
  tags: "database, postgresql, nosql, serverless, real-time, offline, cloud"
  context7-libraries: "/neondatabase/neon, /supabase/supabase, /firebase/firebase-docs"
  related-skills: "moai-platform-auth, moai-lang-typescript, moai-domain-backend"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4500

# MoAI Extension: Triggers
triggers:
  keywords: ["neon", "supabase", "firestore", "cloud database", "serverless postgresql", "real-time database", "offline sync", "pgvector", "rls", "database branching", "vector database", "nosql", "mobile database"]
  agents: ["expert-backend", "expert-devops", "manager-spec"]
  phases: ["plan", "run"]
  languages: ["typescript", "javascript", "python", "go"]
---

# moai-platform-database-cloud: Cloud Database Platform Specialist

## Quick Reference

Cloud Database Platform Coverage: Consolidated expertise for Neon (serverless PostgreSQL), Supabase (PostgreSQL 16 with real-time), and Firebase Firestore (NoSQL with offline sync).

### Platform Comparison

Neon provides serverless PostgreSQL with auto-scaling, database branching, and scale-to-zero compute for cost optimization. Best for serverless applications, preview environments, and edge deployment with connection pooling.

Supabase provides PostgreSQL 16 with pgvector for AI/ML, Row-Level Security for multi-tenant apps, real-time subscriptions, and integrated auth/storage. Best for full-stack apps requiring real-time features and vector search.

Firestore provides NoSQL document database with real-time sync, offline caching, Security Rules, and mobile-optimized SDKs. Best for mobile-first apps, offline-first architecture, and cross-platform development.

### Quick Decision Guide

Need serverless PostgreSQL with auto-scaling: Use Neon.

Need database branching for CI/CD: Use Neon branching.

Need edge-compatible database: Use Neon with connection pooling.

Need vector search for AI/ML: Use Supabase with pgvector.

Need Row-Level Security: Use Supabase RLS policies.

Need real-time subscriptions: Use Supabase real-time or Firestore listeners.

Need offline-first mobile app: Use Firestore with offline persistence.

Need SQL with real-time features: Use Supabase.

Need NoSQL flexibility: Use Firestore.

### Database Type Selection

SQL vs NoSQL Decision: Choose SQL (Neon, Supabase) for structured data with complex relationships, ACID transactions, and query flexibility. Choose NoSQL (Firestore) for flexible schemas, offline-first mobile apps, and real-time sync across clients.

PostgreSQL Variants: Choose Neon for serverless auto-scaling and branching. Choose Supabase for integrated features (auth, storage, real-time) and pgvector.

---

## Platform Selection Matrix

### Use Case Alignment

Serverless Applications benefit from Neon's auto-scaling and scale-to-zero that reduce costs significantly.

Multi-tenant SaaS benefits from Supabase Row-Level Security providing automatic tenant isolation.

AI/ML Applications benefit from Supabase pgvector for vector embeddings and similarity search.

Real-time Collaboration benefits from Supabase Postgres Changes or Firestore real-time listeners.

Mobile-First Apps benefit from Firestore offline caching and mobile-optimized SDKs.

Preview Environments benefit from Neon database branching for per-PR databases.

Edge Deployment benefits from Neon connection pooling for edge runtime compatibility.

Cross-Platform Apps benefit from Firestore unified SDKs across iOS, Android, Web, and Flutter.

### Feature Comparison

Serverless Compute: Neon (auto-scaling, scale-to-zero), Supabase (Supavisor pooling), Firestore (built-in serverless).

Database Branching: Neon (instant copy-on-write), Supabase (not available), Firestore (not available).

Vector Search: Neon (via pgvector extension), Supabase (native pgvector with HNSW), Firestore (not available).

Real-time Subscriptions: Neon (via logical replication), Supabase (native Postgres Changes), Firestore (native listeners).

Offline Support: Neon (not available), Supabase (limited), Firestore (first-class with IndexedDB).

Security Model: Neon (connection-level), Supabase (Row-Level Security), Firestore (Security Rules).

Mobile SDKs: Neon (community drivers), Supabase (TypeScript/JS native), Firestore (first-party mobile SDKs).

### Pricing Comparison

Neon Free Tier: 3GB storage, 100 compute hours/month, scale-to-zero idle free.

Supabase Free Tier: 500MB database, 1GB file storage, 2GB bandwidth/month, 50K MAU.

Firestore Free Tier: 1GB storage, 50K daily reads, 20K daily writes, real-time listeners included.

---

## Common Database Patterns

### Connection Management

Neon Serverless Driver requires @neondatabase/serverless package with neon function for query execution. Use DATABASE_URL for direct connection and DATABASE_URL_POOLED for serverless/edge compatibility.

Supabase Client uses @supabase/supabase-js with createClient function. Environment variables SUPABASE_URL and SUPABASE_ANON_KEY for client-side, SUPABASE_SERVICE_ROLE_KEY for server-side.

Firestore Client uses firebase/app and firebase/firestore with initializeFirestore. Enable offline persistence with persistentLocalCache and persistentMultipleTabManager for multi-tab support.

### Migration Strategy

Use Supabase CLI with supabase migration new and supabase db push for Supabase schema management.

Use neonctl or Neon API for database branching and reset operations in Neon.

Use Firebase CLI with firebase deploy --only firestore:rules for Firestore Security Rules deployment.

### ORM Integration

Neon supports Drizzle ORM with drizzle-orm/neon-http adapter, Prisma with @prisma/adapter-neon, and direct SQL with @neondatabase/serverless.

Supabase supports direct SQL with @supabase/supabase-js client, Drizzle ORM with Postgres driver, and Prisma with connection string.

Firestore SDK is the primary interface with no ORM abstraction layer needed.

---

## Context7 Documentation Access

For latest platform documentation, use the Context7 MCP tools:

Neon: Use mcp__context7__resolve-library-id with query "neondatabase/neon" to get the library ID, then mcp__context7__get-library-docs with topics like "branching", "connection pooling", or "auto-scaling".

Supabase: Use mcp__context7__resolve-library-id with query "supabase" to get the library ID, then mcp__context7__get-library-docs with topics like "postgresql pgvector", "row-level-security", or "realtime".

Firestore: Use mcp__context7__resolve-library-id with query "firebase" to get the library ID, then mcp__context7__get-library-docs with topics like "firestore security-rules", "firestore offline", or "firestore real-time".

---

## Platform-Specific Deep Dives

For platform-specific implementation patterns, architecture details, and advanced features, consult the reference files:

Neon Serverless PostgreSQL at reference/neon.md covers database branching workflows, auto-scaling configuration, PITR and backups, serverless driver usage, Drizzle and Prisma ORM integration, and edge deployment patterns.

Supabase PostgreSQL 16 at reference/supabase.md covers pgvector for AI/ML, Row-Level Security policies, real-time subscriptions and presence, Edge Functions with Deno, Storage with CDN, auth integration, and TypeScript client patterns.

Firebase Firestore at reference/firestore.md covers NoSQL document modeling, real-time listeners with metadata, offline caching and sync, Security Rules with custom claims, transactions and batch operations, composite indexes, and mobile SDK patterns.

For comparative analysis and migration guidance, see reference/comparison.md which covers SQL vs NoSQL decision matrix, PostgreSQL variant comparison, migration strategies between platforms, feature parity mapping, and cost optimization strategies.

---

## Works Well With

- moai-platform-auth for authentication integration with Supabase Auth or Firebase Auth
- moai-lang-typescript for TypeScript patterns across all platforms
- moai-lang-flutter for Firestore mobile SDK patterns
- moai-domain-backend for backend architecture with database integration
- moai-domain-mobile for mobile-first database patterns
- moai-quality-security for security best practices (RLS policies, Security Rules)

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2026-02-09
Version: 2.0.0 (Consolidated)
Platforms: Neon, Supabase, Firestore
