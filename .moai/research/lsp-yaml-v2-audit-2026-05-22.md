---
created: 2026-05-22
updated: 2026-05-22
tags: [research, config, audit, lsp, dead-code, v2-audit]
status: active
authors: [Claude Opus 4.7 — applying v2 6-step procedure]
related_specs:
  - SPEC-LSP-001 (Language Server Protocol Client System) status:superseded
  - SPEC-LSP-CORE-002 (LSP Core foundation) status:in-progress v1.0.0
  - SPEC-LSP-AGG-003 (LSP Aggregator) status:completed v1.0.0
  - SPEC-LSP-QGATE-004 (LSP Quality Gates) status:completed v1.0.0
  - SPEC-GOPLS-BRIDGE-001 (Gopls Bridge) status:completed v1.0.0
parent_audit: .moai/research/config-audit-2026-05-22.md
---

# LSP yaml v2 Audit — 2026-05-22

> Applies the 6-step procedure from `.moai/research/config-audit-2026-05-22.md` §4 to `lsp.yaml`. v1 audit classified ~80% of lsp.yaml as "dead" based on grep symbol reference alone. This v2 re-verification finds the classification is **incorrect or premature** — the keys are owned by completed SPECs but lack yaml→Go binding (hardcoded constants used instead). Decision required: WontDo + yaml removal, or forward-compat retention.

---

## 1. Summary

- **Total lsp.yaml top-level keys**: 6 (`enabled`, `aggregator`, `circuit_breaker`, `client_impl`, `delegate_to_astgrep`, `discovery`, `servers`)
- **Total Go consumer files** (excluding tests/templates): 5 keys with consumers
- **Owner SPECs identified**: 5 (LSP-001 superseded / LSP-CORE-002 in-progress / LSP-AGG-003 completed / LSP-QGATE-004 completed / GOPLS-BRIDGE-001 completed)
- **v1 audit "dead" claim**: 80% (`aggregator.*`, `circuit_breaker.*`, `delegate_to_astgrep.*`, `discovery.*`, `enabled`)
- **v2 verdict**: 0% confirmed dead. Mixed disposition required.

---

## 2. Step-by-Step Analysis

### Step 1 — Symbol reference (necessary)

```
lsp.aggregator           → 0 Go consumer files (struct field with yaml:"aggregator" tag)
lsp.circuit_breaker      → 0
lsp.delegate_to_astgrep  → 0
lsp.discovery            → 0
lsp.enabled              → 1 (internal/lsp/gopls/config.go)
lsp.client_impl          → 2 (internal/lsp/config/types.go + core/manager.go)
lsp.servers              → 1 (internal/lsp/core/manager.go)
```

The `lspYAMLRoot` struct in `internal/lsp/config/loader.go:13-18` declares only `ClientImpl` + `Servers` — the other 4 sub-keys never bind to Go structs. v1 audit conclusion derived from this signal.

### Step 2 — SPEC catalog cross-check (mandatory)

| Key | Owner SPEC(s) | Status |
|-----|---------------|--------|
| `lsp.enabled` | SPEC-GOPLS-BRIDGE-001, SPEC-LSP-CORE-002 | completed / in-progress |
| `lsp.client_impl` | SPEC-LSP-CORE-002 | in-progress (AC10 feature flag) |
| `lsp.servers` | SPEC-LSP-001, SPEC-LSP-AGG-003 | superseded / completed |
| `lsp.aggregator.cache_ttl_seconds` | SPEC-LSP-AGG-003 | **completed** |
| `lsp.aggregator.diagnostics_debounce_ms` | SPEC-LSP-AGG-003 | **completed** |
| `lsp.aggregator.max_parallel_servers` | SPEC-LSP-AGG-003 | **completed** |
| `lsp.aggregator.request_timeout_seconds` | SPEC-LSP-AGG-003 | **completed** |
| `lsp.aggregator.diagnostic_pull_enabled` | SPEC-LSP-AGG-003 | **completed** |
| `lsp.aggregator.server_startup_timeout_seconds` | SPEC-LSP-AGG-003 | **completed** |
| `lsp.aggregator.shutdown_timeout_seconds` | SPEC-LSP-AGG-003 | **completed** |
| `lsp.circuit_breaker.threshold` | SPEC-LSP-AGG-003 | **completed** |
| `lsp.circuit_breaker.open_duration_seconds` | SPEC-LSP-AGG-003 | **completed** |
| `lsp.delegate_to_astgrep.enabled` | (none found by grep — needs deeper check) | unknown |
| `lsp.delegate_to_astgrep.languages` | (none found) | unknown |
| `lsp.delegate_to_astgrep.rules_dir` | (none found) | unknown |
| `lsp.discovery.auto_detect_language` | (none found) | unknown |
| `lsp.discovery.on_missing` | (none found) | unknown |

