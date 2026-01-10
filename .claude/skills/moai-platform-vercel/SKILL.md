---
name: "moai-platform-vercel"
description: "Vercel edge deployment specialist covering Edge Functions, Next.js optimization, preview deployments, ISR, and storage solutions. Use when deploying Next.js applications, implementing edge computing, or configuring Vercel platform features."
version: 2.1.0
category: "platform"
modularized: true
user-invocable: false
tags: ['vercel', 'edge', 'nextjs', 'isr', 'preview', 'cdn', 'kv', 'blob', 'postgres']
context7-libraries: "/vercel/next.js, /vercel/vercel"
related-skills: "moai-platform-railway, moai-lang-typescript, moai-domain-frontend"
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

# moai-platform-vercel: Vercel Edge Deployment Specialist

## Quick Reference

Vercel Optimization Focus: Edge-first deployment platform with global CDN, Next.js optimized runtime, managed storage solutions, and developer-centric preview workflows.

### Core Capabilities

Edge Functions provide global low-latency compute at 30+ edge locations with sub-50ms cold start for optimal user experience. Features include geo-based routing and personalization, and edge middleware for request/response transformation.

Next.js Optimized Runtime provides first-class Next.js support with automatic optimizations. Features include Server Components and App Router integration, streaming SSR for improved TTFB, and built-in image optimization with next/image.

Preview Deployments provide automatic PR-based preview URLs with branch-specific environment variables. Features include comment integration for PR reviews and instant rollback capabilities.

ISR (Incremental Static Regeneration) enables on-demand revalidation for dynamic content with stale-while-revalidate caching strategy. Features include tag-based cache invalidation and background regeneration without user impact.

Managed Storage includes Vercel KV for Redis-compatible key-value store, Vercel Blob for object storage for files, and Vercel Postgres for managed PostgreSQL database.

### Quick Decision Guide

Choose Vercel when Next.js is primary framework, edge performance is critical requirement, preview deployments needed for team collaboration, Web Vitals monitoring is priority, or serverless storage solutions needed.

---

## Module Index

This skill uses modular documentation for progressive disclosure. Load modules as needed:

### Edge Functions and Middleware

Module at modules/edge-functions.md covers Edge Runtime configuration, middleware patterns, geo-based content delivery, A/B testing at edge, and authentication at edge. Includes region selection, supported APIs, and cold start optimization.

### ISR and Caching Strategies

Module at modules/isr-caching.md covers Incremental Static Regeneration patterns, time-based and on-demand revalidation, tag-based cache invalidation, CDN cache headers, streaming with Suspense, and cache debugging techniques.

### Deployment Configuration

Module at modules/deployment-config.md covers vercel.json configuration, environment variables management, preview and production deployments, function memory and duration settings, cron jobs, monorepo setup, domain configuration, and security headers.

### Analytics and Speed Insights

Module at modules/analytics-speed.md covers Vercel Analytics integration, Speed Insights for Web Vitals, custom performance monitoring, Core Web Vitals optimization, and third-party analytics integration.

### KV, Blob, and Postgres Storage

Module at modules/kv-storage.md covers Vercel KV for Redis-compatible storage, Vercel Blob for object storage, Vercel Postgres for managed database, edge-compatible usage patterns, and common storage patterns.

---

## Implementation Guide

### Essential Configuration

The vercel.json configuration file contains a schema reference to openapi.vercel.sh/vercel.json, framework set to nextjs, regions array for deployment regions such as iad1, sfo1, and fra1, and functions configuration for API routes specifying memory in MB and maxDuration in seconds.

### Edge Function Quick Start

For an Edge Function route, create an api route file and export runtime as "edge". Export preferredRegion as an array of region codes. Export an async GET handler that takes a Request parameter, extracts the country from request.geo defaulting to "Unknown", and returns a Response.json with the country data.

### ISR Quick Start

For ISR implementation, export revalidate constant set to the revalidation interval in seconds. In the async page component, fetch data with the next configuration option containing tags array for cache invalidation. Return the component with fetched data.

### Analytics Integration

For Analytics integration in the root layout, import Analytics from @vercel/analytics/react and SpeedInsights from @vercel/speed-insights/next. In the RootLayout component body, render children followed by Analytics and SpeedInsights components.

---

## Advanced Patterns

### Blue-Green Deployment

Deploy new version, run smoke tests on preview URL, then switch production alias using Vercel SDK for zero-downtime releases.

### Monorepo with Turborepo

For monorepo configuration, set buildCommand to run turbo build from root with filter for the web package. Set installCommand to run pnpm install from root. Set framework to nextjs.

### Context7 Integration

Use Context7 MCP tools for latest documentation:

Step 1: Resolve library ID using mcp__context7__resolve-library-id with "vercel" or "next.js".

Step 2: Fetch documentation using mcp__context7__get-library-docs with resolved ID, specific topic, and appropriate token allocation of 5000 to 10000 tokens.

---

## Works Well With

- moai-platform-railway for container-based deployment alternative
- moai-lang-typescript for TypeScript patterns for Next.js
- moai-domain-frontend for React and Next.js component patterns
- moai-foundation-quality for deployment validation and testing
- moai-domain-database for database patterns for Vercel Postgres

---

## Additional Resources

- Reference Guide at reference.md provides complete CLI commands and configuration options
- Code Examples at examples.md provides production-ready code snippets

---

Status: Production Ready
Version: 2.1.0
Updated: 2026-01-11
