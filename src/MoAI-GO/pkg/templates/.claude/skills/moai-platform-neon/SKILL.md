---
name: moai-platform-neon
description: >
  Neon serverless PostgreSQL specialist covering auto-scaling, database branching, PITR,
  and connection pooling. Use when building serverless apps needing PostgreSQL, implementing
  preview environments, or optimizing database costs.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Bash Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.1.0"
  category: "platform"
  modularized: "true"
  status: "active"
  updated: "2026-01-11"
  tags: "neon, postgresql, serverless, branching, auto-scaling"
  context7-libraries: "/neondatabase/neon"
  related-skills: "moai-platform-supabase, moai-lang-typescript, moai-domain-database"

# MoAI Extension: Triggers
triggers:
  keywords: ["neon", "serverless", "postgresql", "database branching", "auto-scaling", "pitr", "connection pooling"]
---

# moai-platform-neon: Neon Serverless PostgreSQL Specialist

## Quick Reference

Neon Serverless PostgreSQL Expertise: Specialized knowledge for Neon serverless PostgreSQL covering auto-scaling, scale-to-zero compute, database branching, Point-in-Time Recovery, and modern ORM integration.

### Core Capabilities

Serverless Compute provides auto-scaling PostgreSQL with scale-to-zero for cost optimization.

Database Branching enables instant copy-on-write branches for dev, staging, and preview environments.

Point-in-Time Recovery offers 30-day PITR with instant restore to any timestamp.

Connection Pooling provides built-in connection pooler for serverless and edge compatibility.

PostgreSQL 16 ensures full PostgreSQL 16 compatibility with extensions support.

### Quick Decision Guide

Need serverless PostgreSQL with auto-scaling: Use Neon.

Need database branching for CI/CD: Use Neon branching.

Need edge-compatible database: Use Neon with connection pooling.

Need instant preview environments: Use Neon branch per PR.

Need vector search: Consider Supabase with pgvector instead.

### Context7 Documentation Access

To fetch the latest Neon documentation:

Step 1: Resolve the library ID using mcp__context7__resolve-library-id with library name "neondatabase/neon".

Step 2: Fetch documentation using mcp__context7__get-library-docs with the resolved Context7 ID, specifying topics like "branching", "connection pooling", or "auto-scaling".

---

## Module Index

This skill is organized into focused modules for progressive disclosure:

### Core Modules

Database Branching at modules/branching-workflows.md covers copy-on-write branches for development, preview environments, and CI/CD integration with GitHub Actions.

Auto-Scaling and Compute at modules/auto-scaling.md covers compute unit configuration, scale-to-zero settings, and cost optimization strategies.

Connection Pooling at modules/connection-pooling.md covers serverless connection pooling for edge runtimes, WebSocket configuration, and pool sizing.

PITR and Backups at modules/pitr-backups.md covers point-in-time recovery, branch restoration, and backup strategies.

### Supporting Files

Reference Guide at reference.md provides API reference, environment configuration, and provider comparison.

Code Examples at examples.md provides complete working examples for common integration patterns.

---

## Implementation Guide

### Setup and Configuration

Package Installation requires @neondatabase/serverless from npm. Optionally install drizzle-orm for Drizzle ORM integration or @prisma/client and prisma for Prisma ORM integration.

Environment Configuration requires DATABASE_URL for direct connection used for migrations, formatted as postgresql://user:pass@ep-xxx.region.neon.tech/dbname?sslmode=require. DATABASE_URL_POOLED provides pooled connection for serverless and edge, formatted as postgresql://user:pass@ep-xxx-pooler.region.neon.tech/dbname?sslmode=require. NEON_API_KEY provides the Neon API key for branching operations. NEON_PROJECT_ID provides the project identifier.

### Serverless Driver Usage

For basic query execution, import neon from @neondatabase/serverless. Create the sql function by calling neon with the DATABASE_URL environment variable. Execute simple queries using tagged template literals with the sql function. For parameterized queries that are SQL injection safe, include variables inside the template literal. The driver supports transaction operations using sql.transaction with an array of SQL statements.

### Drizzle ORM Integration

For schema definition, import table and column types from drizzle-orm/pg-core. Define tables using pgTable with column definitions. Use uuid for UUIDs with primaryKey and defaultRandom, text for strings with notNull and unique modifiers, timestamp for dates with defaultNow, and jsonb for JSON columns.

For Drizzle client setup, import neon from @neondatabase/serverless, drizzle from drizzle-orm/neon-http, and the schema module. Create the sql function with neon and the DATABASE_URL. Export the db instance created with drizzle passing sql and the schema. Execute queries using db.select().from(schema.tableName).

### Prisma ORM Integration

For Prisma with Neon Serverless Driver, import Pool and neonConfig from @neondatabase/serverless, PrismaNeon from @prisma/adapter-neon, and PrismaClient from @prisma/client. Set neonConfig.webSocketConstructor to the ws module. Create a Pool with the DATABASE_URL connection string. Create an adapter with PrismaNeon passing the pool. Export the prisma instance created with PrismaClient passing the adapter.

---

## Provider Decision Guide

### When to Use Neon

Serverless Applications benefit from auto-scaling and scale-to-zero that reduce costs significantly.

Preview Environments benefit from instant branching that enables per-PR databases with production data.

Edge Deployment benefits from connection pooling that provides edge runtime compatibility.

Development Workflow benefits from branching from production for realistic development data.

Cost Optimization benefits from paying only for active compute time with scale-to-zero.

### When to Consider Alternatives

Need Vector Search: Consider Supabase with pgvector or dedicated vector database.

Need Real-time Subscriptions: Consider Supabase or Convex for real-time features.

Need NoSQL Flexibility: Consider Firestore or Convex for document storage.

Need Built-in Auth: Consider Supabase for integrated authentication.

### Pricing Reference

Free Tier provides 3GB storage and 100 compute hours per month.

Pro Tier provides usage-based pricing with additional storage and compute.

Scale-to-Zero incurs no charges during idle periods.

---

## Works Well With

- moai-platform-supabase for alternative when RLS or pgvector needed
- moai-lang-typescript for TypeScript patterns for Drizzle and Prisma
- moai-domain-backend for backend architecture with database integration
- moai-domain-database for general database patterns and optimization

---

Status: Production Ready
Version: 2.1.0
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2026-01-11
Technology: Neon Serverless PostgreSQL
