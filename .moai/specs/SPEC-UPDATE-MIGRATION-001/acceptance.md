# acceptance.md — SPEC-UPDATE-MIGRATION-001 (card t1547)

Verification layer. Each AC is Given-When-Then, binary-testable. Requirement obligations live in spec.md (REQ-UPM-*); this file does not restate them.

## §D AC Matrix

### AC-UPM-001 — Classification completeness (REQ-UPM-001/002)
Given a fixture project containing, under managed roots: an unmodified template-carried file, a user-modified template-carried file, a local-only file, a local-only file with a non-ASCII (e.g. Korean) filename, and a file dropped from the current template.
When the classifier runs over the managed-root set.
Then every regular file is assigned exactly one class (`template-owned` / `user-modified` / `user-owned` / `stale`), the four sets are disjoint and exhaustive, the non-ASCII-named file classifies by the same byte-semantic rules and survives per its class, and the assignment is carriage-gated — verified by a test asserting a `moai-`-named file the template does not carry classifies `user-owned`, not removal-eligible.

### AC-UPM-002 — Namespace force-preserve (REQ-UPM-003)
Given fixture files under `IsUserOwnedNamespace` prefixes (`hns-*` skill, `.claude/agents/harness/`, user skill names) inside managed-root scope.
When classification runs.
Then all of them classify `user-owned` regardless of manifest state, and update leaves them byte-for-byte identical (hash-compared before/after).

### AC-UPM-003 — Symlink non-dereference (REQ-UPM-004)
Given a fixture containing a live file symlink, a live directory symlink, and a dangling symlink under a managed root.
When classification and reconciliation run.
Then links appear in the classification's symlink list, no link target is read or written, and the link dispositions match the existing SPEC-CLI-CLEAN-SYMLINK-001 notes (greppable `symlink` token).

### AC-UPM-010 — Refresh of clean template-owned files (REQ-UPM-010, NFR-UPM-003)
Given a clean fixture project (no user modifications, no local-only files).
When update runs.
Then the end state equals the current wipe-and-redeploy end state (directory-diff equal) and the summary reports the refreshed set. Proven by M1 characterization baseline vs M4 pipeline output.

### AC-UPM-020 — HAZARD: local-only file survives (REQ-UPM-013/015) — RELEASE-BLOCKING
Given a fixture project with a tracked local-only file under a managed root (e.g. `.claude/rules/moai/dev-only-rule.md`, `.moai/config/sections/custom.yaml`).
When update runs over the fixture.
Then the file exists after the run with byte-identical content, appears in the summary's preserved list, and no backup-restore step was needed.
RED measurement protocol (bound to plan.md M2; the plan phase writes no code, so the RED is taken at M2 landing, not now): the M2 hazard test **`TestUpdate_LocalOnlyFileSurvives`** (`internal/cli/update/reconcile_test.go`) runs against the current wipe-first flow, and its RED capture lands in progress.md §E.2 with ALL FOUR verification-completeness §2.1 elements BEFORE M4 — (1) the command `go test -run TestUpdate_LocalOnlyFileSurvives -count=1 ./internal/cli/update/`; (2) that run's verbatim stdout, which must show the test EXECUTING (a run whose executed-test count is zero — `ok ... [no tests to run]` — is not a RED and is not admissible evidence); (3) the exit code as its own field (non-zero for RED); (4) baseline attribution naming the tree SHA the measurement ran on plus `(this run, this tree)`. If M2 cannot produce all four elements, this criterion auto-demotes to regression-guard and loses release-blocking eligibility (§2.1 undecidable disposition). Historical context only — NOT this criterion's baseline: the same hazard measured 2026-08-15 (12 files) and `.moai/reports/t1159/measurement.md`.

