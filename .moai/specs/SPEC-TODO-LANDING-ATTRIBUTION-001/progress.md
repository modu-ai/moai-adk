# Progress — SPEC-TODO-LANDING-ATTRIBUTION-001

Card t472. Worktree `.claude/worktrees/t472`, branch `WT-landed-drift-detect`.

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored at tree `4bcac7079`: `spec.md`, `plan.md`, `acceptance.md`, this file.
- **Version 0.2.0 — plan-audit iteration-1 remediation**, re-measured at HEAD `e227871b4` against
  `origin/develop` `7835148d3`. D1 (BLOCKING) closed by repairing §A.4 form 3 from an occurrence test
  into two positional shapes (3a, 3b) plus a non-attribution rule for absorb-direction merges, with
  `MUT-MERGE-ANY-TOKEN` added to the mutant set and AC-TLA-003 gaining the falsifying third clause.
  D2/D3/D4/D5/D6/D7/D8/D9 dispositions recorded in the artifacts. Counts unchanged: 12 REQ, 12 AC
  (Tier M ceiling 16/16).
- **Version 0.3.0 — plan-audit iteration-2 remediation**, measured at HEAD `75e63d6f2` against the
  pinned corpus commit `7835148d3` with a binary built from this tree
  (`go build -o <scratch>/moai-spec3 ./cmd/moai`, rc=0 — not the installed build, ~190 commits
  behind; VCI §2.2). Dispositions:
  - **iter-2 D1 (major) — ACCEPTED.** Form 3b had no falsifier (`grep -c t412 acceptance.md` → 0)
    and AC-TLA-005 claimed a protection that did not exist. AC-TLA-003 gains clauses 4-5 on the
    `t412` fixture set; AC-TLA-005 now carries a per-form falsifier map with the measured
    "ids this form alone attributes" column. Auditing that column found form **3a** was equally
    unfalsified — closed with `t244` (AC-TLA-003b). §A.4 form 3b restricted to a single card token
    in the group, removing the contradiction with §A.4's own preamble.
  - **iter-2 D2 (major) — ACCEPTED, and the corrected figure is larger than the audit's.** "Exactly
    one under-count (t250)" is withdrawn; measured **≥ 19** across four named shapes. The audit
    measured 7 and missed the largest family (`merge: <card>`, 31 subjects). Form 3c adopted →
    residual **7**. `plan.md` §D re-argues the tolerance at 7 with the loud-vs-silent trade stated.
  - **iter-2 D3 (major) — SPLIT: window claim REFUTED on measurement, derivation defect ACCEPTED.**
    Measured on `origin/main` `7ad9f8534`: 0 of 101 merge subjects target `develop` or `main` with a
    card-bearing trailing group, so in the M1-only window a hardcoded-`develop` implementation and a
    ref-derived one are behaviourally identical — the window does not distinguish them, and the
    audit's downstream inference (which it recorded as its own Gap) does not hold here. The
    permanent downstream defect is real and is closed by REQ-TLA-013 + AC-TLA-003 clause 6. The
    [HARD] M1→M2 ordering is untouched.
  - **iter-2 D4 (minor) — ACCEPTED, with wider grounding.** The `WT-` prefix rule is contraposed to
    a target-mismatch rule. The audit measured 6 non-`WT-` worktree targets; re-measured here there
    are **9**, three carrying no prefix at all (`t403`, `t78`, `t86`).
  - **iter-2 D5 (minor) — ACCEPTED.** The diff-stat figure re-measured 1569 at this HEAD (1101 →
    1301 → 1569 across three trees) and is now in R4 form: command first, value parenthesized and
    dated as a reference.
