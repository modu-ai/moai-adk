# Supabase PostgreSQL 16 Specialist

## Core Capabilities

PostgreSQL 16 provides latest PostgreSQL with full SQL support, JSONB for flexible document storage, advanced indexing, and extension ecosystem.

pgvector Extension enables AI embeddings storage with 1536-dimension OpenAI embeddings support, HNSW and IVFFlat indexes for fast similarity search, and vector similarity functions.

Row-Level Security provides automatic multi-tenant data isolation at database level using policies based on JWT claims.

Real-time Subscriptions enable live data sync via Postgres Changes for INSERT, UPDATE, DELETE events and Presence for online status tracking.

Edge Functions provide serverless Deno functions at the edge with JWT authentication, CORS support, and database connectivity.

Storage provides file storage with automatic image transformations, CDN delivery, and RLS integration.

Auth provides built-in authentication with JWT integration, social providers, email/password, and phone auth.

## PostgreSQL and pgvector

Extension Setup requires executing CREATE EXTENSION IF NOT EXISTS vector; in SQL. Create documents table with embedding column as vector(1536) for OpenAI embeddings.

Vector Storage uses vector data type with dimensionality matching embedding model. OpenAI text-embedding-3-small uses 1536 dimensions.

Index Strategy involves creating HNSW index using CREATE INDEX ON documents USING hnsw (embedding vector_cosine_ops). HNSW provides faster search with higher memory usage. IVFFlat provides lower memory usage with slightly slower search.

Similarity Search uses <=> operator for cosine distance. Query finds nearest vectors using ORDER BY embedding <=> query_vector LIMIT k.

Hybrid Search combines vector similarity with full-text search using PostgreSQL tsvector. Weight and combine scores for ranking.

## Row-Level Security

Basic Tenant Isolation involves enabling RLS with ALTER table projects ENABLE ROW LEVEL SECURITY. Create policy using CREATE POLICY tenant_isolation ON projects FOR ALL USING (tenant_id = (auth.jwt()->>'tenant_id')::uuid).

Hierarchical Access uses policies with subqueries to check organization hierarchy. Allow access if user belongs to parent organization.

Role-Based Modification creates separate policies for SELECT, INSERT, UPDATE, DELETE operations. Allow read access to team members, write access to owners.

Service Role Bypass uses service_role key in server-side operations to bypass RLS for administrative tasks. Never expose service_role key to clients.

Policy Testing involves testing policies with different JWT tokens. Use supabase.auth.setAuth() to simulate different users.

## Real-time and Presence

Postgres Changes Subscription uses supabase.channel('custom-channel').on('postgres_changes', { event: '*', schema: 'public', table: 'messages' }, payload => console.log(payload)).subscribe().

Filtered Changes specify event type (INSERT, UPDATE, DELETE) and filter by column using filter: { column: 'value' }.

Presence Tracking uses channel.track({ online: true }) and channel.subscribe().presence events include join, leave, and sync.

Collaborative Features include cursor position sharing, typing indicators, and online status. Presence state syncs across all connected clients.

Subscription Management involves storing unsubscribe functions and calling them on component unmount. Use channel.unsubscribe() to clean up.

## Edge Functions

Basic Edge Function uses Deno runtime with serve() function. Import createClient from '@supabase/supabase-js' using esm.sh CDN.

CORS Configuration sets Access-Control-Allow-Origin header. Use Deno.env.get() for environment variables.

JWT Verification extracts token from Authorization header. Verify token using supabase.auth.getUser(jwt) in server-side context.

Rate Limiting implements in-memory rate limiting using Map. Track request count per IP and return 429 status for exceeded limits.

Database Access uses service_role key for admin operations in Edge Functions. Never use anon key in server-side context.

## Storage and CDN

File Upload uses supabase.storage.from('bucket').upload(path, file). Generate signed URLs for private files.

Image Transformation uses transformation URLs with resize, quality, and format options. Example: /storage/v1/object/public/bucket/image.jpg?width=400&quality=80.

Cache Control sets Cache-Control headers on upload. Use public buckets for CDN-accessible files.

RLS Integration applies RLS policies to storage buckets. Policy can reference storage.folder_path column for folder-level access.

## Auth Integration

Server-side Client uses createClient with SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY. Use service role for admin operations.

Cookie-based Sessions uses supabase.auth.signInWithPassword() and stores session in httpOnly cookie. Pass session token to server for validation.

Auth State Sync involves listening to auth state changes with supabase.auth.onAuthStateChange. Update UI and database connections on auth change.

Protected Routes check authentication status before rendering. Use supabase.auth.getUser() to validate session token.

## TypeScript Patterns

Next.js App Router uses createServerClient for server components and createClientComponentClient for client components. Use createRouteHandlerClient for route handlers.

Service Layer Abstraction creates database service functions that abstract Supabase client. Provides type safety and reusable queries.

Subscription Management uses React hooks to manage real-time subscriptions. Handle connection lifecycle and error recovery.

Type-safe Operations uses TypeScript types generated from database schema. Use supabase gen types typescript command.

## Best Practices

Performance optimization includes using HNSW indexes for vectors, Supavisor for connection pooling in serverless environments, and prepared statements for repeated queries.

Security practices include always enabling RLS on all tables, verifying JWT tokens in server-side code, and using service_role key only in Edge Functions.

Migration management uses Supabase CLI with supabase migration new and supabase db push commands. Never modify schema directly in dashboard for production.

Error handling includes checking for errors in response data. Use try-catch for async operations and handle network errors gracefully.

---

Status: Production Ready
Platform: Supabase PostgreSQL 16
Version: 2.1.0
Coverage: PostgreSQL 16, pgvector, RLS, Real-time, Edge Functions, Storage
