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

All measurements below were taken by this author (manager-develop, run phase) in
this tree. **Document-level tree pin for every cell: `f28a7424f`** (the M4 commit)
unless a cell names its own. The working tree at measurement time carried no
tracked modification — only the untracked evidence directory
`.moai/reports/t380/verify-run/`. Nothing here is carried over from the plan audit
or from t299; where a figure originates elsewhere, it says so.

### Commits

| Milestone | SHA | Subject |
|---|---|---|
| M3 | `971e8b480` | `test(SPEC-SYNCSHA-BAND-BOUNDARY-001): M3 band-boundary fixtures (t380)` |
| M4 | `f28a7424f` | `test(SPEC-SYNCSHA-BAND-BOUNDARY-001): M4 register the band boundary cases (t380)` |

M5 planted and reverted four transient mutants and produced no commit of its own;
M6 carries the evidence export and this write-up.

### The four-element mutation observations (completing ledger R-1..R-4)

Each mutant replaced the `isCommitSHAToken(token)` call at
`internal/spec/lint_syncsha.go:103` with an inlined
`regexp.MustCompile("^[0-9a-fA-F]{N,M}$").MatchString(token)` and added the
`regexp` import. **One mutant was in the tree at a time**; each was reverted by
restoring `internal/spec/lint_syncsha.go` from HEAD before the next was planted,
and the revert was confirmed by `grep -c 'regexp' internal/spec/lint_syncsha.go`
returning `0` plus an empty diff against HEAD.

| Id | Mutant | Direction | Command | Verbatim stdout (FAIL lines) | Exit code | Revert confirmed |
|---|---|---|---|---|---|---|
| R-1 | `{8,40}` | floor narrowed — the t299 surviving mutant, verbatim | `go test ./internal/spec/ -run 'TestSyncSHASlot' -count=1` | `--- FAIL: TestSyncSHASlot_SilentOnSHA (1.53s)` — `lint_syncsha_test.go:140: sha-min7: expected 0 findings on a well-formed SHA, got 1: [{testdata/syncsha/sha-min7/progress.md 9 warning SyncSHASlotFormat ...}]` — `FAIL ... internal/spec 3.498s` | `1` | yes — `grep -c 'regexp'` → `0`; empty diff |
| R-2 | `{7,39}` | ceiling narrowed | same | `--- FAIL: TestSyncSHASlot_SilentOnSHA (1.60s)` — `lint_syncsha_test.go:140: sha-full: expected 0 findings on a well-formed SHA, got 1: [{testdata/syncsha/sha-full/progress.md 9 warning SyncSHASlotFormat ...}]` — `FAIL ... internal/spec 3.652s` | `1` | yes — same two checks |
| R-3 | `{6,40}` | floor widened | same | `--- FAIL: TestSyncSHASlot_FlagsOutOfBand (0.71s)` — `lint_syncsha_test.go:191: sha-below6: expected exactly 1 SyncSHASlotFormat finding, got 0: []` — `FAIL ... internal/spec 3.450s` | `1` | yes — same two checks |
| R-4 | `{7,41}` | ceiling widened | same | `--- FAIL: TestSyncSHASlot_FlagsOutOfBand (0.69s)` — `lint_syncsha_test.go:191: sha-above41: expected exactly 1 SyncSHASlotFormat finding, got 0: []` — `FAIL ... internal/spec 3.475s` | `1` | yes — same two checks |

Full untruncated output per mutant:
`.moai/reports/t380/verify-run/mutant-01-floor-narrow-8-40.txt`,
`mutant-02-ceiling-narrow-7-39.txt`, `mutant-03-floor-widen-6-40.txt`,
`mutant-04-ceiling-widen-7-41.txt` — each carrying an `exit_code:` line as its own
field.

Ledger entries R-3 and R-4, which `acceptance.md` §B recorded as three-element
(verbatim stdout and exit code absent, attributed to the plan audit), are
**completed here** by this author's own measurement. Nothing from the audit's
table was carried forward as if it were mine — note that the audit's R-2 output
named `TestSyncSHASlot_FlagsProse` under its pre-ruling design, whereas the
observation above lands in `TestSyncSHASlot_FlagsOutOfBand`, the delivered design.

