---
spec: SPEC-ENVKEY-ANTHROPIC-SSOT-001
phase: run
---

# Progress - SPEC-ENVKEY-ANTHROPIC-SSOT-001

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored at base commit `76d9a8f3b` on branch
`plan/SPEC-ENVKEY-ANTHROPIC-SSOT-001`.

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- SPEC ID regex pre-write check: executed as Bash, observed `PASS`
- Duplicate SPEC ID check: observed `NO_DUPLICATE_SPEC_ID`
- `moai spec lint` on `spec.md`: observed `✓ No findings — all SPEC documents are valid`
- Baseline measurements re-verified in this worktree (not carried over):
  83 bare literals / 10 files / 9 unique values / 6 existing constants /
  46 reachable + 37 unreachable by the existing guard / 291 `_test.go` refs /
  12 template refs / `go vet ./...` exit 0
- 15 acceptance criteria authored, each carrying either an executed verdict
  command with its observed pre-change baseline, or an explicit
  `not runnable at base` marking with the reason
- 4 open questions (OQ-1..OQ-4) recorded with recommendations, pending
  resolution at the Implementation Kickoff Approval gate

### Audit iteration 2 (plan-auditor FAIL 0.72 -> remediation)

An independent plan-auditor returned FAIL (0.72) with 5 MUST-FIX and 6 SHOULD-FIX
findings. Remediation landed in `spec.md` v0.2.0, `plan.md`, and `acceptance.md`.
Every remediating code fact was re-verified by executed command in this worktree:

- `lifecycle` corrected `spec-first` -> `spec-anchored` (canonical enum,
  `.claude/rules/moai/development/spec-frontmatter-schema.md:48`)
- AC-EAS-006 rebuilt as a **runtime** banned-set assertion; the prior source-grep
  baseline was fabricated (executed: exit 0 with 6 lines, not exit 2) and
  contradicted plan.md's derive-from-constants design
- AC-EAS-013 lint verdict rebuilt to capture the exit code before the pipe;
  vacuity proven by executed `bash -c 'false | tail -5; echo $?'` -> `0`
- M6 falsification retargeted `internal/hook/session_start.go` ->
  `internal/sandbox/env.go:32` (the sole production `ANTHROPIC_API_KEY` site,
  confirmed by repo-wide grep)
- Guard root-finder retargeted to `findProjectRoot(t)` at
  `internal/config/required_checks_test.go:162`; the
  `findRepoRoot`-redeclaration hazard against
  `internal/config/token_budget_guard.go:45` recorded in plan.md F/M2
- AC-EAS-007 gained a positive scope assertion (sibling-file plant); AC-EAS-012
  gained a `cmd/` plant, a 9-name falsification table, and residual-diff checks
- Missing baselines executed and recorded: per-package `^FAIL` counts (all `0`),
  `golangci-lint run --timeout=3m` -> `0 issues.` / `lint_exit=0`, template
  changed-file count `0`
- Census re-measurement gate added as a [HARD] run-phase entry precondition
  (spec.md A.5, plan.md C.1)

## §E.2 Run-phase Evidence

### M3-M5 — literal-to-constant transition (AC-EAS-008 / 009 / 010 / 011)

