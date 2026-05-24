---
id: SPEC-V3R6-MULTI-SESSION-COORD-001
title: "Multi-Session Coordination — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: "GOOS행님"
priority: P1
phase: "v3.0.0"
module: "internal/session"
lifecycle: spec-anchored
tags: "multi-session, coordination, registry, hook, race-mitigation"
---

# SPEC-V3R6-MULTI-SESSION-COORD-001 — Acceptance Criteria

## §A AC Summary Matrix

| ID | Layer | Description | Severity | REQ Trace |
|----|-------|-------------|----------|-----------|
| **AC-COORD-001** | L1 | `RegisterSession` atomically appends entry; concurrent invocations don't corrupt registry | MUST | REQ-COORD-001..003, REQ-COORD-008 |
| **AC-COORD-002** | L1 | `Heartbeat` updates `last_heartbeat` field without modifying other fields | MUST | REQ-COORD-004 |
| **AC-COORD-003** | L1 | `DeregisterSession` removes matching entry; idempotent on missing entry | MUST | REQ-COORD-005 |
| **AC-COORD-004** | L1 | `PurgeStale(30)` removes entries with `last_heartbeat` older than 30 minutes; returns purged count | MUST | REQ-COORD-007 |
| **AC-COORD-005** | L2 | `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format contains `source_session_id` requirement | MUST | REQ-COORD-009 |
| **AC-COORD-006** | L2 | `.claude/output-styles/moai/moai.md` §8 Session Handoff directs orchestrator to populate `source_session_id` | MUST | REQ-COORD-010, REQ-COORD-011 |
| **AC-COORD-007** | L3 | `internal/hook/session_start.go` invokes `RegisterSession` + `PurgeStale` + `QueryActiveWork` on every SessionStart event | MUST | REQ-COORD-013..015 |
| **AC-COORD-008** | L3 | `.claude/hooks/moai/handle-session-start.sh` passes session_id from stdin JSON to `moai hook session-start` | MUST | REQ-COORD-016 |
| **AC-COORD-009** | L4 | `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check contains `moai session list --json` literal | MUST | REQ-COORD-017 |
| **AC-COORD-010** | L4 | Rule body specifies AskUserQuestion options (wait / override / abort) on non-empty `moai session list --json` output | MUST | REQ-COORD-019 |
| **AC-COORD-011** | Cross | Cross-platform build matrix (linux/amd64 + darwin/amd64 + darwin/arm64 + windows/amd64) exits 0 | MUST | REQ-COORD-022 |
| **AC-COORD-012** | Cross | `go vet ./internal/session/... ./cmd/moai/...` + `golangci-lint run --timeout=2m` exit 0 issues | MUST | REQ-COORD-023 |

Total ACs: 12 (10 layer-specific + 2 cross-cutting).

### §A.1 Out of Scope

- AC for cross-machine coordination scenarios (single-machine assumption only)
- AC for distributed lock failure modes (registry is advisory, no strong mutex)
- AC for race conditions exceeding 30-minute stale heartbeat threshold (out of scope by design — REQ-COORD-007 cutoff)
- AC for paste-ready format breaking change migration (backward compat tolerated per REQ-COORD-012)
- AC for `moai session migrate` schema versioning (REQ-COORD-024 freezes schema)
- AC for `MEMORY.md` concurrent write coordination (paste-ready tagging only — L2 scope)
- AC for hook events beyond `SessionStart` (no `Stop` / `Notification` / `PostToolUse` ACs)
- AC for windows/arm64 build matrix (only windows/amd64 in AC-COORD-011)
- AC for GUI / TUI / web dashboard (CLI text + `--json` only)
- AC for performance benchmarks of registry operations (no throughput / latency targets in this SPEC)

## §B AC Verification Details

### §B.1 AC-COORD-001 — RegisterSession Atomic Append

**Verification**:
```bash
# Unit test
go test -race -run TestRegisterSession ./internal/session/

# Concurrent stress test (10 goroutines × 100 iterations each)
go test -race -run TestRegisterSessionConcurrent ./internal/session/
```

**Expected**:
```
=== RUN   TestRegisterSession
--- PASS: TestRegisterSession (0.0Xs)
=== RUN   TestRegisterSessionConcurrent
--- PASS: TestRegisterSessionConcurrent (0.Xs)
PASS
ok  	github.com/modu-ai/moai-adk/internal/session	0.XXXs
```

**Pass criterion**: 0 race detector warnings + 0 test failures + final registry entry count = 1000 (10 × 100, no lost updates).

### §B.2 AC-COORD-002 — Heartbeat Field-Selective Update

