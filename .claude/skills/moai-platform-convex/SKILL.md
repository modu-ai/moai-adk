---
name: moai-platform-convex
description: >
  Convex real-time backend specialist covering TypeScript-first reactive patterns,
  optimistic updates, server functions, and file storage. Use when building real-time
  collaborative apps, implementing reactive queries, or integrating with Clerk/Auth0.
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
  tags: "convex, realtime, reactive, typescript, optimistic-updates"
  context7-libraries: "/get-convex/convex"
  related-skills: "moai-platform-supabase, moai-lang-typescript, moai-domain-frontend"

# MoAI Extension: Triggers
triggers:
  keywords: ["convex", "real-time", "reactive", "optimistic updates", "collaborative", "server functions"]
---

# Convex Real-time Backend Specialist

Convex is a real-time reactive backend platform with TypeScript-first design, automatic caching, and optimistic updates.

---

## Quick Reference

### When to Use Convex

Use Convex for real-time collaborative applications including docs, whiteboards, and chat. Choose Convex for apps requiring instant UI updates without manual refetching. Select Convex for TypeScript-first projects needing end-to-end type safety. Use Convex for applications with complex optimistic update requirements.

### Core Concepts

Server Functions include queries for read operations, mutations for write operations, and actions for external API calls.

Reactive Queries automatically re-execute when underlying data changes.

Optimistic Updates provide instant UI updates before server confirmation.

Automatic Caching provides built-in query result caching with intelligent invalidation.

### Quick Start

Initialize a new Convex project using npm create convex@latest. Start the development server with npx convex dev.

### Context7 Library

Use mcp__context7__resolve-library-id with "convex" to get the library ID, then use mcp__context7__get-library-docs for latest documentation.

---

## Module Index

This skill is organized into specialized modules for detailed implementation guidance:

### Reactive Queries Module

Location: modules/reactive-queries.md

Covers real-time reactive query patterns including basic query structure, index-based queries, search indexes, pagination patterns, React integration with useQuery hooks, and optimistic updates.

### Server Functions Module

Location: modules/server-functions.md

Covers mutations for write operations, actions for external API integration, internal functions for server-to-server calls, scheduled functions using crons, and HTTP endpoints for webhooks.

### Authentication Integration Module

Location: modules/auth-integration.md

Covers Clerk integration, Auth0 integration, server-side authentication patterns, authorization and role-based access control, and session management.

### File Storage Module

Location: modules/file-storage.md

Covers file upload workflows, storage URL generation, client-side upload with progress, file display components, and file management operations.

---

## Implementation Guide

### Project Structure

A Convex project contains a convex directory with _generated subdirectory for auto-generated types and API, schema.ts for database schema definition, a functions subdirectory for server functions organized by domain, optional http.ts for HTTP endpoints, and optional crons.ts for scheduled jobs. The src directory contains ConvexProvider.tsx for client setup.

### Schema Definition

The schema.ts file imports defineSchema and defineTable from convex/server and v from convex/values. The default export uses defineSchema to define tables. Each table is defined with defineTable specifying field types such as v.string(), v.boolean(), and v.number(). Tables can have indexes defined using the index method with index name and field array. Search indexes use the searchIndex method specifying searchField and filterFields.

### Validators

The v module from convex/values provides primitive validators including v.string(), v.number(), v.boolean(), v.null(), v.int64(), and v.bytes().

Complex types include v.array() for arrays of a type, v.object() for objects with specified fields, v.union() for union types with v.literal() options, and v.optional() for optional fields.

Reference validators include v.id() for references to table documents.

### React Client Setup

The ConvexProvider component imports ConvexProvider and ConvexReactClient from convex/react, ConvexProviderWithClerk from convex/react-clerk, and useAuth from @clerk/clerk-react. Create a ConvexReactClient instance with the VITE_CONVEX_URL environment variable. The Providers component wraps children with ConvexProviderWithClerk passing the client and useAuth hook.

### React Hooks Usage

Import useQuery and useMutation from convex/react and api from the generated API module. In a component, call useQuery with the query function and parameters to get reactive data. Call useMutation to get a mutation function. Handle the undefined state during loading. Render the data and use the mutation function in event handlers.

---

## Best Practices

Query Optimization:

- Use indexes for all filtered queries
- Prefer paginated queries for large datasets
- Use search indexes for full-text search
- Leverage automatic caching

Mutation Design:

- Keep mutations focused and atomic
- Use internal mutations for multi-step operations
- Validate all inputs with the v module
- Always check authorization

Error Handling:

- Use ConvexError for structured errors
- Check for undefined during loading states
- Handle optimistic update rollbacks

---

## Works Well With

- moai-platform-supabase for alternative PostgreSQL-based backend
- moai-lang-typescript for TypeScript patterns and best practices
- moai-domain-frontend for React integration patterns
- moai-platform-clerk for Clerk authentication patterns
- moai-platform-auth0 for Auth0 authentication patterns

---

## Resources

- Official Documentation at docs.convex.dev
- Context7 Library at /get-convex/convex
- GitHub at github.com/get-convex/convex

---

Status: Production Ready
Version: 2.1.0
Last Updated: 2026-01-11
Platform: Convex Real-time Backend