- **Version 0.4.0 — plan-audit iteration-3 remediation, D3-1 ONLY (operator decision).** Iteration 3
  returned PASS-WITH-DEBT **0.8375**, down 0.0375 from iteration 2, which fired `spec-workflow.md:160`'s
  score-regression STOP condition alongside the three-iteration ceiling. The operator scoped this
  remediation to D3-1 and left D3-2 through D3-5 as accepted debt. Measured at HEAD `012d9680a` over the
  pinned corpus `7835148d3` with a binary built from this tree
  (`go build -o <scratch>/moai-spec4 ./cmd/moai`, rc=0 — not the installed build, ~190 commits behind;
  VCI §2.2).
  - **iter-3 D3-1 (major, blocking) — ACCEPTED.** Form 2's `)$` anchor silently excluded its own
    position with a pull-request reference group appended. Reproduced independently here: `43`
    subjects, `40` distinct ids, `39` attributed by no other form. **Form 2b** adopted (single card
    token in the card-bearing group, exactly one trailing `(#NNNN)` reference group). Attributed set
    `270 → 309`; unattributed-but-subject-present `77 → 38`. AC-TLA-002 gains clause 3 with the `t210`
    fixture and the new mutant `MUT-PAREN-END-ANCHOR`; the AC-TLA-005 map gains form 2b's row (39) and
    form 1's exclusive column is corrected `42 → 41` (`t230` is now shared). `plan.md` §D's tolerance
    ground is **withdrawn**, not repaired — see the debt list below. §F Definition of Done gains a
    form-2b positive control, because version 0.3.0's controls (`t401`, `t440`) were both shapes the
    enumeration already handled.

### Debt this SPEC enters run-phase with

**[HARD] No fourth plan-audit will verify the 0.4.0 change.** The audit budget is spent (three
iterations, score regressing). A reader who needs to check the D3-1 fix runs these two commands
against the pinned corpus `7835148d3` instead — they are the whole verification surface for it:

    grep -cE '\([^()]*t[0-9]+[^()]*\) \(#[0-9]+\)$' <pinned-corpus subjects>   → 43
    ... ids extracted, sort -u, comm -23 against the five-form attributed set  → 39

Five items are carried into run-phase unfixed and are recorded here so no reader mistakes silence
for absence:

1. **D3-2 — `t311` unclassified.** Its sole subject occurrence is
   `merge(WT-codex-init): integrate card t340 … (closes t311)`. `acceptance.md` §D claims the
   multi-card trailing groups were ruled on "id by id"; `t311` is ruled on nowhere. It is neither
   confirmed as a further under-count nor as a correctly-unattributed note.
2. **D3-3 — the merge denominator is wrong.** `spec.md` §A.4's preamble says "414 of them merges".
   414 is the count of subjects *beginning* `Merge`; the actual merge-commit count at that corpus is
   **661**. 247 merge commits — all 31 form-3c subjects among them — sit outside the denominator the
   merge-shape survey is described against.
3. **D3-4 — the "nine targets" list prints eleven.** Two entries
   (`worktree-agent-a205e7a01ec2e0f27`, `worktree-agent-a350b7a40faaf39c6`) are merge **sources**, not
   targets, and do not reproduce under the extraction command quoted beside the list. The rule the
   list illustrates is unaffected; the list is.
4. **D3-5 — `plan.md` §D overstates the sixth-form rejection rule.** It reads "reject it if that count
   is 0", which is stronger than this SPEC's own practice: REQ-TLA-013's row reads `n/a` and is
   falsified by a constructed `release/v9` fixture. A zero column means "no *corpus* falsifier
   exists", not "no falsifier exists".
5. **The standing residual — classified only to a floor.** The named-shape under-count is **at least
   10** and the subject-present-but-unattributed population is **38**. Neither number is a total:
   nobody — author, auditor, or lane — has classified the 38 exhaustively, and every round so far has
   raised the figure (1 → 19 → 7 → ≥10). The **live-queue status of the newly-surfaced ids is
   unmeasured** by this author. The lane's committed evidence
   (`.moai/reports/t472/axis-bf-measurement.md`, iter-3 section) records that 38 of them appear in
   neither table of the disk store — a reading consistent with the known `moai todo` / disk-store
   split, and therefore establishing nothing either way about whether those cards are live.