Each mutant killed **exactly one** fixture, which is `spec.md` §B.1's per-fixture
mapping confirmed rather than assumed: R-1 killed `sha-min7` and left `sha-full`
silent; R-2 the reverse; R-3 and R-4 each dropped exactly one outside-band
fixture's count to 0.

### Test-first evidence (RED before GREEN)

The new function and the `sha-min7` registration were written **before** the
fixtures existed, and observed red:

```
--- FAIL: TestSyncSHASlot_FlagsOutOfBand (0.00s)
    lint_syncsha_test.go:191: sha-below6: expected exactly 1 SyncSHASlotFormat finding, got 0: []
    lint_syncsha_test.go:191: sha-above41: expected exactly 1 SyncSHASlotFormat finding, got 0: []
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/spec	2.618s
```

Command `go test ./internal/spec/ -run 'TestSyncSHASlot' -count=1`, exit code `1`,
file `.moai/reports/t380/verify-run/red-01-fixtures-absent.txt`. After the
fixtures were created the same selector returned `ok ... 3.645s`, exit `0`, with
all four `TestSyncSHASlot_*` functions PASS.

**What that RED also revealed, stated rather than buried**: the `sha-min7`
registration did **not** turn red with its fixture absent. `syncSHAFindings`
returns zero findings without error when the fixture directory does not exist, and
`TestSyncSHASlot_SilentOnSHA` expects zero — so a deleted case directory is
indistinguishable there from a passing one. This is a pre-existing property of
that function (it applies equally to the four fixtures t299 registered), not one
introduced by this card, and repairing it would mean changing an existing
function's assertions, which this card's scope forbids. It is recorded as a
residual risk in §E.3, and it is why AC-SBB-001's real evidence is the R-1
mutation observation rather than the green run.

### AC PASS/FAIL matrix — four-element cells

| AC | Verdict | Command | Verbatim observation | Exit code | Tree |
|---|---|---|---|---|---|
| AC-SBB-001 — band floor inside is silent | **PASS** | `go test ./internal/spec/ -run TestSyncSHASlot_SilentOnSHA -count=1` (clean tree); mutation evidence R-1 | clean tree: `ok ... internal/spec 3.645s` with `--- PASS: TestSyncSHASlot_SilentOnSHA (1.68s)`. Under the `{8,40}` mutant: `sha-min7: expected 0 findings on a well-formed SHA, got 1` | `0` clean / `1` mutant | `f28a7424f` |
| AC-SBB-002 — band ceiling inside is silent | **PASS** | same selector; mutation evidence R-2 | clean tree as above. Under the `{7,39}` mutant: `sha-full: expected 0 findings on a well-formed SHA, got 1` | `0` clean / `1` mutant | `f28a7424f` |
| AC-SBB-003 — one below the floor is flagged | **PASS** | `go test ./internal/spec/ -run TestSyncSHASlot_FlagsOutOfBand -count=1`; mutation evidence R-3 | clean tree: `--- PASS: TestSyncSHASlot_FlagsOutOfBand (0.57s)`. Under the `{6,40}` mutant: `sha-below6: expected exactly 1 SyncSHASlotFormat finding, got 0: []` | `0` clean / `1` mutant | `f28a7424f` |
| AC-SBB-004 — one above the ceiling is flagged | **PASS** | same selector; mutation evidence R-4 | clean tree as above. Under the `{7,41}` mutant: `sha-above41: expected exactly 1 SyncSHASlotFormat finding, got 0: []` | `0` clean / `1` mutant | `f28a7424f` |
| AC-SBB-005 — delivery shape + no production source change | **PASS** | diff-stat against base `3f03d9c36`, first scoped to the three production files, then to `internal/spec/` | Production-scoped diff produced **empty output** — none of `syncsha.go` / `lint_syncsha.go` / `closer.go` appears. Package-scoped diff: 7 files, `213 insertions(+), 3 deletions(-)`, every path under `internal/spec/testdata/syncsha/` except `internal/spec/lint_syncsha_test.go`. Clause 2 (no duplicated assertion body): `TestSyncSHASlot_FlagsOutOfBand` carries its own `[]struct{fixture, prefix string}` case list against `syncSHAFindings` and parameterizes the flagged-line prefix, where `TestSyncSHASlot_FlagsProse` hardcodes `"sync_commit_sha: TBD"` — read off the diff, not counted by lines. Clause 3: the test-file diff shows the prose function's body unchanged; only its neighbours and the doc comment above `TestSyncSHASlot_SilentOnSHA` moved | `0` | `f28a7424f` |
| AC-SBB-006 — a band-WIDENING mutant is observed red across the outside pair | **PASS** | R-3 (`{6,40}`) and R-4 (`{7,41}`), planted and observed **separately**, one run per direction | Both FAIL blocks quoted above; both exit `1`; both reverted with `grep -c 'regexp'` → `0`; suite re-observed GREEN after the final revert (`ok ... 39.898s`, exit `0`, full package) | `1` each | `f28a7424f` |
| AC-SBB-007 — a band-NARROWING mutant is observed red across the inside pair | **PASS** | R-1 (`{8,40}`) and R-2 (`{7,39}`), planted and observed separately | Both FAIL blocks quoted above; both exit `1`; both reverted and confirmed. R-1 is the t299 surviving mutant verbatim — observing it red is what closes debt D1 | `1` each | `f28a7424f` |
| AC-SBB-008 — AC-SSF-007 untouched, distinction preserved | **PASS** | diff-stat against base scoped to `.moai/specs/SPEC-SYNC-SHA-SLOT-FORMAT-001/` | **Empty output** — that SPEC directory, `acceptance.md` included, is unchanged. `spec.md` §C's "Out of Scope — AC-SSF-007" topic records the CI-automation gap as distinct from the fixture-corpus gap, unedited by this card | `0` | `f28a7424f` |

