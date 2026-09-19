# SPEC-ROSTER-NUMERAL-AXIS-001 — Progress

card t930 · branch `WT-numeral-roster-guard` · base `690dfe369`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M, Class C).
- Requirements: REQ-RNA-001 … REQ-RNA-013 (GEARS). Acceptance: AC-RNA-001 … AC-RNA-015
  (15 criteria, under the Tier M ceiling of 16).
- Decisions D1 (noun class) and D2 (discharge rule) are ADOPTED and recorded in `spec.md` §D
  with their rejected alternatives and measured cost. Both are **operator-unanswered** — put to
  the operator, no answer inside the window — so they are the lead's / this SPEC's judgment and
  stay reviewable at the Implementation Kickoff Approval gate.
- **D3 (mirror derivation) — WITHDRAWN by operator decision, 2026-09-18.** Kept in `spec.md` §D
  as a recorded rejected alternative with its four findings (wrong premise measured at 13 rows /
  33 result rather than 20 / 26; derivation blinds the guard to sibling divergence;
  `readmeSite()` precedent overstated; withdrawal shrinks the SPEC). M5 stands at 46 authored
  rows with no folding.
- Plan-audit iteration 1 FAIL (0.775) → revision v0.2.0 cleared D1-D12. Iteration 2 FAIL (0.825,
  score clears; verdict rested on an introduced Must-vs-Must contradiction plus the partially
  resolved D3) → revision v0.3.0, authorised by the operator as a third iteration past the Tier
  M cap of two. Verdicts at `.moai/reports/t930/plan-audit.md` and
  `.moai/reports/t930/plan-audit-iter2.md` — both gitignored by operator directive, never
  committed.
- Cost arithmetic (measured by the dispatching lead in this tree at HEAD `6abcc85fa`): 63 hit
  paths, 20 registry paths carrying `ClaimCount`, 17 hits discharged, **46 residual**, 1 extra
  path the rejected any-row rule would free. Mirror-fold measurement at HEAD `d8f140b25`: 17 of
  the 46 under `internal/template/templates/`, 14 with a local twin on disk, **13 foldable
  pairs** → 33 rows, which is what withdrew D3.
- The 16-vs-17 `ClaimCount` disagreement between the auditor's pass-1 parse and the lead's
  figures is SETTLED in favour of the lead: a text-level `Path:`/`Claims:` parse cannot see the
  four rows `readmeSite()` builds from a `path` parameter. AC-RNA-013 requires the run-phase
  re-derivation to read `Registry()` rather than the file for this reason.

### Implementation Kickoff Approval

- **Given 2026-09-18 by the operator, CONDITIONAL on this SPEC reaching plan-audit PASS.**
- Recorded here so run-phase entry rests on a written approval rather than a remembered one.
- **Condition met 2026-09-18.** The confirmation read over revision v0.3.0 returned four of the
  lead's five axes clean and one minor finding (`plan.md` citing a stale `AC-RNA-014` after the
  16 → 15 renumbering), which was verified and repaired in `89a4831e5` together with the
  unlabelled 17-vs-20 `ClaimCount` pair in §A. The SPEC is PASS; run-phase entry rests on this
  written approval and that recorded condition.
- One criterion needed repair AFTER the audit, for a reason the audit could not have seen:
  `AC-RNA-010` anchored `$IMPL` to the `internal/harness/rosterguard/` directory, which resolves
  to card t922's own commit `bdaafe6fe` once develop is absorbed. Re-anchored to this card's own
  new file in `fd17f4bb0`. The audit was correct at `c9a1e3e1e`; the tree moved underneath it.
- Plan-phase measurements attributed in `spec.md` §A: population figures supplied by the
  dispatching lead (this tree, base `690dfe369`); independently re-measured here —
  `go test -count=1 ./internal/harness/rosterguard/...` → `ok … 1.096s`,
  `profileMatrixAgentOrder` = 13 names, `registry.go` = 30 `Site` rows.
- Status: `draft`.

## §E.2 Run-phase Evidence

### Run-phase baseline — measured BEFORE any implementation edit