| AC | Milestone | Status | Verification Command | Actual Output |
|----|-----------|--------|----------------------|---------------|
| AC-EAS-008 | M3 | PASS | `grep -rn '"ANTHROPIC_[A-Z_]*"' internal/cli/ --include='*.go' --exclude='*_test.go' \| wc -l` | `0` |
| AC-EAS-008 | M3 | PASS | `go test ./internal/cli/... -count=1 2>&1 \| grep -c '^FAIL'` | `0` (test_exit=0) |
| AC-EAS-009 | M4 | PASS | `grep -rn '"ANTHROPIC_[A-Z_]*"' internal/hook/ --include='*.go' --exclude='*_test.go' \| wc -l` | `0` |
| AC-EAS-009 | M4 | PASS | `go test ./internal/hook/... -count=1 2>&1 \| grep -c '^FAIL'` | `0` (test_exit=0) |
| AC-EAS-010 | M5 | PASS | `grep -rn '"ANTHROPIC_[A-Z_]*"' internal/tmux/ internal/statusline/ internal/sandbox/ --include='*.go' --exclude='*_test.go' \| wc -l` | `0` |
| AC-EAS-010 | M5 | PASS | `go test ./internal/tmux/... ./internal/statusline/... ./internal/sandbox/... -count=1 2>&1 \| grep -c '^FAIL'` | `0` (`ok` for all three packages) |
| AC-EAS-011 | M5 | PASS | `grep -rn '"ANTHROPIC_[A-Z_]*"' internal/ pkg/ cmd/ --include='*.go' --exclude='*_test.go' \| grep -v '^internal/config/envkeys.go' \| wc -l` | `0` |
| AC-EAS-011 | M5 | PASS | `go test ./internal/config/ -run 'TestNoBareAnthropicEnvVarLiteralsInProduction' -v -count=1` | `--- PASS: TestNoBareAnthropicEnvVarLiteralsInProduction (0.29s)` / `ok  github.com/modu-ai/moai-adk/internal/config  6.862s` (guard_exit=0) |

### Guard trajectory (the RED-window progress meter)

| Point | Offender count | Guard verdict |
|-------|----------------|---------------|
| M2 landed (base of this run) | 83 | `--- FAIL` |
| after M3 (`internal/cli`, 46 replaced) | 37 | `--- FAIL`, zero `internal/cli/` paths |
| after M4 (`internal/hook`, 29 replaced) | 8 | `--- FAIL`, zero `internal/hook/` paths |
| after M5 (remaining 8 replaced) | 0 | `--- PASS` — RED window closed |

Offender count measured as
`grep -cE '\.go:[0-9]+:[0-9]+: ANTHROPIC_' <guard-log>`.

### Per-file replacement counts (46 + 29 + 8 = 83)

| File | Replaced |
|------|---------:|
| `internal/cli/glm.go` | 32 |
| `internal/cli/launcher.go` | 7 |
| `internal/cli/settings.go` | 6 |
| `internal/cli/worktree/tmux_integration.go` | 1 (prefix) |
| `internal/hook/session_end.go` | 12 |
| `internal/hook/glm_tmux.go` | 9 |
| `internal/hook/session_start.go` | 8 |
| `internal/tmux/cg_detect.go` | 4 |
| `internal/statusline/metrics.go` | 3 |
| `internal/sandbox/env.go` | 1 |

`internal/config` imports were added to `glm_tmux.go`, `cg_detect.go`,
`metrics.go`, and `env.go` (no import cycle: `internal/config` depends only on
`internal/defs` + `pkg/models`).

### Behaviour-preservation evidence

- Each literal was replaced by the constant whose declared value in
  `envkeys.go` is byte-identical (all 9 pairs machine-verified).
- Forward attestation: applying the same value-preserving mapping to the base
  (`f4055ca36`) source of all 10 files and gofmt-comparing against the working
  tree reports `IDENTICAL` for all 10 (`FORWARD_ALL_IDENTICAL`) — the tree is
  exactly the mechanical transform of base, so no wrong-but-compiling constant
  was substituted.
- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go vet ./...` no output, exit 0.
- `golangci-lint run --timeout=3m` → `0 issues.`, exit 0 (identical to the
  pre-M3 baseline; zero NEW issues).

### Untouched-surface confirmation (AC-EAS-014 / AC-EAS-015 spot-check)

- `_test.go` ANTHROPIC_* occurrences: `291` (unchanged).
- `internal/template/templates/` ANTHROPIC_ references: `12`; changed-file
  count vs `f4055ca36`: `0`.

M6 (falsifiability proof + full-suite verification) is NOT part of this run.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
