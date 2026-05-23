---
id: SPEC-V3R6-LEGACY-CLEANUP-003
title: "SPEC-V3R6-LEGACY-CLEANUP-003 — Implementation Plan"
version: "0.2.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P3
tags: "cleanup, legacy, terminology, sprint-2, plan"
issue_number: null
tier: M
phase: "v3.0.0"
module: "internal/runtime"
lifecycle: spec-anchored
---

# SPEC-V3R6-LEGACY-CLEANUP-003 — Implementation Plan

## §1 — Approach

Surgical Wave→Round rename across 16 production Go files + 1 test file with one API parameter rename. No new functionality. No public symbol removal. No SPEC behavior change beyond terminology alignment with `.claude/rules/moai/development/sprint-round-naming.md` v2.0.0.

Execution model: `manager-develop` Tier L 4-milestone sequential cycle. Each milestone has a deterministic deliverables list, file-by-file edit map, and a per-milestone verification command.

## §2 — Milestones

### M1 — Comment-only Wave→Round renames (14 files, low-risk batch)

**Deliverable**: 14 files with Wave→Round comment renames applied per REQ-LCL-001 file-by-file map.

**Edit map** (file → line range → operation):

| File | Lines | Operation |
|------|-------|-----------|
| `internal/ciwatch/classifier.go` | 1 | Rename `(Wave 2, T2)` → `(Round 2, T2)` |
| `internal/ciwatch/handoff.go` | 38, 52, 55 | Rename `Wave 3` → `Round 3` (3 occurrences) |
| `internal/cli/hook.go` | 104 | Rename `Wave A` → `Round A` |
| `internal/cli/pr/watch.go` | 17, 75 | Rename `Wave 2` → `Round 2`, `Wave 3` → `Round 3` |
| `internal/cli/worktree/guard.go` | 19 | Rename `Wave 5` → `Round 5` (surrounding text; SPEC-V3R3-CI-AUTONOMY-001 reference retained) |
| `internal/cli/worktree/new.go` | 50 | Rename `Wave 7` → `Round 7` (surrounding text) |
| `internal/config/required_checks.go` | 29, 54 | Rename `Wave 2` → `Round 2` (2 occurrences) |
| `internal/harness/types.go` | 33 | Rename `Wave C` → `Round C` |
| `internal/hook/session_start.go` | 153 | Rename `Wave 3: W3-T3` → `Round 3: W3-T3` |
| `internal/hook/spec_status.go` | 49, 77 | Rename `Wave 2: AC-02.b` → `Round 2: AC-02.b`, `Wave 1` → `Round 1` |
| `internal/runtime/budget.go` | 166 | Rename `smaller waves` → `smaller rounds` |
| `internal/spec/lint.go` | 956 | Rename `Implements Wave 3: W3-T4` → `Implements Round 3: W3-T4` |
| `internal/worktree/doc.go` | 11 | Rename `Wave 5 (T6 Worktree State Guard)` → `Round 5 (T6 Worktree State Guard)`. Retain `SPEC: SPEC-V3R3-CI-AUTONOMY-001` verbatim (immutable historical SPEC-ID per REQ-LCL-001 decision rule) |
| `internal/worktree/state_guard.go` | 15, 39 | Line 15: retain `@MX:SPEC: SPEC-V3R3-CI-AUTONOMY-001 Wave 5` verbatim (historical SPEC-ID artifact per REQ-LCL-001 decision rule §a). Line 39: **retain `strategy-wave5.md §7` verbatim** — pre-flight verified file exists on disk at `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave5.md` (40155 bytes, mtime May 9 23:58); renaming the comment would create a dead reference. This is the §b immutable exemption per REQ-LCL-001 decision rule. NO edit applied to line 39. |

**M1 verification**:

