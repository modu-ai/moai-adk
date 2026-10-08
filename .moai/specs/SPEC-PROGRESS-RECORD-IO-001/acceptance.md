# SPEC-PROGRESS-RECORD-IO-001 — Acceptance Criteria

## §A Purpose

Machine-verifiable acceptance layer for the F14 darwin seeder fix. Every criterion names its
command and its expected observable; where the decisive evidence can only exist on CI
(linux/windows GOOS-tagged runs), the criterion names the CI surface and no local command — a
local darwin run is never cited for a GOOS-tagged family.

Two-cell discipline: AC-PRI-002 is the RED-now cell (observed at M2 entry on the pre-fix tree,
tree SHA pinned in `progress.md` §E.2); AC-PRI-003 is its green path (the same family on the
post-fix tree).

## §B AC Matrix (traceability)

| AC | REQ | Evidence surface | Severity | Command / observable |
|----|-----|------------------|----------|----------------------|
| AC-PRI-001 | REQ-PRI-005, REQ-PRI-006 | local (darwin) + record | High | Probe artifact exists; fork decision recorded with verbatim measurement |
| AC-PRI-002 | REQ-PRI-007 | local (darwin, pre-fix) | High (RED cell) | Promoted family observed FAIL on pre-fix tree |
| AC-PRI-003 | REQ-PRI-001, REQ-PRI-002, REQ-PRI-007 | local (darwin, post-fix) | High (GREEN cell) | Promoted family passes `-race -count=2`, no SKIP on the three axes |
| AC-PRI-004 | REQ-PRI-002 | local | High | 0 `exec.Command` hits in the darwin seeder |
| AC-PRI-005 | REQ-PRI-001, REQ-PRI-003 | local (darwin) | High | Append regression family passes `-race -count=2` |
| AC-PRI-006 | REQ-PRI-003 | local (darwin) | Medium | Affected package passes `-timeout 30m` |
| AC-CI-007 | REQ-CI-008 | **CI only** (card PR runs) | Medium | GOOS-tagged families decisive-PASS on CI; run URLs recorded |
| AC-PRI-008 | REQ-DOC-009 | local | Medium | Pre-fix exec-exception comment superseded; post-fix disposition note present |

## §C Given-When-Then scenarios

- **AC-PRI-001 (probe artifact + fork decision)**
  Given the M1 probe ran on the base tree, When the probe artifact is read, Then
  `.moai/state/verify/t1598/probe-darwin-fd-xattr.md` exists and records (a) the exact xattr
  name the kernel stores ACLs under (or the decisive null result), (b) the blob layout where a
  surface exists, (c) the non-root fd-set attempt result, and (d) the selected fork branch with
  evidence.
  Command: `test -f .moai/state/verify/t1598/probe-darwin-fd-xattr.md && grep -c "fork decision"
  .moai/state/verify/t1598/probe-darwin-fd-xattr.md`
  Expected: exit 0, count ≥ 1. The decisive verbatim lines are carried into `progress.md` §E.2
  (`grep -n "probe-darwin-fd-xattr" .moai/specs/SPEC-PROGRESS-RECORD-IO-001/progress.md` → ≥1
  hit).

- **AC-PRI-002 (RED cell — held family fails pre-fix)**
  Given the held family is promoted verbatim into `internal/runtime` (darwin tag) and the seeder
  is UNCHANGED, When the family runs, Then it FAILS because the exec-based seeder cannot run with
  `PATH=""`.
  Command: `go test -run TestAppendProgressRecordPreservesAllMetadataAxes ./internal/runtime/`
  (on the pre-fix tree)
  Expected: exit 1, FAIL output naming the exec/PATH failure. Verbatim stdout + exit code + tree
  SHA recorded in `progress.md` §E.2 at M2 entry — RED observed, not inferred
  (`grep -n "FAIL.*PreservesAllMetadataAxes"
  .moai/specs/SPEC-PROGRESS-RECORD-IO-001/progress.md` → ≥1 hit at close).

- **AC-PRI-003 (GREEN cell — held family passes post-fix)**
  Given the F14 fix is implemented, When the promoted family runs on the post-fix tree, Then all
  three axes (umask 0600, `group:_guest deny read` ACL, `user.t1560-axis=seeded` xattr) survive
  the replace with `PATH` stripped.
  Command: `go test -race -count=2 -v -run
  TestAppendProgressRecordPreservesAllMetadataAxes ./internal/runtime/`
  Expected: exit 0, `ok` line, and NO `SKIP` line for that test (a skipped axis is an
  observation, not a pass — the decisive run must exercise all three axes on the darwin authoring
  platform).

