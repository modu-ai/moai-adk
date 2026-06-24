---
name: moai-ref-secops
description: >
  DevSecOps, container, and API operational defensive security reference: CI/CD
  pipeline hardening, secret scanning, IaC misconfiguration detection, SAST/DAST
  integration, container image scanning, Kubernetes RBAC hardening, container-escape
  defense, runtime threat detection, OWASP API Top 10 operational defense, WAF rule
  tuning, and GraphQL/REST depth and rate limiting. Agent-extending skill that
  amplifies backend, security, and platform-engineering work with production-grade
  defensive patterns for pipelines, containers, and running APIs.
  NOT for: offensive techniques (exploit execution, container-escape attack steps,
  privilege-escalation procedures, attack tooling), dev-time web-app OWASP Top 10
  (see moai-ref-owasp-checklist), LLM/AI security (see moai-ref-llm-security),
  supply-chain provenance and signing (see moai-ref-supply-chain), or general API
  design (see moai-ref-api-patterns).

when_to_use: >
  Use when hardening a CI/CD pipeline, scanning infrastructure-as-code for
  misconfiguration, hardening a container image or Kubernetes cluster, defending
  against container escape, writing runtime-detection rules, enforcing operational
  API defenses (BOLA detection in production, rate-limit enforcement, server-side
  authorization, WAF tuning), or limiting GraphQL/REST query depth and complexity.
  Loads as background knowledge for DevSecOps review, container-security hardening,
  and API operational-defense tasks across any language ecosystem.

user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-06-24"
  tags: "devsecops, container, kubernetes, rbac, api-security, owasp-api, cicd, iac-scanning, runtime-detection, waf, reference"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000
---

# DevSecOps, Container, and API Operational Security Reference

Defensive practitioner reference for the operational layer of a system — the
pipeline that builds it, the container and orchestrator that run it, and the API
surface it exposes at runtime. Every section is framed as defense, hardening,
detection, or verification: it describes the misconfiguration, how to detect it,
and how to prevent it, never how to exploit it.

This skill is split into three modules by sub-domain. The overview below gives the
shared threat model and an entry point; the depth lives in the modules. The threat
model is operational: a pipeline can be subverted to inject a build step, a container
can be misconfigured into a host breakout, and a running API can leak one tenant's
data to another. The defenses establish least privilege, isolation, detection, and
runtime authorization at each layer.

## Three Sub-Domains (modules)

| Sub-domain | Module | Covers |
|------------|--------|--------|
| DevSecOps | [modules/devsecops.md](modules/devsecops.md) | CI/CD pipeline hardening, secret scanning, IaC misconfiguration detection (Terraform / CloudFormation), SAST/DAST integration |
| Container | [modules/container.md](modules/container.md) | Image scanning, Kubernetes RBAC hardening, container-escape defense (seccomp / AppArmor / read-only root / non-root), runtime threat detection |
| API operational | [modules/api-ops.md](modules/api-ops.md) | OWASP API Top 10 operational defense (BOLA / broken-auth detection in production, rate-limit enforcement, server-side authorization), WAF rule tuning, GraphQL/REST depth and complexity limiting |

## Operational Trust Boundaries

The core defensive insight: each operational layer is a boundary where an attacker
who has reached it can move to the next. Defense-in-depth means each boundary
assumes the one before it may have failed.

| Boundary | Operational risk | Primary defense | Module |
|----------|------------------|-----------------|--------|
| Source → pipeline | Injected build step, leaked secret, poisoned runner | Signed pipeline, secret scanning, runner isolation | DevSecOps |
| Pipeline → infra config | Misconfigured cloud resource (open bucket, permissive IAM) | IaC scanning before apply | DevSecOps |
| Image → registry | Vulnerable base image, embedded secret | Image scanning + admission control | Container |
| Container → host | Container escape to the node | seccomp, AppArmor, read-only root, non-root, drop capabilities | Container |
| Cluster identity → resources | Over-privileged ServiceAccount, broad RoleBinding | Least-privilege RBAC, PodSecurity admission | Container |
| Client → API | Broken object/function authorization, resource exhaustion | Server-side authorization, rate limiting, WAF | API operational |

## Ecosystem-Neutral Tooling

This skill names tools at the category level. Where a tool is the de-facto
cross-ecosystem standard for its category, it is named but framed as "the standard
\<category\> tool" — the concept transfers to any equivalent tool. No single language
ecosystem is privileged; pipeline examples are CLI-shape, not language-specific source.

## Distinction from moai-ref-owasp-checklist (operational, not dev-time)

