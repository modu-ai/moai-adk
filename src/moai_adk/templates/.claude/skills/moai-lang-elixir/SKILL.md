---
name: moai-lang-elixir
description: >
  Elixir 1.17+ development specialist covering Phoenix 1.7, LiveView, Ecto,
  and OTP patterns. Use when developing real-time applications, distributed
  systems, or Phoenix projects.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob Bash(mix:*) Bash(elixir:*) Bash(iex:*) Bash(erl:*) mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "language, elixir, phoenix, liveview, ecto, otp, genserver"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["Elixir", "Phoenix", "LiveView", "Ecto", "OTP", "GenServer", ".ex", ".exs", "mix.exs"]
  languages: ["elixir"]
---

## Quick Reference (30 seconds)

Elixir 1.17+ Development Specialist - Phoenix 1.7, LiveView, Ecto, OTP patterns, and functional programming.

Auto-Triggers: `.ex`, `.exs` files, `mix.exs`, `config/`, Phoenix/LiveView discussions

Core Capabilities:

- Elixir 1.17: Pattern matching, pipes, protocols, behaviours, macros
- Phoenix 1.7: Controllers, LiveView, Channels, PubSub, Verified Routes
- Ecto: Schemas, Changesets, Queries, Migrations, Multi
- OTP: GenServer, Supervisor, Agent, Task, Registry
- ExUnit: Testing with setup, describe, async
- Mix: Build tool, tasks, releases
- Oban: Background job processing

### Quick Patterns

Phoenix Controller: Define a module using MyAppWeb with :controller. Create alias for the context module like MyApp.Accounts. Define action functions like show taking conn and params map with destructured id. Fetch data using the context function with bang like get_user! and render the template passing the data as assigns.

For create actions, pattern match on the context result tuple. On ok tuple, use pipe operator to put_flash with success message and redirect using ~p sigil for verified routes. On error tuple with Ecto.Changeset, render the form template passing the changeset.

Ecto Schema with Changeset: Define a module using Ecto.Schema and importing Ecto.Changeset. Define the schema block with field declarations including types like :string and virtual fields. Create a changeset function taking the struct and attrs, using pipe operator to chain cast with the list of fields to cast, validate_required, validate_format with regex, validate_length with min option, and unique_constraint.

GenServer Pattern: Define a module using GenServer. Create start_link taking initial_value and calling GenServer.start_link with __MODULE__, initial_value, and name option. Define client API functions that call GenServer.call with __MODULE__ and the message atom. Implement init callback returning ok tuple with state. Implement handle_call callbacks for each message, returning reply tuple with response and new state.

---

## Implementation Guide (5 minutes)

### Elixir 1.17 Features

Pattern Matching Advanced: Define function heads with pattern matching on map keys and types. Use guard clauses with when to add constraints like is_binary or byte_size checks. Add a catch-all clause returning error tuple for invalid inputs.

Pipe Operator with Error Handling: Use the with special form for chaining operations that may fail. Pattern match each step with left arrow, and on successful completion of all steps, return the final ok tuple. In the else block, handle error tuples by returning them unchanged.

Protocols for Polymorphism: Define a protocol with @doc and function specification using defprotocol. Implement the protocol for specific types using defimpl with for: option. Each implementation provides the specific behavior for that type.

### Phoenix 1.7 Patterns

LiveView Component: Define a module using MyAppWeb with :live_view. Implement mount callback taking params, session, and socket, returning ok tuple with assigned state. Implement handle_event callback for user interactions, returning noreply tuple with updated socket using update helper. Implement render callback with assigns, using ~H sigil for HEEx templates with assigns accessed via @ prefix.

LiveView Form with Changesets: In mount, create initial changeset and assign using to_form helper. Implement validate event handler that creates changeset with :validate action and reassigns the form. Implement save event handler that calls context create function, on success using put_flash and push_navigate, on error reassigning the form with error changeset.

Phoenix Channels: Define a module using MyAppWeb with :channel. Implement join callback matching on topic pattern with angle brackets for dynamic segments. Use send with self() for after_join messages. Implement handle_info for server-side messages using push. Implement handle_in for client messages using broadcast! to send to all subscribers.

Verified Routes: Define routes in router.ex within scope blocks using live macro for LiveView routes. Use ~p sigil for verified routes in templates and controllers, with interpolation syntax for dynamic segments.

### Ecto Patterns

Multi for Transactions: Use Ecto.Multi.new() and pipe through operations using Ecto.Multi.update with name atoms and changeset functions. Use Ecto.Multi.insert with function callback when needing results from previous steps. Pipe final Multi to Repo.transaction() which returns ok or error tuple with named results.

Query Composition: Create a query module with composable query functions. Define a base function returning from expression. Create filter functions with default parameter for query, returning modified from expression with where clause. Chain functions with pipe operator before passing to Repo.all.

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:

- Production deployment with releases
- Distributed systems with libcluster
- Advanced LiveView patterns (streams, components)
- OTP supervision trees and dynamic supervisors
- Telemetry and observability
- Security best practices
- CI/CD integration patterns

See:

- [Advanced Patterns](modules/advanced-patterns.md) - Complete advanced patterns guide

---

## Context7 Library Mappings

- /elixir-lang/elixir - Elixir language documentation
- /phoenixframework/phoenix - Phoenix web framework
- /phoenixframework/phoenix_live_view - LiveView real-time UI
- /elixir-ecto/ecto - Database wrapper and query language
- /sorentwo/oban - Background job processing

---

## Works Well With

- `moai-domain-backend` - REST API and microservices architecture
- `moai-domain-database` - SQL patterns and query optimization
- `moai-workflow-testing` - DDD and testing strategies
- `moai-essentials-debug` - AI-powered debugging
- `moai-platform-deploy` - Deployment and infrastructure

---

## Troubleshooting

Common Issues:

Elixir Version Check: Run elixir --version to verify 1.17+ and mix --version for Mix build tool version.

Dependency Issues: Run mix deps.get to fetch dependencies, mix deps.compile to compile them, and mix clean to remove build artifacts.

Database Migrations: Run mix ecto.create to create database, mix ecto.migrate to run migrations, and mix ecto.rollback to rollback last migration.

Phoenix Server: Run mix phx.server to start the server, iex -S mix phx.server to start with IEx shell, and MIX_ENV=prod mix release to build production release.

LiveView Not Loading:

- Check websocket connection in browser developer console
- Verify endpoint configuration includes websocket transport
- Ensure Phoenix.LiveView is listed in mix.exs dependencies

---

Last Updated: 2026-01-11
Status: Active (v1.1.0)