- **AC-PRI-004 (no shell-out)**
  Given the fix landed, When the darwin seeder source is scanned, Then no external-process
  invocation remains.
  Command: `grep -c "exec.Command" internal/runtime/progress_metadata_darwin.go`
  Expected: output `0` (baseline on a2a184ad3: `2`), exit 1 (grep found nothing) — the zero-count
  is the pass.

- **AC-PRI-005 (append regression family — F15/F16 guarantees survive)**
  Given the fix landed without touching the swap/close semantics, When the whole append family
  runs, Then every promoted regression guard stays green — including
  `TestAppendProgressRecordSwapKeepsForeignFile` (F15, `audit_ceiling_replace_test.go:220`) and
  the close-hygiene paths (F16).
  Command: `go test -race -count=2 -run '^TestAppendProgressRecord' ./internal/runtime/`
  Expected: exit 0, `ok` line.

- **AC-PRI-006 (affected-package regression)**
  Given the fix landed, When the affected package runs, Then the full package suite passes
  (no local full-suite run — lane discipline).
  Command: `go test -timeout 30m ./internal/runtime/`
  Expected: exit 0, `ok` line.

- **AC-CI-007 (GOOS-tagged decisive verdicts — CI-only evidence)**
  Given the fix PR is open, When the linux and windows GOOS-tagged seeder/append families run,
  Then their decisive PASS verdicts come from the card PR's CI runs (F9/F10/6b/6c/6d
  accumulation).
  Evidence surface: the PR's `gh pr checks` / `gh run view --job` records for the linux and
  windows jobs carrying the tagged families. **No local command exists for this criterion** —
  a local darwin run is structurally unable to execute GOOS-tagged families and is never cited.
  Observable: run URLs + per-family verdicts recorded in `progress.md` §E.2.

- **AC-PRI-008 (ruling (i) disposition re-documented)**
  Given the fix landed, When the darwin seeder source is read, Then the pre-fix exec-exception
  comment is gone and a post-fix disposition note stands: the residual fd-verify→rename window
  documented at the reduced harm class, plus the kauth_filesec follow-up disposition.
  Commands: `grep -c "LEADER-ACCEPTED darwin exception"
  internal/runtime/progress_metadata_darwin.go` → `0`; `grep -n "ruling (i)"
  internal/runtime/progress_metadata_darwin.go` → ≥1 line reflecting the post-fix harm class.

## §D Edge cases

- **ACL seeding unavailable in the test environment**: the family's `Skipf` guards fire (no
  `_guest` group, restricted FS). A skip is an observation, not a pass — AC-PRI-003 requires the
  decisive darwin run to show no SKIP; a skipped decisive run re-runs on a capable environment
  before close.
- **APFS exposes no ACL xattr**: M1's null result is decisive — route killed, decision-index Q2
  escalates; the fix then re-scopes only per the operator verdict (default: document-residual).
- **`x/sys` fd-xattr wrappers absent on darwin at v0.48.0**: M1 measures the wrapper surface
  first; if absent, the raw syscall route via `unix.Syscall` is measured before the route is
  declared dead — the probe records which of the two was attempted.
- **Non-root EPERM on the ACL xattr namespace**: a measured EPERM on fd-set kills route (i) even
  where the xattr exists — recorded as the fd-set result, not silently retried.

## §E Quality gates

- TRUST 5 Tested: AC-PRI-003/005/006 green with `-race -count=2` / `-timeout 30m`.
- Secured: the F14 window is closed by construction (no exec, fd-anchored writes); fail-closed
  contract preserved (AC-PRI-005).
- Trackable: conventional commits per milestone; card id t1598 in commit messages and evidence
  paths (`.moai/reports/t1598/`).

## §F Definition of Done

All eight ACs carry observed evidence in `progress.md` §E.2 (verbatim outputs, exit codes, tree
SHAs); AC-CI-007 carries CI run URLs. No [NEEDS CLARIFICATION] markers remain (decision-index
Q1-Q4 resolved or operator-escalated per its verdicts). SPEC frontmatter transitions
`draft → in-progress` at M1 commit start (manager-develop owns the transition).