- Tier M. **13** requirements (REQ-TLA-001..013), **13** unique acceptance-criterion identifiers
  (AC-TLA-001..012 plus AC-TLA-003b, paired to AC-TLA-003 per the AC sub-ID convention). Ceiling
  16/16 — both in budget. Version 0.4.0 added no REQ and no AC identifier: the new form is covered by
  REQ-TLA-001/002/004 and falsified by a third clause absorbed into AC-TLA-002.
- Milestone order fixed [HARD]: M1 (axis F, the attribution predicate) before M2 (axes A+B, the ref
  chain and its disclosure) — ground in `spec.md` §A.7 and `plan.md` §A.3.
- Ref-chain option **A3** selected; A1, A2, A4 rejected with recorded grounds in `plan.md` §A.1.
- Evidence base: `.moai/reports/t472/premise-recheck.md` and `.moai/reports/t472/axis-bf-measurement.md`,
  both committed in this tree by this lane; fixtures re-run at `4bcac7079` before citation.
- `status: draft`. Run-phase not entered; no implementation code written.

**Lint evidence.** `moai spec lint` and `moai spec lint --strict`, scoped to this SPEC and run with a
binary built from this tree (`go build ./cmd/moai`, rc=0 — not the installed build, which is 190
commits behind; VCI §2.2), both returned `✓ No findings — all SPEC documents are valid`. The lane
observed both; the parent lane session independently re-ran the `--strict` form from the worktree
root with the same binary and observed the same result, so the pass is confirmed rather than reported.

**Gap — whole-corpus lint not observed, and deliberately not pursued.** Three attempts at a
repository-wide `moai spec lint` produced no usable measurement: the first was piped through
`tail -40`, discarding all but the last 40 lines and leaving the result unattributable, and was in any
case taken before the `(maps REQ-…)` clauses were added; the second and third were killed by the parent
lane session (`rc=143`, then exit `144`) because this machine's load average was 27 while
lanes share it, and each left a 0-byte output file. Per lane direction the item is **closed as a gap,
not resolved** — no further attempt is to be made from this lane. A grep over those empty files
returns zero for any pattern and establishes nothing.

The residual this would have covered is bounded by an **argument, not a measurement**: `git status
--short` returned exactly `?? .moai/specs/SPEC-TODO-LANDING-ATTRIBUTION-001/` — four new files in a new
directory, no existing file modified — so no mechanism is apparent by which this change could alter
another SPEC's findings. That reasoning is recorded as an argument and must not be cited as a corpus
measurement.

**Stale, and kept rather than deleted (plan-audit D9b).** The `git status --short` reading above was
true when written, at tree `4bcac7079`, before the SPEC directory was committed. It was committed at
`62cbfdf77`, so the command now returns empty and the argument is no longer reproducible as stated.
The argument's *substance* survives, and it is re-measured at read time rather than quoted.
**Run this, do not read the number below as current** (VCI §2.1 remedy R4 — `HEAD` is a moving
coordinate, so the command is the criterion and any value is a dated reference):

    git diff --stat 4bcac7079 HEAD | tail -1
    git diff --name-only 4bcac7079 HEAD

*Reference values, measured 2026-09-04 at HEAD `75e63d6f2`:* `6 files changed, 1569 insertions(+)`,
and the name list is exactly the four SPEC artifacts plus the two evidence files under
`.moai/reports/t472/`. **The figure has now drifted twice** — 1101 recorded at `165d69d14`'s
predecessor, 1301 read by the plan-audit at `165d69d14`, 1569 here — which is the whole reason it is
demoted to a dated reference rather than restated as a fact (plan-audit iter-2 D5: the fix for a
stale-figure defect had reproduced that defect in miniature by leading with the value).