**Critical finding**: SPEC-LSP-AGG-003 is status `completed` v1.0.0. The aggregator/circuit_breaker yaml keys belong to a completed SPEC, but the runtime uses **hardcoded constants**:

```go
// internal/lsp/aggregator/aggregator.go (verified 2026-05-22)
const defaultQueryTimeout = 5 * time.Second
const defaultCacheTTL = 5 * time.Second
const defaultCBThreshold = 3
const defaultCBTimeout = 30 * time.Second
```

The yaml values for `aggregator.cache_ttl_seconds: 5`, `circuit_breaker.threshold: 3`, `circuit_breaker.open_duration_seconds: 30` happen to **match** the hardcoded defaults — suggesting the yaml was authored alongside the SPEC as documentation/intent but the implementer chose to inline the constants rather than wire them through. This is **not dead config** per the v2 procedure definition — it is **WontDo or deferred yaml override**.

### Step 3 — Owner SPEC disposition

| Key category | Owner SPEC | Status | Disposition |
|--------------|-----------|--------|-------------|
| `client_impl` + `servers` + `enabled` | LSP-CORE-002 / GOPLS-BRIDGE-001 | live | **Active, keep** |
| `aggregator.*` (8 keys) | LSP-AGG-003 | completed | **WontDo or deferred** — yaml present, code uses hardcoded constants. Requires SPEC owner decision: (a) remove yaml keys + comment "intentional, see SPEC-LSP-AGG-003 completion notes"; or (b) wire through yaml→Go binding to honor the yaml values |
| `circuit_breaker.*` (2 keys) | LSP-AGG-003 | completed | Same as aggregator (`defaultCBThreshold = 3`, `defaultCBTimeout = 30s` are hardcoded counterparts) |
| `delegate_to_astgrep.*` (3 keys) | unknown | unknown | Investigation required — search astgrep integration code for whether yaml is read |
| `discovery.*` (4 keys) | unknown | unknown | Investigation required — search server discovery code |

### Step 4 — Consumer fan_in vs `@MX:ANCHOR` declaration

Verified `@MX:ANCHOR` in `internal/lsp/`:
- `ServersConfig` (config/types.go:56) — `top-level config type returned by Load() and used as a Manager initialization argument`
- `Load` (config/loader.go) — `primary entry point for LSP config; all server startup paths call this`
- 8+ additional anchors in core/, transport/, models.go

No `@MX:ANCHOR` declares fan_in for the orphan keys (`aggregator/circuit_breaker/delegate_to_astgrep/discovery`). This is **consistent with "yaml documents intent, code uses hardcoded"** — no ANCHOR contradiction, just a yaml↔Go binding gap.

### Step 5 — Opt-in pattern detection

`lsp.enabled: false` (default in yaml). The single consumer is `internal/lsp/gopls/config.go`. Confirmed yaml→Go binding exists. **Not dead** — it gates whether LSP infrastructure starts at all. Default `false` is consistent with the project's posture (LSP optional, opt-in via yaml flip).

`lsp.aggregator.diagnostic_pull_enabled: true` — if implementation were yaml-wired, this would be opt-in. But hardcoded → no opt-in semantics. WontDo signature.

### Step 6 — Cross-SPEC dependency graph

```
SPEC-LSP-001 (superseded by ...) ──────┐
                                       ▼
                              SPEC-LSP-CORE-002 (in-progress, defines client_impl + servers)
                                       │
                                       ├──▶ uses ──▶  config.Load() reading client_impl + servers
                                       │
                                       └──▶ delegates to ──▶ SPEC-LSP-AGG-003 (completed)
                                                              │
                                                              ├──▶ aggregator.go const defaultQueryTimeout = 5s
                                                              ├──▶ aggregator.go const defaultCacheTTL = 5s
                                                              ├──▶ aggregator.go const defaultCBThreshold = 3
                                                              └──▶ aggregator.go const defaultCBTimeout = 30s
                                                                   (yaml aggregator.* + circuit_breaker.*
                                                                    values match but are not consumed)

SPEC-LSP-QGATE-004 (completed) — separate quality-gate concern, lsp.servers consumer only
SPEC-GOPLS-BRIDGE-001 (completed) — gopls-specific consumer of lsp.enabled
```