### AC-UPM-021 — git-strategy operator values survive (REQ-UPM-020/021) — RELEASE-BLOCKING
Given a fixture project whose `.moai/config/sections/git-strategy.yaml` sets operator values on keys the template ALSO carries with neutral defaults — the template (`git-strategy.yaml.tmpl`) carries `git_strategy.worktree_base_branch: ""` (nested under `git_strategy:`, not top-level) and `git_strategy.manual.workflow: github-flow`; the fixture sets `git_strategy.worktree_base_branch: develop` and `git_strategy.manual.workflow: git-flow`.
When update runs.
Then both operator VALUES survive verbatim: the post-update file reads `worktree_base_branch: develop` (not the template's empty default) and `workflow: git-flow` (not `github-flow`), with zero manual steps. The hazard is VALUE reversion, not key absence — an assertion of key presence alone would pass the reverted state and is not admissible; the check asserts the operator values.
RED measurement protocol (bound to plan.md M2): the M2 hazard test **`TestUpdate_GitStrategyValuesSurvive`** (`internal/cli/update/reconcile_test.go`) runs against the current wipe-first flow, captured with the same FOUR §2.1 elements as AC-UPM-020 (command `go test -run TestUpdate_GitStrategyValuesSurvive -count=1 ./internal/cli/update/`; verbatim stdout with non-zero executed-test count; exit code as its own field; tree SHA + `(this run, this tree)`), recorded in progress.md §E.2 BEFORE M4, with the same auto-demotion rule. Historical context only: the 2026-09-24 incident (6 card worktrees cut from `main`), `.moai/reports/t1159/measurement.md`.

### AC-UPM-030 — Conflict semantics (REQ-UPM-011/012)
Given a fixture with a user-modified mergeable file whose edit conflicts with the new render (non-mergeable case), one whose edit merges cleanly, and a leftover `<path>.moai-new` sibling from a prior run (or an independent user file of that name) blocking the conflict case's sidecar path.
When update runs.
Then the clean-merge case: merged file written, reported `merged`. The conflict case: on-disk file byte-identical to pre-run, the sidecar lands at the FIRST UNUSED numbered sibling (`<path>.moai-new.2`, `<path>.moai-new.3`, … — the pre-existing sibling is never overwritten, hash-verified), and the summary carries a conflict row naming both paths including the collision. Verified by content hashes and summary-structure assertions.

### AC-UPM-031 — Stale archive-then-remove (REQ-UPM-014)
Given a fixture with a file a prior template carried and the current template does not.
When update runs.
Then the file is copied into the migration archive root (layout preserved) and removed from place, the archive copy is byte-identical, and the summary lists the removal. A forced archive-write failure aborts the removal (file still in place).

### AC-UPM-032 — Summary honesty (REQ-UPM-030/031/032)
Given a fixture run that refreshes, merges, conflicts, preserves, and archive-removes at least one file each.
When the summary is emitted.
Then it reports counts and per-path lists for all five categories, includes every deletion, and the counted total includes managed-root files (the `plan.go:73` exclusion is gone). Structure assertions only — no output-format vocabulary that pre-commits t1527.

### AC-UPM-033 — Wholesale-path guard (REQ-UPM-015/040)
Given the retained legacy wholesale path invoked (directly in a test) over a fixture containing a `user-owned` file and an unresolved user-modified file.
When it runs.
Then it refuses to delete both (error or skip-with-report), while template-owned and stale files still clear — asserting the guard on BOTH paths, default and legacy.

### AC-UPM-040 — Deletion-free observable (card completion criterion a)
Given a fixture project whose local-only files are git-tracked (fixture is a git repo).
When update runs, then `git -C <fixture> status --porcelain` contains no ` D` (deletion) lines for local-only files — the §2.3 manual check, automated inside the AC-UPM-020 test.

### AC-UPM-041 — Abort leaves unprocessed paths intact (NFR-UPM-002)
Given a reconciliation forced to fail on path k of n (injected write error).
When update runs.
Then paths 1..k-1 stay in their post-step state (each backed up), paths k..n are byte-identical to pre-run, and the run reports the failure with the failing path named.

### AC-UPM-050 — Determinism (NFR-UPM-001)
Given the same fixture tree and template set.
When classification runs twice.
Then both outputs are byte-identical (JSON-serialized comparison).

## §D.1 Severity

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-UPM-020, 021, 030 | Critical | The deletion/value-reversion/overwrite hazards this SPEC exists to close |
| AC-UPM-031, 033, 040, 041 | High | Data-safety guarantees |
| AC-UPM-001, 002, 003, 010, 032, 050 | Medium | Correctness of the new pipeline |

## §D.2 Traceability

REQ-UPM-001→AC-UPM-001 · 002→001 · 003→002 · 004→003 · 010→010 · 011/012→030 · 013→020 · 014→031 · 015→020,033,040 · 016→031,041 · 020/021→021 · 030/031/032→032 · 040→033. NFR-UPM-001→050 · 002→041 · 003→010.

## §D.3 Indirect verification

- Symlink and backup-ordering guarantees are additionally protected by the EXISTING suites (`deploy_symlink_*_test.go`, `deploy_preclean_backup_test.go`, `deploy_contract_test.go`) — they must remain green unmodified in intent.
- Windows path handling: `GOOS=windows GOARCH=amd64 go build ./...` gate per milestone (no windows runner locally; CI matrix is the final verdict).

## §D.4 Quality gates

- `go test -count=1 ./internal/cli/update/...` green; coverage ≥85% per touched package (`internal/cli/update`, `internal/cli/update/deploy`, `internal/cli/update/plan`, `internal/cli/update/report`).
- `golangci-lint run` — zero NEW issues vs pre-work baseline.
- RED evidence for AC-UPM-020/021 captured verbatim before M4 (plan.md M2).
- No test writes outside `t.TempDir()`; no OTEL env in tests.

## §D.5 Definition of Done

All ACs PASS with observed command output in progress.md §E.2; the two card completion checks (AC-UPM-020's git-status form, AC-UPM-021's grep form) demonstrated against a fixture project; docs updated (M6); §E.4 closed by sync.