```bash
# Expected: 14 files modified, ~22 lines changed
git diff --stat internal/ciwatch/ internal/cli/ internal/config/required_checks.go \
  internal/harness/types.go internal/hook/session_start.go internal/hook/spec_status.go \
  internal/runtime/budget.go internal/spec/lint.go internal/worktree/

# Expected: 0 Wave-keyword occurrences remaining EXCEPT 3 immutable exemptions (REQ-LCL-001 §a/b/c): state_guard.go:15+doc.go:11 historical SPEC-ID + state_guard.go:39 strategy-wave5.md file reference (file exists verbatim on disk)
grep -ni "wave" internal/ciwatch/ internal/cli/hook.go internal/cli/pr/watch.go \
  internal/cli/worktree/guard.go internal/cli/worktree/new.go internal/config/required_checks.go \
  internal/harness/types.go internal/hook/session_start.go internal/hook/spec_status.go \
  internal/runtime/budget.go internal/spec/lint.go internal/worktree/ --include="*.go"

# Expected: PASS (no API change, comment-only)
go build ./...
go vet ./internal/ciwatch/... ./internal/cli/... ./internal/config/... \
  ./internal/harness/... ./internal/hook/... ./internal/runtime/... \
  ./internal/spec/... ./internal/worktree/...
```

### M2 — Public API parameter rename (`internal/runtime/persist.go`)

**Deliverable**: `PersistProgress` and `buildResumeMessage` carry `roundLabel` parameter name; section template literal uses `- Round: %s` heading; `strings.ReplaceAll` operates on `{round_label}`.

**Edit map** (`internal/runtime/persist.go`):

| Line | Before | After |
|------|--------|-------|
| 18 | `func (t *Tracker) PersistProgress(specID, waveLabel, approach, nextStep string) (string, error) {` | `func (t *Tracker) PersistProgress(specID, roundLabel, approach, nextStep string) (string, error) {` |
| 38 | `section := fmt.Sprintf("\n## Auto-saved at %s (75%% threshold)\n\n- Wave: %s\n- Approach: %s\n- Next step: %s\n",` | `section := fmt.Sprintf("\n## Auto-saved at %s (75%% threshold)\n\n- Round: %s\n- Approach: %s\n- Next step: %s\n",` |
| 39 | `timestamp, waveLabel, approach, nextStep)` | `timestamp, roundLabel, approach, nextStep)` |
| 58 | `resumeMsg := buildResumeMessage(cfg.ResumeMessageFormat, specID, waveLabel, approach, progressPath, nextStep)` | `resumeMsg := buildResumeMessage(cfg.ResumeMessageFormat, specID, roundLabel, approach, progressPath, nextStep)` |
| 89 | `func buildResumeMessage(format, specID, waveLabel, approach, progressPath, nextStep string) string {` | `func buildResumeMessage(format, specID, roundLabel, approach, progressPath, nextStep string) string {` |
| 93 | `msg = strings.ReplaceAll(msg, "{wave_label}", waveLabel)` | `msg = strings.ReplaceAll(msg, "{round_label}", roundLabel)` |

**M2 verification**:

```bash
# Expected: 0 waveLabel/wave_label/Wave: occurrences in persist.go
grep -n "waveLabel\|wave_label\|- Wave:" internal/runtime/persist.go

# Expected: 6 roundLabel/round_label/Round: occurrences
grep -n "roundLabel\|round_label\|- Round:" internal/runtime/persist.go

# Compiles
go build ./internal/runtime/...
go vet ./internal/runtime/...

# Tests will fail until M3+M4; explicit expected to fail here
```

### M3 — ResumeMessageFormat default + DefaultFallback constant + budget.go message text

**Deliverable**: `internal/runtime/config.go` line 28 + 136 updated; `internal/runtime/budget.go` line 166 string updated.

**Edit map**:

| File:Line | Before | After |
|-----------|--------|-------|
| `internal/runtime/config.go:28` | `DefaultFallback = "split_into_waves"` | `DefaultFallback = "split_into_rounds"` |
| `internal/runtime/config.go:136` | `ResumeMessageFormat:   "ultrathink. {wave_label} 이어서 진행. SPEC-{spec_id}부터 {approach_summary}. progress.md 경로: {progress_path}. 다음 단계: {next_step}.",` | `ResumeMessageFormat:   "ultrathink. {round_label} 이어서 진행. SPEC-{spec_id}부터 {approach_summary}. progress.md 경로: {progress_path}. 다음 단계: {next_step}.",` |
| `internal/runtime/budget.go:166` | `"Consider splitting the work into smaller waves. /clear is NOT auto-triggered."` | `"Consider splitting the work into smaller rounds. /clear is NOT auto-triggered."` |