Tree: `29a4266f3` on `WT-numeral-roster-guard`, after absorbing develop (`0 8` against it,
working tree clean). The absorption was done first ON PURPOSE: a baseline measured on a
pre-absorption tree would not be a baseline for the tree the implementation actually lands on.

Green-before-edit, scoped to the affected package:

```
go build ./internal/harness/rosterguard/...            exit 0
go vet   ./internal/harness/rosterguard/...            exit 0
go test -count=1 ./internal/harness/rosterguard/...    ok  4.287s
```

`internal/harness/rosterguard/` currently holds `axis.go`, `check.go`, `registry.go`,
`root_test.go`, `rosterguard_test.go`. **`numeral.go` does not exist**, so `AC-RNA-010`'s
`$IMPL` is currently empty — which is the criterion failing, as designed, until this card
creates that file.

### Population re-derived in-run (REQ-RNA-012, AC-RNA-013)

```
hit paths                       : 64
registry paths carrying Count   : 20
hits discharged by a Count row  : 17
hits needing a NEW row/exempt   : 47
extra paths the any-row rule would free : 1   (internal/web/agentfm.go)
```

**The residual is 47, not the 46 carried in `spec.md` §A.** The plan-phase figure was measured
at `6abcc85fa`; absorbing develop brought in one further hit path,
`.claude/skills/moai-foundation-quality/references/reference.md` — the local twin of a template
file that was already residual, having gained the same claim during those commits.

This is the obligation working rather than an error in either number: a plan-phase figure is not
a run baseline, and had the run reused 46 it would have under-registered by one row and the
layer would have reported an undeclared hit. `spec.md` §A stays at 46 as the attributed
plan-phase figure; 47 is the run baseline, measured here, on this tree.

The residual's shape is unchanged: 47 paths, of which 17 sit under
`internal/template/templates/` as mirrors, and the rest split between genuinely stale claims
(the `11 retained agents` / `11-agent catalog` family), legitimate historical citations (the
`then-8-agent catalog` family, `17->8` test comments, dated research), measured false positives
(`172); MoAI retained agents`, `05-25), the agent catalog`, `001): agent catalog`), and
rosterguard's own files describing themselves.

### Scope adjudications

Recorded at `.moai/reports/t930/adjudication.md` (disk-only per the operator directive in
`.gitignore:225-229`): (a) roster drift living as prose in agent definition files is already
reached by the adopted noun class — measured, no widening made; (b) the tier axis in
`advanced/no-haiku-3tier.md` is out of scope on two independent grounds — `docs-site/` is
outside the swept tree, and all four locales carry zero roster-noun hits because the claims are
about tiers, which is cellguard's subject matter.

### Implementation — what landed

Commit `e577f1ce6` (`feat(rosterguard): numeral-adjacency layer for count-only roster
claims (card t930)`), 5 files, +1496/-12, all inside `internal/harness/rosterguard/`:

- `numeral.go` (NEW — the path `AC-RNA-010` anchors `$IMPL` to): scope, noun class,
  both axes, nearest-preceding selection, neutralisation, and the D2 discharge rule.
- `numeral_test.go`, `numeral_rederivation_test.go` (NEW).
- `axis.go`: `Site.NumeralUnreachable` + a package-doc section naming the second
  DISCOVERY axis and the breadth/finding distinction.
- `registry.go`: 23 new `ClaimCount` rows over 19 paths, 27 numeral exemptions.

**No prose was repaired**, as `spec.md` §E requires: the commit's file list carries no
document outside this package, so the `11 retained agents` / `11-agent catalog` family
and every historical citation are byte-unchanged.

### RED evidence (E8) — captured BEFORE the implementation

`go test -count=1 ./internal/harness/rosterguard/...` against a declarations-only
skeleton, exit 1, 70 lines. Verbatim head:

