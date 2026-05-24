---
id: SPEC-V3R6-MULTI-SESSION-COORD-001
title: "Multi-Session Coordination — Implementation Plan"
version: "0.1.0"
status: implemented
created: 2026-05-24
updated: 2026-05-25
author: "GOOS행님"
priority: P1
phase: "v3.0.0"
module: "internal/session"
lifecycle: spec-anchored
tags: "multi-session, coordination, registry, hook, race-mitigation"
---

# SPEC-V3R6-MULTI-SESSION-COORD-001 — Plan

## §A Tier Classification + LOC Envelope

### §A.1 Tier M (Medium)

- **Plan-auditor PASS threshold**: 0.80 (Tier M baseline)
- **Workflow envelope**: Standard 5-milestone (M1-M5) lifecycle, standard verification batch
- **Justification**: 4-layer architecture spanning Go primitive + CLI + hook + multi-rule extension exceeds Tier S minimal (1-2 file mechanical edit) but stays under Tier L thorough (cross-cutting refactor with breaking changes). New Go package + new CLI subcommand classifies as Tier M.

### §A.2 LOC Envelope (Estimated, run-phase will report actual)

| Component | File | LOC (est) | Type |
|-----------|------|-----------|------|
| Go primitive — registry | `internal/session/registry.go` | ~150 | NEW |
| Go primitive — tests | `internal/session/registry_test.go` | ~180 | NEW |
| CLI subcommand | `cmd/moai/session.go` | ~80 | NEW |
| CLI tests | `cmd/moai/session_test.go` | ~60 | NEW |
| Hook integration | `internal/hook/session_start.go` | ~30 (modify) | EXTEND |
| Hook wrapper script | `.claude/hooks/moai/handle-session-start.sh` | ~5 (modify) | EXTEND |
| Rule extension — pre-spawn | `.claude/rules/moai/core/agent-common-protocol.md` | ~30 (extend) | EXTEND |
| Rule extension — handoff | `.claude/rules/moai/workflow/session-handoff.md` | ~15 (extend) | EXTEND |
| Output-style extension | `.claude/output-styles/moai/moai.md` | ~10 (extend) | EXTEND |
| CLI registration | `internal/cli/root.go` | ~3 (modify) | EXTEND |
| **Code subtotal** | | **~563** | |
| Plan artifact — spec.md | (this SPEC) | ~310 | NEW |
| Plan artifact — plan.md | (this SPEC) | ~230 | NEW |
| Plan artifact — acceptance.md | (this SPEC) | ~250 | NEW |
| Plan artifact — progress.md | (this SPEC) | ~110 | NEW |
| **Docs subtotal** | | **~900** | |
| **TOTAL** | | **~1463 lines** | |

Within Tier M envelope (typical Tier M: 500-2000 LOC code + 600-1200 LOC docs).

## §B File Scope

### §B.1 EXTEND List (Modification Target)

```
internal/session/registry.go             # NEW — Go primitive
internal/session/registry_test.go        # NEW — unit tests
cmd/moai/session.go                      # NEW — CLI subcommand entry
cmd/moai/session_test.go                 # NEW — CLI smoke tests
internal/hook/session_start.go           # MODIFY — add 3-step protocol
.claude/hooks/moai/handle-session-start.sh   # MODIFY — pass session_id
.claude/rules/moai/core/agent-common-protocol.md      # EXTEND — §Pre-Spawn Sync Check 3rd command
.claude/rules/moai/workflow/session-handoff.md        # EXTEND — Canonical Format source_session_id
.claude/output-styles/moai/moai.md       # EXTEND — §8 Session Handoff template
internal/cli/root.go                     # MODIFY — register session subcommand
```

### §B.2 Mirror Targets (template parity per CLAUDE.local.md §2 [HARD] Template-First Rule)

When `.claude/rules/...` or `.claude/output-styles/...` are extended, the corresponding template mirror under `internal/template/templates/.claude/...` must be updated byte-identical (per L46 TEMPLATE-MIRROR-DRIFT family). Run-phase MUST include mirror cp + `TestRuleTemplateMirrorDrift` verification.

```
internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md
internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
internal/template/templates/.claude/output-styles/moai/moai.md
internal/template/templates/.claude/hooks/moai/handle-session-start.sh
```

### §B.3 Plan Artifacts (under `.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/`)

