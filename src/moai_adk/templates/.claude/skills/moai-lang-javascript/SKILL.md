---
name: moai-lang-javascript
description: Enterprise JavaScript for Node.js 22 LTS and browser development with ES2025 features, async/await patterns, Express 4.21+, and modern runtime optimization
---

## Quick Reference

Enterprise JavaScript with Node.js 22 LTS (Jod, support until 2027), npm 11, ES2025 features, async operations, Express framework patterns, and module systems for server-side and browser development.

**Key Facts**:
- **Runtime**: Node.js 22.11.0 LTS (Oct 2024 - Apr 2027)
- **Package Manager**: npm 11.x with workspace support and provenance attestation
- **Web Framework**: Express 4.21+ (mature), Fastify 5.x (performance-focused)
- **Module Systems**: ES Modules (standard), CommonJS (legacy compatibility)
- **Testing**: Vitest 2.x, Jest 30.x for comprehensive testing
- **When to Use**: Server-side applications, scripting, tooling, REST APIs, real-time applications

---

## What It Does

JavaScript provides a universal platform for server-side and browser development. It excels in:

- **Server-Side Development**: Node.js for REST APIs, microservices, real-time applications
- **Async Programming**: Native async/await, Promises, EventEmitter for concurrent operations
- **Package Ecosystem**: npm registry with 2M+ packages for rapid development
- **Full-Stack Development**: Single language for frontend and backend
- **Real-Time Applications**: WebSockets, Server-Sent Events, event-driven architecture
- **Build Tools**: Webpack, Turbopack, esbuild for module bundling and optimization

JavaScript's event-driven, non-blocking I/O model makes it ideal for I/O-bound applications, real-time systems, and full-stack development with shared code between client and server.

---

## When to Use

**Use JavaScript when**:
- Building REST APIs and microservices with Node.js (Express, Fastify, Hapi)
- Need full-stack development with shared client/server code
- Creating real-time applications (WebSockets, chat, collaboration tools)
- Developing command-line tools and build scripts
- Rapid prototyping with extensive npm package ecosystem

**Avoid JavaScript when**:
- CPU-intensive computations required (use Go, Rust, C++ instead)
- Strong type safety is mandatory (consider TypeScript instead)
- Predictable memory usage is critical (Node.js has garbage collection overhead)

---

## Key Features

1. **Async/Await**: Clean asynchronous programming without callback hell
2. **ES Modules**: Standard module system with import/export syntax
3. **Event-Driven Architecture**: EventEmitter for decoupled component communication
4. **npm Ecosystem**: World's largest package registry with 2M+ reusable modules
5. **Promise API**: Native promise support for async operations and error handling
6. **Arrow Functions**: Concise function syntax with lexical `this` binding
7. **Destructuring**: Extract values from arrays/objects with clean syntax
8. **Template Literals**: String interpolation with embedded expressions

---

## Works Well With

- `moai-lang-typescript` — TypeScript for type-safe JavaScript development
  - Best for: Large codebases requiring type safety and better IDE support

- `moai-domain-backend` — Backend architecture patterns and REST API design
  - Best for: Express/Fastify server architecture, middleware patterns

- `moai-domain-frontend` — Frontend frameworks (React, Vue, Angular)
  - Best for: Full-stack JavaScript applications with shared code

- `moai-domain-database` — MongoDB, PostgreSQL integration patterns
  - Best for: Database operations with Node.js drivers and ORMs

- `moai-domain-cloud` — AWS Lambda, Google Cloud Functions deployment
  - Best for: Serverless JavaScript applications and edge functions

---

## Core Concepts

### Event-Driven Programming
JavaScript uses an event loop for non-blocking I/O. Operations like HTTP requests, file I/O, and timers don't block the main thread. Instead, callbacks or promises handle completion, allowing thousands of concurrent operations with a single thread.

### Callback vs Promise vs Async/Await
JavaScript async patterns evolved from callbacks (error-prone "callback hell") to Promises (chainable `.then()`) to async/await (synchronous-looking async code). Modern JavaScript prefers async/await for readability and error handling.

### Module Systems
JavaScript supports two module systems: ES Modules (standard `import/export`) and CommonJS (legacy `require/module.exports`). Node.js 22 defaults to ES Modules for `.mjs` files or when `"type": "module"` is in package.json.

---

## Best Practices

### ✅ DO

1. **Use Async/Await**: Prefer async/await over raw promises or callbacks for readability
   - Reason: Cleaner code, better error handling with try/catch

2. **Handle Promise Rejections**: Always catch unhandled promise rejections to prevent crashes
   - Reason: Node.js will exit on unhandled rejections in future versions

3. **Use ES Modules**: Adopt ES Modules (`import/export`) for new projects
   - Reason: Standard syntax, better tree-shaking, future-proof

4. **Environment Variables**: Use environment variables for configuration (dotenv)
   - Reason: Separates configuration from code, enables different environments

5. **Error Handling**: Implement proper error handling with try/catch and error middleware
   - Reason: Prevents crashes, provides debugging information

### ❌ DON'T

1. **Blocking Operations**: Never use synchronous I/O in production (fs.readFileSync)
   - Reason: Blocks event loop, kills concurrency for all requests

2. **Callback Hell**: Avoid deeply nested callbacks (pyramid of doom)
   - Reason: Hard to read, maintain, and debug; use async/await instead

3. **Global Variables**: Don't pollute global scope with variables
   - Reason: Causes naming conflicts, makes code unpredictable

4. **Ignoring Errors**: Never swallow errors with empty catch blocks
   - Reason: Hides bugs, makes debugging impossible

5. **Mixed Module Systems**: Don't mix ES Modules and CommonJS in same file
   - Reason: Causes compatibility issues, complicates build process

---

## Implementation Guide

(See previous content for implementation details)

---

## Advanced Patterns

(See previous content for advanced patterns)

---

## Context7 Integration

### Related Libraries & Tools
- [Node.js](/nodejs/node): JavaScript runtime for server-side development
- [Express](/expressjs/express): Fast, unopinionated web framework for Node.js
- [Fastify](/fastify/fastify): High-performance web framework for Node.js

### Official Documentation
- [Node.js Documentation](https://nodejs.org/docs/latest/api/)
- [Express Documentation](https://expressjs.com/)
- [Fastify Documentation](https://fastify.dev/)

### Version-Specific Guides
Latest stable version: Node.js 22.11.0 LTS, Express 4.21.x, Fastify 5.x
- [Node.js 22 Release Notes](https://nodejs.org/en/blog/release/v22.11.0)
- [npm 11 Changelog](https://github.com/npm/cli/releases)
- [Express 4.21 Release Notes](https://github.com/expressjs/express/releases)

---

**Last Updated**: 2025-11-22  
**Status**: Production Ready  
**Version**: 4.0.0