**Verification**:
```bash
go test -race -run TestHeartbeat ./internal/session/
```

**Expected**:
```
=== RUN   TestHeartbeat
    registry_test.go:XXX: original entry: started_at=2026-05-24T20:00:00Z, last_heartbeat=2026-05-24T20:00:00Z
    registry_test.go:XXX: after Heartbeat: started_at=2026-05-24T20:00:00Z (preserved), last_heartbeat=2026-05-24T20:15:00Z (updated)
--- PASS: TestHeartbeat (0.XXs)
PASS
```

**Pass criterion**: `started_at` unchanged + `last_heartbeat` updated + all other fields preserved.

### §B.3 AC-COORD-003 — DeregisterSession Idempotent

**Verification**:
```bash
go test -race -run TestDeregisterSession ./internal/session/
go test -race -run TestDeregisterSessionIdempotent ./internal/session/
```

**Expected**: Both PASS. Idempotent test calls `DeregisterSession(sessionID)` twice; second call returns nil error.

**Pass criterion**: First call removes entry. Second call no-op without error.

### §B.4 AC-COORD-004 — PurgeStale 30-Minute Cutoff

**Verification**:
```bash
go test -race -run TestPurgeStale ./internal/session/
```

**Expected**: Test sets up 3 entries: (a) fresh `last_heartbeat = now`, (b) 25 minutes old, (c) 35 minutes old. `PurgeStale(30)` returns `purgedCount = 1` (only entry c removed).

**Pass criterion**: Fresh + 25min entries preserved. 35min entry removed. Returned count = 1.

### §B.5 AC-COORD-005 — session-handoff.md Contains source_session_id

**Verification**:
```bash
grep -E "source_session_id" .claude/rules/moai/workflow/session-handoff.md | head -3
```

**Expected** (non-empty output, at least 1 occurrence):
```
source_session_id: <UUID>                  # Required. Claude Code session_id from current orchestrator turn.
```

**Pass criterion**: `grep` exit code 0 + at least 1 match line.

### §B.6 AC-COORD-006 — output-styles/moai/moai.md Directs source_session_id Population

**Verification**:
```bash
grep -E "source_session_id" .claude/output-styles/moai/moai.md | head -3
```

**Expected** (non-empty output, at least 1 occurrence):
```
... populate `source_session_id` ...
```

**Pass criterion**: `grep` exit code 0 + at least 1 match line.

### §B.7 AC-COORD-007 — SessionStart Hook 3-Step Protocol

**Verification**:
```bash
# Inspect hook source
grep -E "RegisterSession|PurgeStale|QueryActiveWork" internal/hook/session_start.go

# Run hook handler with mock stdin
echo '{"session_id":"test-uuid-1234"}' | go run ./cmd/moai hook session-start
```

**Expected**: grep finds all 3 function calls in hook source. Test execution: registry file at temp path contains entry for `test-uuid-1234` with `spec_id="(none)"`, `phase="(none)"`.

**Pass criterion**: All 3 calls present in source + test invocation creates registry entry.

### §B.8 AC-COORD-008 — handle-session-start.sh Passes session_id

**Verification**:
```bash
grep -E 'INPUT|session_id|session-start' .claude/hooks/moai/handle-session-start.sh
```

**Expected**:
```
INPUT=$(cat)
moai hook session-start <<< "$INPUT"
```

**Pass criterion**: Wrapper preserves stdin JSON (containing session_id) and pipes to `moai hook session-start` subcommand.

### §B.9 AC-COORD-009 — agent-common-protocol.md Contains `moai session list --json` Literal

**Verification**:
```bash
grep -E "moai session list --json" .claude/rules/moai/core/agent-common-protocol.md
```

**Expected** (non-empty output):
```
moai session list --json --filter-spec=<SPEC-ID>
```

**Pass criterion**: `grep` exit code 0 + literal command present in rule body.

### §B.10 AC-COORD-010 — Rule Body Specifies AskUserQuestion (wait/override/abort)

**Verification**:
```bash
# Locate the §Pre-Spawn Sync Check section + verify 3 options
awk '/Pre-Spawn Sync Check/,/Cross-reference/' .claude/rules/moai/core/agent-common-protocol.md | grep -E "wait|override|abort"
```

**Expected**: All 3 keywords (wait, override, abort) found within the §Pre-Spawn Sync Check section.

**Pass criterion**: 3 keyword matches within section bounds.

### §B.11 AC-COORD-011 — Cross-Platform Build Matrix

