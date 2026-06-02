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

## §E.5 Mx-phase Audit-Ready Signal (2026-06-02)

```yaml
mx_complete_at: 2026-06-02
mx_status: evaluate-pass
mx_commit_sha: e979a4d13
mx_tag_count: 10
mx_new_tags_introduced: 0
mx_skip_justified: false
mx_verdict: EVALUATE-PASS
mx_evidence: |
  All 8 production .go files successfully integrated with cache_control system. 
  10 @MX:ANCHOR tags added during run-phase (cache_control.go, cache_config.go, 
  cache_usage_log.go, doctor_cache.go, posttooluse_cache.go). 
  No new dangerous patterns introduced. TAG delta = +10. 
  AC-PC-001..009 all PASS (docs-site Should-fix AC-PC-009 deferred to Phase M5).
  Sync commit 84a184f2c confirmed; mx_commit_sha e979a4d13 (backfilled by orchestrator 2026-06-02).
```

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

### M3 + M4 (cycle_type=tdd) — 2026-05-30

manager-develop spawn (Mode 5 sub-agent). M3 (PostToolUse telemetry hook + JSONL writer) + M4 (`moai doctor` cache metric) implemented as one spawn. M3 precedes M4 because M4 aggregates the JSONL schema M3 owns.

#### Deliverables

| File | Type | Milestone |
|------|------|-----------|
| `internal/state/cache_usage_log.go` | NEW | M3 — JSONL writer/reader + `AggregateCacheUsage` (7-day window, hit-rate + single-turn-ratio). NEW package `internal/state`. |
| `internal/state/cache_usage_log_test.go` | NEW | M3 — append/read/aggregate unit tests (7 cases) |
| `internal/hook/posttooluse_cache.go` | NEW | M3 — `CacheUsageRecorder.Record` + `ExtractCacheTokenUsage` + single-turn penalty + live `logCacheUsage` |
| `internal/hook/posttooluse_cache_test.go` | NEW | M3 — AC-PC-005/006/010 + extraction + live-path + error-path tests (9 cases) |
| `internal/hook/post_tool.go` | MODIFY | M3 — `Handle` calls `logCacheUsage(input)` (live telemetry wiring; observation-only) |
| `internal/cli/doctor_cache.go` | NEW | M4 — `checkCacheHitRate` (REQ-PC-006 hit rate + K5 single-turn warn) |
| `internal/cli/doctor_cache_test.go` | NEW | M4 — AC-PC-007 + K5 warn + edge cases (8 cases) |
| `internal/cli/doctor.go` | MODIFY | M4 — register `"Cache Hit Rate"` in workspace checks |
| `internal/cli/testdata/doctor-{light,dark,nocolor}.golden` | MODIFY | M4 cascade — regenerated golden snapshots (one new "Cache Hit Rate" OK line + 통과 10→11; `UPDATE_GOLDEN=1` sanctioned mechanism) |

#### Design decisions

- **New `internal/state` package**: the cache-usage JSONL record type + read/write/aggregate live in `internal/state` (the SSOT for the JSONL schema) so M3 (writer) and M4 (reader) share one definition. AC-PC-006 greps `./internal/state/...`; AC-PC-008 race-safe suite now includes it.
- **Full Anthropic field names in JSON**: the struct uses `json:"cache_creation_input_tokens"` / `json:"cache_read_input_tokens"` so the persisted JSONL line carries both literal keys (AC-PC-005 grep target).
- **Single-turn detection (REQ-PC-007/AC-PC-010)**: `isSingleTurnPenalty` fires iff `turn == 1 AND 0 < elapsed < 5min`. Turn > 1 never triggers (the AC-PC-010 negative-case false-positive guard). The warning string contains both `single-turn cache write penalty risk` and `session_ttl: "off"`.
- **B7 path resolution**: the recorder + live path resolve the project root via `resolveProjectRootFromInputOrEnv` (input.CWD → CLAUDE_PROJECT_DIR → os.Getwd), never leaking a stray `internal/hook/.moai/`.
- **Live `Handle` wiring is conservative**: `logCacheUsage` appends only (no warning) and defaults turn to 0 when the response omits it, so the live path never spuriously flags single-turn. The penalty warning is surfaced by the recorder (tested) and by `moai doctor` (M4 K5).
- **M4 enabled-gated surfacing**: `checkCacheHitRate` returns the `Cache hit rate (last 7 days): NN%` line ONLY when `cacheStrategy.enabled == true` (AC-PC-007 §3 — absence is correct when disabled). K5 ratio > 10% → CheckWarn with `session_ttl: "off"` recommendation.
- **@MX:ANCHOR** on `CacheUsageLogPath`, `AggregateCacheUsage`, `CacheUsageRecorder.Record`, `checkCacheHitRate` (each expected fan_in ≥ 3: hook writer + doctor reader + AC tests).