What the command establishes, and what is load-bearing, is the **name list, not the insertion
count**: no file outside this SPEC's directory and this card's report directory is touched, and no
existing line is modified anywhere, so no mechanism alters another SPEC's findings. The present-tense
`git status` output above is stale and is marked so rather than quietly rewritten. The gap itself
remains closed as a gap; no further whole-corpus lint attempt is made from this lane.

## §E.2 Run-phase Evidence

Run by manager-develop (cycle_type=tdd), 2026-09-06, in worktree `.claude/worktrees/t472`, branch
`WT-landed-drift-detect`, starting at HEAD `c43c07c3d`. Implementation commits: M1 `2b07aa010`
→ M2 `00148e239` → M3 `c779a0a15`, in that order ([HARD] ordering held; no combined M1+M2 commit).
No push, no PR, no branch creation (repo-local override, dispatch B9).

### E.2.1 Open design decision (plan.md §A.2) — RESOLVED: Go-side subject filter

The six shapes are expressed as a Go-side positional matcher over `git log <ref> --format=%s`,
NOT as a widened `--grep`. Grounds, all measured or structural:

1. Forms 2b/3b require "exactly one DISTINCT card token in the group" — a set property (count of
   distinct tokens), not a pure regex property. Form 3b's target comparison is against the branch
   DERIVED from the resolved ref (REQ-TLA-013) — a comparison, not a pattern.
2. git has no subject-only `--grep`; a widened whole-message pattern re-admits bodies.
3. A regex's failure mode is silent (unmatched pattern ≡ not-landed — the hazard
   `prlink_landed.go`'s own header was written about). A Go-side matcher is inspectable and
   unit-testable per form.

REQ-TLA-005 holds: ONE exported argv builder, `LandedSubjectArgs(ref)`; the tripwire
(`TestLandedSubjectArgs_Tripwire`) calls that symbol and asserts its output.

