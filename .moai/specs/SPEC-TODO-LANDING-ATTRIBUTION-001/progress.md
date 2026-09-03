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
- Tier M. **13** requirements (REQ-TLA-001..013), 12 acceptance criteria (AC-TLA-001..012, with
  AC-TLA-003b paired to AC-TLA-003 per the AC sub-ID convention). Ceiling 16/16 — both in budget.
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