```
--- FAIL: TestNumeralSelectionIsNearestPreceding (0.00s)
    --- FAIL: .../live:_CLAUDE.md_section-4_citation
        numeral_test.go:83: want exactly one hit, got []
--- FAIL: TestNumeralDischargeFollowsTheD2Rule (0.00s)
    --- FAIL: .../a_membership-only_row_does_NOT_discharge_a_count_claim
        numeral_test.go:209: want a finding containing "a membership registration
        does not discharge a count claim", got:
--- FAIL: TestNumeralAxisFindsNoUndeclaredCountClaim (0.00s)
    numeral_test.go:287: the numeral layer produced an EMPTY breadth set ...
```

### Population re-derived in-run — the run baseline (AC-RNA-013)

`go test -count=1 -v -run TestNumeralResidualArithmeticCloses ./internal/harness/rosterguard/...`,
run in THIS tree in THIS run. It reads `Registry()` and `NumeralExemptions()` at
RUNTIME — never a text parse of `registry.go`, which cannot see the four rows
`readmeSite()` builds from a path parameter. Verbatim:

```
numeral re-derivation (in-run, this tree)
  breadth-set hits                : 115
  breadth-set paths               : 63
  registry rows carrying Count    : 43
  registry paths carrying Count   : 39
  ...declared NumeralUnreachable  : 3
  hits discharged by a Count row  : 36
  hits discharged by an exemption : 27
  residual (undeclared)           : 0
  numeral exemptions declared     : 27
```

Arithmetic closes: 36 + 27 + 0 = 63 breadth paths, finding set empty.

**Residual measured before the registry expansion: 46 paths** (63 breadth − 17
discharged by the 17 reachable pre-existing count rows). That is one BELOW the 47 the
run baseline above this section recorded from the lead's own probe, and four below
nothing — the two probes are different implementations of the same rule, and the
delta is not attributable line-by-line because the lead's probe is not in the tree to
re-run. The layer's own figure is the one the criteria are decided against; the
lead's stands as the independent order-of-magnitude check it was.

**Authored against that residual: 50 declarations** — 23 `ClaimCount` rows over 19
paths (INDEX.md and its mirror each state the same size three times, so each takes
three rows: a `CountPattern` must match exactly once) plus 27 exemptions.

### §E.2 Run-phase Evidence — AC matrix