The API-operational module covers the **runtime/operational** half of API security —
detecting Broken Object Level Authorization in production traffic, enforcing rate
limits at the gateway, tuning a WAF, limiting query depth on a live endpoint. The
**dev-time** half — secure-coding patterns a developer applies while writing the
endpoint (parameterized queries, input validation, authentication design, security
headers) — lives in `moai-ref-owasp-checklist`. This skill keeps its coverage
operational and does not duplicate them.

## Cross-References

- `moai-ref-owasp-checklist` — dev-time web-app OWASP Top 10, authentication
  patterns, input validation, HTTP security headers (the development-time surface;
  this skill is the operational surface).
- `moai-ref-supply-chain` — SBOM, SLSA provenance, Sigstore signing, dependency
  hygiene (the supply-chain surface that image scanning and pipeline signing here
  draw on for artifact provenance).
- `moai-ref-llm-security` — AI/LLM defensive security (the LLM-specific operational
  surface, distinct from the general API/container/pipeline surface here).
- `moai-ref-api-patterns` — REST/GraphQL API design and error handling (the
  API-design surface, distinct from the API-operational-defense surface here).

## Defensive Severity Levels

| Level | Label | Action | Example |
|-------|-------|--------|---------|
| P0 | CRITICAL | Block release | Container runs privileged with host mounts; API endpoint has no server-side object-ownership check |
| P1 | HIGH | Fix before merge | Pipeline secret exposed in build logs; ServiceAccount bound to cluster-admin |
| P2 | MEDIUM | Fix within iteration | No image scan in the pipeline; no rate limit on an expensive endpoint |
| P3 | LOW | Track in backlog | WAF in detection-only mode where enforcement is warranted; no GraphQL depth limit on a shallow schema |

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "The cluster is internal, so RBAC hardening is optional" | An internal cluster is reachable from any compromised pod. Least-privilege RBAC is the boundary that stops a single compromised workload from owning the cluster. |
| "The container runs our own trusted image, so escape defense is overkill" | A trusted image with a vulnerable dependency is still an escape risk. seccomp, non-root, and read-only root are cheap and assume the image may be compromised. |
| "Authorization is handled in the frontend, the API can trust the client" | The client is never a trust boundary. Every operational API defense — BOLA in particular — requires server-side object-ownership checks on every request. |
| "We scan images at build time, runtime detection is redundant" | Build-time scanning misses a zero-day disclosed after deploy and misses runtime behavior. Runtime detection catches what static scanning cannot. |
| "Rate limiting hurts legitimate users, we will add it if abused" | Unbounded consumption is an OWASP API Top 10 risk that is exploited before it is noticed. Rate limits and query-complexity caps are the control, not an afterthought. |

**Assume-breach-by-layer**: each operational layer assumes the one before it may have
failed — the pipeline assumes a poisoned source, the container a vulnerable image, the
API a hostile client. Defense is the control at each layer, not trust in the previous.

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- A container runs as privileged, as root, with a writable root filesystem, or with host paths mounted
- A Kubernetes ServiceAccount is bound to cluster-admin or a wildcard Role
- An API endpoint returns an object by ID with no server-side ownership check (Broken Object Level Authorization)
- The CI/CD pipeline echoes secrets into build logs or runs untrusted code on a shared, non-isolated runner
- Infrastructure-as-code is applied to production without a misconfiguration scan
- A public-facing API has no rate limit, no query-depth limit, and no request-size cap
- Runtime detection is absent, so a post-deploy compromise produces no alert

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] CI/CD pipeline runs secret scanning and IaC misconfiguration scanning before deploy; runners are isolated (see [modules/devsecops.md](modules/devsecops.md))
- [ ] Container images are scanned and gated by admission control; containers run non-root with a read-only root filesystem, dropped capabilities, and a seccomp profile (see [modules/container.md](modules/container.md))
- [ ] Kubernetes RBAC follows least privilege; no ServiceAccount is bound to cluster-admin; PodSecurity admission is enforced (see [modules/container.md](modules/container.md))
- [ ] Runtime-detection rules alert on container-escape and credential-access behavior; alerts route to a responder (see [modules/container.md](modules/container.md))
- [ ] Every API endpoint enforces server-side object-level and function-level authorization; BOLA is detectable in production traffic (see [modules/api-ops.md](modules/api-ops.md))
- [ ] Rate limits, request-size caps, and GraphQL/REST depth and complexity limits are enforced at the gateway (see [modules/api-ops.md](modules/api-ops.md))
- [ ] Operational coverage does not duplicate the dev-time patterns in `moai-ref-owasp-checklist`; the two are consulted together

<!-- moai:evolvable-end -->
