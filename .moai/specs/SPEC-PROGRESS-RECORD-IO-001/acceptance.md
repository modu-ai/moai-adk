# SPEC-PROGRESS-RECORD-IO-001 — Acceptance Criteria

## §A Purpose

Machine-verifiable acceptance layer for the F14 darwin seeder fix. Every criterion names its
command and its expected observable; where the decisive evidence can only exist on CI
(linux/windows GOOS-tagged runs), the criterion names the CI surface and no local command — a
local darwin run is never cited for a GOOS-tagged family.

Two-cell discipline: AC-PRI-002 is the RED-now cell (observed at M2 entry on the pre-fix tree,
tree SHA pinned in `progress.md` §E.2); AC-PRI-003 is its green path (the same family on the
post-fix tree). Evidence is **branch-conditional**: on route (i) (pure-Go fd route measured
writable) all nine ACs stand; on route (ii) (no writable route) the unachievable ACs are
dispositioned per §D's route-(ii) disposition — no AC is claimed unconditionally across both
branches.

## §B AC Matrix (traceability)

| AC | REQ | Evidence surface | Severity | Command / observable |
|----|-----|------------------|----------|----------------------|
| AC-PRI-001 | REQ-PRI-005, REQ-PRI-006 | local (darwin) + record | High | Probe artifact exists; fork decision recorded with verbatim measurement |
| AC-PRI-002 | REQ-PRI-007 | local (darwin, pre-fix) | High (RED cell) | Promoted family observed FAIL on pre-fix tree |
| AC-PRI-003 | REQ-PRI-001, REQ-PRI-002, REQ-PRI-007 | local (darwin, post-fix) | High (GREEN cell) | Promoted family passes `-race -count=2`, no SKIP on the three axes |
| AC-PRI-004 | REQ-PRI-002 | local | High | 0 `exec.Command` hits in the darwin seeder |
| AC-PRI-005 | REQ-PRI-001, REQ-PRI-003 | local (darwin) | High | Append regression family + close-hygiene probe pass `-race -count=2` with zero SKIP |
| AC-PRI-006 | REQ-PRI-003 | local (darwin) | Medium | Affected package passes `-timeout 30m` |
| AC-CI-007 | REQ-CI-008 | **CI only** (`release-pr-multi-os.yml` 3-OS leg: `release/*`→main PR or `workflow_dispatch`) | Medium | GOOS-tagged families decisive-PASS from the `-json` stream (SKIP is not PASS); run URLs recorded |
| AC-PRI-008 | REQ-DOC-009 | local | Medium | Stale exec-exception AND `victim-overwrite` text absent; post-fix harm-class note present |
| AC-PRI-009 | REQ-PRI-004 | local (darwin) | High (fd-anchoring) | REAL seeder fails closed under a mid-seed name swap; victim content AND metadata untouched |

## §C Given-When-Then scenarios

- **AC-PRI-001 (probe artifact + fork decision — structured fields, content-gated)**
  Given the M1 probe ran on the base tree, When the probe artifact is read, Then
  `.moai/state/verify/t1598/probe-darwin-fd-xattr.md` carries the FOUR fixed field headings
  promised in plan M1 step 6 — `## xattr name`, `## blob layout`, `## fd-set result`,
  `## fork decision` — each populated with the measured value (or the decisive null result).
  The gated observable is the §E.2 carrier: `progress.md` §E.2 must quote the selected branch
  token (`route (i)` or `route (ii)`) AND the decisive raw output line verbatim — a marker-only
  artifact (headings present, values absent) fails this AC.
  Command: `grep -c '^## \(xattr name\|blob layout\|fd-set result\|fork decision\)$'
  .moai/state/verify/t1598/probe-darwin-fd-xattr.md` → `4`;
  `grep -cE 'route \((i|ii)\)' .moai/specs/SPEC-PROGRESS-RECORD-IO-001/progress.md` → ≥ 1
  (the hit must quote the decisive command output, not just name the branch).

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

