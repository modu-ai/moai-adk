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
| AC-STV-016 non-overlap with `StatusValueEnumRule` | PASS (with a stated caveat) | `TestStatusTokenUnrecognizedDoesNotDuplicateStatusValueInvalid`, plus the corpus intersection below | intersection **0** — but one operand is empty, so read the caveat; the non-duplication property is now pinned by a constructed both-fire fixture |
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

The non-overlap that *is* substantively established comes from the corpus messages: all seven tokens
are observed in **git history** (`"Completed" → "completed"`, `"approved" → "completed"`, ...) while
every one of those documents' current frontmatter status is a valid enum member — which is precisely
the disjointness REQ-STV-015 describes.

**A second leg previously cited here has been struck as a mis-citation.** It named `AC-STV-015`'s
paired assertion as supporting evidence; that assertion pins non-overlap between the two NEW codes
(`StatusTokenUnrecognized` must not also emit `StatusTransitionInvalid`), whereas `AC-STV-016` and
REQ-STV-015 concern non-overlap with `StatusValueInvalid`. Different pair, so it supported nothing
here. Leg (a) above stands alone.

**And the property is now pinned by construction, not left corpus-conditional.** The two rules are
not structurally disjoint: a SPEC whose frontmatter carries an invalid status AND whose history
carries an unrecognized token lands in both sets, and the corpus intersection of 0 holds only
because `StatusValueInvalid` has population 0 today — it would lapse silently the first time such a
document appeared. `TestStatusTokenUnrecognizedDoesNotDuplicateStatusValueInvalid`
(`internal/spec/lint_transition_overlap_test.go`) constructs exactly that document and asserts what
REQ-STV-015 actually prohibits, which is **duplication** — the same fact reported twice — rather
than co-occurrence. On the overlapping fixture each finding names its own subject: `StatusValueInvalid`
names the frontmatter's current value (`"Completed"`) and never the history-only token, and
`StatusTokenUnrecognized` names the history-only token (`"approved"`, which appears nowhere in the
frontmatter). The pair check stays silent, as AC-STV-015 requires.

The fixture is not vacuous — three mutants were run against it and killed:

| Mutant | Change | Result |
|---|---|---|
| A | the rule declines to judge a document whose frontmatter status is invalid | KILLED — `StatusTokenUnrecognized findings = 0, want 1` |
| D | the token check reads `fm.Status` instead of the two history tokens, in both the message and the token list | KILLED — message names `"Completed" → "Completed"`, not the history-only token |
| E | `StatusValueEnumRule` drops the offending value from its message | KILLED — `StatusValueInvalid message does not name the frontmatter value "Completed"` |

Two narrower mutants **survived** and are recorded rather than hidden: mutating only the message's
`%q → %q` pair (B), or only the token list (C), leaves the other channel still naming the
history-only token, so the assertion still holds. That is a real property of the message — it
reports its subject through two redundant channels — not a gap in the assertion; killing it requires
removing both, which mutant D does.

**Scope note.** `AC-STV-016` is a corpus-level AC and is untouched by this fixture: the corpus
intersection remains **0**, so neither of the AC's two remedies is owed. If a corpus document ever
lands in both sets, the AC's second remedy (enumerate the bounded allowed-overlap set and amend the
AC to name it) applies, and amending `acceptance.md` is manager-spec's artifact, not run-phase's.

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

**What shelters the 97, measured rather than described.** The two shelters are not interchangeable,
and the split decides who begins gating if either narrows:

```
$ jq -r '.[]|select(.code=="StatusTransitionInvalid")|.file' lint-after-f1.json | sort -u | wc -l
      97
   (frontmatter status read from each of those 97 files)
      49 completed      → sheltered by terminal frontmatter status (terminalStatusEnum)
      48 implemented    → sheltered SOLELY by the grandfathered-era exemption
```

`implemented` is **not** in `terminalStatusEnum` (`internal/spec/lint.go:1339-1344`, which carries
`superseded` / `archived` / `rejected` / `completed`), so those 48 rest on the era exemption alone.
They are the population that **begins gating** the moment that exemption narrows — which card t382
(SPEC-ERA-H3-NARROWING-001) is actively doing in another lane. Recording the split is what turns
"the corpus is all sheltered" from a reassurance into a number someone can act on.

### Deviation from plan.md M2, recorded

plan.md M2 says "register both rules". One rule is registered
(`StatusTransitionValidityRule`) emitting both codes. A rule emitting more than one code has
precedent in this file (`BreakingChangeIDRule` emits `BreakingChangeMissingID` and `OrphanBCID`),
and `Code()` has no consumer outside the rules' own bodies. The AC set binds on codes, not on rule
count.

**The reason first recorded here was wrong, and is corrected rather than left standing.** It argued
the walk count as "one rule = one walk vs two rules = two". The real baseline was *one rule = two
walks* — `OwnershipTransitionRule` already performed one, and this card's rule added a second — so
the card doubled the corpus's most expensive query while citing plan.md §F's "reuse the existing
lookup" mitigation as satisfied. It was not: the implementation reused the *function*, not its
*result*. Registering one rule instead of two never addressed that, and the one-rule choice now
rests on the precedent above alone.

### The doubling, closed (F1)

`cachedOwnershipTransition` (`internal/spec/gitquery_cache.go`) memoizes the history walk per
document for the duration of one `Lint()` run, alongside the existing per-run `git rev-parse` cache,
and both rules now read through it. This removes the doubling this card introduced **and** the
pre-existing duplication, so the corpus pays one walk per document rather than one per rule.