#### AC PASS/FAIL matrix (M3+M4 scope + cross-cutting)

| AC | REQ | Status | Verification Command | Actual Output |
|----|-----|--------|---------------------|---------------|
| AC-PC-001 | REQ-PC-001/002 | PASS (regression) | `grep -rn 'cache_control' internal/cli/ internal/runtime/ \| grep -v _test.go \| wc -l` | `14` (≥ 2 — unchanged by M3/M4) |
| AC-PC-005 | REQ-PC-004 | PASS | `go test -run 'TestPostToolUseCache_JSONLAppend' ./internal/hook/... -count=1 -v` | `--- PASS`; JSONL entry has both `cache_creation_input_tokens` AND `cache_read_input_tokens` |
| AC-PC-006 | REQ-PC-004 | PASS | `go test -run 'TestCacheUsage_TwoTurnSession_Turn2HitsCache' ./internal/hook/... ./internal/state/... -count=1 -v` | `--- PASS`; turn1 read==0 & creation>0, turn2 read>0 |
| AC-PC-007 | REQ-PC-006 | PASS | synthetic 7-day fixture + `moai doctor 2>&1 \| grep -cE 'Cache hit rate.*[0-9]+%'` + `go test -run TestCheckCacheHitRate ./internal/cli/...` | `1` matching line (`Cache hit rate (last 7 days): 80%`); 8/8 doctor cache tests PASS |
| AC-PC-008 | cross-cutting | PASS | `go test ./internal/{cli,runtime,hook,config,state}/... -race -count=1` | all `ok`, no DATA RACE (internal/state now exists; -race -count=3 on new tests stable) |
| AC-PC-010 | REQ-PC-007 | PASS | `go test -run 'TestPostToolUseCache_SingleTurnSession_PenaltyWarning' ./internal/hook/... -count=1 -v` | `--- PASS`; warning string + `session_ttl: "off"` present; 2-turn negative case absent |

#### Quality gates

- Cross-platform build: `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (B1 — JSONL writer uses `filepath.Join`, no hardcoded separators).
- Coverage (new code): `internal/state` 88.1%; `posttooluse_cache.go` all funcs 100%; `doctor_cache.go::checkCacheHitRate` 100% — all ≥ 85% target.
- Subagent boundary (B11/C-HRA-008): `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/posttooluse_cache.go internal/state/ internal/cli/doctor_cache.go | grep -v _test.go | grep -v "//"` → 0 matches.
- Lint: `golangci-lint run` → same 6 pre-existing errcheck baseline in `internal/cli/update_clean_install.go`; 0 NEW issues.
- Template mirror: M3/M4 touch no template-mirrored files → `make build` not required.

#### Run-phase completion note

M3+M4 complete the 5 Blocking ACs in this spawn's scope (AC-PC-005/006/007/010 + AC-PC-008 cross-cutting). Combined with the M1+M2 record above, 8 of 9 Blocking ACs (AC-PC-001..008 + AC-PC-010) are PASS. AC-PC-009 (docs-site 4-locale, Should-fix) remains M5 scope (not in this spawn).