- **AC-PRI-005 (append regression family — F15/F16 guarantees survive, no-SKIP decisive)**
  Given the fix landed without touching the swap/close semantics, When the whole append family
  runs verbosely, Then every promoted regression guard stays green AND executed — including
  `TestAppendProgressRecordSwapKeepsForeignFile` (F15, `audit_ceiling_replace_test.go:220`), the
  close-hygiene probe `TestAppendProgressRecordSeedCloseHygiene` (F16 — see below), and the
  ACL/xattr per-axis tests.
  Command: `go test -race -count=2 -v -run '^TestAppendProgressRecord' ./internal/runtime/`
  Expected: exit 0, `ok` line, and **zero `SKIP` lines**. A SKIP (e.g. the ACL-axis `Skipf` at
  `audit_ceiling_acl_test.go:65` on an inherited-ACL-restricted environment) is a non-decisive
  observation, never evidence: a run with any SKIP does not satisfy this AC — the decisive run
  re-executes on a capable environment before close (the fallback is recorded in §E.2 as
  non-decisive, never as a pass).
  The close-hygiene probe (F16 guard, plan M3): a `seedFileMetadataFn` wrapper captures the held
  descriptor; after `appendProgressRecord` returns, a second `Close` must report already-closed
  on the normal path (close-before-rename, `audit_ceiling.go:716-720`, plus the deferred close
  `:676-684`) and on an abort path (deferred close only) — a refactor dropping the close is
  locally observed.

- **AC-PRI-006 (affected-package regression)**
  Given the fix landed, When the affected package runs, Then the full package suite passes
  (no local full-suite run — lane discipline).
  Command: `go test -timeout 30m ./internal/runtime/`
  Expected: exit 0, `ok` line.

- **AC-CI-007 (GOOS-tagged decisive verdicts — CI-only evidence, real surface)**
  Given the fix is on the card branch, When the linux/windows GOOS-tagged seeder/append families
  run, Then their decisive verdict comes from the **`release-pr-multi-os.yml` 3-OS leg**
  (`go test -json -race -timeout 35m ./...`, linux+macos+windows) — triggered by a `release/*`→`main`
  PR, or by `workflow_dispatch` on the card branch (run-anytime; the dispatch is the path when
  the card must record its verdict before a release PR exists). `ci.yml` is ubuntu-only
  (ci.yml:94) and `pr-multi-os-gate.yml:112` excludes `internal/runtime` — neither is a decisive
  surface for these families; this repository's git-flow has NO card PR (cards merge to develop,
  leader batch-pushes), so "card PR CI" does not exist as a surface.
  Recording rule: the per-family verdict is read from the `-json` stream — decisive-PASS
  requires an explicit per-test pass Action for EVERY test in the family; a `skip` Action is
  recorded as SKIP and is NOT a PASS. §E.2 records the run URL, per-family executed/skipped
  counts, and verdicts. **No local command exists for this criterion** — a local darwin run is
  structurally unable to execute GOOS-tagged families and is never cited.

- **AC-PRI-008 (ruling (i) disposition re-documented — stale text absent, content-gated)**
  Given the fix landed, When the darwin seeder source is read, Then the STALE pre-fix text is
  gone (both the exec-exception marker AND the stale victim-overwrite wording at
  `progress_metadata_darwin.go:34-38`) and the post-fix note stands with its pinned content (plan
  M3 step 3): the residual window named `fd-verify→rename`, the reduced harm class — own append
  fails / the foreign temp entry is replaced — and the kauth_filesec follow-up disposition.
  Commands: `grep -c "victim-overwrite" internal/runtime/progress_metadata_darwin.go` → `0`;
  `grep -c "LEADER-ACCEPTED darwin exception" internal/runtime/progress_metadata_darwin.go` →
  `0`; `grep -n "fd-verify" internal/runtime/progress_metadata_darwin.go` → ≥1;
  `grep -n "kauth_filesec" internal/runtime/progress_metadata_darwin.go` → ≥1.
  A marker-only edit (replacing the exception token while the stale harm-class text survives)
  fails the `victim-overwrite` → 0 predicate.

