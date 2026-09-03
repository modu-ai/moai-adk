# SPEC-PRECOMMIT-VET-MONOREPO-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-04
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md
baseline_sha: b9298de32
reference_patch: t312-precommit-vet @ b6f478b1a (read + judged; adoption-by-re-authoring, spec.md §A)
twin_pair:
  - internal/cli/hook_install_precommit.go   # constant preCommitHookContent, vet block :78-103
  - internal/template/templates/.git_hooks/pre-commit   # vet block :35-59
```

Plan-phase notes for the auditor:

- **Approach settled, no OPEN decision**: adopt reference patch `b6f478b1a`'s logic (module-root walk → MODROOTS grouping → per-module subshell vet → module-root naming in the failure message), re-authored in place with generalized comments. The Implementation Kickoff Approval gate is unchanged and still mandatory before run-phase entry.
- **Local repro impossibility is structural**: this repository's `go.mod` is at the repo root, so the defect cannot be reproduced against this tree. The RED mechanism is the two contrast tests (AC-PVM-002/003) with a `submod/`-only fixture inside `t.TempDir()`, run through the existing `runPreCommitHook` harness.
- **RED ordering is load-bearing**: the RED cells must be captured at M1's opening act against the UNCHANGED constant pair, before the twin edit lands (plan.md §F M1; verification-completeness §2). AC-PVM-003's RED discriminates via the stderr naming assertion, not the exit code (exit 1 occurs on old for the wrong reason — stated in acceptance.md so the RED is red for the right stated reason).
- **Release coordination (one line, for the lead)**: t230's landing precondition is satisfied (`32d2221fa` + `539349c5b` are develop ancestors); the remaining "at least one release must pass after t230's landing before this ships" is a deployment-time concern owned by release card t204 and does not block this SPEC's phases (AC-PVM-012).

## §E.2 Run-phase Evidence

> Placeholder skeleton — populated by manager-develop (run-phase owner). Every row carries the attribution triple per VCI §2: (a) the exact command, (b) the verbatim observed output, (c) the tree SHA of the run. An empty row asserts nothing.

### M1 pre-flight baseline (to measure before the first change commit)

| Check | Command | Observed |
|---|---|---|
| Branch + HEAD | `git branch --show-current && git rev-parse --short HEAD` | _pending_ |
| Build | `go build ./...` | _pending_ |
| Cross-platform | `GOOS=windows GOARCH=amd64 go build ./...` | _pending_ |
| Pre-change guards | `go test -count=1 ./internal/cli/ -run 'TestPreCommit'` | _pending_ |
| Lint baseline | `golangci-lint run --timeout=2m ./internal/cli/...` | _pending_ |

### M1 opening act — RED cells (captured against the PRE-EDIT constant pair)

| Test | Failing command | Verbatim RED output + exit code | Fixture + tree SHA |
|---|---|---|---|
| `TestPreCommitHook_SubmodulePassesClean` | `go test -count=1 ./internal/cli/ -run 'TestPreCommitHook_SubmodulePassesClean'` | _pending — expected RED via exit code (got 1, expected 0: clean submodule file blocked)_ | _submod/-only fixture, t.TempDir(); tree SHA pre-edit_ |
| `TestPreCommitHook_SubmoduleVetBlocks` | `go test -count=1 ./internal/cli/ -run 'TestPreCommitHook_SubmoduleVetBlocks'` | _pending — expected RED via the stderr discriminator (`module root: submod` absent; exit 1 occurs on old for the wrong reason)_ | _same fixture; tree SHA pre-edit_ |

### AC matrix (populated as milestones close)

| AC | Status | Verification (command → observed, this run, tree SHA) |
|---|---|---|
| AC-PVM-001 (twin identity, guard) | _pending_ | _pending_ |
| AC-PVM-002 (submodule clean pass) | _pending_ | _pending_ |
| AC-PVM-003 (submodule vet block + naming) | _pending_ | _pending_ |
| AC-PVM-004 (repo-root fallback guard) | _pending_ | _pending_ |
| AC-PVM-005 (build tags before cd) | _pending_ | _pending_ |
| AC-PVM-006 (subshell isolation, both surfaces) | _pending_ | _pending_ |
| AC-PVM-007 (gofmt guard) | _pending_ | _pending_ |
| AC-PVM-008 (skip-bypass guard) | _pending_ | _pending_ |
| AC-PVM-009 (toolchain-absent / no-staged-Go guards) | _pending_ | _pending_ |
| AC-PVM-010 (package sweep, ≥600s, rc unpiped) | _pending_ | _pending_ |
| AC-PVM-011 (cross-platform build) | _pending_ | _pending_ |
| AC-PVM-012 (release-coordination line) | _pending_ | _pending_ |

## §E.3 Run-phase Audit-Ready Signal

> Placeholder — populated by manager-develop at run-phase close. Expected content: the completed AC matrix reference, the two-cell adoption record (RED from M1's opening act + GREEN flip), sweep outputs (go test / build / coverage / lint delta), and the release-coordination line naming t204.

```yaml
run_status: pending
run_complete_at: null
```

## §E.4 Sync-phase Audit-Ready Signal

> Placeholder — populated by manager-docs on the single sync commit (3-phase close).

```yaml
sync_status: pending
sync_commit_sha: pending-backfill-sync
```
