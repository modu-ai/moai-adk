---
id: SPEC-HOOK-SESSIONSTART-PROBE-001
title: "SessionStart Probe — Progress"
version: "0.1.0"
status: in-progress
created: 2026-07-07
updated: 2026-07-07
author: manager-spec
priority: P2
phase: "v14.4.0"
module: ".claude/hooks/moai"
lifecycle: spec-anchored
tags: "hook, sessionstart, probe, mode-a, genuine-risk"
---

# Progress — SPEC-HOOK-SESSIONSTART-PROBE-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-07
plan_artifacts:
  - spec.md          # 7 REQs (REQ-HOOK-001..007), GEARS format
  - plan.md          # Branch A selected, single M1 milestone, 8 anti-patterns
  - acceptance.md    # 7 ACs (AC-HOOK-001..007), all MUST-PASS, machine-verifiable
  - progress.md      # this file
tier: S              # minimal, single-pass, additive-only
lifecycle: plan-phase-complete
next_phase_gate: plan-auditor (Phase 0.5)
iter_2_corrections_applied: 2026-07-07  # plan-auditor iter-1 PASS-WITH-DEBT 0.84 → 6 defects corrected (D1 surface over-claim, D2 resume-coverage gap documented per Option 1, D3 tier:S frontmatter, D4 grep -F fixed-string, D5 REQ-HOOK-002 GEARS compound, D6 channel-role separation merged with D1)
```

**Plan-phase deliverables complete.** All 3 body artifacts authored; SPEC ID
decomposition PASSed (`SPEC | HOOK | SESSIONSTART | PROBE | 001`); frontmatter
12-canonical-field schema validated; Out of Scope section carries 7 `### Out of Scope —`
H3 sub-headings; GEARS notation used throughout (no `IF/THEN` legacy modality).

## §E.2 Run-phase Evidence

**Run-phase complete (Tier S, single-pass M1).** Branch A (wrapper-only,
self-contained bash) implemented in both the local wrapper and its template
mirror. All 7 ACs verified GREEN via the TDD harness at
`internal/hook/sessionstart_probe_test.go` (AC-002/003/004/006/007 + §D.7 edge
cases) and via the verbatim acceptance.md bash commands (reproduced below).

### AC PASS/FAIL Matrix

| AC | REQ | Status | Verification Command (acceptance.md verbatim) | Actual Output |
|----|-----|--------|-----------------------------------------------|---------------|
| AC-HOOK-001 | REQ-HOOK-001 | PASS | `diff <(git show HEAD~1:.claude/hooks/moai/handle-session-start.sh \| awk '/^# Try moai command in PATH/,/^# Try default/') <(awk '/^# Try moai command in PATH/,/^# Try default/' .claude/hooks/moai/handle-session-start.sh)` | empty diff — tier-1/2/3 chain byte-identical (awk range `/^# Try moai.../,/^# Try default/` targets line 17-22, untouched by this SPEC; only the fallback branch line 32+ changed) |
| AC-HOOK-002 | REQ-HOOK-002 | PASS | stubbed PATH/HOME invocation + stdout/stderr capture | `STDERR-OK` (hook-stderr.log contains 'moai' + 'PATH'/'go/bin'/'.local/bin'); `STDOUT-OK` (stdout carries `hookSpecificOutput` + `additionalContext`) |
| AC-HOOK-003 | REQ-HOOK-003 | PASS | `echo $?` after fallback invocation | `exit=0` |
| AC-HOOK-004 | REQ-HOOK-004 | PASS | 4-source parameterized invocation + grep count | `startup: WARN-EMITTED (expected)`; `resume/clear/compact: SUPPRESSED (expected)` |
| AC-HOOK-005 | REQ-HOOK-005 | PASS | `git diff HEAD~1 --name-only -- '.claude/hooks/moai/handle-*.sh' \| grep -v handle-session-start.sh \| wc -l` | `0` — only `handle-session-start.sh` modified; the other 30 wrappers byte-identical (verified pre-commit via `git status --porcelain`) |
| AC-HOOK-006 | REQ-HOOK-006 | PASS | `diff <(awk '/^# Not found/,/^exit 0/' local.sh) <(awk ... template.sh.tmpl)` | empty diff — `PARITY OK` (local and template fallback branches byte-identical; wrapper uses no template vars) |
| AC-HOOK-007 | REQ-HOOK-007 | PASS | warning-text content grep (4 mandated elements) | `(a) TIERS-OK` (PATH + go/bin + .local/bin); `(b) CONSEQUENCE-OK` (31 wrappers); `(c) FRAMING-OK` (non-blocking + advisory); `(d) REMEDIATION-OK` (reinstall + restore PATH + make build) |

### §D.7 Edge Cases Covered

| Edge case | Behavior | Verification |
|-----------|----------|--------------|
| Empty stdin | warning emitted (safer-to-warn) | `TestSessionStartProbe_EdgeCases_EmptyAndMalformedStdin/empty` PASS |
| Malformed JSON (`{malformed}`) | warning emitted | `.../malformed` PASS |
| Non-JSON stdin | warning emitted | `.../not-json` PASS |
| `$HOME/.moai/logs/` unwritable | graceful degrade — stdout JSON + exit 0 (stderr `2>/dev/null \|\| true`) | covered by `2>/dev/null \|\| true` suffixes (not separately tested — design.8 graceful degrade) |
| stdout closed (redirected to /dev/null) | graceful degrade — stderr log + exit 0 | covered by `2>/dev/null \|\| true` suffix on printf stdout |

### Dual-Channel Warning Text Emitted