- **AC-PRI-009 (fd anchoring — the REAL seeder under a mid-seed name swap)**
  Given the REAL darwin seeder (not a stub) is reached through the `seedFileMetadataFn` seam
  (`audit_ceiling.go:731`) and a wrapper performs a mid-seed name swap — renaming the temp away
  and planting a symlink to a victim file at the temp's name — before delegating to the real
  implementation, When the replace runs, Then it fails closed AND the victim is untouched in
  content AND metadata. A path-based regression — a Go-native seeder that re-opens the swapped
  NAME for its xattr/mode writes (`unix.Setxattr` is path-based; this is the natural mutant
  shape) — follows the symlink and clobbers the victim, which this AC detects. Unlike
  `TestAppendProgressRecordTempSwapFailsClosed` (`audit_ceiling_replace_test.go:174-182`), which
  stubs the seeder and observes none of this, this AC exercises the real implementation.
  Command: `go test -race -count=2 -run TestAppendProgressRecordRealSeedMidSwap
  ./internal/runtime/` (darwin)
  Expected: exit 0, `ok`. Characterization posture: it passes on the current seeder (the
  pre-check rejects the symlink) and MUST keep passing post-fix — wired into M2's GREEN gate and
  M3's sweep.

## §D Edge cases

- **ACL seeding unavailable in the test environment**: the family's `Skipf` guards fire (no
  `_guest` group, restricted FS). A skip is an observation, not a pass — AC-PRI-003 requires the
  decisive darwin run to show no SKIP; a skipped decisive run re-runs on a capable environment
  before close.
- **APFS exposes no ACL xattr**: M1's null result is decisive — route killed, decision-index Q2
  escalates; the fix then re-scopes only per the operator verdict (default: document-residual).
- **Route-(ii) disposition (decision-index Q2 fires — branch-conditional completion)**: on route
  (ii) (no writable pure-Go route), AC-PRI-003 (family GREEN), REQ-PRI-002's no-exec face, and
  AC-PRI-009's post-fix guard are UNACHIEVABLE by construction — the current exec seeder cannot
  pass `PATH=""` (measured: `exec: "chmod" not found`). The completion condition becomes:
  AC-PRI-001 (probe + null-result record), AC-PRI-002 (RED observed), AC-PRI-005/006 (existing
  behaviors unregressed), AC-PRI-008 (the re-documentation IS the route-(ii) deliverable), and
  AC-CI-007 stand; AC-PRI-003, AC-PRI-004's no-exec face, and AC-PRI-009's post-fix leg are
  dispositioned as NOT-EVIDENCE and replaced by the operator's Q2 verdict record quoted in
  `progress.md` §E.2. The SPEC re-enters plan phase for the re-scope amendment before closing —
  the DoD is never satisfied unconditionally across both branches.
- **`x/sys` fd-xattr wrappers absent on darwin at v0.48.0**: M1 measures the wrapper surface
  first; if absent, the raw syscall route via `unix.Syscall` is measured before the route is
  declared dead — the probe records which of the two was attempted; the probe, like all darwin
  changes, stays behind `//go:build darwin`.
- **Non-root EPERM on the ACL xattr namespace**: a measured EPERM on fd-set kills route (i) even
  where the xattr exists — recorded as the fd-set result, not silently retried.

## §E Quality gates

- TRUST 5 Tested: AC-PRI-003/005/006 green with `-race -count=2` / `-timeout 30m`.
- Secured: the F14 window is closed by construction (no exec, fd-anchored writes); fail-closed
  contract preserved (AC-PRI-005).
- Trackable: conventional commits per milestone; card id t1598 in commit messages and evidence
  paths (`.moai/reports/t1598/`).

## §F Definition of Done

Route (i): all nine ACs carry observed evidence in `progress.md` §E.2 (verbatim outputs, exit
codes, tree SHAs); AC-CI-007 carries the `release-pr-multi-os.yml` run URL with per-family
executed/skipped counts. Route (ii): the branch-conditional set per §D's route-(ii) disposition
— the dispositioned ACs are replaced by the Q2-verdict record, never silently dropped. No
[NEEDS CLARIFICATION] markers remain (decision-index Q1-Q4 resolved or operator-escalated per
its verdicts). SPEC frontmatter transitions `draft → in-progress` at M1 commit start
(manager-develop owns the transition).
