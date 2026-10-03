# SPEC-TPL-AST-GUARD-001 — Acceptance Criteria

All ACs are mechanically verifiable. Mutation ACs follow the cycle: apply mutation → run guard → capture verbatim failing output → revert → re-observe clean pass. Evidence: `.moai/reports/t1377/` (primary; citation-target). Guard invocation shorthand below: `GUARD` = `go test ./internal/template/ -run '^TestWorkflowWorktreeKeyHonesty$' -count=1 -v` (exact test name set at M1; the anchored `-run` selector names only the new test — package-level gates are separate ACs).

All 12 ACs are regression-guards pending run-phase mutation evidence (spec §C C-5): none carries a plan-phase RED-now cell, and observed-failure evidence is produced by the M3 mutation cycles (plan.md §F M3).

## §D AC Matrix

| AC | Requirement(s) | Severity | Verification |
|----|----------------|----------|--------------|
| AC-001 | REQ-001, REQ-003 | must | clean-tree pass |
| AC-002 | REQ-003 | must | dropped-reader mutation |
| AC-003 | REQ-002, REQ-003 | must | unnamed-reader mutation |
| AC-004 | REQ-004 | must | either-site mutation (2 arms) |
| AC-005 | REQ-001, REQ-002, REQ-009 | must | alias + write-only fixture characterization |
| AC-006 | REQ-006 | must | table-entry deletion mutation |
| AC-007 | REQ-005 | must | reserved-key mutation + table-invalid mutation |
| AC-008 | REQ-007 | must | type-error mutation |
| AC-009 | REQ-001 | must | package test gate |
| AC-010 | REQ-001 (TRUST 5 Readable/Unified gate anchor) | must | lint gate |
| AC-011 | REQ-001 (TRUST 5 Unified gate anchor) | must | format gate |
| AC-012 | REQ-008 | must | test-file exclusion negative test |

## §D.1 AC-001 — Clean tree passes

- **Given** the card worktree at baseline with no mutations, **When** `GUARD` runs, **Then** exit 0, the test reports ok, **at least one test is executed** (an executed-test count of zero means the selector matched nothing — a vacuous pass that fails this AC), and the reader index is logged (packages/files scanned > 0 — non-vacuity).

## §D.2 AC-002 — Dropped expected reader fails

- **Given** baseline, **When** the `cfg.Workflow.Worktree.AutoCleanup` read in `internal/cli/session_worktree.go` is deleted (mutation), **When** `GUARD` runs, **Then** the guard FAILS with a finding naming `auto_cleanup` and `internal/cli/session_worktree.go` as an expected reader no longer reading the field.

## §D.3 AC-003 — Unnamed reader fails

- **Given** baseline, **When** a production read of a `WorkflowWorktreeConfig` field is added in a file absent from the expectation table (e.g. `_ = cfg.Workflow.Worktree.AutoMerge` inside an unrelated `internal/cli` file), **When** `GUARD` runs, **Then** the guard FAILS naming that file as an unnamed reader of the field.

## §D.4 AC-004 — Either auto_cleanup site alone fails (both-sites requirement)

- **Given** baseline, **When** arm 1 deletes the `AutoCleanup` read in `internal/cli/session_worktree.go` only, **Then** `GUARD` FAILS; **When** arm 2 (from clean tree) deletes the read in `internal/cli/session_worktree_prmerge.go` only, **Then** `GUARD` FAILS; **When** arm 3 (from clean tree) deletes the `prMergeCleanup` read AND removes the `internal/cli/session_worktree_prmerge.go` table entry in the same mutation — the maintenance path plan §C pre-flight endorses — **Then** `GUARD` still FAILS via the independent table-content assertion (REQ-004), naming the missing mandatory site. Each arm names the file removed. (Traceability: PR #1707 review — either-one-alone acceptance was the be8b2dc defect; arm 3 closes the same-commit bypass that set equality alone cannot catch.)

## §D.5 AC-005 — Alias-copy detection (characterization)

- **Given** the fixture package `internal/template/testdata/worktreekeyaliasprobe/` whose file `aliasprobe.go` reads `AutoCleanup` through a local alias copy and whose second file `writeonly.go` touches the same field ONLY as a pure write (`w.AutoCleanup = false`), **When** the guard's fixture-mode scan loads the package by explicit pattern with the production matcher, **Then** (a) `aliasprobe.go` is reported as a reader of `AutoCleanup`; (b) `writeonly.go` is NOT reported as a reader (REQ-002's write exclusion); and **When** the legacy text accessor string `.Workflow.Worktree.AutoCleanup` is matched against `aliasprobe.go`, **Then** (c) zero matches — documenting the defect class the retained t682 text scan (still on develop) cannot see; the PR #1707 review's MAJOR drove the PR to fix exactly this in its final AST-based state, and this AC re-proves the property on the adopted mechanism.