```
spec.md          # NEW — EARS requirements + architecture
plan.md          # NEW — this file
acceptance.md    # NEW — AC matrix + verification commands
progress.md      # NEW — lifecycle status + run/sync/Mx placeholders
```

### §B.4 Out of Scope

- Modification of `.moai/state/` runtime-managed files outside `active-sessions.json` (`session-memo.md`, `last-session-state.json`, `audit-gate-merge-at.txt` etc preserved)
- Migration of any legacy session-tracking file (parallel coexistence acceptable — no auto-migration in run-phase)
- Hook event additions beyond `SessionStart` (no `Stop` / `Notification` / `PostToolUse` integration in this SPEC)
- Cross-platform Windows-specific lockfile syscall handling beyond `golang.org/x/sys/windows.LockFileEx` baseline (defer to follow-up SPEC if portability tests fail on windows/arm64)
- Performance tuning of registry read/write (no caching, no in-memory mirror, no batching — naïve atomic-write per call)
- Schema migration tooling (REQ-COORD-024 freezes schema; no `moai session migrate` subcommand)
- Multi-machine federation (no remote registry sync, no shared NFS lock)
- @MX tag canonical taxonomy for `internal/session/` beyond plan-phase recommendation (auto-tagging via `/moai mx` Step C, out of run-phase scope)

## §C Milestones (M1-M5)

### §C.1 M1 — Go Primitive (Registry Package)

**Deliverable**: `internal/session/registry.go` + `internal/session/registry_test.go`

**Tasks**:
1. Create package skeleton with package doc comment
2. Define `Entry` struct matching REQ-COORD-002 schema
3. Implement `RegisterSession(sessionID, specID, phase string) error` with atomic-write semantics
4. Implement `Heartbeat(sessionID string) error` with atomic update
5. Implement `DeregisterSession(sessionID string) error` (idempotent)
6. Implement `QueryActiveWork(optSpecID string) ([]Entry, error)`
7. Implement `PurgeStale(thresholdMinutes int) (purgedCount int, err error)`
8. Implement internal `atomicWrite` + `withLock` helpers
9. Unit tests for all 5 public functions (table-driven, table size ≥ 3 cases each)
10. Concurrency test: spawn 10 goroutines doing Register + Heartbeat, verify no corruption (via `-race`)
11. Cross-platform path handling (use `filepath.Join`, not `path.Join`)

**Verification**: AC-COORD-001..004 + AC-COORD-011 + AC-COORD-012 PASS

### §C.2 M2 — CLI Subcommand

**Deliverable**: `cmd/moai/session.go` + `cmd/moai/session_test.go` + `internal/cli/root.go` registration

**Tasks**:
1. Create `sessionCmd` cobra command with 5 subcommands (register / heartbeat / deregister / list / purge)
2. Each subcommand: accept positional args + `--json` flag for machine-readable output
3. `register <session_id> <spec_id> <phase>` — calls `session.RegisterSession`
4. `heartbeat <session_id>` — calls `session.Heartbeat`
5. `deregister <session_id>` — calls `session.DeregisterSession`
6. `list [--filter-spec=<SPEC-ID>] [--json]` — calls `session.QueryActiveWork`
7. `purge [--threshold-minutes=N]` (default 30) — calls `session.PurgeStale`
8. Register `sessionCmd` under root command in `internal/cli/root.go`
9. Smoke test each subcommand (table-driven, verify exit code 0)
10. JSON output marshal stability test (parse output back to verify)

**Verification**: AC-COORD-013 PASS via `moai session --help` + per-verb smoke

### §C.3 M3 — Hook Integration

**Deliverable**: `internal/hook/session_start.go` modification + `.claude/hooks/moai/handle-session-start.sh` modification

**Tasks**:
1. In `internal/hook/session_start.go`, add 3-step protocol at start of handler:
   - Step 1: parse `session_id` from stdin JSON → `session.RegisterSession(sessionID, "(none)", "(none)")`
   - Step 2: `purged, _ := session.PurgeStale(30)` (log purged count if > 0)
   - Step 3: `entries, _ := session.QueryActiveWork("")` → filter out current session_id → emit stderr system-reminder if remaining > 0
2. Preserve all existing SessionStart hook behavior (GLM credential injection, teammateMode, etc per CLAUDE.local.md §22)
3. Modify `.claude/hooks/moai/handle-session-start.sh`:
   - Ensure session_id is preserved when piping stdin to `moai hook session-start`
   - Current wrapper already does `INPUT=$(cat); moai hook session-start <<< "$INPUT"` — verify session_id field present in INPUT JSON
