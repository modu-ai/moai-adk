---
name: moai-baas-cloudflare-ext
version: 2.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: Cloudflare Workers edge computing, CDN optimization, and security services integration. Use when implementing edge functions, CDN strategies, or global application deployment.
keywords: ['cloudflare', 'edge-computing', 'cdn', 'workers', 'security']
allowed-tools:
  - Read
  - Bash
  - WebFetch
---

# Cloudflare Extension Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-cloudflare-ext |
| **Version** | 2.0.0 (2025-11-11) |
| **Allowed tools** | Read, Bash, WebFetch |
| **Auto-load** | On demand when Cloudflare integration detected |
| **Tier** | BaaS Extension |

---

## What It Does

Cloudflare Workers edge computing, CDN optimization, and security services integration.

**Key capabilities**:
- ✅ Edge function deployment
- ✅ Global CDN optimization
- ✅ DDoS protection
- ✅ SSL/TLS management
- ✅ DNS and routing

---

## When to Use

- ✅ Implementing edge functions
- ✅ Optimizing CDN strategies
- ✅ Global application deployment
- ✅ Enhancing security posture

---

## Core Cloudflare Patterns

### Edge Computing Architecture
1. **Workers**: Serverless functions at the edge
2. **KV Storage**: Global key-value storage
3. **Durable Objects**: Persistent edge state
4. **Pages**: Jamstack deployment
5. **R2 Storage**: S3-compatible object storage

### Performance & Security
- **CDN**: Global content delivery
- **Argo Smart Routing**: Optimized traffic routing
- **WAF**: Web application firewall
- **Bot Management**: Automated bot protection
- **Load Balancing**: Traffic distribution

---

## Dependencies

- Cloudflare account and domain
- Workers runtime knowledge
- API integration
- Security configuration

---

## Works Well With

- `moai-baas-foundation` (BaaS patterns)
- `moai-domain-security` (Security patterns)
- `moai-essentials-perf` (Performance optimization)

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, edge computing patterns
- **v1.0.0** (2025-10-22): Initial Cloudflare integration

---

**End of Skill** | Updated 2025-11-11
