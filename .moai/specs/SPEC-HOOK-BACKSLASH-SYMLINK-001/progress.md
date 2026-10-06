# SPEC-HOOK-BACKSLASH-SYMLINK-001 — Progress Record

status: draft (plan phase)
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

Additional baseline observation (recorded for the run phase): the pre-fix package
run of `go test -count=1 ./internal/hook/` fails on THREE tests unrelated to this
SPEC — `TestStaleRunNoticeLegacyLeaderSpelling`, `TestStaleRunNoticeLegacySessionRecord`,
`TestStaleRunNoticeFactoryLegacyLabel` — confirmed failing in isolation on this tree
(exit 1) with this SPEC's changes being docs + a new test file only. Classified as
PRE-EXISTING BASELINE (not caused by this work, not repairable within this scope);
AC-HBS-006's post-fix verdict is scoped to "no NEW failure beyond the recorded trio"
rather than a bare exit 0.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