## §D.6 AC-006 — Table completeness against the live struct

- **Given** baseline, **When** one key's entry is deleted from the expectation table (mutation), **When** `GUARD` runs, **Then** the guard FAILS reporting the field with no expectation entry.

## §D.7 AC-007 — Reserved-key integrity

- **AC-007a**: **Given** baseline, **When** a production read of `SessionNamePattern` or `TmuxPreferred` is added anywhere in non-test code (mutation), **Then** `GUARD` FAILS — a reserved key gained an unnamed reader. The added read may use the direct or the alias-copy shape (`w := cfg.Workflow.Worktree; _ = w.SessionNamePattern`) — both must be caught (REQ-002's type-resolved attribution is shape-independent; set equality catches both at the table level).
- **AC-007b**: **Given** baseline, **When** the expectation table names a file under a reserved key (mutation), **Then** the guard FAILS at table validation — the table contradicts its own empty-reader claim (the PR review Minor, made structural).

## §D.8 AC-008 — Type errors fail the guard

- **Given** baseline, **When** a scanned package carries a type error (mutation: e.g. an undefined identifier in an `internal/cli` file), **Then** `GUARD` FAILS rather than passing on a possibly-incomplete reader index.

## §D.9 AC-009..011 — Package quality gates

- **AC-009**: `go test ./internal/template/ -count=1` → ok (whole package, no selector).
- **AC-010**: `golangci-lint run ./internal/template/...` → 0 issues.
- **AC-011**: `gofmt -l internal/template/workflow_worktree_key_honesty_test.go internal/template/testdata/worktreekeyaliasprobe/aliasprobe.go` → empty output.

## §D.10 AC-012 — Test-file exclusion (negative)

- **Given** baseline, **When** a read of a `WorkflowWorktreeConfig` field is added ONLY inside a `*_test.go` file (mutation), **When** `GUARD` runs, **Then** the guard passes — no unnamed-reader finding is emitted for test files (REQ-008).

## §D.11 Edge Cases

- Alias chains longer than one hop (`c := cfg.Workflow; w := c.Worktree; w.AutoCleanup`) — same resolution path as AC-005; covered by the fixture if trivial to add, otherwise a research-note limitation (type resolution handles arbitrary chains; the fixture demonstrates one).
- A reader file renamed (not deleted): appears as dropped-expected + unnamed-reader simultaneously — two findings, both naming files (REQ-003 shape).
- `packages.Load` returning zero packages (wrong Dir): the non-vacuity log in AC-001 must make this visible — zero packages/files scanned is a failure precondition, not a silent pass.

## §D.12 Quality Gate Criteria (TRUST 5)

- **Tested**: the deliverable is a test; mutation matrix AC-002..008, AC-012 provides its RED evidence; AC-001/009 the GREEN.
- **Readable**: findings name key + file + reason (every AC's Then clause names files); no bare booleans in output.
- **Unified**: gofmt clean (AC-011); lint clean (AC-010); style follows the existing guard files (package-level expectation maps, `t.Helper()` helpers).
- **Secured**: test-only change; fixture reads no user data; no network; no env dependence beyond the Go toolchain.
- **Trackable**: conventional commits naming SPEC-TPL-AST-GUARD-001; evidence under `.moai/reports/t1377/`.

## §D.13 Definition of Done

1. All 12 ACs observed passing (or failing-then-reverted, for mutations) with verbatim captured output in `.moai/reports/t1377/`.
2. spec.md §C Exclusions hold: template untouched (`git diff --name-only` over run-phase commits lists only the two new files), t682 + shipped-key guards untouched.
3. `make build` not required and not run for template reasons (contract A).
4. Sync phase closes with the standard gates.
