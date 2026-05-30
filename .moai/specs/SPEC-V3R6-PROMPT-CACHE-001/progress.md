# SPEC-V3R6-PROMPT-CACHE-001 — Progress

## §A — Status

- SPEC version: 0.1.1 (run-phase entry)
- Status: draft → in-progress (manager-develop M1 first commit performs the transition)
- Tier: M (PASS threshold 0.80)
- Run-phase entered: 2026-05-30
- Plan-auditor verdict (Phase 0.5 iter-2): **PASS 0.89** (+0.07 monotonic vs iter-1 0.82; D1/D2 resolved, no blocking defects). Report: `.moai/reports/plan-audit/SPEC-V3R6-PROMPT-CACHE-001-2026-05-30.md`
- GATE-2 (plan-to-implement HUMAN GATE): approved in predecessor session (carried via paste-ready resume).

## §E — Phase 0.95 Mode Selection

### Input parameters

| Parameter | Value |
|-----------|-------|
| tier | M |
| scope (file count) | ~14 files (M1: 2 Go · M2: 3 [Go+yaml+mirror] · M3: 3 Go · M4: 2 Go · M5: 4 markdown) |
| domain count | predominantly 1 (Go `internal/{runtime,config,hook,state,cli}`) + config yaml + template mirror + docs-site markdown |
| file language mix | ~70% Go source, config yaml, 4-locale markdown (M5) |
| concurrency benefit | LOW — coding-heavy with strict sequential milestone dependencies (M1↔M2 bidirectional, M1→M3→M4→M5) |
| Agent Teams prereqs | NOT met (coding-heavy; harness not `thorough`+team-enabled) |

### Mode evaluation

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | Multi-file Go implementation with TDD; not a typo/single-line change |
| 2 background | no | Writes code (Write/Edit) — background agents auto-deny writes |
| 3 agent-team | no | Coding-heavy; Agent Teams capability gate (thorough + team.enabled + env) not met |
| 4 parallel | no | Finding A4 caveat — coding tasks have fewer parallelizable units; milestone chain is sequential |
| 5 sub-agent | **YES** | Sequential `manager-develop` spawn(s) per milestone group; cycle_type=tdd |

### Decision: sub-agent (Mode 5)

### Justification

Per Anthropic Finding A4 (most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at real-time coordination), coding-heavy run-phase work defaults to Mode 5 sequential sub-agent. The milestone chain has hard dependencies (M1↔M2 bidirectional cache_control↔config; M3 telemetry requires M1/M2; M4 doctor requires M3 JSONL; M5 docs requires M1-M4 understanding), so parallel spawn (Mode 4) would race on shared `internal/` files. Spawn grouping: Spawn 1 = M1-M4 (all Go, 9 Blocking ACs); Spawn 2 = M5 (docs-site 4-locale markdown, AC-PC-009 Should-fix). Tier M is below the SSE-stall Round threshold (≥30 tasks), so no Round split required; the Go/markdown domain split keeps each spawn within safe context.

## §R — Run-phase Evidence

### M1 + M2 (cycle_type=tdd, atomic) — 2026-05-30

manager-develop spawn 1 (Mode 5 sub-agent). M1 (cache_control injection) + M2 (cache.yaml config schema) implemented together as a single logical unit due to the M1↔M2 bidirectional coupling documented in plan.md §D.

**Status transition**: `draft → in-progress` performed on this M1 commit (spec.md frontmatter `status:` + `updated: 2026-05-30`).

#### Deliverables

| File | Type | Milestone |
|------|------|-----------|
| `internal/runtime/cache_control.go` | NEW | M1 — `InjectCacheControl` + payload types |
| `internal/runtime/cache_control_test.go` | NEW | M1 — AC-PC-003/004 + 5 fallback tests |
| `internal/config/cache_config.go` | NEW | M2 — `LoadCacheConfig` + `Validate` + safe-default fallback |
| `internal/config/cache_config_test.go` | NEW | M2 — AC-PC-002 schema validation (3 fixtures) |
| `internal/config/testdata/cache-{valid,malformed,invalid-ttl}/cache.yaml` | NEW | M2 — test fixtures |
| `.moai/config/sections/cache.yaml` | NEW | M2 — runtime config |
| `internal/template/templates/.moai/config/sections/cache.yaml` | NEW | M2 — template mirror |
| `internal/config/audit_loader_completeness_test.go` | MODIFY | M2 cascade — `cache` added to `acknowledgedDedicatedLoaders` (LoadCacheConfig is a dedicated loader outside Loader.Load chain, mirroring `runtime`→LoadRuntime) |

#### Design decisions

- **M1 location**: NEW `internal/runtime/cache_control.go` (not an extension of `cc.go`). Rationale: cache_control placement is a pure-Go payload transformation, fully unit-testable without SDK/network wiring. AC-PC-001 greps `internal/cli/ internal/runtime/`; the literals live in `internal/runtime/`.
- **Threshold-agnostic fallback (R4 decoupled)**: `estimateSystemTokens` uses the len(text)/4 heuristic; below `MinCacheableTokens` (default 2048) the session breakpoint is omitted. No dependency on SPEC-V3R6-AGENT-MODEL-ROUTING-001.
- **Safe-default-on-any-failure (M2)**: `LoadCacheConfig` NEVER returns an error — file-not-found / malformed YAML / invalid enum all degrade to `enabled: false`, so a misconfigured cache.yaml can never block launch (plan.md M2 contract).
- **@MX:ANCHOR** added on both `InjectCacheControl` and `LoadCacheConfig` (expected fan_in ≥ 3: SDK wrapper + `/moai run` SPEC bundle path + doctor metric path + AC tests).

#### AC PASS/FAIL matrix (M1+M2 scope)

| AC | REQ | Status | Verification Command | Actual Output |
|----|-----|--------|---------------------|---------------|
| AC-PC-001 | REQ-PC-001/002 | PASS | `grep -rn 'cache_control' internal/cli/ internal/runtime/ \| grep -v _test.go \| wc -l` | `14` (≥ 2) |
| AC-PC-002 | REQ-PC-005 | PASS | `grep -E '^\s*(cacheStrategy\|enabled\|session_ttl\|spec_ttl):' .moai/config/sections/cache.yaml ... \| grep -cE ...` | `4` (≥ 4) |
| AC-PC-003 | REQ-PC-001/003 | PASS | `go test -run 'TestCacheControl_SessionStart_OneHour\|TestCacheControl_AnthropicPayloadSchema' ./internal/runtime/...` | `--- PASS` ×2, exit 0 |
| AC-PC-004 | REQ-PC-003 | PASS | `go test -run 'TestCacheControl_GLMMode_NoInjection' ./internal/runtime/...` | `--- PASS`, exit 0 |
| AC-PC-008 | cross-cutting | PASS (M1/M2 scope) | `go test ./internal/{cli,runtime,hook,config}/... -race -count=1` | all `ok`; `internal/state/...` is an M3 deliverable (path does not yet exist) |

AC-PC-005/006/007/009/010 are M3/M4/M5 scope (not in this spawn). AC-PC-008 over `internal/state/...` completes when M3 creates that package.

#### Quality gates

- Cross-platform build: `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- Coverage: `internal/runtime` 90.2%, `internal/config` cache loader covered (package aggregate ≥ 76.9%; new cache_config.go fully exercised).
- Subagent boundary: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/runtime/ internal/config/ | grep -v _test.go | grep -v "//"` → 0 matches.
- Lint: `golangci-lint run` → 6 pre-existing errcheck issues in `internal/cli/update_clean_install.go` (baseline, NOT introduced by this change); 0 NEW issues.