**8 PASS, 0 FAIL, 0 PASS-WITH-DEBT.**

### Close-out measurements (clean tree, `f28a7424f`)

| Claim | Command | Observed | Exit | Evidence file |
|---|---|---|---|---|
| Affected-package suite GREEN | `go test ./internal/spec/... -count=1` | `ok ... internal/spec 39.898s` | `0` | `m6-gotest-internal-spec.txt` |
| Vet clean | `go vet ./internal/spec/...` | no output | `0` | `m6-govet-internal-spec.txt` |
| Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | no output | `0` | `m6-build-windows.txt` |
| Cross-platform vet (test files compile) | `GOOS=windows GOARCH=amd64 go vet ./internal/spec/...` | no output | `0` | `m6-vet-windows-internal-spec.txt` |
| Binary builds | `go build -o bin/moai ./cmd/moai` | no output | `0` | `m6-gobuild.txt` |
| The delivered test file is gofmt-clean | `gofmt -l internal/spec/` | 12 files listed; **`lint_syncsha_test.go` is NOT among them**. All 12 are unchanged by this card — the diff against base scoped to `internal/spec/*.go` names `internal/spec/lint_syncsha_test.go` alone — so they are pre-existing and not attributable here | — | — |

### §G prediction — CONFIRMED, and by a stronger check than the total

| Claim | Command | Observed |
|---|---|---|
| `moai spec lint` total | `./bin/moai spec lint` | `0 error(s), 1096 warning(s)`, exit `0` |
| By-rule split | `grep -E '^(ERROR\|WARNING)' <out>` piped through `awk '{print $1, $2}'`, `sort`, `uniq -c`, `sort -rn` | `846 CoverageIncomplete` / `114 MovingRefUnpinned` / `43 LegacyEARSKeyword` / `25 ModalityMalformed` / `24 MissingExclusions` / `18 StatusGitConsistency` / `14 FrontmatterInvalid` / `6 InvalidREQID` / `5 SyncSHASlotFormat` / `1 OwnershipTransitionInvalid` — summing to 1096 |
| The new fixtures are outside the linted corpus | `grep -c 'testdata' <out>` and `grep -c 'SSFA' <out>` | `0` and `0` |

The zero-occurrence check is the load-bearing one. An unmoved total is consistent
BOTH with the fixtures being outside the corpus AND with their being inside it and
clean; zero occurrences of `testdata` and of the fixtures' shared id `SSFA` across
the whole output establishes that `moai spec lint` never reached them, so this
card **cannot** have moved the count. `plan.md` §G's prediction is confirmed on
that ground rather than on a coincidence of totals.

