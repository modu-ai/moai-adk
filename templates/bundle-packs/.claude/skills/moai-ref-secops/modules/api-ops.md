# API Operational — Runtime Defensive Hardening

Defensive reference for an API in production. The focus is **operational**: detecting and
enforcing authorization on live traffic, capping resource consumption at the gateway,
tuning a WAF, and bounding query depth and complexity on running endpoints. The dev-time
secure-coding patterns — parameterized queries, input validation while writing the
handler, authentication design, security headers — live in `moai-ref-owasp-checklist`.
This module does not duplicate them; the two are consulted together.

Back to the overview: [SKILL.md](../SKILL.md).

## Operational vs Dev-Time (the split)

| Concern | Operational (this module) | Dev-time (`moai-ref-owasp-checklist`) |
|---------|---------------------------|---------------------------------------|
| Authorization | Detecting Broken Object Level Authorization in production traffic; enforcing server-side ownership checks on every request | Designing the authorization model while writing the endpoint |
| Resource limits | Rate-limit enforcement and request-size caps at the gateway | Pagination defaults in the handler |
| Input handling | WAF rule tuning against live traffic | Parameterized queries and input validation in code |
| Query shape | GraphQL/REST depth and complexity limiting on the live endpoint | Schema and resolver design |

The operational layer assumes the dev-time layer may have a gap and enforces a runtime
boundary in front of it. Both are needed; neither replaces the other.

## OWASP API Top 10 — Operational Defense

The OWASP API Security Top 10 indexes the risks specific to APIs. Below, each risk is
mapped to its **operational** defense — what you detect and enforce on live traffic.
Citations are for defensive correlation; this is not an attack catalog.

| ID | Risk | Operational detection | Operational enforcement |
|----|------|-----------------------|-------------------------|
| API1 | Broken Object Level Authorization (BOLA) | Monitor for one identity accessing another's object IDs in production | Server-side ownership check on every object access — the single most important operational API control |
| API2 | Broken Authentication | Detect credential-stuffing, token reuse, and abnormal auth-failure rates | Enforce token validation, short token lifetimes, and lockout/rate-limit on auth endpoints |
| API3 | Broken Object Property Level Authorization | Detect responses returning fields a role should not see, or accepting fields a client should not set | Field-level response filtering and request allowlisting at the gateway/handler boundary |
| API4 | Unrestricted Resource Consumption | Detect requests that exhaust tokens, memory, or compute | Rate limits, request-size caps, pagination enforcement, timeouts |
| API5 | Broken Function Level Authorization | Detect a non-privileged identity calling an admin function | Enforce role/permission checks on every privileged operation server-side |
| API6 | Unrestricted Access to Sensitive Business Flows | Detect automated abuse of a sensitive flow (bulk signup, scraping, purchase automation) | Flow-level rate limits, anomaly detection, and friction (challenge) on abuse signals |
| API7 | Server-Side Request Forgery (SSRF) | Detect outbound requests to internal/metadata endpoints triggered by user-supplied URLs | Egress allowlist, block internal/metadata ranges, validate fetch targets |
| API8 | Security Misconfiguration | Detect debug mode, verbose errors, missing security headers, or open management endpoints in production | Production-hardened config, security headers (see `moai-ref-owasp-checklist`), inspect deployed config |
| API9 | Improper Inventory Management | Detect undocumented, old-version, or shadow API endpoints reachable in production | Maintain an API inventory; deprecate and decommission old versions; gate undocumented routes |
| API10 | Unsafe Consumption of APIs | Detect blind trust of an upstream/third-party API response | Validate and bound responses from upstream APIs; set timeouts; treat upstream output as untrusted |

### Broken Object Level Authorization (the priority control)

BOLA is the highest-frequency API risk and is fundamentally operational: it occurs when an
endpoint returns an object by ID without checking that the requesting identity owns or may
access that object. The defense is a **server-side ownership check on every object access**
— never trust an object ID from the client as proof of authorization. Detect it in
production by monitoring for an identity accessing object IDs outside its expected set;
enforce it by making the ownership check a non-optional gate in every object-accessing
endpoint. A frontend-side check is not a control; the client is not a trust boundary.

## Rate-Limit and Resource Enforcement

Unrestricted resource consumption (API4) is exploited before it is noticed. Enforce limits
at the gateway so they apply uniformly, ahead of the application.