| AC | Status | Command | Actual output |
|----|--------|---------|---------------|
| AC-RNA-001 | PASS | `go test -run TestUndeclaredCountClaimNamesBothDischargePaths` | `ok … 0.107s`; the message names the numeral, the phrase, `ClaimCount` and `numeral exemption` |
| AC-RNA-002 | PASS | `go test -run TestNumeralDischargeFollowsTheD2Rule` | `--- PASS` on 7 rows: count row discharges, exemption discharges, membership-only does NOT, other-path row does NOT |
| AC-RNA-003 | PASS | `go test -run TestNumeralSelectionIsNearestPreceding` | `--- PASS`; the LIVE `CLAUDE.md §4 (the 13 retained agents` selects `13`, and the synthetic variant likewise |
| AC-RNA-004 | PASS | same test, row `an empty exempt reason is a finding` | `--- PASS`; a whitespace-only reason yields `the declaration is incomplete` and does NOT discharge |
| AC-RNA-005 | PASS | `go test -run 'TestNeutralisationRunsBeforeMatching\|TestLiveWordAxisBreadthSetIsEmptyOutsideThisPackage'` | `--- PASS`. (a) synthetic `thirteen retained agents` fires; (b) both live shapes reach neither set; (c) live word-axis hits outside this package: 0, cause named — `.claude/agents/harness/workflow-specialist.md:52` `All four are retained agents.`, read from disk and removed by subset-predication neutralisation |
| AC-RNA-006 | PASS | (a) `go test -count=1 ./internal/harness/rosterguard/... -v \| grep -c '^digit-axis hit '` → `111`; (b) `go test -run TestNumeralBreadthSetEqualsTheDeclaredUnion` | (a) `111` (+4 word-axis = 115 hits, matching the re-derivation); (b) `--- PASS` — set equality in both directions. **Deviation, declared:** the RHS as written in `acceptance.md` is unsatisfiable, and the resolution is recorded in §E.2 Deviations below |
| AC-RNA-007 | PASS | `go test -count=1 -v -run TestNumeralGuardFiresOnDeliberatelyWrongInput …` then the full package with no `-run` | `=== RUN TestNumeralGuardFiresOnDeliberatelyWrongInput` + 4 sub-runs, `--- PASS` (observed RUN, not `no tests to run`); full package `ok … 17.642s`, 20/20 PASS |
| AC-RNA-008 | PASS | `go test -run TestNumeralAxisFindsNoUndeclaredCountClaim` | `--- PASS`; the empty-breadth branch `t.Fatal`s with "measurement failure, not a clean tree". Exercised live by the RED run above, which took exactly that branch |
| AC-RNA-009 | PASS | `go test -run TestRegisteredSitesMatchTheirDeclaredAxis`; `git show --stat HEAD` | `--- PASS`; every newly-reached stale site carries a `CountPattern` matching exactly once plus a `KnownStale{DeclaredCount}`; the commit touches no document outside `internal/harness/rosterguard/`, so the stale prose is byte-unchanged |
| AC-RNA-010 | PASS | `git merge-base --is-ancestor ac540e554 e577f1ce6` | exit `0`; `$IMPL` resolved by the criterion's own command to `e577f1ce64b107dea86c54f690d6a5952566b40c`, distinct from `$BASE` `ac540e554` |
| AC-RNA-011 | PASS | `go test -count=1 -v ./internal/harness/rosterguard/...` | all five t922 tests `--- PASS` with their assertions unmodified; `SweepUnreachable` semantics unchanged (the new rows USE it, none alters it) |
| AC-RNA-012 | PASS | `go test -run TestNumeralResidualArithmeticCloses` | the block above; 36 + 27 + 0 = 63, finding set empty, 50 authored declarations reported against the plan-phase residual of 46 |
| AC-RNA-013 | PASS | same | produced in-run, in this tree, reading `Registry()` at runtime. The plan-phase figures are cited nowhere as the baseline |
| AC-RNA-014 | PASS | `go test -run TestHistoricalCitationsAreNotFindings`; `git show --stat HEAD` | `--- PASS`; the four `then-8-agent catalog` sites enter no finding, are unrepaired, and each is handled by a per-path exemption whose reason names the class |
| AC-RNA-015 | PASS | `go test -run TestNumeralScopeIsTheSweepScopePlusResearch` | `--- PASS`; `.moai/research/anthropic-best-practices-2026-05-24.md` is asserted OUT OF SCOPE (not exempted), and the exclusion set is the sweep's list plus exactly one entry |

Quality gates: `gofmt -l` clean · `go vet` exit 0 · `golangci-lint run --timeout=3m` →
`0 issues.` · `go build` and `GOOS=windows GOARCH=amd64 go build` exit 0 ·
`go test -cover` → **93.8% of statements** (target 85) · subagent-boundary grep over
the package: no matches.

### §E.2 Deviations

**AC-RNA-006(b)'s right-hand side is unsatisfiable as written, and is satisfied with
a declared subtrahend.** The criterion asks the breadth set to equal
`registered ClaimCount paths ∪ exempt paths` exactly. Measured: three registered
`ClaimCount` paths — `README.ko.md`, `README.ja.md`, `README.zh.md` — write their
count claim with a LOCALIZED roster noun (`에이전트 카탈로그` / `エージェント・カタログ`
/ `智能体目录`), and the adopted noun class is English-only by decision D1, so the
layer cannot reach them. Strict equality therefore fails on three paths for a reason
neither over-eager neutralisation nor a truncated breadth set.

The resolution follows the precedent already in the package rather than inventing a
mechanism: `Site.NumeralUnreachable`, the exact converse of the existing
`Site.SweepUnreachable`, a per-path DECLARATION carrying a mandatory reason. The
assertion becomes `breadth == (ClaimCount paths ∪ exempt paths) \ NumeralUnreachable
paths`, plus the converse check that a declared-unreachable path is NOT reached.

Both terms stay registry-derived, so the criterion's stated purpose holds: no
layer-produced term appears, and an over-eager neutraliser still cannot shrink both
sides together — it would make a declared path fall out of the breadth set, which the
first direction of the equality reports.

