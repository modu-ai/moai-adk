---
name: "moai-platform-supabase"
description: "Supabase specialist covering PostgreSQL 16, pgvector, RLS, real-time subscriptions, and Edge Functions. Use when building full-stack apps with Supabase backend."
version: 2.1.0
category: "platform"
modularized: true
user-invocable: false
tags: ['supabase', 'postgresql', 'pgvector', 'realtime', 'rls', 'edge-functions']
context7-libraries: ['/supabase/supabase']
related-skills: ['moai-platform-neon', 'moai-lang-typescript']
updated: 2026-01-11
status: "active"
allowed-tools:
  - Read
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# moai-platform-supabase: Supabase Platform Specialist

## Quick Reference

Supabase Full-Stack Platform: PostgreSQL 16 with pgvector for AI/vector search, Row-Level Security for multi-tenant apps, real-time subscriptions, Edge Functions with Deno runtime, and integrated Storage with transformations.

### Core Capabilities

PostgreSQL 16 provides latest PostgreSQL with full SQL support, JSONB, and advanced features.

pgvector Extension enables AI embeddings storage with HNSW and IVFFlat indexes for similarity search.

Row-Level Security provides automatic multi-tenant data isolation at database level.

Real-time Subscriptions enable live data sync via Postgres Changes and Presence.

Edge Functions provide serverless Deno functions at the edge.

Storage provides file storage with automatic image transformations.

Auth provides built-in authentication with JWT integration.

### When to Use Supabase

Use Supabase for multi-tenant SaaS applications requiring data isolation, AI/ML applications needing vector embeddings and similarity search, real-time collaborative features including presence and live updates, full-stack applications needing auth, database, and storage, and projects requiring PostgreSQL-specific features.

### Context7 Documentation Access

For latest Supabase API documentation, use the Context7 MCP tools:

Step 1 - Resolve library ID: Use mcp__context7__resolve-library-id with query "supabase" to get the Context7-compatible library ID.

Step 2 - Fetch documentation: Use mcp__context7__get-library-docs with the resolved library ID, specifying topic and token allocation.

Example topics include "postgresql pgvector", "row-level-security policies", "realtime subscriptions presence", "edge-functions deno", "storage transformations", and "auth jwt".

---

## Module Index

This skill uses progressive disclosure with specialized modules for detailed implementation patterns.

### Core Modules

postgresql-pgvector covers PostgreSQL 16 with pgvector extension for AI embeddings and semantic search. Topics include vector storage with 1536-dimension OpenAI embeddings, HNSW and IVFFlat index strategies, semantic search functions, and hybrid search combining vector and full-text.

row-level-security covers RLS policies for multi-tenant data isolation. Topics include basic tenant isolation patterns, hierarchical organization access, role-based modification policies, and service role bypass for server operations.

realtime-presence covers real-time subscriptions and presence tracking. Topics include Postgres Changes subscription patterns, filtered change listeners, presence state management, and collaborative cursor and typing indicators.

edge-functions covers serverless Deno functions at the edge. Topics include basic Edge Function with authentication, CORS header configuration, JWT token verification, and rate limiting implementation.

storage-cdn covers file storage with image transformations. Topics include file upload patterns, image transformation URLs, thumbnail generation, and cache control configuration.

auth-integration covers authentication patterns and JWT handling. Topics include server-side client creation, cookie-based session management, auth state synchronization, and protected route patterns.

typescript-patterns covers TypeScript client patterns and service layers. Topics include server-side client for Next.js App Router, service layer abstraction pattern, subscription management, and type-safe database operations.

---

## Quick Start Patterns

### Database Setup

To enable pgvector extension and create embeddings table, execute SQL to create extension vector if not exists. Create a documents table with id as UUID primary key defaulting to gen_random_uuid(), content as TEXT not null, embedding as vector with 1536 dimensions, metadata as JSONB defaulting to empty object, and created_at as TIMESTAMPTZ defaulting to NOW(). Create an index on documents embedding using hnsw with vector_cosine_ops operator class.

### Basic RLS Policy

To create a tenant isolation policy, alter the projects table to enable row level security. Create a policy named "tenant_isolation" on projects for all operations using a condition where tenant_id equals the tenant_id extracted from the auth.jwt() function cast to UUID.

### Real-time Subscription

For real-time subscription, create a channel using supabase.channel with a channel name. Chain the on method for postgres_changes with event set to asterisk for all events, schema set to public, and table set to messages. Provide a callback that logs the payload. Chain subscribe to activate the subscription.

### Edge Function Template

For Edge Function template, import serve from deno.land/std/http/server.ts and createClient from esm.sh/@supabase/supabase-js. Call serve with an async handler that creates a Supabase client using environment variables SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY. Process the request and return a new Response with JSON stringified success object.

---

## Best Practices

Performance: Use HNSW indexes for vectors and Supavisor for connection pooling in serverless environments.

Security: Always enable RLS, verify JWT tokens, and use service_role only in Edge Functions.

Migration: Use Supabase CLI with supabase migration new and supabase db push commands.

---

## Works Well With

- moai-platform-neon for alternative PostgreSQL for specific use cases
- moai-lang-typescript for TypeScript patterns for Supabase client
- moai-domain-backend for backend architecture integration
- moai-foundation-quality for security and RLS best practices
- moai-workflow-testing for test-driven development with Supabase

---

## Module References

For detailed implementation patterns, see the modules directory:

- modules/postgresql-pgvector.md covers complete vector search implementation
- modules/row-level-security.md covers multi-tenant RLS patterns
- modules/realtime-presence.md covers real-time collaboration features
- modules/edge-functions.md covers serverless function patterns
- modules/storage-cdn.md covers file storage and transformations
- modules/auth-integration.md covers authentication patterns
- modules/typescript-patterns.md covers TypeScript client architecture

For API reference summary, see reference.md. For full-stack templates, see examples.md.

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2026-01-11
Version: 2.1.0 (Modularized)
Coverage: PostgreSQL 16, pgvector, RLS, Real-time, Edge Functions, Storage