**Verification**:
```bash
GOOS=linux   GOARCH=amd64 go build ./... && echo "linux/amd64 OK"
GOOS=darwin  GOARCH=amd64 go build ./... && echo "darwin/amd64 OK"
GOOS=darwin  GOARCH=arm64 go build ./... && echo "darwin/arm64 OK"
GOOS=windows GOARCH=amd64 go build ./... && echo "windows/amd64 OK"
```

**Expected**: 4 "OK" lines.

**Pass criterion**: All 4 build commands exit 0.

### §B.12 AC-COORD-012 — Lint Clean

**Verification**:
```bash
go vet ./internal/session/... ./cmd/moai/...
golangci-lint run --timeout=2m
```

**Expected**:
```
(go vet produces no output on success)
(golangci-lint output ends with "0 issues.")
```

**Pass criterion**: `go vet` exit 0 + no output + `golangci-lint` reports `0 issues.`

## §C Invariants (PRESERVE)

### §C.1 PRESERVE List (11 entries — verbatim through plan/run/sync/Mx)

[HARD] `git status --short | sort` MUST contain exactly these 11 lines through all phase boundaries:

```
 M .moai/config/sections/git-convention.yaml
 M .moai/config/sections/language.yaml
 M .moai/config/sections/quality.yaml
 M .moai/harness/usage-log.jsonl
?? .moai/harness/learning-history/
?? .moai/harness/observations.yaml
?? .moai/research/anthropic-best-practices-2026-05-24.md
?? .moai/research/v3.0-redesign-2026-05-23.md
?? .moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/
?? i18n-validator
```

Plus the new SPEC artifacts under `.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/` which transition from `??` (post-Write, pre-commit) to tracked (post-commit).

### §C.2 @MX Tag Delta = 0 (Documentation-Only File Edits)

For files in M4 (rule + output-style extension): @MX tag count source vs mirror MUST remain delta = 0 byte-identical. Verified by `TestRuleTemplateMirrorDrift` test suite.

### §C.3 No Touching SPEC-V3R6-HARNESS-PROPOSAL-GEN-001

[HARD] `?? .moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/` is another active session's work. Any modification (including reading the directory contents beyond `ls`) is forbidden.

Verification: `git status --short | grep HARNESS-PROPOSAL-GEN` MUST show `?? .moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/` only (unchanged).

### §C.4 No `git add -A` / `git add .` (Path-Specific Only)

All run-phase commits MUST use path-specific `git add <path>` invocations. Verification via post-commit `git log --stat HEAD~1..HEAD` showing exactly the intended files (no PRESERVE list contamination).

### §C.5 Frontmatter 12-Field Canonical Schema

All 4 artifacts MUST contain exactly the 12 canonical fields (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags). Verified via:

```bash
for f in spec.md plan.md acceptance.md progress.md; do
  echo "=== $f ==="
  head -20 .moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/$f | grep -cE "^(id|title|version|status|created|updated|author|priority|phase|module|lifecycle|tags):"
done
```

Expected: Each artifact reports `12` (one match per canonical field).

## §D Forward-Looking Behavioral Tests (Post-Merge)

### §D.1 Race Detection Validation

On the next ARR/SIV-class race scenario after merge, the new coordination layer SHOULD:

1. **Layer 1 detection**: Two sessions concurrent on same SPEC → `moai session list --json --filter-spec=SPEC-X` returns 2 entries (one per session)
2. **Layer 4 surface**: orchestrator pre-spawn batch 3rd command detects the other entry → STOP + AskUserQuestion (wait/override/abort)
3. **Layer 3 stderr signal**: each SessionStart emits stderr system-reminder if other sessions active

**Validation method (post-merge, manual scenario)**:
- Open 2 terminals on same project root
- In terminal 1: run `moai session register session-a SPEC-TEST-001 plan`
- In terminal 2: run `moai session list --json --filter-spec=SPEC-TEST-001`
- Expect: terminal 2 sees session-a entry
- In terminal 2: simulate orchestrator pre-spawn check — see entry surfaced

### §D.2 Backward Compatibility Validation

Pre-existing memory files without `source_session_id` field MUST continue to be parseable by orchestrator. Validation:

- Read 5 random pre-existing `project_*.md` files from `~/.claude/projects/{hash}/memory/`
- Verify orchestrator does not error on missing `source_session_id`
- Expected: graceful degradation (treat as `source_session_id: unknown`)

### §D.3 Stale Purge Effective

After 30 minutes of inactivity (no Heartbeat), entry SHOULD be auto-purged on next SessionStart hook invocation. Validation:

- Manually insert entry with `last_heartbeat = now() - 31min`
- Trigger SessionStart hook (or call `moai session purge`)
- Expected: stale entry removed, `purged_count = 1` reported

