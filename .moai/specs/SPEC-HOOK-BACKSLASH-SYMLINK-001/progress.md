# SPEC-HOOK-BACKSLASH-SYMLINK-001 — Progress Record

status: completed (sync phase)
card: t1556 (factory lane-12)
worktree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1556 (branch WT-backslash-symlink, HEAD cad44a751, clean at authoring)

## Plan-phase Evidence (what was read, what was decided)

Read and verified in this tree (cad44a751):

- `internal/hook/pre_tool.go:1357-1480` — `resolveThroughExistingParent` /
  `resolvePhysicalWalk`: the t1530 physical-walk repair. Defect line confirmed at
  1397: `strings.Split(strings.ReplaceAll(filepath.ToSlash(p), "\\", "/"), "/")` —
  unconditional backslash→slash conversion before segmentation.
- `internal/hook/pre_tool.go:1487-1593` — `absoluteUncleaned` + `checkFileAccess`:
  boundary check, NFC normalization, project-root EvalSymlinks symmetry — untouched
  surfaces.
- `internal/hook/protected_zone_path.go:34-49` — `zoneSlash` carries the same
  unconditional conversion; classified OUT OF SCOPE (PowerShell display-form matching,
  separate follow-up), recorded in spec.md §F so its silence is deliberate.
- Defect evidence: `.moai/worktrees/t1533/.moai/reports/t1533/codex-review-gate-1.md`
  (card t1533 codex review gate round 1, "[P1] POSIX 경로의 실제 백슬래시를 보존 —
  internal/hook/pre_tool.go:1397" — reproduced twice; external write of `"escaped"`
  observed pre-fix; blocking works with the base interpretation function).

Decisions made at plan phase:

1. The defect is an input-normalization error INSIDE the t1530 walking pattern; the
   walk's structure (probe existing components, rejoin missing tail, physical `..`
   pop, depth-bound fail-closed) is retained. The leader hint (does
   `resolveThroughExistingParent` apply?) evaluated to: it IS the defective code; the
   repair narrows to platform-appropriate segmentation. Concrete mechanism (runtime
   GOOS check vs build tags) left to the repair agent, constrained by
   REQ-HBS-001/REQ-HBS-004.
2. Tier M artifact set (spec/plan/acceptance/progress). SPEC ID
   `SPEC-HOOK-BACKSLASH-SYMLINK-001` — regex check PASS, no collision in
   `.moai/specs/`.
3. Scope: ONLY the pre_tool.go:1397 bypass. The other 12 findings of the t1533 round-1
   review belong to the t1454 residual ledger (out of scope, spec.md §F).

## §E.1 Plan-phase Audit-Ready Signal

Iteration 1 verdict: **FAIL**, score 0.875, MP-8 blocking — verdict
`.moai/reports/t1556/plan-audit-1.md` (receipt rcpt-9dae86459575ea64124cd7e5;
audited SHA 785cfaaff). Revision applied same day, addressing D1–D6:

- **D1 (critical, MP-8)** — the reproduction tests are now AUTHORED AND EXECUTED at
  plan phase: `internal/hook/pre_tool_backslash_repro_test.go`
  (`TestCheckFileAccessPosixBackslashSymlinkEscape` + a second resolver-level test
  added for AC-HBS-003's own observation). RED observed on the pre-fix tree
  (`785cfaaff`, code identical to `cad44a751`), exit code 1, verbatim stdout and
  exit code recorded in the four-element cells `RED-HBS-001` / `RED-HBS-003` in
  acceptance.md. The test file ships with this SPEC revision commit, before any fix
  exists.
- **D2** — AC-HBS-002 rewritten: the fixture's non-deny branch performs the actual
  Write through the literal path, so pre-fix the external write is observed (it is —
  the verbatim RED line records the landing path outside the project root) and
  post-fix the deny branch skips it.
- **D3** — AC-HBS-004..007 reclassified as regression-guard ONLY (release-blocking
  list now AC-001..003); AC-HBS-006's false depth-bound coverage claim removed
  (branch `pre_tool.go:1444` measured 0x by the audit) and scoped to the existing
  tests' rejoin/`..`-pop pass; AC-HBS-007 given two explicit fixtures
  (depth-bound-exceeded chain, unresolvable `..`) with concrete asserts and a
  verification command form, authored in M3.
- **D4** — spec.md §F zoneSlash rationale corrected: it feeds
  `resolveZoneTarget` → `checkProtectedZone`, a deny-decision path — recorded as a
  LIVE sibling bypass of the same defect family, deferred to a named follow-up card
  in the t1454 residual ledger (not dismissed). plan.md §B aligned.
- **D5/D6** — finding count corrected to 13 (3 P1 + 10 P2); `.moi` typo fixed; the
  primary evidence citation now points at the in-tree
  `.moai/reports/t1556/codex-review-gate-1.md` with the t1533 file as provenance
  (spec.md §B, plan.md §A).

Plan artifacts revised; awaiting plan-audit iteration 2.

Iteration 2 verdict: **FAIL**, score 0.94 (D1–D6 cores verified fixed) — verdict
`.moai/reports/t1556/plan-audit-2.md` (receipts rcpt-422f7db791a4653a98efb415 + r1).
Three mechanical residues revised same day:

- **D1-r** — AC-HBS-003's cell now records the SOLO run of exactly its one cited
  command: raw stdout unelided, own exit code 1, tree `785cfaaff`. AC-HBS-001's cell
  was re-observed the same way after the test edit below.
- **D2-r** — `pre_tool_backslash_repro_test.go` non-deny branch: the WriteFile error
  is now fatal (`t.Fatalf`), and the external file is read back asserting content
  `"escaped"` before the failure message — the landing claim carries observed
  content (`external write VERIFIED …, content "escaped"`), closing the
  claimed-landing-that-did-not-happen (ENOENT) hole codex demonstrated.
- **D3-r** — AC-HBS-007's fixtures respec'd from the walk code so they ENGAGE the
  two named branches: depth bound via a chain of > `zoneSymlinkDepthBound` symlinks
  with a NON-EXISTENT terminal target (forces hop-by-hop recursion; an
  existing-terminal chain resolves wholesale at the first EvalSymlinks and never
  reaches the bound), and the `..` pop via `<project>/linked/../leaf` with `linked`
  an EXISTING outside-pointing symlink (the `<missing>/../leaf` form returns at the
  unresolved-tail branch `pre_tool.go:1435` and never reaches `pre_tool.go:1413`).
  plan.md M3's write-down matches; §C.2/§C.3 baseline text updated to the observed
  state (repro RED + pre-existing StaleRunNotice trio; depth bound uncovered by
  existing tests).

Awaiting plan-audit iteration 3.

Additional baseline observation (recorded for the run phase): the pre-fix package
run of `go test -count=1 ./internal/hook/` fails on THREE tests unrelated to this
SPEC — `TestStaleRunNoticeLegacyLeaderSpelling`, `TestStaleRunNoticeLegacySessionRecord`,
`TestStaleRunNoticeFactoryLegacyLabel` — confirmed failing in isolation on this tree
(exit 1) with this SPEC's changes being docs + a new test file only. Classified as
PRE-EXISTING BASELINE (not caused by this work, not repairable within this scope);
AC-HBS-006's post-fix verdict is scoped to "no NEW failure beyond the recorded trio"
rather than a bare exit 0.

## §E.2 Run-phase Evidence

Run executed 2026-10-07 in the card worktree (branch `WT-backslash-symlink`),
starting at HEAD `9d78421a4`. All commands run with the lane env scrubbed in ONE
compound `unset` invocation per DEBT-HBS-BASELINE-ENV (scrubbed vars:
`MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_NAME
MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_BACKEND MOAI_KANBAN_SETTINGS_INJECTED
MOAI_FACTORY_ROLE MOAI_FACTORY_LANE`).

**Mechanism decision (M2).** Platform-appropriate segmentation via a
platform-parameterized helper, NOT a runtime.GOOS read inside the split body:
`pathSegments(p string, windows bool)` in `internal/hook/pre_tool.go` — the
single production caller passes `runtime.GOOS == "windows"`. The parameter
form (over a global flag) was chosen because AC-HBS-005 requires the Windows
branch unit-tested at the string level on every platform; a test-overridable
global would be mutable state. The walker body (probe / rejoin / `..` pop /
depth bound) is unchanged, per plan §B.

**AC matrix** (command + observed output verbatim, this run, this tree):

| AC | Status | Verification command | Actual output (verbatim) |
|----|--------|---------------------|--------------------------|
| AC-HBS-001 | PASS | `go test -count=1 ./internal/hook/ -run 'TestCheckFileAccessPosixBackslashSymlinkEscape'` (env scrubbed) | `--- PASS: TestCheckFileAccessPosixBackslashSymlinkEscape (0.01s)` … `ok github.com/modu-ai/moai-adk/internal/hook 0.968s`, EXIT=0 (joint run with AC-HBS-003's test, log `/tmp/t1556_green.log`) |
| AC-HBS-002 | PASS | same command (the test's deny branch runs the `os.Stat` external-absence assert) | PASS with EXIT=0 — the non-deny write branch did not execute; the external-file assert inside the test is the file-existence observation |
| AC-HBS-003 | PASS | `go test -count=1 ./internal/hook/ -run 'TestResolveThroughExistingParentPosixBackslashSymlinkDivergence'` (joint run with AC-HBS-001's test) | resolver log line: `resolved="/private/var/.../1369638387/002/escaped.txt" ok=true` — resolution lands OUTSIDE (the symlink's destination), test PASS |
| AC-HBS-004 | PASS | `go test -count=1 ./internal/hook/ -run 'TestCheckFileAccessPosixBackslashLegitNameAllowed'` | `--- PASS: TestCheckFileAccessPosixBackslashLegitNameAllowed (0.00s)` |
| AC-HBS-005 | PASS | `go test -count=1 ./internal/hook/ -run 'TestPathSegmentsPlatformSeparatorSemantics'` | both subtests PASS; subtest caught the ToSlash-is-a-noop-on-POSIX hazard (first windows-branch draft failed: `pathSegments(windows) = ["C:\\proj\\linked\\..\\x"]`) and forced the explicit `ReplaceAll` |
| AC-HBS-006 | PASS | `go test -timeout 30m -count=1 ./internal/hook/` (env scrubbed) | `ok github.com/modu-ai/moai-adk/internal/hook 572.539s`, EXIT=0 — ZERO failures, stronger than the allowed delta (the StaleRunNotice trio passed with the scrubbed env, confirming DEBT-HBS-BASELINE-ENV: env-dependent, not regressions; a CI trio failure remains a regression signal) |
| AC-HBS-007 | PASS | `go test -count=1 ./internal/hook/ -run 'TestResolvePhysicalWalkBranchPreservation'` | `--- PASS: TestResolvePhysicalWalkBranchPreservation (0.02s)` — depth-bound fixture (33 links, non-existent terminal) returns ok=false fail-closed + decision-level fallback NOT a deny; `..`-pop fixture (`projectDir + "/linked/../leaf"`, concatenated) resolves OUTSIDE with ok=true + checkFileAccess deny |

**RED re-observation** (E8, pre-fix tree HEAD `9d78421a4`, this run, before the
fix — matching the plan-phase cells `RED-HBS-001`/`RED-HBS-003`):

```text
--- FAIL: TestCheckFileAccessPosixBackslashSymlinkEscape (0.01s)
    pre_tool_backslash_repro_test.go:68: guard allowed the backslash-symlink escape: decision="" reason="" (external write VERIFIED at /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestCheckFileAccessPosixBackslashSymlinkEscape2670212849/002/escaped.txt, content "escaped")
--- FAIL: TestResolveThroughExistingParentPosixBackslashSymlinkDivergence (0.01s)
    pre_tool_backslash_repro_test.go:115: walk validated a fictional IN-PROJECT path ... — validated path diverges from the path the OS walks
FAIL
```

**Coverage (focused run, the modified resolution surface)** — command:
`go test -count=1 -coverprofile=/tmp/t1556_cover.out ./internal/hook/ -run
'Backslash|PathSegments|ResolvePhysicalWalkBranchPreservation|ResolveThroughExistingParent|CheckFileAccess'`,
EXIT=0:

```text
resolveThroughExistingParent  100.0%
resolvePhysicalWalk            82.1%
pathSegments                  100.0%
pathHasDotDotSegment          100.0%
absoluteUncleaned              33.3%   (untouched function; its relative-input branch is outside this test scope)
checkFileAccess                91.1%
```

**Quality gates**: `GOOS=windows GOARCH=amd64 go build ./...` EXIT=0 ·
`go vet ./internal/hook/` EXIT=0 · `gofmt -l internal/hook/` empty ·
`golangci-lint run internal/hook/...` EXIT=0, `0 issues.` · scope check
`git diff --stat` vs `9d78421a4`: ONLY `internal/hook/pre_tool.go` (+25/-1) and
`internal/hook/pre_tool_backslash_repro_test.go` (+170) — no `zoneSlash` touch.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-10-07"
run_commit_sha: "51481f355"
run_status: audit-ready
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 0   # spec.md §F surfaces untouched: zoneSlash, file_changed.go, agentmemory ToSlash uses
new_warnings_or_lints_introduced: 0
cross_platform_build:
  goos_windows_build_exit: 0
  command: "GOOS=windows GOARCH=amd64 go build ./..."
total_run_phase_files: 2   # internal/hook/pre_tool.go + internal/hook/pre_tool_backslash_repro_test.go
m1_to_mN_commit_strategy: "single run-phase commit (M2+M3+M4 land together; M1 repro pre-existed at plan phase)"
env_notes: "DEBT-HBS-BASELINE-ENV: all measurements taken with lane env scrubbed (MOAI_KANBAN* / MOAI_FACTORY_* unset in one compound invocation); the StaleRunNotice legacy trio is env-dependent and PASSES scrubbed — treat CI trio failures as regressions, not baseline"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: "2026-10-07"
sync_commit_sha: "pending-backfill-sync"   # backfilled with the real SHA by a following commit (D3 exemption)
status_transition: "in-progress -> implemented -> completed, spec.md frontmatter `status:` only, riding the single sync commit"
changelog_entry_position: "CHANGELOG.md [Unreleased] / Fixed, first entry"
b12_self_test_a: pre_emission_grep_0   # grep -c SPEC-HOOK-BACKSLASH-SYMLINK-001 CHANGELOG.md returned 0 before emission
b12_self_test_b: "live AC count 7 of 7 (acceptance.md: AC-HBS-001..007, no [RETIRED]/[REF] markers), matching the entry's declared count"
b12_self_test_c: all_entry_paths_ls_verified
canary_compliance_check: "n/a — this SPEC defines no forward-looking policy with own sync tests"
```

What sync changed: (1) `CHANGELOG.md` — one `[Unreleased]`/`Fixed` entry describing the
POSIX backslash/symlink boundary bypass fix (internally neutral; card id + SPEC link per
the file's existing convention, no lane/session detail). (2) `spec.md` frontmatter —
`status: in-progress → completed` on this single sync commit; `updated:` already carried
today's date (2026-10-07), so `status:` is the only changed line. (3) This file — §E.4 and
the header status line. `plan.md` / `acceptance.md` carry no frontmatter (stateless
artifacts), so there is no `updated:` to refresh. NO README or docs-site changes: the fix
lives entirely inside the pre-tool-use hook's path interpretation — no user-facing surface
moved. Codemap regeneration not run (owned by the periodic codemaps cards).

Verification run at sync phase (this tree, HEAD `939e39cf4`): working tree clean before the
sync edits; B12 pre-emission grep returned 0 (no duplicate entry from a parallel sync); the
B12 AC counter (canonical grammar, prefix `AC`) counted exactly 7 live identifiers
(`AC-HBS-001..007`), 0 excluded, 0 ambiguous — reconciled by hand grep against
`acceptance.md` and matching §E.3 `ac_pass_count: 7`; every file path named in the
CHANGELOG entry (`internal/hook/pre_tool.go`, the SPEC directory) exists in this tree. No
Go files touched by sync — gofmt not applicable.

Declared evidence (local, gitignored, cited by name only): `.moai/reports/t1556/card-review.md`,
`.moai/reports/t1556/plan-audit-3.md`, `.moai/reports/t1556/decision-record.md`.