| Control | Defensive rationale |
|---------|---------------------|
| Per-identity and per-IP rate limits | Cap requests per identity and per source so a single client cannot exhaust the service |
| Request-size caps | Reject oversized request bodies before the application parses them |
| Pagination enforcement | Enforce a maximum page size server-side so a client cannot request an unbounded result set |
| Timeouts and circuit breakers | Bound how long any request (and any upstream call) may run; shed load under pressure |
| Tighter limits on expensive endpoints | An endpoint that runs a heavy query or a costly model call gets a stricter limit than a cheap read |
| Flow-level limits on sensitive business flows | Beyond per-request limits, cap the rate of a sensitive flow (signup, password reset, purchase) to defend API6 |

## WAF Rule Tuning

A Web Application Firewall sits in front of the API and inspects traffic against a rule
set (commonly a core rule set derived from OWASP CRS). The defensive value is in tuning —
an untuned WAF either blocks legitimate traffic (false positives) or is left in
detection-only mode and enforces nothing. This section is about rule-shape and tuning, not
a vendor-specific configuration.

| Tuning step | Defensive rationale |
|-------------|---------------------|
| Start in detection mode, then enforce | Run the rule set in detection-only first to measure false positives against real traffic, then switch to enforcement once tuned |
| Tune false positives narrowly | Suppress a confirmed false-positive rule for a specific path/parameter — never disable a whole rule category globally, which blinds the WAF |
| Set the paranoia/sensitivity level deliberately | A higher sensitivity catches more but raises false positives; choose the level per the application's risk and tune from there |
| Rate-limit and geo/IP rules at the WAF | Use the WAF for coarse rate limiting and source-based blocking ahead of the application |
| Log blocked and allowed-with-anomaly traffic | Feed WAF logs into detection so tuning is data-driven and an evasion attempt is visible |
| Treat the WAF as defense-in-depth, not the only control | A WAF reduces noise and blocks known patterns; it does not replace server-side authorization and input validation |

## GraphQL / REST Depth and Complexity Limiting

A flexible query interface — GraphQL especially, but also REST with expansion/include
parameters — lets a client request a deeply nested or highly complex response that
exhausts the server. The defense is to bound query depth and complexity on the live
endpoint.

| Control | Defensive rationale |
|---------|---------------------|
| Query depth limit | Reject a query nested beyond a maximum depth so a client cannot request a deeply recursive expansion |
| Query complexity / cost analysis | Assign a cost to each field and reject a query whose total cost exceeds a budget, defending against expensive-but-shallow queries |
| Persisted-query allowlist | In production, accept only a pre-registered set of queries (by hash) so an arbitrary expensive query cannot be sent |
| Disable introspection in production | Turn off schema introspection on the production endpoint to reduce the exposed surface (keep it in development) |
| Pagination + max page size on list fields | Bound the size of any list a query can return, the same control as REST pagination enforcement |
| Per-query timeout and rate limit | Combine depth/complexity limits with a timeout and rate limit so even an in-budget query cannot run unbounded or be sent at volume |

For REST, the analogous controls bound `include`/`expand` depth, cap the number of related
resources expanded in one request, and enforce a maximum page size — the same defense
applied to REST's expansion surface.

## MITRE ATT&CK — Defensive Correlation (API surface)

The skill cites MITRE ATT&CK technique IDs to correlate a defense with the technique it
counters — never to provide an attack procedure. Representative correlations for the API
surface:

| ATT&CK technique | Technique name | Defensive control that counters it |
|------------------|----------------|------------------------------------|
| T1190 | Exploit Public-Facing Application | WAF tuning, input validation (dev-time), server-side authorization, patch hygiene |
| T1499 | Endpoint Denial of Service | Rate limits, request-size caps, query depth/complexity limits, timeouts |
| T1078 | Valid Accounts | Auth-anomaly detection, token-lifetime limits, server-side function-level authorization |

Each technique is cited to state how to detect and defend, never how to execute. A
technique ID accompanied by an exploitation procedure is the anti-pattern this skill forbids.

## API Operational Verification Checklist

- [ ] Every object-accessing endpoint enforces a server-side ownership check (BOLA defense); the client object ID is never trusted as authorization
- [ ] Every privileged operation enforces a server-side role/permission check (function-level authorization)
- [ ] Rate limits, request-size caps, and pagination maximums are enforced at the gateway, with tighter limits on expensive and sensitive flows
- [ ] SSRF defense blocks user-supplied URLs from reaching internal/metadata endpoints (egress allowlist)
- [ ] The WAF is tuned (false positives suppressed narrowly) and in enforcement mode, not detection-only where enforcement is warranted
- [ ] GraphQL/REST queries are bounded by depth and complexity limits; production introspection is disabled; a persisted-query allowlist is used where applicable
- [ ] An API inventory tracks all endpoints and versions; old and undocumented endpoints are deprecated and gated
- [ ] Operational defenses are layered with the dev-time patterns in `moai-ref-owasp-checklist`, not in place of them