`delegate_to_astgrep.*` and `discovery.*` have no identifiable owner SPEC by grep — likely planned by SPEC-LSP-001 (superseded). Verify via SPEC-LSP-CORE-002 successor coverage.

---

## 3. v2 Verdict — Mixed Disposition

| Group | v1 claim | v2 verdict | Action |
|-------|----------|-----------|--------|
| `enabled` + `client_impl` + `servers` | "live" | **CONFIRMED LIVE** | Keep |
| `aggregator.*` (8 keys) | "dead 80%" | **WontDo or deferred — NOT dead** | Owner SPEC LSP-AGG-003 completed but yaml→Go binding intentionally omitted; either (a) remove yaml keys with SPEC amendment, or (b) wire through as separate cleanup SPEC |
| `circuit_breaker.*` (2 keys) | "dead 80%" | Same as aggregator | Same |
| `delegate_to_astgrep.*` (3 keys) | "dead" | **Unknown — needs owner SPEC discovery** | Search SPEC catalog and astgrep integration code |
| `discovery.*` (4 keys) | "dead" | **Unknown — needs owner SPEC discovery** | Search server discovery code |

---

## 4. Recommended Next Steps (NOT executed in this audit)

1. **For `aggregator.*` + `circuit_breaker.*`**: Open AskUserQuestion to owner: should yaml be honored (wire through Aggregator constructor options) OR be retired (remove yaml + add SPEC-LSP-AGG-003 HISTORY note "yaml override not implemented, hardcoded defaults retained")?

2. **For `delegate_to_astgrep.*`**: Investigate `internal/lsp/aggregator/` or new astgrep integration package for yaml consumer. If none found, locate the SPEC that declared the integration (possibly SPEC-LSP-001 superseded). Verify successor coverage.

3. **For `discovery.*`**: Investigate `internal/lsp/core/manager.go` or `subprocess/launcher.go` for auto-detection logic. If yaml-blind, repeat (a) wire-or-retire decision.

4. **DO NOT create a "CONFIG-DEAD-CLEANUP-LSP-001" SPEC** without resolving steps 1-3 first — repeating the v1 audit anti-pattern (premature dead classification) will produce a second manager-spec blocker report.

---

## 5. Lessons Reinforced (apply to future audits)

1. **Hardcoded constants matching yaml values** is a strong signal of **deferred wire-through**, not dead config.
2. **SPEC-LSP-AGG-003 status: completed** + yaml binding absent = the SPEC was completed using hardcoded defaults; the yaml may have been authored as design intent before the binding decision. This is **WontDo**, not dead.
3. The 6-step procedure correctly identifies the mismatch — Step 2 (SPEC catalog cross-check) + Step 5 (opt-in pattern) reveal the wire-through gap.

---

## 6. Disposition

- **No yaml or Go changes made in this audit.** This document is analysis only.
- **No SPEC created** — premature.
- **Follow-up**: orchestrator should surface 3 questions (aggregator/circuit_breaker wire-or-retire, delegate_to_astgrep ownership, discovery ownership) via AskUserQuestion when the user is ready to act on this category.

---

## 7. References

- `internal/lsp/config/loader.go` lines 13-18 — `lspYAMLRoot` struct with only ClientImpl + Servers
- `internal/lsp/config/types.go` lines 54-79 — `ServersConfig` definition with `@MX:ANCHOR`
- `internal/lsp/aggregator/aggregator.go` lines 18-27 — hardcoded constants (`defaultQueryTimeout`, `defaultCacheTTL`, `defaultCBThreshold`, `defaultCBTimeout`)
- `internal/lsp/gopls/config.go` — `lsp.enabled` consumer (verified via grep)
- `.moai/specs/SPEC-LSP-AGG-003/` v1.0.0 status:completed — yaml owner without wire-through
- `.moai/specs/SPEC-LSP-CORE-002/` v1.0.0 status:in-progress — client_impl + servers consumer
- `.moai/specs/SPEC-LSP-001/` status:superseded — predecessor, possible owner of `delegate_to_astgrep` and `discovery`
- `.moai/research/config-audit-2026-05-22.md` — parent v2 procedure document (6-step)