**Non-attribution rule keying — run-phase resolution, measured.** The rule ("a merge subject that
names a target attributes no card unless the target is the branch the resolved landed ref names")
keys on git's merge-COMMIT subject spelling (capital-`Merge`). Keying on lowercase `merge:`
subjects too would lose **t78** — `merge: wire audit_multi cross-model convergence into /moai
review verdict step (t78)` — whose " into " is prose and whose sole attribution is form 2's
trailing group: measured, the lowercase-inclusive gate scores **308** attributed / **39**
unattributed over the pinned corpus, against the contract's 309/38. The capital-only gate
reproduces 309/38 exactly AND satisfies AC-TLA-003 clause 3 (t284 not attributed by the absorb
merge `c4ae1ecbd`). Both variants were measured with throwaway transcriptions before the
implementation was written; the enumeration lives in `landedSubjectForms` + `subjectAttribution`
(one named place, REQ-TLA-002), with the rule's keying documented there.

### E.2.2 Corpus reproduction (constraint D8) — REPRODUCED EXACTLY

Reference (Python transcription, `.moai/reports/t482/forms.py`), run from the worktree root:

    $ git log 7835148d3 --format=%s | python3 .moai/reports/t482/forms.py
    subjects                              : 5837
    ids appearing anywhere in a subject   : 347
    ids attributed (forms 1/2/2b/3a/3b/3c): 309
    subject-present but unattributed      : 38
    per-form subject counts               : {'1': 290, '2': 869, '2b': 43, '3a': 5, '3b': 77, '3c': 31}

    RESIDUAL: t2 t21 t40 t46 t68 t73 t74 t80 t94 t121 t123 t124 t128 t129 t131 t132 t133 t134 t135
    t137 t139 t141 t142 t143 t144 t147 t148 t149 t155 t157 t158 t216 t225 t250 t311 t409 t443 t460

Implementation (Go predicate over the same subject stream), committed as
`TestLandedPredicate_PinnedCorpusReproduction` (skips honestly when the pinned commit is absent —
a shallow clone yields a skip, never a vacuous green):

    $ go test ./internal/kanban/ -run TestLandedPredicate_PinnedCorpusReproduction -v
    === RUN   TestLandedPredicate_PinnedCorpusReproduction
    --- PASS: TestLandedPredicate_PinnedCorpusReproduction (0.17s)

The test asserts 347 mentioned / 309 attributed / 38 residual AND pins the residual id SET (the
38 ids above, embedded) — the partition, not merely the counts. **No divergence; no id list to
report.** The attributed set is also invariant to single-attribution-per-subject semantics (the
predicate answers for ONE queried card) versus the reference's union semantics — measured equal
before implementation, via a throwaway transcription of both semantics.

### E.2.3 RED evidence (TDD; acceptance.md §E gate)

Pre-GREEN, against the tree as received (whose shipped predicate IS MUT-WHOLE-MESSAGE, whose
resolver IS MUT-CONFIG-ONLY, whose silent fallback IS MUT-SILENT-FALLBACK):

    $ go test ./internal/kanban/ -run 'TestLandedPredicate|TestLandedRefFor_Chain'
    --- FAIL: TestLandedPredicate_Form1_ConventionalScope (0.27s)
        prlink_landed_forms_test.go:96: t237 = "landed", want "not-landed" — a body mention is not an attribution
    --- FAIL: TestLandedPredicate_Form2_TrailingParenthetical (0.28s)
        prlink_landed_forms_test.go:120: t443 = "landed", want "not-landed" — a mid-subject mention attributes nothing (MUT-SUBJECT-ONLY)
    --- FAIL: TestLandedPredicate_Form3_Merges (0.95s)
        --- FAIL: .../clause_1_and_2: t216 = "landed", want "not-landed"
        --- FAIL: .../clause_3: t386 = "landed", want "not-landed" — MUT-MERGE-ANY-TOKEN
                          t387 = "landed", want "not-landed"
                          t284 = "landed", want "not-landed"
        --- FAIL: .../clause_5: t412 = "landed", want "not-landed" — MUT-NO-TARGET-TEST
        --- FAIL: .../clause_6: t901 = "landed", want "not-landed" — MUT-HARDCODED-DEVELOP fixture
    --- FAIL: TestLandedPredicate_Form3a_3c_CardLedMerges (0.36s)
        prlink_landed_forms_test.go:237: t80 = "landed", want "not-landed" — a branch name inside a group attributes nothing
    --- FAIL: TestLandedPredicate_BodyMentionNeverAttributes (0.39s)
        t555 = "landed", want "not-landed" ; t777 = "landed", want "not-landed"
    --- FAIL: TestLandedRefFor_ChainLevel2_OriginHEAD (0.02s → fixed fixture, then):
        prlink_landedref_chain_test.go:53: LandedRefFor = "origin/main", want "origin/develop" — level 2 reads the repository's own recorded default (MUT-CONFIG-ONLY)

    $ go test ./internal/cli/ -run 'TestTodoDone_VerdictNamesRef|TestTodoDone_DisclosesLevel3|TestTodoDone_NoDisclosureAtLevel1|TestTodoDone_WithoutFlagNoRefNamed'
    --- FAIL: TestTodoDone_VerdictNamesRefAndDisclosesLevel2 (0.28s)
        todo_done_ref_test.go:67: stdout = "done t1 landing=landed\n", want the answering ref named (AC-TLA-010)
        todo_done_ref_test.go:70: stderr = "", want the answering chain level disclosed (AC-TLA-011)
    --- FAIL: TestTodoDone_DisclosesLevel3 (0.21s)
        stdout = "done t1 landing=landed\n", want the default ref named as the answerer ; stderr = ""
    --- FAIL: TestTodoDone_NoDisclosureAtLevel1 (0.20s)
        stdout = "done t1 landing=landed\n", want the configured ref named

(Verbatim lines preserved above except where ellipsed for width; every elided line is a repeat of
a shown line for a sibling id. The full raw output existed in the run transcript.)

**Honesty note on the mutant corpus.** The pre-repair tree embodies exactly ONE named mutant
(MUT-WHOLE-MESSAGE), so only its RED is genuinely pre-GREEN. The remaining mutants are
INTERMEDIATE defective implementations that do not exist on any tree before GREEN; their failures
were observed as post-GREEN **mutation probes** — the implementation was temporarily mutated to
each defective shape, the named criterion was run, the verbatim red was captured, and the mutant
was reverted (`git status` clean before and after; `git diff` empty at the end of the probe
sequence). This satisfies verification-completeness §1.1 (each instrument's failure observed on a
known failing input) and is recorded here as probes, not as pre-GREEN runs.

| Probe (mutant injected) | Criterion that went red (verbatim tail) |
|---|---|
| MUT-SUBJECT-ONLY (any token in subject attributes) | `t443 = "landed", want "not-landed" — a mid-subject mention attributes nothing (MUT-SUBJECT-ONLY)` + `t461 = "not-landed", want "landed"` + clause 3/5/6 reds |
| MUT-GROUP-ANY-TOKEN + MUT-NO-TARGET-TEST (3b group read as occurrence set; target test dropped) | `t387 = "landed" ... (MUT-MERGE-ANY-TOKEN)` ; `t284 = "landed" ...` ; clause 5: `t412 = "landed", want "not-landed" — same card, wrong target ... (MUT-NO-TARGET-TEST)` |
| MUT-NO-FORM-2B / MUT-PAREN-END-ANCHOR (form 2b disabled) | `t210 = "not-landed", want "landed" — form 2b: a card group before one reference group (MUT-NO-FORM-2B)` |
| MUT-HARDCODED-DEVELOP (literal `develop` in 3b + rule) | clause 6: `t900 = "not-landed", want "landed"` ; `t901 = "landed", want "not-landed"` — both directions |
| MUT-NO-FORM-3A + MUT-NO-FORM-3C (card-led shapes disabled) | `t244 = "not-landed", want "landed" ... (MUT-NO-FORM-3A)` ; `t79 = "not-landed", want "landed" ... (MUT-NO-FORM-3C)` |
| MUT-FIRST-TOKEN-OF-GROUP (form 2 widened to first token of any trailing group) | `t80 = "landed", want "not-landed" — a branch name inside a group attributes nothing` — the silent false positive §A.4.2 declines (also mutated form-2's anchor shape: `t401 = "not-landed"`, `t461 = "not-landed"`) |

All probes reverted; final tree byte-identical to the M3 commit (`git status --short` → only the
pre-existing untracked `.moai/reports/t472/run-handoff-2026-09-06.md`).

### E.2.4 Definition-of-Done checks (acceptance.md §F)

Live `todo pr` re-run in this tree (binary built from THIS tree, `go build -o /tmp/moai-t472-verify
./cmd/moai`; the verb is read-only and wrote nothing):

    $ /tmp/moai-t472-verify todo pr --json   (filtered to the DoD ids)
    t237 no-link        (was the measured false positive — no longer reads landed)
    t312 no-link        (was the second origin/main false positive)
    t401 landed         (true positive survives — against origin/develop, which the
                         M2 chain now selects via refs/remotes/origin/HEAD)
    rows_total 57

t440, t210, t216, t443 are absent from the live queue (57 rows scanned), so their DoD evidence is
carried by the pinned-corpus test (neither t440 nor t210 is in the pinned residual — t210's
form-2b positive control holds corpus-wide) and by the fixture tests that assert t440 and t210 in
both directions. t440's `docs(t440): …` subject is form 1's own example in spec.md §A.4.

### E.2.5 Verification batch (attributable; HEAD `c779a0a15` unless noted)

- **Full scoped suites** (this tree; NOT a local full-repo run — CI's verdict is the full suite):

      $ go test ./internal/kanban/   → ok  github.com/modu-ai/moai-adk/internal/kanban  139.796s
      $ go test ./internal/cli/      → ok  github.com/modu-ai/moai-adk/internal/cli     309.916s

- **Coverage** (E3): `go test -cover ./internal/kanban/` → `coverage: 86.7% of statements` — ≥ 85%
  package target. (cli package not re-measured with -cover; its suite is 310s and the dispatch
  scoped E3 to kanban.)
- **Cross-platform** (E2): `go build ./...` rc=0; `GOOS=windows GOARCH=amd64 go build ./...` rc=0
  (both re-run after M1 and after M2; final at M3 tree).
- **Lint** (E5): `golangci-lint run --timeout=2m ./internal/kanban/... ./internal/cli/...` →
  `0 issues.` — identical to the pre-flight baseline (`0 issues.`), so **zero new lint issues**.
- **Subagent boundary** (E4): `grep -rn 'AskUserQuestion' internal/kanban internal/cli/todo.go
  internal/cli/todo_pr.go | grep -v _test.go | grep -v '// '` → empty (rc=1).
- **go vet**: clean on both touched packages.
- **Push state** (E6): three commits on `WT-landed-drift-detect`, NOT pushed — by design (repo-local
  override; integration is the lead's develop-merge window).
- **Pre-existing fixture update, SPEC-justified**: three CLI test stubs keyed landing answers on
  the card id inside the query argv. The subject-stream query no longer carries the id (the
  predicate matches stream-side — the defect being the argv's per-card grep), so the stubs plan
  answers per call; one stub subject (`abc1234 landed t2`) was rewritten to an attributing
  position — its old mid-subject mention is precisely what M1 stopped reading as a landing. All
  other cli tests pass unmodified.

### E.2.6 Debt and scope discipline

- D3-2 (t311), D3-3 (merge denominator 661), D3-4 (nine-targets list), D3-5 (plan §D wording), and
  the standing residual floor remain debt, untouched. The corpus test PINS the residual as
  measured; it does not classify it.
- SPEC body content of spec.md/plan.md/acceptance.md: unmodified. Frontmatter writes limited to
  spec.md `status: draft→in-progress` + `updated: 2026-09-06` on the M1 commit. progress.md §E.2/§E.3
  authored here; §E.4 untouched (manager-docs).
- PRESERVE held: internal/statusline/** untouched (its `LandedRefFor` call site inherits the chain
  through the shared resolver — a behavior change of the resolver, not of its code);
  .moai/reports/t472/** untouched except the pre-existing untracked handoff note;
  .moai/reports/t482/** only read (forms.py executed; it writes residual.json under its own
  directory); template trees untouched. Files changed overall: 3 implementation files
  (prlink_landed.go, prlink_landedref.go, todo.go) + 1 doc string (todo_pr.go help body) + 8 test
  files.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-06
run_commit_sha: c779a0a15   # M3, the last implementation commit; the evidence commit carrying this section follows it
run_status: complete
ac_pass_count: 13           # AC-TLA-001..012 + AC-TLA-003b
ac_fail_count: 0
preserve_list_post_run_count: 0   # files under PRESERVE modified: none
l44_pre_commit_fetch: n/a   # repo-local override: no push path; lead performs develop integration
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0   # baseline 0 issues -> 0 issues
cross_platform_build:
  host: pass                # go build ./... rc=0
  windows_amd64: pass       # GOOS=windows GOARCH=amd64 go build ./... rc=0
total_run_phase_files: 12   # 4 impl (incl. 1 new) + 8 test (3 new, 5 updated)
m1_to_mN_commit_strategy: "M1 2b07aa010 -> M2 00148e239 -> M3 c779a0a15; strict [HARD] order held; no combined commit; no push (repo-local override)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-06
sync_commit_sha: "775f97fc1"   # the sync close commit; backfilled in the immediately following commit (a commit cannot cite its own SHA)
sync_status: complete
changelog_entry: none   # repo convention carries NO card-level CHANGELOG entries (release-time surface); grep count for this SPEC-ID in CHANGELOG.md = 0, deliberately kept at 0
docs_sync: docs-site 4-locale moai-todo.md — verdict-line format (ref suffix + stderr level disclosure) and the attribution-position predicate description updated in ko/en/ja/zh in one commit
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (on the sync commit; the 3-phase close)"
  updated: "2026-09-06"
  plan_acceptance_frontmatter: none — both artifacts are stateless on the status axis per spec-frontmatter-schema.md (no status field present, none added)
b12_self_test:
  changelog_preemission_grep: 0   # grep -c 'SPEC-TODO-LANDING-ATTRIBUTION-001' CHANGELOG.md -> 0; emission correctly skipped
  acceptance_ac_count: 13         # AC-TLA-001..012 + AC-TLA-003b (sub-ID convention; the generic numeric-only grep pattern reports 12 and misses the 003b sub-ID)
  changelog_ac_match: n/a         # no CHANGELOG entry emitted (repo convention), so no AC-count cross-check applies
canary_compliance_check:
  body_edits: 0   # spec.md/plan.md/acceptance.md body untouched; frontmatter status/updated only
  preserve: internal/**, templates, .moai/reports/t472/**, .moai/reports/t482/** untouched
```

## §F Phase 4 Mode Selection

Logged by lane-8 (card t472, lead-1 dispatch) before the first run-phase `Agent()` spawn.

**Kickoff approval: PASSED.** Granted by the operator during the plan phase but undeliverable —
the owning lane-3 session was gone and the lead's dispatch to it failed unreachable. Re-delivered
via the lead's t472 dispatch (2026-09-06). The operator verdict riding the same gate: remediation
scope was D3-1 only (already closed as form 2b in v0.4.0), the five recorded debts stay debt, and
there is **no fourth plan-audit round**.

**Phase 1 Plan Audit Gate: skip taken.** The three skip conditions, all satisfied:
1. Final-iteration verdict is **PASS-WITH-DEBT 0.8375** (iteration 3 of 3; the score-regression
   STOP fired and the operator ruled no fourth round — the explicit-override path of the retry
   contract).
2. Score 0.8375 ≥ the Tier M threshold 0.80.
3. **Plan-artifact hash unchanged since the verdict** — the develop absorb merge (develop
   `3084f1071` → HEAD `c43c07c3d`) touched no SPEC artifact; the measured explosion radius is
   exactly 7 files: this SPEC's 4 artifacts plus the 3 report files under `.moai/reports/t472/`,
   none modified by the absorb.

| Input | Value |
|---|---|
| tier | M (13 REQ / 13 AC) |
| scope (files) | 3 Go source files + their tests: `internal/kanban/prlink_landed.go`, `internal/cli/todo.go`, `internal/cli/todo_pr.go` |
| domain count | 1 (Go backend CLI) |
| file language mix | Go |
| concurrency benefit | LOW (coding-heavy) |
| agent-team prereqs | not requested |

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | semantic multi-file change, not a typo-scale fix |
| **serial** | **YES** | coding-heavy (Anthropic parallelism caveat); single domain, 3 files; one manager-develop carries M1→M2→M3 with per-milestone commits under the [HARD] M1-before-M2 ordering |
| fanout | no | no independent multi-domain research surface |
| sweep | no | not a ≥30-file mechanical-uniform transform |

**Decision: serial**

Justification: the acceptance criteria are per-form unit tests over one package boundary, so a
single coherent author beats fan-out reconciliation; the [HARD] milestone order (M1 predicate
before M2 ref chain, `spec.md` §A.7) makes the work serial by construction.
