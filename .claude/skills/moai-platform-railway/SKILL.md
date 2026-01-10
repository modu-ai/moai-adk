---
name: "moai-platform-railway"
description: "Railway container deployment specialist covering Docker, multi-service architectures, persistent volumes, and auto-scaling. Use when deploying containerized full-stack applications, configuring multi-region deployments, or setting up persistent storage."
version: 2.1.0
category: "platform"
modularized: true
user-invocable: false
tags: ['railway', 'docker', 'containers', 'multi-service', 'auto-scaling']
context7-libraries: ['/railwayapp/railway']
related-skills: ['moai-platform-vercel', 'moai-domain-backend']
updated: 2026-01-11
status: "active"
allowed-tools:
  - Read
  - Write
  - Bash
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# moai-platform-railway: Container Deployment Specialist

## Quick Reference

Railway Platform Core: Container-first deployment platform with Docker and Railpack builds, multi-service architectures, persistent volumes, private networking, and auto-scaling capabilities.

### Railway Optimal Use Cases

Container Workloads:

- Full-stack containerized applications with custom runtimes
- Multi-service architectures with inter-service communication
- Backend services requiring persistent connections such as WebSocket and gRPC
- Database-backed applications with managed PostgreSQL, MySQL, and Redis

Infrastructure Requirements:

- Persistent volume storage for stateful workloads
- Private networking for secure service mesh
- Multi-region deployment for global availability
- Auto-scaling based on CPU, memory, or request metrics

### Build Strategy Selection

Docker Build is optimal for custom system dependencies, multi-stage builds, and specific base images.

Railpack Build is optimal for standard runtimes including Node.js, Python, and Go, providing zero-config and faster builds.

Note: Nixpacks is deprecated. New services default to Railpack.

### Key CLI Commands

Railway CLI workflow begins with railway login, railway init, and railway link commands. Use railway up to deploy the current directory. Use railway up with the detach flag to deploy without waiting for logs. Use railway variables with the set flag to configure environment variables. Use railway logs with the service flag to view service logs. Use railway rollback with the previous flag to revert to the previous deployment.

---

## Module Index

### Docker Deployment

Module at modules/docker-deployment.md covers multi-stage Dockerfiles for Node.js, Python, Go, and Rust. Includes build optimization, image size reduction, health checks, and Railpack migration.

### Multi-Service Architecture

Module at modules/multi-service.md covers monorepo deployments, service communication, variable references, private networking, and message queue patterns.

### Volumes and Storage

Module at modules/volumes-storage.md covers persistent volume configuration, file storage service, SQLite on volumes, and backup/restore patterns.

### Networking and Domains

Module at modules/networking-domains.md covers private networking, custom domains with SSL, multi-region deployment, auto-scaling, and WebSocket support.

---

## Implementation Guide

### Phase 1: Project Setup

Begin with railway login, railway init, and railway link commands to authenticate, initialize, and link to a project.

The railway.toml configuration file specifies build settings with builder set to DOCKERFILE and dockerfilePath pointing to the Dockerfile. Deploy settings include healthcheckPath for the health endpoint, healthcheckTimeout in seconds, restartPolicyType for restart behavior, and numReplicas for instance count. Resource settings under deploy.resources specify memory allocation and CPU limits.

### Phase 2: Container Configuration

For a quick start Node.js Dockerfile, the builder stage uses node:20-alpine as base, sets WORKDIR to /app, copies package files, runs npm ci, copies source files, and runs npm run build.

The runner stage also uses node:20-alpine, sets WORKDIR to /app, sets NODE_ENV to production, creates a nodejs group and appuser user with specific IDs, copies node_modules and dist from the builder stage, switches to appuser, exposes port 3000, and runs the node command with dist/main.js.

For detailed patterns, see modules/docker-deployment.md.

### Phase 3: Multi-Service Setup

Service Variable References use the double curly brace syntax with dollar sign prefix. Reference Postgres.DATABASE_URL for database connections, Redis.REDIS_URL for Redis connections, and api.RAILWAY_PRIVATE_DOMAIN for internal service communication.

For Private Networking, create a getInternalUrl helper function that takes a service name and optional port defaulting to 3000. The function retrieves the RAILWAY_PRIVATE_DOMAIN environment variable for the service name in uppercase, returning the internal URL with http protocol and port, or localhost fallback.

For detailed patterns, see modules/multi-service.md.

### Phase 4: Storage and Scaling

Volume Configuration in railway.toml uses a volumes array section specifying mountPath for the directory path, name for the volume identifier, and size for the storage allocation.

Auto-Scaling Configuration in deploy.scaling section specifies minReplicas, maxReplicas, and targetCPUUtilization percentage.

Multi-Region Configuration uses deploy.regions array with entries containing name for the region identifier and replicas for the instance count per region.

---

## CI/CD Integration

For GitHub Actions integration, create a workflow file named Railway Deploy with trigger on push to main branch. The deploy job runs on ubuntu-latest, checks out code, installs Railway CLI globally using npm, and runs railway up with detach flag using the RAILWAY_TOKEN secret from GitHub secrets.

---

## Context7 Documentation Access

Step 1: Use mcp__context7__resolve-library-id with "railway" to get the library ID.

Step 2: Use mcp__context7__get-library-docs with the resolved ID and specific topic.

---

## Works Well With

- moai-platform-vercel for edge deployment for frontend
- moai-domain-backend for backend architecture patterns
- moai-lang-python for Python FastAPI deployment
- moai-lang-typescript for TypeScript Node.js patterns
- moai-lang-go for Go service deployment

---

## Additional Resources

- reference.md provides extended documentation and configuration
- examples.md provides working code examples for common scenarios

---

Status: Production Ready
Version: 2.1.0
Updated: 2026-01-11