4. Mirror the .sh change to `internal/template/templates/.claude/hooks/moai/handle-session-start.sh`
5. Add unit test for `session_start.go` 3-step protocol (use `t.TempDir()` for isolated registry)

**Verification**: AC-COORD-007..008 PASS

### §C.4 M4 — Rule + Output-Style Extension

**Deliverable**: 3 rule/output-style files + 4 template mirror cp

**Tasks**:
1. Edit `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check:
   - Add 3rd command after existing 2-command batch: `moai session list --json --filter-spec=<SPEC-ID>`
   - Add interpretation matrix row: `[]` → proceed, non-empty → STOP + AskUserQuestion
   - Cross-reference `.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/spec.md` §D.1
2. Edit `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format:
   - Add `source_session_id: <UUID>` requirement in Block 2 preamble (or new metadata block)
   - Update example to show source_session_id
   - Update Anti-Patterns to include "missing source_session_id"
3. Edit `.claude/output-styles/moai/moai.md` (locate §8 Session Handoff section):
   - Direct orchestrator to populate `source_session_id` from current session_id
   - Direct project_*.md auto-memory writer to include `source_session_id` field
   - Direct MEMORY.md index entry to include `(session: <UUID-8-char-prefix>)` annotation
4. Mirror cp ALL 3 edited files to `internal/template/templates/.claude/...` paths
5. Verify `TestRuleTemplateMirrorDrift` passes for the 3 mirrored files (and TestManifestHashFormat for catalog.yaml regen if hash affected)

**Verification**: AC-COORD-005..006 + AC-COORD-009..010 PASS

### §C.5 M5 — Progress Finalization

**Deliverable**: progress.md run/sync placeholder fills + frontmatter status transitions

**Tasks**:
1. Fill progress.md §D Run-Phase Evidence (AC verification command outputs)
2. Fill progress.md §E Run-Phase Audit-Ready Signal (run_complete_at, run_commit_sha, AC PASS/FAIL roster)
3. Transition all 4 artifact frontmatter `status: draft` → `status: implemented`
4. Final B12 self-test:
   - (a) CHANGELOG `[Unreleased]` `### Added` entry count for SPEC-V3R6-MULTI-SESSION-COORD-001
   - (b) acceptance.md `^\| \*\*AC-COORD-[0-9]+` row count matches §E.2 AC count summary
   - (c) `^status: implemented` count = 4 (spec.md + plan.md + acceptance.md + progress.md)

**Verification**: progress.md §D complete + B12 PASS

## §D Implementation Order

```
M1 (Go primitive + tests)
  ├─→ M2 (CLI subcommand)           ┐
  └─→ M3 (Hook integration)         ├─→ M5 (Progress finalization)
M4 (Rule + output-style extension)  ┘
```

