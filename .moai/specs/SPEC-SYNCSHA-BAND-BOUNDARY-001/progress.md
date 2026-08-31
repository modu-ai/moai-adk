# SPEC-SYNCSHA-BAND-BOUNDARY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

**Card**: t380 · **Tier**: S · **Era**: V3R6 · **Module**: `internal/spec`
**Tree**: `.claude/worktrees/t380`, branch `WT-shared-predicate-guard`,
base `3f03d9c36` (= `origin/develop` at dispatch)
**Plan audit**: PASS-WITH-DEBT **0.9216** (Tier S threshold 0.75), 1/1 iteration,
`.moai/reports/t380/plan-audit.md`. Central claim REPRODUCED under the auditor's
own hands — all four band mutants red, no direction green.

### Artifacts

| File | State |
|---|---|
| `spec.md` | v0.2.0 — 8 GEARS requirements (REQ-SBB-001..008), per-fixture mutation map (§B.1), §B.2 records the test-function ruling, 7 out-of-scope topics |
| `plan.md` | 6 milestones, reversibility-ordered; M1 and M2 open points both CLOSED |
| `acceptance.md` | **8** criteria (AC-SBB-001..008) at the Tier S ceiling, plus a RED-now evidence ledger (§B, R-1..R-4) |
| `progress.md` | this file |

### Measurements taken at plan phase (this tree, `3f03d9c36`)

| Claim | Command | Observed |
|---|---|---|
| SHA fixture token lengths are 9 / 40 / 9 / 9 — none at the band floor | `awk '/^sync_commit_sha:/ {v=$2; gsub(/"/,"",v); print FILENAME, v, length(v)}' internal/spec/testdata/syncsha/{sha-short,sha-full,sha-quoted,sha-annotated}/progress.md` | `a6bbbf82b 9` / `…d5e6 40` / `a6bbbf82b 9` / `a6bbbf82b 9` |
| Band coverage exists at the predicate level (7 / 40 / 6 / 41 rows) | `sed -n '52,78p' internal/spec/syncsha_test.go` | `TestIsCommitSHAToken_LengthBand` carries all four boundary rows |
| The mutation site is a call, not a pattern | `grep -n 'isCommitSHAToken' internal/spec/lint_syncsha.go` | `103: if isCommitSHAToken(token) {` |
| The `{40}` comment sits on line 123, not 124 (audit D6) | `grep -n 'd5e6f\|{40}' internal/spec/lint_syncsha_test.go` | `123: // Mutation that must turn it red: narrow the SHA pattern to {40} …` |
| `lint_syncsha_test.go` has 3 test functions (pre-work baseline) | `grep -c '^func Test' internal/spec/lint_syncsha_test.go` | `3` |
| Targeted suite GREEN on the clean tree | `go test ./internal/spec/ -run 'TestSyncSHASlot\|TestIsCommitSHAToken' -count=1` | `ok … 2.496s`, exit 0 |
| Working tree carries no source change | `git status --short` | before authoring: `?? .moai/reports/t380/` alone. After authoring (current): that entry plus `?? .moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/`. Both describe the same fact — no source change. |
| SPEC ID conforms to `specIDPattern` | `[[ "SPEC-SYNCSHA-BAND-BOUNDARY-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` | `PASS` |
| SPEC ID is unique | `ls -d .moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001` | `No such file or directory` (before authoring) |
| Criterion count is at the Tier S ceiling | `grep -c '^### AC-SBB-' .moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/acceptance.md` | `8` |
| No dangling AC ID after renumbering | `grep -ohE 'AC-SBB-[0-9]{3}' <4 artifacts> \| sort -u` | `AC-SBB-001` … `AC-SBB-008`, contiguous, no `009` |

### Figures quoted from elsewhere, with provenance stated

- `go test ./internal/spec/... -count=1` GREEN at exit 0 in 34.439s under the
  surviving mutant — **quoted from SPEC-SYNC-SHA-SLOT-FORMAT-001 §E.4**, not
  re-measured here.
