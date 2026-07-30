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

run_status: audit-ready
run_complete_at: 2026-07-31
milestones_complete: M1, M2, M3, M4, M5, M6

### M6 — falsifiability proof + full-suite verification (executed by the orchestrator)

Precondition verified before any probe: `git status --porcelain` on the four probe
targets was empty, i.e. every M1-M5 transition was already committed, so
`git checkout --` could not discard uncommitted work.

**AC-EAS-007 — SSOT exclusion is present AND narrow — PASS**

| Step | Observed |
|------|----------|
| (a) clean tree | `--- PASS: TestNoBareAnthropicEnvVarLiteralsInProduction (0.28s)` |
| (b) plant in sibling `internal/config/log.go` | `exit=1`; `--- FAIL:` line count `1`; offender naming `config/log.go` count `1`; guard-ran evidence `anthropic_env_ssot_test.go` count `1` |
| (c) revert | `git status --porcelain internal/config/log.go` → 0 lines; `--- PASS:` again |

The sibling plant being detected is the positive proof that the `envkeys.go`
exclusion is an exact-path match and not a package-wide exemption.

**AC-EAS-012 — guard falsifiable in all three enumerated roots — PASS**

| Leg | Root | Observed |
|-----|------|----------|
| (b) | `internal/` — `internal/sandbox/env.go` | `exit=1`, `--- FAIL` ×1, file named ×1, revert residue 0 |
| (d) | `cmd/` — `cmd/moai/main.go` | `exit=1`, `--- FAIL` ×1, file named ×1, revert residue 0 |
| (f) | `pkg/` — `pkg/version/version.go` | `exit=1`, `--- FAIL` ×1, file named ×1, revert residue 0 |
| (a)/(post) | clean tree before and after | `--- PASS:` both times |

(e) per-name falsification, all 9 banned names planted in turn into
`internal/sandbox/env.go` and reverted: **9/9 produced a non-zero guard exit**,
including the three names with zero natural offenders (`ANTHROPIC_`,
`ANTHROPIC_DEFAULT_FABLE_MODEL`, `ANTHROPIC_REASONING_EFFORT`) that no other AC
exercises. Total residual diff after the loop: `git status --porcelain` → 0 lines.

**AC-EAS-013 — behaviour preservation — PASS, with a disclosed flaky observation**

| Leg | Observed |
|-----|----------|
| `go vet ./...` | no output, `vet_exit=0` |
| `golangci-lint run --timeout=3m` | `0 issues.`, `lint_exit=0` (exit captured directly, not after a pipe) |
| `go test ./...` run 1 | `exit=1`; 2 failing packages: `internal/lsp/subprocess` (`TestSupervisor_MultipleWatchers`), `internal/web`; 104 `ok` |
| `go test -count=1 ./...` run 2 | `exit=0`, 0 `^FAIL`, 106 `ok` |
| `go test -count=1 ./...` run 3 | `exit=0`, 0 `^FAIL`, 106 `ok` |

Run 1's failure was investigated rather than assumed away, and is attributed to
pre-existing load-sensitive flakiness, not to this refactor. Evidence for that
attribution:

- Both packages pass **in isolation** on this tree (`go test ./internal/lsp/subprocess/... ./internal/web/...` → `exit=0`, both `ok`).
- Both packages contain **zero** `ANTHROPIC` references (`grep -rn 'ANTHROPIC'` → 0 each).
- **Zero** files changed by M1-M5 live in either package (`git diff --name-only 76d9a8f3b..HEAD -- internal/lsp internal/web` → 0).
- Two subsequent cache-disabled full-suite runs were clean.
- `TestSupervisor_MultipleWatchers` is a ~10 s multi-watcher timing test — the shape that degrades under a 106-package parallel run.

Counter-evidence recorded for honesty: a single full-suite run at base
`76d9a8f3b` (in a temporary probe worktree, since removed) was clean (0 `^FAIL`,
106 `ok`). That is one sample against this tree's three, so it does not establish
that base is immune; it is disclosed rather than omitted.

**Stability ACs re-confirmed at M6**

| AC | Command | Observed | Required |
|----|---------|----------|----------|
| AC-EAS-011 | production bare literals, `envkeys.go` excluded | `0` | `0` |
| AC-EAS-014 | `git diff --name-only 76d9a8f3b -- internal/template/templates/ \| wc -l` | `0` | `0` |
| AC-EAS-015 | `_test.go` ANTHROPIC occurrences | `291` | `291` (unchanged) |
| AC-EAS-001 | `envkeys.go` ANTHROPIC literals | `9` | `9` (6 base + 3 from M1) |

Cross-platform: `go build ./...` → 0; `GOOS=windows GOARCH=amd64 go build ./...` → 0.

**Gaps at run-phase close**

- No Windows test *run* (cross-compile only); `internal/hook` and `internal/tmux` carry platform-conditional code.
- No end-to-end runtime exercise of the GLM/tmux/hook paths against a live tmux
  session or a real endpoint. Behaviour preservation rests on the value-identity
  attestation (forward round-trip from base, all 10 files `IDENTICAL`) plus the
  per-package suites.
- Coverage was not measured; this refactor adds no production branches.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