Primary channel (stdout JSON, model-context surfacing):
```json
{"hookSpecificOutput":{"additionalContext":"moai binary not found in PATH, $HOME/go/bin, or $HOME/.local/bin — all 31 wrappers (handle-*.sh) are silently no-op (non-blocking, advisory). Reinstall moai or restore PATH (e.g. make build, then go install ./cmd/moai)."}}
```

Secondary channel (stderr audit log at `$HOME/.moai/logs/hook-stderr.log`, after-the-fact greppable):
```
[sessionstart-probe] moai binary not found in PATH, $HOME/go/bin, or $HOME/.local/bin — all 31 wrappers (handle-*.sh) are silently no-op (non-blocking, advisory). Reinstall moai or restore PATH (e.g. make build, then go install ./cmd/moai).
```

### Files Modified (run-phase, single M1 commit)

| File | Change |
|------|--------|
| `.claude/hooks/moai/handle-session-start.sh` | fallback branch (line 32+) — probe added; tier chain (line 17-30) byte-identical |
| `internal/template/templates/.claude/hooks/moai/handle-session-start.sh.tmpl` | same fallback-branch edit (byte-identical to local — AC-006) |
| `internal/hook/sessionstart_probe_test.go` | NEW — TDD harness for AC-002/003/004/006/007 + §D.7 edge cases |
| `.moai/specs/SPEC-HOOK-SESSIONSTART-PROBE-001/progress.md` | §E.2 + §E.3 populated (this file) |

### Scope-Discipline Verification

- **30 non-SessionStart wrappers**: byte-identical (AC-005 `git diff` count = 0).
- **`internal/hook/session_start.go`**: NOT modified (Branch A is wrapper-only; Go handler is §F Out of Scope).
- **3 governance gates** (`status-transition-ownership.sh`, `sync-phase-quality-gate.sh`, `team-ac-verify.sh`): NOT modified.
- **Settings.json**: SessionStart registration (matcher, timeout 30s) NOT modified.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-07
run_commit_sha: "<to-be-backfilled-after-M1-commit>"  # M1 single commit (Tier S); orchestrator-side backfill if agent context is lost
run_status: pass
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 0   # no PRESERVE-list items modified
l44_pre_commit_fetch: performed   # git rev-list --left-right origin/main...HEAD = "0 N" (local ahead, clean)
l44_post_push_fetch: pending      # Hybrid Trunk main-direct — push state recorded post-commit
new_warnings_or_lints_introduced: 0   # golangci-lint clean (0 issues on internal/hook)
cross_platform_build:
  darwin_amd64: pass              # go build ./... exit 0
  linux_amd64: not-tested         # bash wrapper is POSIX; Go test harness skipOnWindows guards the wrapper-execution tests
  windows_amd64: skip             # wrapper tests skip on Windows (POSIX bash dependency); wrapper itself is not invoked on Windows hooks
total_run_phase_files: 4          # 2 wrapper edits + 1 new test + 1 progress.md
m1_to_mn_commit_strategy: single-M1   # Tier S — no milestone split; single run-phase commit carries draft→in-progress
```

### Self-Verification Deliverables (§E.2 of manager-develop-prompt-template.md)

- **E1 AC Matrix**: 7/7 PASS (see §E.2 AC PASS/FAIL Matrix above).
- **E2 Cross-Platform Build**: `go build ./...` exit 0; `go vet ./internal/hook/...` exit 0.
- **E3 Coverage**: bash wrapper is not Go source — no per-package coverage delta applicable; the new Go test harness exercises the wrapper via `exec.Command("bash", ...)` and contributes to `internal/hook` test breadth (package already ≥85% baseline maintained).
- **E4 Scope-Boundary Grep**: `git diff HEAD --name-only -- '.claude/hooks/moai/handle-*.sh' | grep -v handle-session-start.sh` → empty (only the SessionStart wrapper touched).
- **E5 Lint**: `golangci-lint run --timeout=2m ./internal/hook/...` → 0 issues. shellcheck not available on this host (advisory skip).
- **E6 Commit/Push**: single M1 commit via pathspec `git add` (only the 4 files above); push state recorded post-commit.
- **E7 Blocker**: NONE.

### Pre-existing FAIL Tests (NOT caused by this SPEC)

The full `go test ./...` run surfaced 3 pre-existing failures unrelated to this
SPEC's scope (verified — each fails on a file this SPEC did not touch):

1. `internal/statusline` `TestBuild_WritesContextUsageWithSessionID` — context_window_size mismatch (1000000 vs 256000); statusline builder concern, not hook-layer.
2. `internal/template` `TestTemplateNoInternalContentLeak` — `templates/.moai/config/sections/cache.yaml` carries `SPEC-V3R6-PROMPT-CACHE-001` (pre-existing, cache.yaml untouched by this SPEC). This SPEC's template edit was verified §25-clean: `grep -r 'SPEC-HOOK-SESSIONSTART-PROBE-001' internal/template/templates/` → empty.
3. `internal/template` `TestSettingsTemplateRequiredEnvVars` — missing `CLAUDE_CODE_FILE_READ_MAX_OUTPUT_TOKENS` env var; settings.json.tmpl concern, untouched by this SPEC.

The `internal/hook` package — the package this SPEC modifies — passes its full
test suite cleanly (`go test ./internal/hook/...` → all PASS).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates this section with the sync-phase
completion signal, including `sync_commit_sha:` field, on the single sync
commit that carries the `implemented → completed` transition>_

---

Version: 0.1.0
Status: in-progress (run-phase complete — M1 single commit, Tier S)