## §E Quality Gate Criteria

### §E.1 Definition of Done (DoD)

[HARD] Run-phase complete when ALL of the following hold:

1. ✅ All 12 ACs PASS (B.1 ~ B.12 verification commands all exit 0 with expected outputs)
2. ✅ All 5 invariants hold (§C.1 PRESERVE / §C.2 @MX delta / §C.3 no HARNESS-PROPOSAL-GEN touch / §C.4 path-specific add / §C.5 frontmatter 12-field)
3. ✅ Cross-platform build matrix green (4 GOOS/GOARCH combos)
4. ✅ Test coverage ≥ 85% for `internal/session/` package
5. ✅ Race detector clean (`-race` flag on all `internal/session/` tests)
6. ✅ Template mirror parity verified (`TestRuleTemplateMirrorDrift` PASS for 3 mirrored files)
7. ✅ Catalog hash regen if affected (run `make build-catalog-hashes` if `TestManifestHashFormat` reports drift)

### §E.2 B12 Self-Test (Sync-Phase)

Per MEMORY.md B12 standing-rule (9 consecutive PASS streak as of HCW-001 sync 2026-05-24), sync-phase MUST verify:

(a) **CHANGELOG count**: `[Unreleased]` section contains 1 (and only 1) entry attributing SPEC-V3R6-MULTI-SESSION-COORD-001
```bash
grep -cE "SPEC-V3R6-MULTI-SESSION-COORD-001" CHANGELOG.md
# Expected: ≥ 1 (1 [Unreleased] entry)
```

(b) **AC count match**: acceptance.md AC row count = 12 (matches §A summary matrix)
```bash
grep -cE '^\| \*\*AC-COORD-[0-9]+\*\*' .moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/acceptance.md
# Expected: 12
```

(c) **Frontmatter status transition**: all 4 artifacts have `status: implemented` (sync-phase completes the transition from `draft` → `implemented` via manager-docs)
```bash
grep -lE "^status: implemented" .moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/*.md | wc -l
# Expected: 4
```

### §E.3 Mx-Phase (Step C Judge)

@MX tag delta scan post-run:
- `internal/session/registry.go` (NEW file): expected @MX:NOTE on package doc comment + @MX:ANCHOR on `Entry` struct + @MX:NOTE on `withLock` helper. Total ~3-5 @MX tags.
- `internal/hook/session_start.go` (MODIFY): expected @MX:NOTE on 3-step protocol block. Total +1 @MX tag delta.
- `.claude/rules/...` (MODIFY): documentation-only edits, expected @MX delta = 0.

**Step C verdict**: EVALUATE-PASS if @MX delta within expected range AND all @MX:WARN tags have @MX:REASON sibling.

**Mx-chore backfill**: progress.md §E.5 mx_chore section filled with @MX delta breakdown + mx_commit_sha.

## §F Test Strategy

### §F.1 Unit Tests (`internal/session/registry_test.go`)

- Table-driven tests for each public function (RegisterSession, Heartbeat, DeregisterSession, QueryActiveWork, PurgeStale)
- Each table ≥ 3 cases (happy path + edge case + error case)
- Use `t.TempDir()` for isolated registry path per test
- Use `clock.FakeClock` for time-sensitive tests (PurgeStale)
- Concurrency test with `-race` (10 goroutines × 100 ops)

### §F.2 CLI Smoke Tests (`cmd/moai/session_test.go`)

- Each of 5 subcommands (register / heartbeat / deregister / list / purge) gets a smoke test
- Use `cobra` test pattern with output capture
- Verify exit code 0 + expected stdout content
- `--json` flag test: parse output back via `json.Unmarshal` to verify valid JSON

### §F.3 Hook Integration Test (`internal/hook/session_start_test.go`)

- Mock stdin JSON with `session_id` field
- Invoke `session-start` handler with temp registry path
- Assert registry contains expected entry + stderr contains expected system-reminder format

### §F.4 Cross-Platform CI

- `.github/workflows/*` already configured for linux/darwin/windows matrix per repo (existing pattern)
- Run-phase verification matrix runs all 4 build combos locally before sync

## §G Cross-References

- **spec.md** §C — REQ-COORD-001..024 source of truth
- **spec.md** §D — Architecture (layer breakdown + data flow + atomic-write strategy)
- **plan.md** §C — Milestones M1-M5 (verification ↔ this AC matrix mapping)
- **plan.md** §F — Cross-platform considerations (build matrix per AC-COORD-011)
- **plan.md** §G — Anti-patterns (AP-MSC-001..008)
- **progress.md** — Lifecycle status + run/sync evidence (filled by run-phase)