**M3 verification**:

```bash
# Expected: 0
grep -n "split_into_waves\|smaller waves\|{wave_label}" internal/runtime/config.go internal/runtime/budget.go

# Expected: 3 (DefaultFallback, ResumeMessageFormat, budget.go message)
grep -n "split_into_rounds\|smaller rounds\|{round_label}" internal/runtime/config.go internal/runtime/budget.go

# Compiles
go build ./internal/runtime/...
```

### M4 — Test file alignment + full verification

**Deliverable**: `internal/runtime/budget_test.go` updated to match production rename; `go test ./...` passes; quality gates clean.

**Edit map** (`internal/runtime/budget_test.go`):

| Line | Before | After |
|------|--------|-------|
| 24 | `Fallback:              "split_into_waves",` | `Fallback:              "split_into_rounds",` |
| 27 | `ResumeMessageFormat:   "ultrathink. {wave_label} 이어서 진행. SPEC-{spec_id}부터 {approach_summary}. progress.md 경로: {progress_path}. 다음 단계: {next_step}.",` | `ResumeMessageFormat:   "ultrathink. {round_label} 이어서 진행. SPEC-{spec_id}부터 {approach_summary}. progress.md 경로: {progress_path}. 다음 단계: {next_step}.",` |
| 192-193 | `if !strings.Contains(recommendation, "split_into_waves") { ... "expected fallback=split_into_waves in recommendation, got %q" ...` | `if !strings.Contains(recommendation, "split_into_rounds") { ... "expected fallback=split_into_rounds in recommendation, got %q" ...` |
| 212 | `msg, err := tracker.PersistProgress(specID, "Wave 1", "budget tracker 구현", "다음: /moai sync")` | `msg, err := tracker.PersistProgress(specID, "Round 1", "budget tracker 구현", "다음: /moai sync")` |
| 274 | `msg, err := tracker.PersistProgress("SPEC-NONEXISTENT", "Wave 1", "test", "next step")` | `msg, err := tracker.PersistProgress("SPEC-NONEXISTENT", "Round 1", "test", "next step")` |
| 382 | `    fallback: split_into_waves` (inside YAML test fixture) | `    fallback: split_into_rounds` |
| 450 | `_, err := tracker.PersistProgress(specID, "Wave 2", "budget fix", "next: sync")` | `_, err := tracker.PersistProgress(specID, "Round 2", "budget fix", "next: sync")` |
| 485 | `msg, err := tracker.PersistProgress(specID, "Wave 3", "test approach", "/moai sync")` | `msg, err := tracker.PersistProgress(specID, "Round 3", "test approach", "/moai sync")` |
| **491** | `required := []string{"ultrathink", specID, "Wave 3", "test approach", "/moai sync", "progress.md"}` | `required := []string{"ultrathink", specID, "Round 3", "test approach", "/moai sync", "progress.md"}` (assertion expectation paired with line 485 input rename; renaming 485 without 491 produces green-broken-build per REQ-LCL-005 §B.5 + plan-auditor B3 finding 2026-05-24) |

