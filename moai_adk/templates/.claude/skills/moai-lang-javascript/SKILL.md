---
name: moai-lang-javascript
description: >
  JavaScript ES2024+ development specialist covering Node.js 22 LTS, Bun 1.x (serve, SQLite, S3, shell, test), Deno 2.x, testing (Vitest, Jest), linting (ESLint 9, Biome), and backend frameworks (Express, Fastify, Hono). Use when developing JavaScript APIs, web applications, or Node.js projects.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob Bash(node:*) Bash(npm:*) Bash(npx:*) Bash(yarn:*) Bash(pnpm:*) Bash(bun:*) Bash(deno:*) Bash(jest:*) Bash(vitest:*) Bash(eslint:*) Bash(prettier:*) Bash(biome:*) mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.2.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "false"
  tags: "language, javascript, nodejs, bun, deno, vitest, eslint, express"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["JavaScript", "Node.js", "Bun", "Deno", "Express", "Fastify", "Hono", "Vitest", "Jest", "ESLint", ".js", "package.json"]
  languages: ["javascript", "js"]
---

## Quick Reference (30 seconds)

JavaScript ES2024+ Development Specialist - Modern JavaScript with Node.js 22 LTS, multiple runtimes, and contemporary tooling.

Auto-Triggers: Files with .js, .mjs, or .cjs extensions, package.json, Node.js projects, JavaScript discussions

Core Stack:

- ES2024+: Set methods, Promise.withResolvers, immutable arrays, import attributes
- Node.js 22 LTS: Native TypeScript, built-in WebSocket, stable watch mode
- Runtimes: Node.js 20 and 22 LTS, Deno 2.x, Bun 1.x
- Testing: Vitest, Jest, Node.js test runner
- Linting: ESLint 9 flat config, Biome
- Bundlers: Vite, esbuild, Rollup
- Frameworks: Express, Fastify, Hono, Koa

Quick Commands:

Create a Vite project using npm create vite with latest tag, project name, and vanilla template. Initialize with modern tooling using npm init and npm install with D flag for vitest, eslint, and eslint/js. Run with Node.js watch mode using node with watch flag. Run TypeScript directly in Node.js 22+ using node with experimental-strip-types flag.

---

## Implementation Guide (5 minutes)

### ES2024 Key Features

Set Operations:

Create setA with values 1, 2, 3, 4 and setB with values 3, 4, 5, 6. Call setA.intersection with setB to get Set containing 3 and 4. Call setA.union with setB to get Set containing 1 through 6. Call setA.difference with setB to get Set containing 1 and 2. Call setA.symmetricDifference with setB to get Set containing 1, 2, 5, and 6. Call isSubsetOf, isSupersetOf, and isDisjointFrom methods for set comparisons.

Promise.withResolvers:

Create a createDeferred function that destructures promise, resolve, and reject from Promise.withResolvers call. Return an object with these three properties. Create a deferred instance, set a timeout to resolve with done after 1000 milliseconds, and await the promise for the result.

Immutable Array Methods:

Create original array with values 3, 1, 4, 1, 5. Call toSorted to get new sorted array without mutating original. Call toReversed to get new reversed array. Call toSpliced with index 1, delete count 2, and insert value 9. Call with method at index 2 with value 99 to get new array with replaced element. The original array remains unchanged.

Object.groupBy and Map.groupBy:

Create items array with objects containing type and name properties. Call Object.groupBy with items and a function that returns item.type to get an object with arrays grouped by type. Call Map.groupBy with the same arguments to get a Map with type keys and array values.

### ES2025 Features

Import Attributes for JSON Modules:

Import config from config.json with type attribute set to json. Import styles from styles.css with type attribute set to css. Access config.apiUrl property.

RegExp.escape:

Create userInput string with special characters like parentheses. Call RegExp.escape with userInput to get escaped pattern string. Create new RegExp with the safe pattern.

### Node.js 22 LTS Features

Built-in WebSocket Client:

Create new WebSocket with wss URL. Add event listener for open event that sends JSON stringified message. Add event listener for message event that parses event.data as JSON and logs the received data.

Native TypeScript Support Experimental:

Run .ts files directly in Node.js 22.6+ using node with experimental-strip-types flag. In Node.js 22.18+, type stripping is enabled by default so files can be run directly.

Watch Mode Stable:

Use node with watch flag for auto-restart on file changes. Use watch-path flag multiple times to watch specific directories like src and config.

Permission Model:

Use node with permission flag and allow-fs-read set to a specific path to restrict file system access. Use allow-net flag with domain name to restrict network access.

### Backend Frameworks

Express Traditional Pattern:

Import express. Create app by calling express function. Use express.json middleware. Create a get endpoint at api/users that awaits database query and responds with json. Create a post endpoint that creates a user and responds with status 201 and json. Call listen on port 3000 with callback logging server running.

Fastify High Performance Pattern:

Import Fastify. Create fastify instance with logger set to true. Define userSchema with body containing type object, required array with name and email, and properties with validation constraints. Create a post endpoint with schema option and async handler that creates user and returns with code 201. Call listen with port 3000.

Hono Edge-First Pattern:

Import Hono and middleware functions. Create app instance. Use logger middleware for all routes. Use cors middleware for api routes. Create get endpoint at api/users that awaits database query and returns c.json. Create post endpoint with validator middleware that checks for required fields, then creates user and returns c.json with status 201. Export app as default.

### Testing with Vitest

Configuration:

Create vitest.config.js with defineConfig. Set test object with globals true, environment node, and coverage with provider v8 and reporters for text, json, and html.

Test Example:

In test file, import describe, it, expect, vi, and beforeEach from vitest. Import functions to test. Create describe block for User Service. In beforeEach, call vi.clearAllMocks. Create it block for should create a user that awaits createUser and expects result to match object with name and email, and id to be defined. Create it block for should throw on invalid email that expects createUser to reject with Invalid email error.

### ESLint 9 Flat Config

Create eslint.config.js. Import js from eslint/js and globals. Export array with js.configs.recommended followed by object with languageOptions containing ecmaVersion 2025, sourceType module, and globals merged from globals.node and globals.es2025. Set rules for no-unused-vars with error and args ignore pattern, no-console with warn and allowed methods, prefer-const as error, and no-var as error.

### Biome All-in-One

Create biome.json with schema URL. Enable organizeImports. Set linter enabled with recommended rules. Set formatter enabled with indentStyle space and indentWidth 2. Under javascript.formatter, set quoteStyle to single and semicolons to always.

---

## Advanced Patterns

For comprehensive documentation including advanced async patterns, module system details, performance optimization, and production deployment configurations, see:

- reference.md for complete API reference, Context7 library mappings, and package manager comparison
- examples.md for production-ready code examples, full-stack patterns, and testing templates

### Context7 Integration

For Node.js documentation, use context7 get library docs with nodejs/node and topics like esm modules async. For Express, use expressjs/express with middleware routing. For Fastify, use fastify/fastify with plugins hooks. For Hono, use honojs/hono with middleware validators. For Vitest, use vitest-dev/vitest with mocking coverage.

---

## Works Well With

- moai-lang-typescript for TypeScript integration and type checking with JSDoc
- moai-domain-backend for API design and microservices architecture
- moai-domain-database for database integration and ORM patterns
- moai-workflow-testing for DDD workflows and testing strategies
- moai-foundation-quality for code quality standards
- moai-essentials-debug for debugging JavaScript applications

---

## Quick Troubleshooting

Module System Issues:

Check package.json for type field. ESM uses type module with import and export. CommonJS uses type commonjs or omits the field and uses require and module.exports.

Node.js Version Check:

Run node with version flag for 20.x or 22.x LTS. Run npm with version flag for 10.x or later.

Common Fixes:

Clear npm cache with npm cache clean using force flag. Delete node_modules and package-lock.json then run npm install. Fix permission issues by setting npm config prefix to home directory npm-global folder.

ESM and CommonJS Interop:

To import CommonJS from ESM, import the default then destructure named exports from it. For dynamic import in CommonJS, use await import and destructure the default property.

---

Last Updated: 2026-01-11
Status: Active (v1.2.0)