**Attribution note**: `1096` is this author's own measurement on this tree at
`f28a7424f`. The delegation prompt supplied the same figure as an expectation; it
was not carried forward — the run produced it independently, and the two `grep -c`
checks are what decide the prediction either way.

### Evidence export

Everything cited above lives under `.moai/reports/t380/verify-run/` (tracked), NOT
`.moai/state/verify/` (gitignored, and disposed with this worktree). File count by
`find .moai/reports/t380/verify-run -maxdepth 1 -mindepth 1 | wc -l`: **11**.

One exception, handled by extract rather than by export: the raw `moai spec lint`
output is 341,577 bytes / 1,101 lines, too large to commit whole. It is NOT
committed. `m6-speclint-after-extract.txt` records its sha256
(`178729c73535a25de9a35c2b127ad8d8d07ae2259975388a261288a5d1ada689`), its byte and
line counts, its exit code, the command that produced it, the command to
regenerate and re-verify it, and the three extracts actually cited above.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-31
run_commit_sha: pending-backfill-run     # the M5/M6 evidence commit; backfilled once it lands
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 0          # no production source file modified at close
l44_pre_commit_fetch: not-performed      # isolated worktree, no push; the lead owns integration
l44_post_push_fetch: not-applicable      # this lane does not push (dispatch constraint)
new_warnings_or_lints_introduced: 0      # spec lint 0 error / 1096 warning; 0 testdata + 0 SSFA occurrences
cross_platform_build:
  darwin_build: pass                     # go build -o bin/moai ./cmd/moai, exit 0
  windows_build: pass                    # GOOS=windows GOARCH=amd64 go build ./..., exit 0
  windows_vet: pass                      # GOOS=windows GOARCH=amd64 go vet ./internal/spec/..., exit 0