**M4 verification** (full 7-item parallel batch per `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution):

```bash
# 1. Wave-keyword final audit (expected: 0 in production .go + budget_test.go, EXCEPT immutable SPEC-V3R3-CI-AUTONOMY-001 Wave 5 reference in state_guard.go:15)
grep -rni "wave" internal/ pkg/ --include="*.go" 2>/dev/null | grep -v "Wave 5" | grep -v "audit_test" | grep -v "hook_opt_in" | grep -v "observability.go" | grep -v "handle-harness-observe"

# 2. Round-keyword forward audit (expected: ~28 occurrences in renamed locations)
grep -rni "round" internal/runtime/ internal/ciwatch/ internal/cli/hook.go internal/cli/pr/watch.go internal/cli/worktree/guard.go internal/cli/worktree/new.go internal/config/required_checks.go internal/harness/types.go internal/hook/session_start.go internal/hook/spec_status.go internal/spec/lint.go internal/worktree/ --include="*.go" | wc -l

# 3. Full test suite
go test ./...

# 4. Vet
go vet ./...

# 5. Race-condition guard on runtime package (PersistProgress concurrency)
go test -race ./internal/runtime/...

# 6. Lint baseline
golangci-lint run --timeout=2m

# 7. CLI smoke
go run ./cmd/moai --version
```

## §3 — Risk Mitigation

| Risk | Mitigation | Detection |
|------|-----------|-----------|
| `strategy-wave5.md` file reference in state_guard.go:39 — resolved at plan-phase | Pre-flight verified file exists at `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave5.md` (40155 bytes, mtime May 9 23:58). Decision: **retain verbatim** as REQ-LCL-001 §b immutable exemption. No edit applied to line 39. AC-LCL-001 grep includes `grep -v "strategy-wave5"` exclusion to exempt from rename audit. | n/a — resolved at plan-phase per plan-auditor B4 finding 2026-05-24 |
| `runtime.yaml` user override silently breaks per D-1 backward-compat | Document in CHANGELOG via `/moai sync` (out of run-phase scope). No code-side mitigation | Users observe `{wave_label}` literal in resume output |
| Test fixture YAML at `budget_test.go:382` is parsed by yaml.Unmarshal — verify schema match | M4 verification runs full test suite catching yaml parse mismatch | `go test ./internal/runtime/...` fails |
| Plan-auditor REVISE on iter-1 (precedent: 0.78 → fix-forward) | Iter-2 fix-forward orchestrator-direct edits per L32 precedent if textual/regex only | iter-1 verdict < 0.85 |
| Multi-session race during M2 API rename (per CLAUDE.local.md §23.8) | Pre-spawn `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` 4-case matrix per agent-common-protocol.md §Pre-Spawn Sync Check | `N 0` output means STOP + AskUser |

## §4 — Run-Phase Spawn Prompt Skeleton (for orchestrator reference)

Tier M Section A-E template (~1500-2000 tokens minimal-to-medium form). Orchestrator constructs from this skeleton when invoking `manager-develop` for `/moai run SPEC-V3R6-LEGACY-CLEANUP-003`:

- Section A: Mission + Scope (from spec.md §A)
- Section B: Verified facts (from spec.md §A.4 — orchestrator's Section B.3 inverted; pre-flight Wave count 28/16; budget_test.go 10 occurrences across 5 line locations 212/274/450/485/491)
- Section C: Pre-flight commands (from acceptance.md per-AC verification commands)
- Section D: Tier classification (Tier M per spec.md §A.5 L40 per-SPEC override, AskUserQuestion-confirmed 2026-05-24)
- Section E: 4-milestone deliverables matrix (from plan.md §2)

Cycle type: `ddd` per `.moai/config/sections/quality.yaml` `development_mode`. (DDD ANALYZE-PRESERVE-IMPROVE applies — existing code, no new behavior; characterization tests are the M4 `go test ./...` baseline confirming behavior preservation.)

## §5 — Definition of Done

- [ ] All 4 milestones M1-M4 commits on main with SPEC-attribution in commit subject
- [ ] All 8 acceptance.md ACs PASS (orchestrator independent trust-but-verify batch)
- [ ] Zero new Wave keyword in production Go (REQ-LCL-006)
- [ ] [Unwanted] §A.6 categories byte-identical (REQ-LCL-007)
- [ ] Self-compliance: no NEW Wave terminology in SPEC artifacts (REQ-LCL-008)
- [ ] `go test ./...` PASS
- [ ] `go vet ./...` 0 errors
- [ ] `golangci-lint run --timeout=2m` 0 errors
- [ ] progress.md status matrix row updated per L28 reinforced mitigation (explicit deliverable to manager-docs)
- [ ] CHANGELOG.md `[Unreleased]` `### Changed` entry added by `/moai sync` (separate phase, out of run scope)