**Parallel-eligible**: M2 and M3 may proceed in parallel after M1 completes (both depend on `internal/session/` package but don't conflict in their own file scope). M4 is documentation-only and may proceed in parallel with any code work.

**Critical path**: M1 → M2 (or M3) → M5. Estimated wall-time: M1 ~ M2 ~ M3 each ~30-45 min implementation + verification. M4 ~ 20 min. M5 ~ 10 min.

## §E PRESERVE List (Pre-existing untracked / modified — DO NOT TOUCH in run-phase)

[HARD] The following 11 entries (current `git status` snapshot) MUST remain verbatim through plan + run + sync + Mx phases of this SPEC. Run-phase `git add` MUST use path-specific commands (NEVER `git add -A` / `git add .`).

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

**Critical**: `?? .moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/` is **another active session's work** — touching it would be the exact race condition this SPEC is designed to prevent.

**Verification post-each-commit**: `git status --short | sort` MUST match the above 11 lines verbatim (plus the freshly committed SPEC artifacts which become tracked, not untracked).

## §F Cross-Platform Considerations

### §F.1 Atomic-Write Portability

- **linux / darwin**: `O_CREATE | O_EXCL | O_WRONLY` for lockfile + `os.Rename` for atomic replace (POSIX guarantees rename atomicity within same filesystem)
- **windows**: `golang.org/x/sys/windows.MoveFileEx` with `MOVEFILE_REPLACE_EXISTING | MOVEFILE_WRITE_THROUGH` flags
- **Path separator**: always use `filepath.Join`, never string concat or `path.Join`
- **Line endings**: registry JSON file uses `\n` only (Go default), no CRLF normalization needed

### §F.2 Lockfile Semantics

- **linux / darwin**: POSIX `flock(LOCK_EX)` via `golang.org/x/sys/unix` package
- **windows**: `LockFileEx` via `golang.org/x/sys/windows` package
- **Wrapper**: `internal/session/lock_unix.go` + `internal/session/lock_windows.go` build-tagged files for platform-specific implementation
- **Timeout**: lock acquisition with 2-second timeout (avoid indefinite blocking; report `ErrLockTimeout` if exceeded)

### §F.3 JSON Marshal Stability

- Use `json.Marshal` with explicit field order (struct field order = JSON field order in Go's encoding/json)
- Pretty-print for human readability: `json.MarshalIndent(entries, "", "  ")`
- Sort entries by `started_at` ascending before write to ensure deterministic output (helpful for git diff if registry ever committed — though .gitignore'd per §F.4)

### §F.4 Registry File Gitignore

- Add `.moai/state/active-sessions.json` and `.moai/state/active-sessions.json.lock` to `.gitignore` (if not already covered by `.moai/state/` blanket ignore per CLAUDE.local.md §2 Local-Only Files section which lists `.moai/state/`)
- Verify: `git check-ignore .moai/state/active-sessions.json` returns the file path (proves gitignored)

### §F.5 Build Verification Matrix

Run-phase verification must include:
- `GOOS=linux GOARCH=amd64 go build ./...`
- `GOOS=darwin GOARCH=amd64 go build ./...`
- `GOOS=darwin GOARCH=arm64 go build ./...`
- `GOOS=windows GOARCH=amd64 go build ./...`

All four MUST exit 0 for AC-COORD-011 PASS.

## §G Anti-Patterns to Avoid

1. **AP-MSC-001 — Global mutable state in registry package**: All state lives in registry file. No package-level vars. `var defaultRegistryPath = ".moai/state/active-sessions.json"` is OK (config constant), but no `var entries []Entry` (mutable state).

2. **AP-MSC-002 — Direct file writes without lock**: Every mutation MUST go through `withLock` helper. Reads can be lock-free (eventually consistent acceptable for QueryActiveWork).

3. **AP-MSC-003 — Hard-coded paths in tests**: Use `t.TempDir()` + dependency injection (accept `registryPath string` parameter on internal functions; package-level `RegisterSession` uses default).

4. **AP-MSC-004 — Time-dependent test flakiness**: Use injectable clock (`clock.Clock` interface with `time.Now()` default + `FakeClock` for tests). Tests on `PurgeStale` MUST use fake clock.

5. **AP-MSC-005 — Block on lock acquisition without timeout**: All `flock` calls MUST have 2-sec timeout. Indefinite block is unacceptable in CLI / hook context.

6. **AP-MSC-006 — Skip cross-platform build verification**: Don't ship without `GOOS=windows` build green. Windows-specific bugs are easy to miss on darwin dev machine.

7. **AP-MSC-007 — Modify other SPEC's files**: `?? .moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/` is OFF-LIMITS. Touching it = the race this SPEC prevents.

8. **AP-MSC-008 — Skip template mirror cp**: Per L46 TEMPLATE-MIRROR-DRIFT family + CLAUDE.local.md §2 [HARD] Template-First Rule. Every `.claude/...` edit MUST have matching `internal/template/templates/.claude/...` mirror cp.

## §H Cross-References

- **spec.md** — companion §C requirements + §D architecture
- **acceptance.md** — companion AC matrix + verification commands
- **progress.md** — companion lifecycle status + commit attribution
- **CLAUDE.local.md** §23.8 Multi-Session Race Mitigation — policy this SPEC operationalizes
- **CLAUDE.local.md** §2 [HARD] Template-First Rule — mirror cp obligation
- **CLAUDE.local.md** §6 Test Isolation — `t.TempDir()` rule + `filepath.Abs` rule
- **CLAUDE.local.md** §14 [HARD] 하드코딩 방지 — registry path constant in `config/defaults.go`
- **`.claude/rules/moai/development/spec-frontmatter-schema.md`** — 12-field canonical schema (used in all 4 artifact frontmatter)
- **MEMORY.md L52** — multi-session race coordination lesson (motivating example)
