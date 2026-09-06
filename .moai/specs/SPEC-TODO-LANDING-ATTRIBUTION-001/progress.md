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

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

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