- Four-mutant reproduction, baseline `ok … 3.201s` exit 0, and the alphabet-mutant
  survival at `ok … 3.082s` exit 0 — **attributed to the plan audit**
  (`.moai/reports/t380/plan-audit.md`), not measured here.
- The probe's verbatim FAIL (`sha-min7: expected 0 findings … got 1`) — read from
  `.moai/reports/t380/probe-05-mutant-red.txt`, produced in this tree before
  authoring and fully reverted.

### Diagnosis correction recorded

`spec.md` §A.1 states explicitly that the originating card's diagnosis
("AC-SSF-007 has no CI carrier") is wrong, and §A.2 adds the sharper measured
distinction: band coverage exists at `isCommitSHAToken` and is absent at
`SyncSHASlotFormatRule.Check`, which is where the mutant lands. `spec.md` §C
preserves AC-SSF-007's remaining CI-automation gap as a **different** gap, not one
this card closes.

### Plan-audit defect disposition (D1-D7)

| Defect | Class | Disposition |
|---|---|---|
| D1 — `TestSyncSHASlot_FlagsProse` name is load-bearing in t299's AC-SSF-001 | blocking | **Operator ruling, option (a)**: new `TestSyncSHASlot_FlagsOutOfBand`; prose function untouched; this card's own delivery-shape criterion relaxed from a `func Test` count to a no-duplicated-assertion-body test. REQ-SBB-005 rewritten to match; `plan.md` M2 closed. |
| D2 — 9 criteria against the Tier S ceiling of 8 | blocking | **Operator ruling**: merged the pre-audit AC-SBB-005 (delivery shape) and AC-SBB-008 (no production source change) into one criterion, both read off the same `git diff --stat`. Count now 8; stays `tier: S`. The pre-audit ninth criterion (AC-SSF-007 untouched) renumbered down to AC-SBB-008. |
| D3 — alphabet residual uncovered at rule level | minor | **Declared, not fixed**: new §C topic "Out of Scope — the alphabet clause of the §D.1 grammar", carrying the audit's measurement (`^[0-9a-f]{7,40}$` mutant survives at `ok … 3.082s` exit 0) attributed to the audit. The incidentally-covered anchor clause is recorded as an observed bonus, explicitly not claimed as a designed property. |
| D4 — no RED-now four-element cells | minor | **Added**: `acceptance.md` §B carries a four-element evidence ledger (R-1..R-4) with a document-level tree pin of `3f03d9c36`; each mutation-bearing criterion cites its entry. R-3/R-4 are three-element and say so — the missing verbatim stdout is named, not inferred. |
| D5 — `priority: P2 Medium` is a hybrid of two enum forms | minor | **Fixed in this SPEC only**: `priority: P2`. No other file touched — the same hybrid in `sha-short/spec.md` and elsewhere is inherited convention outside this card's scope. |
| D6 — `lint_syncsha_test.go:124` cited for text on line 123 | minor | **Fixed**: re-measured and corrected to `:123-124` (the span). |
| D7 — stale `git status --short` figure in three places | minor | **Labelled and re-measured**: `spec.md` §A.3, `plan.md` §C, and this file each now carry the before-authoring figure with its timing label plus the current two-entry value. |

### Prediction to be checked at run phase (NOT an observation)

`moai spec lint` is predicted not to move: the new fixtures live under
`internal/spec/testdata/` rather than `.moai/specs/`, so they should fall outside
the linted corpus. **Unmeasured at plan phase.** Run phase records before/after
finding counts and reports the prediction confirmed or falsified.

### Gaps

- No run-phase mutation observation by this author yet. The plan audit observed
  all four red under its own hands, but two of its ledger entries (R-3, R-4) carry
  no verbatim stdout and none carries an exit code as its own field. AC-SBB-006 /
  AC-SBB-007 require re-measurement at run phase.
- The lint-corpus prediction above is unmeasured.
- The alphabet clause of the §D.1 grammar is uncovered at the rule level, declared
  in `spec.md` §C as a residual this card does not close.

**Status**: `draft` — plan-audit defects dispositioned; awaiting Implementation
Kickoff Approval.

---

## §E.2 Run-phase Evidence

_<pending run-phase>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
