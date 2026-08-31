# Progress — SPEC-STATUS-TRANSITION-VALIDITY-001

Card: **t376**. Tier M.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-31
tier: M
artifacts: [spec.md, plan.md, acceptance.md]
tree_of_record: 3f03d9c36
open_clarifications: 0
decisions_recorded: 5    # spec.md §A.5 D1-D5 (D4/D5 forced by the census, not in the handed-down set)
ac_count: 20             # 19 numeric ids AC-STV-001..019 + AC-STV-007a
plan_audit_iter1: PASS-WITH-DEBT 0.90 (Tier M threshold 0.80)
plan_audit_blocking_closed: [D1, D2, D3, D4, D7]
plan_audit_optional_open: [D5, D6, D8, D9]
```

## §E.2 Run-phase Evidence

### Tree of record and tool provenance

Run-phase work was carried out in worktree `.claude/worktrees/t376` on branch
`WT-status-transition-gap`, entered at `a5d963db6` (the merge of `origin/develop` was already
absorbed and clean). Every corpus number below was produced by a `make build` binary built from
`73bfba170` — the M1+M4 implementation commit — verified by `strings bin/moai | grep -c 73bfba170`
→ **4**. A plain `go build` omits the ldflags and would make that attribution unverifiable, which
is why the check exists. `git diff --stat 73bfba170 -- internal/spec/lint.go internal/spec/lint_transition.go`
is empty at close, so the production code the binary measured is the production code that landed.

### AC PASS/FAIL matrix

| AC | Status | Verification | Observed |
|---|---|---|---|
| AC-STV-001 `draft → completed` caught | PASS | `go test ./internal/spec/ -run 'TestStatusTransitionValidityRule/draft_to_completed_is_caught'` | exactly 1 `StatusTransitionInvalid`, message names `draft`, `completed`, and the transition SHA |
| AC-STV-002 `completed → draft` caught | PASS | same test, `completed_to_draft_is_caught` | exactly 1 finding, both statuses + SHA named |
| AC-STV-003 trailer-independence | PASS | `TestStatusTransitionTrailerIndependence` + `draft_to_completed_without_trailer_is_caught` | both repos emit the finding; after normalizing the SHA and the path the two messages are byte-identical, and severity / advisory / code / line all match |
| AC-STV-004 `implemented → completed` (right owner) silent | PASS | `implemented_to_completed_right_owner_passes` | 0 findings |
| AC-STV-005 `draft → in-progress` silent | PASS | `draft_to_in_progress_passes` (+ the three terminal-target cases) | 0 findings |
| AC-STV-006 `in-progress → implemented` silent | PASS | `in_progress_to_implemented_passes` | 0 findings |
| AC-STV-007 `completed → in-progress` amendment silent | PASS | `completed_to_in_progress_amendment_passes` | 0 findings |
| AC-STV-007a single-sync close + `draft → implemented` silent | PASS | `in_progress_to_completed_single_sync_close_passes`, `draft_to_implemented_passes` | 0 findings each; corpus confirms — neither edge appears among the 97 |
| AC-STV-008 `planned` edges tolerated | PASS | `planned_on_the_left_is_tolerated`, `planned_on_the_right_is_tolerated` | 0 findings; corpus `planned → implemented` (11) and `planned → completed` (4) are absent from the 97 |
| AC-STV-009 no transition / no git silent | PASS | `no_transition_in_history_is_silent`, `non_git_directory_is_silent`, `TestStatusTransitionRuleGuards/git_unreachable_is_silent` | 0 findings, none at error severity |
| AC-STV-010 live control in the same execution | PASS | the `firedAtLeastOnce` guard at the end of `TestStatusTransitionValidityRule` | fired. Its value is on record: in the RED run (`red-evidence.md`) all 20 must-stay-silent cases passed vacuously and this guard is what failed the run |
| AC-STV-011 demotion annotation names its cause | PASS | `TestApplyEraDemotionNamesItsCause`, `TestDemotionCauseEndToEnd` | the two single-cause annotations differ; the terminal-only one reads `[terminal lifecycle status — downgraded to warning]` and does not say "grandfathered". Confirmed on the live corpus: `SPEC-V3R6-LINK-FIX-001` (`era: V3R6` explicit, `status: completed`) carries exactly that annotation |
| AC-STV-012 observation-only | PASS | `TestStatusTransitionRuleIsObservationOnly`; `git status --porcelain` after the full-corpus run; `git diff a5d963db6..HEAD -- internal/spec/lint.go` | no tracked file modified by the run. The diff touches `terminalStatusEnum`, `eraDemotableCodes`, and the `StatusGitConsistencyRule` early return **not at all**; the demotion decision at the former `lint.go:239` keeps both disjuncts and the same OR — only the message moved |
| AC-STV-013 per-code baseline re-measurement | PASS | see the per-code table below | every pre-existing code is unmoved; two new codes appear |
| AC-STV-014 `completed → implemented` reversal caught | PASS | `completed_to_implemented_reversal_is_caught` | 1 finding; 48 on the corpus |
| AC-STV-015 unrecognized token fires its own code | PASS | the four token cases (`synced` / `approved` / `cancelled` / `Completed`) | each emits exactly 1 `StatusTokenUnrecognized` naming the token, and **0** `StatusTransitionInvalid` for the same pair |
| AC-STV-016 non-overlap with `StatusValueEnumRule` | PASS (with a stated caveat) | see below | intersection **0** — but one operand is empty, so read the caveat |
| AC-STV-017 `(none) → X` skipped | PASS | `none_to_completed_is_skipped`, `none_to_draft_is_skipped`, `TestStatusTransitionCheckOrder/none_skip_precedes_token_check` | 0 findings; on the corpus all 136 `(none)` records are absent from both new codes |
| AC-STV-018 the finding actually gates | PASS | `TestStatusTransitionFindingGates` | on a modern-era, non-terminal SPEC the finding carries `advisory: false` and `HasErrors()` is true under `--strict` |
| AC-STV-019 gating population measured and decided | PASS — nothing owed | `jq '[.[] \| select(.advisory != true)] \| length'` | **0** overall and **0** for each new code. See the decision note below |

### AC-STV-013 — per-code comparison

Baseline: `.moai/reports/t376/lint-baseline-merged.json`, measured on this tree at `a5d963db6`
before the change (the §A.1 table's own figures were taken at `3f03d9c36`; per-code they are
identical). After: `.moai/reports/t376/lint-after-m1.json`, `./bin/moai spec lint --json` at
`73bfba170`.

| Code | Baseline | After | Δ | Attribution |
|---|---:|---:|---:|---|
| CoverageIncomplete | 846 | 846 | 0 | untouched |
| MovingRefUnpinned | 114 | 114 | 0 | untouched |
| **StatusTransitionInvalid** | — | **97** | +97 | new — this card |
| LegacyEARSKeyword | 43 | 43 | 0 | untouched |
| ModalityMalformed | 25 | 25 | 0 | untouched |
| MissingExclusions | 24 | 24 | 0 | untouched |
| StatusGitConsistency | 18 | 18 | 0 | untouched |
| FrontmatterInvalid | 14 | 14 | 0 | untouched |
| **StatusTokenUnrecognized** | — | **7** | +7 | new — this card |
| InvalidREQID | 6 | 6 | 0 | untouched |
| SyncSHASlotFormat | 5 | 5 | 0 | untouched |
| OwnershipTransitionInvalid | 1 | 1 | 0 | untouched — the pre-existing rule's own verdict did not move |
| **Total** | **1096** | **1200** | **+104** | 97 + 7, entirely attributable |

No code this card did not touch moved by even one finding.

### The projection miss, explained rather than absorbed

spec.md §A.6 projected **~98** `StatusTransitionInvalid` by hand from the census table
(50 `draft → completed` + 48 `completed → implemented`). Observed: **97**, split
`draft → completed` **49** / `completed → implemented` **48**. The token projection (~7) matched
exactly, and the seven observed tokens are the census's own — `Completed` ×3, `Superseded`,
`approved`, `synced`, `cancelled`.

The one-document gap is identified, not estimated. Re-running the census at this tree still counts
50 `draft → completed`, so the difference is not census drift. Diffing the two sets names the
document: **`SPEC-V3R6-LINK-FIX-001`**. Its frontmatter opens with the rejected snake_case alias
`spec_id:` instead of `id:`, so the YAML decoder yields an empty ID and the rule's first guard
(`fm.ID == "" || fm.Status == ""` — the same guard `OwnershipTransitionRule` uses) skips it. The
census reads the SPEC ID from the *directory name*, which is why it sees a document the linter's
ID guard does not.

Skipping it is correct, not a gap to close in this card: `FrontmatterInvalid` already reports the
missing `id` on that SPEC, and naming a transition on a document whose identity cannot be read
would report a second fact resting on a broken premise. The behavior is pinned as a regression test
(`TestStatusTransitionRuleGuards/unreadable_frontmatter_id_defers_to_frontmatter_rule`) so the
explanation cannot quietly stop being true.

### AC-STV-016 — the intersection, and what it does and does not establish

```
StatusTokenUnrecognized population : 7
StatusValueInvalid      population : 0
intersection (documents in both)   : 0
```

The AC passes: zero documents appear in both sets. **But the intersection is empty because one
operand is empty** — the corpus currently carries no `StatusValueInvalid` finding at all, so this
measurement would read the same against a rule that overlapped completely. Reporting it as measured
disjointness without that sentence would be exactly the vacuous-green shape this SPEC exists to
close.

The non-overlap that *is* substantively established comes from two other places: the corpus messages
show all seven tokens observed in **git history** (`"Completed" → "completed"`, `"approved" →
"completed"`, ...) while every one of those documents' current frontmatter status is a valid enum
member — which is precisely the disjointness REQ-STV-015 describes; and the paired assertion in
`AC-STV-015`'s test cases fails if a token case ever also emits `StatusTransitionInvalid`.

### AC-STV-019 — the gating population, and the decision

```
$ jq '[.[] | select(.advisory != true)] | length' lint-after-m1.json
0
$ jq '[.[] | select(.advisory != true and .code=="StatusTransitionInvalid")] | length'  → 0
$ jq '[.[] | select(.advisory != true and .code=="StatusTokenUnrecognized")] | length'  → 0
```

**The non-advisory count is 0**, overall and for each new code — unchanged from the baseline's 0.
`spec-lint --strict` on `main` / `develop` does not redden. Under the AC's own terms nothing is
owed: the recorded decision is required only when the count is non-zero.

The reason is the accepted limitation already stated in spec.md §C, now measured rather than
predicted: all 104 new findings land on documents `applyEraDemotion` demotes — 97 of the 97
`StatusTransitionInvalid` findings end in `completed` or sit on grandfathered-era directories, and
`completed` shelters itself. The rule sets no emission-site `Advisory` flag (REQ-STV-009), so it
**can** gate — AC-STV-018 demonstrates it gating on a modern-era, non-terminal SPEC — but on
today's corpus no document is in that position. This is worth stating plainly rather than reading
the 0 as a clean bill: the gate is live and the corpus is simply all sheltered.

### Deviation from plan.md M2, recorded

plan.md M2 says "register both rules". One rule is registered
(`StatusTransitionValidityRule`) emitting both codes. Reason: `lookupOwnershipTransitionFromGit`
is not memoized — `gitquery_cache.go` caches only the `git rev-parse` environment probes — so two
rules would mean two `git log --follow -p` walks per document instead of one, against plan.md §F's
own "reuse the existing lookup" mitigation. A rule emitting more than one code has precedent in
this file (`BreakingChangeIDRule` emits `BreakingChangeMissingID` and `OrphanBCID`), and `Code()`
has no consumer outside the rules' own bodies. The AC set binds on codes, not on rule count.

### Verification commands (this run, this tree)

```
$ go test ./internal/spec/... -count=1        → ok  github.com/modu-ai/moai-adk/internal/spec  60.405s
$ go vet ./internal/spec/...                  → exit 0
$ golangci-lint run ./internal/spec/...       → 0 issues.
$ find internal/spec -maxdepth 1 -name 'zz_t376*' | wc -l  → 0
```

Full-suite verdict is CI's; the local run is scoped to the affected package per CLAUDE.local.md §4.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-31
run_commit_sha: pending-backfill-final-run-commit
run_status: complete
ac_pass_count: 20
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 0
scratch_probes_removed: [zz_t376_probe_test.go, zz_t376_census_test.go]
new_warnings_or_lints_introduced: 0     # golangci-lint ./internal/spec/... → 0 issues
per_code_regression: none               # every pre-existing code Δ = 0
new_findings_total: 104                 # StatusTransitionInvalid 97 + StatusTokenUnrecognized 7
non_advisory_findings_total: 0          # AC-STV-019 — strict gate unchanged, no decision owed
measurement_binary_sha: 73bfba170       # strings bin/moai | grep -c 73bfba170 → 4
total_run_phase_files: 5                # lint.go, lint_transition.go, lint_transition_test.go, lint_demotion_cause_test.go, lint_phase_test.go
m1_to_mN_commit_strategy: two commits (M1+M4 implementation; M2+M3+M5 measurement, guards, evidence)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