**The `manager-design` rows register a CORRECT count.** Three paths cite `§ 4 (13
retained agents` and agree with `template.ProfileMatrixAgents()`. They are registered
without a `KnownStale` marker, because an already-correct count is what a guard
protects: left undeclared, the next drift in them goes unreported.

**`.claude/rules/moai/NOTICE.md` is registered, not exempted.** Its sentence carries a
historical half ("8 retained agents at consolidation time") and a LIVE half ("now 10
per CLAUDE.md §4"). The live half gets a `CountPattern` and a `KnownStale{10}` rather
than a path-scoped exemption, which would have silenced a real stale claim.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-19
run_commit_sha: e577f1ce64b107dea86c54f690d6a5952566b40c
run_status: complete
ac_pass_count: 15
ac_fail_count: 0
preserve_list_post_run_count: 5    # the five t922 tests, all passing unmodified
l44_pre_commit_fetch: not-performed — lane-local card work, no push in this phase
l44_post_push_fetch: not-performed — the lane does not push; the lead batches develop
new_warnings_or_lints_introduced: 0    # golangci-lint "0 issues."; gofmt/vet clean
cross_platform_build:
  darwin_arm64: exit 0
  windows_amd64: exit 0
total_run_phase_files: 5
m1_to_mN_commit_strategy: >
  two commits — the baseline record ac540e554 (landed by the lead) and the single
  implementation commit e577f1ce6. M1..M6 were developed as one TDD arc and landed
  together because they are one file plus its registry: splitting them would have
  produced intermediate commits whose tests do not pass, and the baseline-first
  ordering AC-RNA-010 witnesses is between the baseline and the implementation, not
  within the implementation.
```

## §E.4 Sync-phase Audit-Ready Signal

### Adjacent repair on this branch, after the run-phase evidence above

Absorbing `develop` (merge `ca4084a98`) rewrote the false-positive sentence the
`foundation-cc-release-version` exemption was declared against
(`.claude/skills/moai-foundation-cc/SKILL.md` — the CC release version moved from
`2.1.172` to `2.1.197` and the roster noun fell outside the adjacency window), so the
layer stopped reaching that path and the exemption excepted nothing. The layer reported
this itself — `TestNumeralBreadthSetEqualsTheDeclaredUnion` failed naming the path and
prescribing the repair — and commit `7c4d09118` (`fix(rosterguard): drop the numeral
exemption develop made stale (card t930)`) deletes the stale row. Re-verified in this
session: `go build`, `go vet`, `go test -count=1`, `go test -cover`, and
`golangci-lint run --timeout=2m` all green on `./internal/harness/rosterguard/...`
(coverage 93.8%, `0 issues.`) at HEAD `7c4d09118`.

### CHANGELOG B12 self-tests (run before emission)

- **(a) pre-emission grep** — `grep -c 'SPEC-ROSTER-NUMERAL-AXIS-001' CHANGELOG.md` → `0`.
  No duplicate entry; emission proceeds.
- **(b) AC count match** — `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l`
  → `15`. `acceptance.md` §D AC Matrix states "15 criteria, under the Tier M ceiling of
  16" — counts agree.
- **(c) file path verification** — every path named in the CHANGELOG entry was checked
  with `ls`: `internal/harness/rosterguard/{numeral.go,numeral_test.go,
  numeral_rederivation_test.go,axis.go,registry.go}` all exist.

```yaml
sync_complete_at: 2026-09-20
sync_commit_sha: 86b1d886fc29fdd9bcaa3b0dc97182124b3f9a3d
sync_status: complete
b12_self_test_a: pass (grep count 0)
b12_self_test_b: pass (15 == 15)
b12_self_test_c: pass (5/5 paths verified via ls)
changelog_entry_position: "top of [Unreleased] > Added"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
canary_compliance_check: not-applicable — this SPEC does not define a forward-looking policy
```

No SPEC body content (`spec.md` / `plan.md` / `acceptance.md`) was modified in this
sync commit — only `spec.md` frontmatter `status:` + `updated:`, this `progress.md`
§E.4 section, and `CHANGELOG.md`.

🗿 MoAI