total_run_phase_files: 7                 # 6 fixture files + lint_syncsha_test.go
m1_to_mN_commit_strategy: one-commit-per-milestone-not-pushed
mutants_planted: 4
mutants_observed_red: 4
mutants_reverted: 4
production_source_diff_vs_base: empty    # base 3f03d9c36, scoped to syncsha.go + lint_syncsha.go + closer.go
```

### Gaps — what was NOT observed

- **Full-tree suite.** `go test ./...` was NOT run; only `./internal/spec/...`.
  The whole-tree verdict belongs to CI, on a clean environment against the pushed
  head.
- **Lint before/after delta.** No `moai spec lint` run was taken on the base tree
  `3f03d9c36`, so no measured delta exists. What was measured is the after-total
  (1096) plus zero occurrences of the fixtures in the output — which settles §G's
  prediction without a delta, but is not the same thing as one.
- **`golangci-lint`.** Not run. Only `go vet` and `gofmt -l`.
- **Coverage.** Not measured. The delivery adds no production statements, so a
  package coverage figure would move for reasons unrelated to this card.
- **The alphabet clause.** Declared out of scope in `spec.md` §C and not
  re-measured here; the audit's `^[0-9a-f]{7,40}$` survival figure stays
  attributed to the audit.
- **Anchor-dropping mutant.** `spec.md` §C's "observed bonus" was not re-measured
  under this author's hands; it remains attributed to the plan audit, and no
  criterion rests on it.
- **Push / integration / CI.** Nothing pushed, nothing merged, no CI observed.

### Residual risk — what could still be wrong despite what was observed

- **`TestSyncSHASlot_SilentOnSHA` passes vacuously on a deleted fixture.**
  Measured in the M4 RED run: with `sha-min7` absent, `syncSHAFindings` returned
  zero findings without error and the function stayed green. Any of its five case
  directories could be deleted and the suite would not notice — the empty-sweep
  shape of `verification-completeness.md` §1.1. Pre-existing (it applies to the
  four cases t299 registered), not introduced here, and not repaired: fixing it
  means changing an existing function's assertions, outside this card's scope.
  Consequence to hold: the two band-INSIDE cases are guarded by the R-1 / R-2
  mutation observations, not by the green run.
- **The mutants were planted at the call site only.** Each replaced the
  `isCommitSHAToken` call in `lint_syncsha.go`; none altered
  `commitSHATokenPattern` itself. A change to the pattern would be caught by
  `TestIsCommitSHAToken_LengthBand` at the predicate level, but this card measured
  that layer not at all.
- **The write-side gate is unmeasured.** `needsSHABackfill`
  (`internal/spec/closer.go:421`) shares the predicate and was neither mutated nor
  exercised here. A rule-level band change and a write-side band change remain
  independently expressible; this card closes the read side only.
- **Fixture era classification is asserted per run, not pinned.** The fixtures
  rely on era heuristic H-4 holding (`§E.2` + `§E.4` + non-empty
  `sync_commit_sha`). The criteria assert `Advisory == false`, so a demotion
  surfaces as a failure rather than a silent pass — but a change to `era.go`'s
  heuristics would break these fixtures for a reason unrelated to the band.
- **`19b6f76` is a real commit prefix today** (`19b6f7625` shortened). Should that
  object ever cease to resolve, the fixture is unaffected — the rule reads shape,
  not existence — but the value's stated provenance stops being true.

---

## §E.4 Sync-phase Audit-Ready Signal

Authored by manager-docs at sync phase, in this tree
(`.claude/worktrees/t380`, branch `WT-shared-predicate-guard`).
**Document-level tree pin for every cell below: `6672fd40f`** — the run-phase
HEAD this sync phase measured against. Figures produced by another actor say so
and name that actor; nothing from §E.2, from the plan audit, or from t299 is
re-served here as a sync-phase measurement.

### Claim

The 3-phase close (`in-progress → implemented → completed`) is carried on this
single sync commit, together with the `CHANGELOG.md` `[Unreleased] → Added`
entry and this signal. The delivery is test-only: the production source diff
against the base `3f03d9c36` is empty, and the SPEC's own acceptance set is
**8 criteria (AC-SBB-001..008), 8 PASS / 0 FAIL** as recorded in §E.2 by
manager-develop.

### Evidence

Commands run by this author at sync phase, with their observed output:

| # | Claim | Command | Observed | Exit |
|---|---|---|---|---|
| S-1 | Production source unchanged | `diff --stat` against `3f03d9c36`, scoped to the three production files `internal/spec/syncsha.go`, `internal/spec/lint_syncsha.go`, `internal/spec/closer.go` | **empty output** — none of the three appears in the diff | `0` |
| S-2 | Delivery scope | `diff --stat` against `3f03d9c36` for `HEAD` | `28 files changed, 1802 insertions(+), 3 deletions(-)`; the only `.go` path is `internal/spec/lint_syncsha_test.go` (`66 ++++-`), the rest being the three fixture directories, the four SPEC artifacts, and `.moai/reports/t380/**` | `0` |
| S-3 | Criterion count (B12 self-test) | `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-SYNCSHA-BAND-BOUNDARY-001/acceptance.md` piped through `sort -u` and `wc -l` | `10` — of which **8** are this SPEC's own (`AC-SBB-001` … `AC-SBB-008`, contiguous) and 2 are cross-references to the completed t299 SPEC (`AC-SSF-001`, `AC-SSF-007`). The count written into the CHANGELOG is the SPEC-scoped **8**, from the same pipeline anchored on `AC-SBB-[0-9]+`. Non-zero on both counters, so neither is the vacuous `0 == 0` shape | `0` |
| S-4 | No `[RETIRED]` / `[REF]` markers to adjudicate | `grep -cE` for those two tokens over `acceptance.md` | `0` — every identifier occurrence is live, so the B12 ambiguity halt does not fire | `1` (no match) |
| S-5 | No duplicate CHANGELOG entry (B12 pre-emission grep) | `grep -c 'SPEC-SYNCSHA-BAND-BOUNDARY-001' CHANGELOG.md` | `0` before the write — no parallel BATCH-SYNC session had already emitted for this SPEC | `1` (no match) |
| S-6 | Evidence files exist and are tracked | `find .moai/reports/t380/verify-run -maxdepth 1 -mindepth 1` piped to `wc -l` | `11`, matching §E.2's count. (`ls` is aliased in this environment and miscounts; `find` is the counter used) | `0` |
| S-7 | This SPEC produces no lint findings of its own | `./bin/moai spec lint` (redirected; 1,101 lines) then `grep -c 'SPEC-SYNCSHA-BAND-BOUNDARY-001'` over that output | total `0 error(s), 1096 warning(s)`; **`0`** occurrences of this SPEC id anywhere in the output | `0` / `1` (no match) |
| S-8 | Sibling artifacts carry no frontmatter to transition | `head -1` on `plan.md`, `acceptance.md`, `progress.md` | each begins with its `#` H1 heading, not `---`; the status axis lives in `spec.md` alone, per `spec-frontmatter-schema.md` § Artifact Statelessness | `0` |
| S-9 | The lint-discovery citation used in the CHANGELOG is real | `sed -n '325,340p' internal/spec/lint.go` | `discoverSPECs` is declared at `internal/spec/lint.go:333` (doc comment at `:332`) and matches `SPEC-*/spec.md` under `.moai/specs/` or directly under the base dir — the fixture dirs are `sha-*` under `internal/spec/testdata/syncsha/` and are therefore unreachable to it | `0` |

Frontmatter transition applied to `spec.md` in this commit: `status: in-progress`
→ `status: completed`, `version: "0.2.0"` → `"0.3.0"`, `updated: 2026-08-31`
(already today's date, unchanged). No body section of `spec.md`, `plan.md`, or
`acceptance.md` was touched — the ownership boundary of
`spec-frontmatter-schema.md` § Forbidden ownership crossings holds.

### Baseline-attribution

- **S-1 … S-9 above** — measured by this author, at sync phase, on this tree at
  `6672fd40f`. S-7's total of `1096` is this author's own run; §E.2 reports the
  same figure measured independently at `f28a7424f` by manager-develop, and the
  two agreeing is a corroboration, not a carry-over.
- **The mutation table (R-1..R-4), the AC PASS/FAIL matrix, the close-out
  measurements (`go test ./internal/spec/... -count=1` → `ok … 39.898s` exit `0`,
  `go vet` clean, `GOOS=windows` build and vet clean), and the test-first RED
  output** — measured by manager-develop at run phase, tree pin `f28a7424f`,
  recorded in §E.2 with per-cell commands and exit codes and exported to
  `.moai/reports/t380/verify-run/`. **Not re-measured at sync phase.**
- **Plan audit verdict PASS-WITH-DEBT `0.9216`** (Tier S threshold 0.75), the
  four-mutant independent reproduction, the alphabet-residual survival
  (`^[0-9a-f]{7,40}$`, `ok … 3.082s` exit `0`) and the anchor-dropping bonus —
  attributed to `.moai/reports/t380/plan-audit.md`. **Not measured by this
  author.**
- **`go test ./internal/spec/... -count=1` GREEN at exit `0` in 34.439s under the
  surviving mutant** — quoted from SPEC-SYNC-SHA-SLOT-FORMAT-001 §E.4 (card
  t299), where it originated. Not re-measured at any phase of this card.

### Gaps — what was explicitly NOT observed at sync phase

- **No CI verdict.** The branch `WT-shared-predicate-guard` is unpushed and no
  pull request exists, so no check has run against any of these four commits.
  The lead owns integration; this lane does not push.
- **No full-tree test run.** `go test ./...` was not run at sync phase (nor at
  run phase, by dispatch instruction). Verification was scoped to
  `./internal/spec/...`, and that scoping is a deliberate constraint rather than
  an omission — but it means the whole-tree verdict is unobserved here.
- **No coverage measurement**, at either phase. The delivery adds no production
  statements, so a package coverage figure would move for reasons unrelated to
  this card.
- **No `golangci-lint` run**, at either phase. Only `go vet` and `gofmt -l`.
- **No before/after `moai spec lint` delta.** No lint run was ever taken on the
  base tree `3f03d9c36`, so no delta exists to report. What exists is the
  after-total (`1096`) plus §E.2's two zero-occurrence checks (`testdata` → `0`,
  `SSFA` → `0`), which settle `plan.md` §G's prediction without a delta and are
  not the same thing as one.
- **No codemaps regeneration**, and none attempted.
- **The write-side gate is unmeasured.** `needsSHABackfill`
  (`internal/spec/closer.go:421`) shares the predicate and was neither mutated
  nor exercised at any phase of this card.
- **The alphabet clause and the anchor-dropping mutant were not re-measured**
  under this author's hands; both stay attributed to the plan audit.
- **`sync_commit_sha` below is a placeholder at write time.** By construction a
  commit cannot name its own hash; the real SHA is backfilled in the immediately
  following commit, per the SHA placeholder backfill exemption in
  `spec-frontmatter-schema.md`.

### Residual-risk — what could still be wrong despite what was observed

- **`TestSyncSHASlot_SilentOnSHA` passes vacuously on a deleted fixture.**
  Measured by manager-develop during the M4 RED run: with `sha-min7` absent,
  `syncSHAFindings` returned zero findings without error and the function stayed
  green, so any of its case directories could be deleted and the suite would not
  notice — the empty-sweep shape of `verification-completeness.md` §1.1. This is
  **pre-existing** (it applies equally to the four fixtures t299 registered) and
  was **not repaired**, by operator ruling: repairing it means changing an
  existing function's assertions, outside this card's approved scope. The
  consequence to hold: the real guarantee for the two band-INSIDE cases
  (`sha-min7`, `sha-full`) is the R-1 / R-2 mutation observation, **not** the
  green run. **Follow-up card candidate.**
- **§E.1's closing line still reads `Status: draft`.** §E.1 is
  plan-phase-authored and manager-spec owns it; by operator ruling it is left
  as-is. It is a **plan-phase record of what was true when that section was
  written**, not a current-state claim — the `spec.md` frontmatter (`completed`)
  is the authoritative status axis, per `spec-frontmatter-schema.md`. A reader
  comparing the two is looking at two different points in time, not at a
  contradiction.
- **The alphabet clause of the grammar stays uncovered at the rule level.** An
  alphabet-narrowing mutant survives the delivered fixture set because every
  fixture value is lowercase (plan audit, `spec.md` §C). This card closes the
  length-band slice and no other slice.
- **The mutants were planted at the call site only.** None altered
  `commitSHATokenPattern` itself; a change to the pattern would be caught by
  `TestIsCommitSHAToken_LengthBand` at the predicate level, a layer this card
  measured not at all.
- **Fixture era classification is asserted per run, not pinned.** The fixtures
  rely on era heuristic H-4 continuing to hold. The criteria assert
  `Advisory == false`, so a demotion surfaces as a failure rather than a silent
  pass — but a change to `era.go`'s heuristics would break these fixtures for a
  reason having nothing to do with the band.
- **AC-SSF-007's CI-automation gap is still open.** It is a **different** gap
  from the fixture-corpus gap this card closes, and closing this card neither
  satisfies nor supersedes it.

```yaml
sync_complete_at: 2026-08-31
sync_commit_sha: 7f8b607dd
sync_status: complete
b12_self_test_a: pass          # pre-emission grep -c on CHANGELOG.md -> 0 (no duplicate entry)
b12_self_test_b: pass          # AC count 8 (AC-SBB-*, SPEC-scoped) / 10 (incl. 2 cross-SPEC refs); non-zero, not vacuous
b12_self_test_c: pass          # every path claimed in the CHANGELOG entry verified present in this tree
changelog_entry_position: "[Unreleased] -> Added, first entry"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed (merged into this sync commit)"
  plan_md: not-applicable      # no frontmatter block (artifact statelessness)
  acceptance_md: not-applicable
  progress_md: not-applicable  # phase state recorded in body sections, not frontmatter
  version_bump: "0.2.0 -> 0.3.0"
  updated_field: 2026-08-31
canary_compliance_check: not-applicable   # this SPEC defines no forward-looking policy that its own sync tests
ac_pass_count: 8
ac_fail_count: 0
production_source_diff_vs_base: empty     # base 3f03d9c36, scoped to syncsha.go + lint_syncsha.go + closer.go
spec_lint_total: "0 error(s), 1096 warning(s)"
spec_lint_findings_for_this_spec: 0
evidence_dir: .moai/reports/t380/verify-run   # 11 files, tracked
ci_verdict: not-observed                  # branch unpushed; the lead owns integration
```