Measured on this tree, same machine, back to back, by counting calls through
`getOwnershipTransitionRunner` on a full-corpus `Lint()` (the count, not wall-clock, is the property):

| Configuration | lookups | findings | elapsed |
|---|---|---|---|
| call sites reverted to the unmemoized hook | 1428 | 1200 | 8m34s |
| memoized (this tree) | **714** | 1200 | 4m0s |

Exactly halved, and the finding total is byte-for-byte unchanged — a memoization that moved a count
would be a bug, not an optimization. The elapsed figures are supporting context only: a loaded
developer machine measures the machine, and CI is the verdict.

Why it mattered on a deadline: CI's `spec-lint` job runs a shallow clone today, so the walk is
nearly empty there — but card t371 proposes `fetch-depth: 0`, at which point the doubled version's
local 7-8m sits against that job's `timeout-minutes: 10` (`.github/workflows/spec-lint.yml:28`).

Per-run invalidation and the direct-caller path are both pinned
(`internal/spec/lint_transition_memo_test.go`): a second `Lint()` in the same process re-reads git,
callers outside `Lint()` are uncached exactly as before, and the memo is keyed per document so no
two SPECs can share one record.

A comment on `StatusTransitionValidityRule` carried the same wrong claim, citing REQ-STV-012 — the
observation-only requirement, which says nothing about history walks — as its authority. The
citation and the claim are corrected together (F3), so the comment is not merely accidentally
correct now that the memo makes the walk-sharing true.

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
run_commit_sha: 73bfba170      # M1+M4 implementation commit — the tree the §E.2 measurement binary was built from (strings bin/moai | grep -c 73bfba170 → 4); sync-audit independently re-measured at ff8a7dcba with its own make-build binary
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

### Carried sync-audit verdict

The sync audit **already ran** before this close: verdict **PASS, weighted 91.2/100**, both must-pass
dimensions (Functionality 93, Security 95) cleared independently; 0 blocking findings, 4 optional
findings (F1-F4). Committed record: `.moai/reports/t376/sync-audit.md` (landed in `b43d4cb56`;
audited HEAD `ff8a7dcba`, base `a5d963db6`, measurement binary `make build` with
`strings bin/moai | grep -c ff8a7dcba → 4`). This close records that audit plus the close —
it does not order or run a new one.

Finding disposition — nothing the audit asked for is left open:

| Finding | Closed in | What landed |
|---|---|---|
| F1 (unmeasured doubling of the per-document git walk) | `b1bcce4f4` | `cachedOwnershipTransition` memoization — lookups 1428 → 714, findings byte-identical 1200 |
| F3 (rule comment citing REQ-STV-012 for a claim it does not make) | `b1bcce4f4` | citation + claim corrected together |
| F4 (AC-STV-016 mis-citation naming AC-STV-015's evidence) | `b1bcce4f4` | mis-citation struck from progress.md |
| F2 (§C limitation statement incomplete on shelter composition) | `ab114b2cf` | both demotion shelters named (49 terminal-status / 48 era-exemption) |

### Sync-phase changes (this close)

| File | Change |
|---|---|
| `progress.md` §E.3 | `run_commit_sha` backfilled: `73bfba170` (tree the §E.2 measurement binary was built from) |
| `progress.md` §E.4 | this section |
| `spec.md` frontmatter | `status: in-progress → completed` (3-phase close, single sync commit), `updated: 2026-09-01` |
| `.moai/reports/t376/sync-close.md` | close verdict report (this close) |

No CHANGELOG / docs-site / README emission — this branch's precedent is zero doc files, and the
sync audit required none (its 4 findings were all optional and all closed in code/record commits).

### What this close did NOT observe

- No tests, lint, or build were re-run. §E.2 / §E.3 figures and the sync-audit's measurements are
  cited from those runs, not re-measured here. This close's diff is markdown-only (SPEC artifacts
  + reports), touching no Go source.
- No CI verdict — nothing was pushed; integration into `develop` is a lead-granted window.

```yaml
sync_complete_at: 2026-09-01
sync_commit_sha: 77550b1b0   # backfilled in a follow-up commit (a commit cannot name its own hash)
sync_status: complete
run_commit_sha_backfilled: 73bfba170
sync_audit_verdict: "PASS 91.2/100 — committed in b43d4cb56 (.moai/reports/t376/sync-audit.md); audited HEAD ff8a7dcba; must-pass Functionality 93 + Security 95 cleared"
sync_audit_findings_disposition: "F1/F3/F4 closed in b1bcce4f4; F2 closed in ab114b2cf; 0 blocking"
b12_self_test_a: not-applicable — no CHANGELOG emission (branch precedent: zero doc files; audit required none)
b12_self_test_b: not-applicable — no CHANGELOG emission
b12_self_test_c: not-applicable — no CHANGELOG emission
docs_sweep: "not applicable — no CHANGELOG/docs-site/README surface required by this SPEC"
template_mirror: "not applicable — run-phase files are all internal/spec/ Go source; sync-phase touched SPEC artifacts only"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (3-phase close on the single sync commit; updated: 2026-09-01)"
  plan_md: "no status/updated fields (no frontmatter block) — nothing to transition"
  acceptance_md: "no status/updated fields (no frontmatter block) — nothing to transition"
  progress_md: "no frontmatter block — nothing to transition"
mx_tag_validation: "no @MX annotation added or changed; sync-phase touched no Go source"
canary_compliance_check: not-applicable
push_state: "not pushed — integration is a lead-granted window (per dispatch constraints)"
```
