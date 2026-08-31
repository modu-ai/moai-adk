# SPEC Review Report: SPEC-STATUS-TRANSITION-VALIDITY-001

Card: **t376** · Iteration: 1/2 (Tier M ceiling) · Tree of record: `3f03d9c36` (`WT-status-transition-gap`)
Auditor: plan-auditor (independent). Reasoning context ignored per M1 Context Isolation — this audit
read only the four SPEC artifacts, the four evidence files, and the source they cite.

Verdict: **PASS-WITH-DEBT**
Overall Score: **0.90** (Tier M PASS threshold 0.80)

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o '^### REQ-STV-[0-9]*' spec.md` returns
  001,002,003,004,005,006,007,008,009,**012**,010,011,013,014,015. Fifteen ids, contiguous 001-015,
  no gaps, no duplicates, uniform 3-digit padding. `012` is presented out of sequence (between 009
  and 010) — a reading-order defect (D6), not a numbering defect.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-STV-*` in
  spec.md), not against the Given-When-Then ACs, which are the correct verification-layer form. All
  15 carry an explicit pattern label and match it: Ubiquitous (001, 002, 008, 009, 010),
  Event-driven (003, 004, 013), Unwanted (005, 012, 015), Capability gate (006, 007, 014),
  State-driven (010). One blemish recorded as D5: REQ-STV-015's second clause ("shall be shown, by
  measurement rather than by assumption") specifies a verification activity, not a system behaviour.
  The leading clause is a proper `shall not` form, so MP-2 passes.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`id`, `title`, `version: "0.1.0"` quoted, `status: draft`, `created`/`updated` ISO,
  `author`, `priority: P2`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` comma-separated
  string), plus `tier: M`. No rejected snake_case alias. Independently measured, not eyeballed:
  `go run ./cmd/moai spec lint --json <spec.md>` returns `[]` — zero findings, so no
  `FrontmatterInvalid`.
- **[N/A] MP-4 language neutrality** — the SPEC is scoped to this repository's own Go lint engine
  (`internal/spec`), a single-language subject. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** —
  `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` returns exactly one id:
  `SPEC-STATUS-TRANSITION-VALIDITY-001` (itself). No external SPEC reference, so no
  retired/superseded reconciliation is owed. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. Auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' <spec dir>/` → rc=1, no
  match. `progress.md` `open_clarifications: 0` is consistent with the measurement. (`research.md`
  absent; Tier M does not require it.)

---

## Evidence verification (I re-measured rather than trusting the citations)

Every claim below was checked against the file the SPEC cites, in this worktree, at `3f03d9c36`.

| SPEC claim | Verification | Result |
|---|---|---|
| §A.1 baseline 1096 findings, 10 codes | `jq 'length'` → 1096; `jq -r '.[].code' \| sort \| uniq -c` reproduces all 10 rows exactly | **matches** |
| §A.1 "one `OwnershipTransitionInvalid` across the corpus" | same tally → `1 OwnershipTransitionInvalid` | **matches** |
| §A.1 716 dirs at baseline / 717 after | `find .moai/specs -maxdepth 1 -mindepth 1 -type d \| wc -l` → **717** now | **consistent** |
| §A.2 8-row probe table | `probe-transition-gap.log` — all 8 rows, including the two `advisory=false` entries, reproduce the table cell-for-cell | **matches** |
| §A.4 census: 714 scanned / 713 records / 1 none | `transition-census.log` line 51 states exactly that; `find .moai/specs -maxdepth 2 -name spec.md \| wc -l` → **714** | **matches** |
| §A.4 all 26 edge rows | every row and count in the table matches the log; the 26 counts sum to **713** | **matches** |
| §A.3 layer 1 — `lint_ownership.go:80-82` `draft\|planned → in-progress\|implemented` | read the lines: exactly that `case` arm | **exact** |
| §A.3 layer 1 — `:61` comment delegating to `StatusValueEnumRule`, naming status reversal | line 61 verbatim | **exact** |
| §A.3 layer 1 — `:404-407` `expected == ownerNone` early return | lines 404-407 verbatim | **exact** |
| §A.3 layer 2 — `lint.go:1105-1110` `StatusValueEnumRule.Check(doc, _ []*SPECDoc)` | signature sits at line 1110; it carries no previous status | **exact** |
| §A.3 layer 3 — `lint.go:1318-1320` terminal early return; `:1290-1295` `completed: true` | both verbatim; `terminalStatusEnum` = superseded/archived/rejected/completed | **exact** |
| §A.3 layer 4 — `lint.go:239` demote decision; `:296-311` blanket `Advisory = true` on every warning; `:61` `--strict` skips `Advisory` | all three verbatim | **exact** |
| §A.3 — `.github/workflows/spec-lint.yml:31` `actions/checkout@v7`, no `fetch-depth` | `grep -n` → line 31 is `- uses: actions/checkout@v7`; no `fetch-depth` anywhere | **exact** |
| plan.md §B — `:412-415` no-trailer guard | the `if rec.AuthoredByAgent == ""` is at **414**, its block closes at **416**; the cited range starts on a comment and stops before the brace | **off-by-two, substantively correct** (D8) |
| §A.3 layer 1 — `:78-94` for `expectedOwnerForTransition` | the function spans **64-95**; 78-94 starts mid-`case ""` arm | **imprecise range, substantively correct** (D8) |
| spec.md is itself lint-clean | `go run ./cmd/moai spec lint --json .../spec.md` → `[]` | **independently confirmed** |
| `MissingExclusions` compliance | read `OutOfScopeRule.Check` (`lint.go:1023-1056`): needs a `###`-prefixed line containing "out of scope" followed by a `-` bullet. §C supplies seven such headings, each with bullets | **compliant, mechanically** |

The card's coordinate-drift hazard is real but **did not materialise**: the substantive coordinates
are exact. The two imprecise ranges (D8) are range-boundary sloppiness, not wrong claims.

One evidence-hygiene note, non-blocking: `lint-after-spec-v2.json` is **byte-identical** to
`lint-baseline.json` (`cmp` → identical, both 373,334 bytes). That is consistent with "the new SPEC
contributes zero findings", and my independent single-file lint corroborates it — but a
byte-identical artifact is also what a stale copy looks like, so it carries no evidentiary weight on
its own. The independent run is what establishes the claim.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 1.0-band with deductions | Every normative statement has one reading; coordinates verified exact. Deducted for D1 (a decision's stated rationale contradicted by its own census) and D3 (an arithmetic figure inconsistent with its own enumeration) |
| Completeness | 0.90 | 1.0-band with deductions | HISTORY, Overview, §A survey, §B requirements, §C exclusions (7 headed subsections, lint-clean), §D pointer, §E cross-refs. Deducted for D2 (plan.md §F risk row 3 names a mitigation that no AC binds) |
| Testability | 0.80 | 0.75-band | Live control (AC-STV-010), bidirectional set, projections explicitly quarantined. Deducted for D4 (AC-STV-016 cannot fail) and D7 (AC-STV-003 as worded is unsatisfiable) |
| Traceability | 1.00 | 1.0 | All 15 REQs carry ≥1 AC (001→AC-001/002/014; 002→003; 003→001; 004→002; 005→004/005/006/007/007a/008; 006→008; 007→009; 008→011; 009→018; 010→013; 011→010; 012→012; 013→015; 014→017; 015→016). Every AC names a REQ that exists. No orphans, no uncovered REQ |

Aggregate: (0.90 + 0.90 + 0.80 + 1.00) / 4 = **0.90**.

---

## Defects Found

**D1 — spec.md §A.5 D2 — the decision's stated rationale is contradicted by the census it cites.**
Severity: **major**. Class: **blocking**.
§A.5 D2 reads: *"The census is the argument, not preference: the overwhelming majority of what this
rule flags ends in `completed`, which the layer-4 demotion already marks advisory whatever this rule
does. Omitting the emission-site flag therefore adds little CI-regression risk on the existing
corpus."* The SPEC's own §A.6 projection is 98 findings = 50 `draft → completed` + 48
`completed → implemented`. Only the first 50 end in `completed`. The other 48 end in `implemented`,
which I verified is **not** in `terminalStatusEnum` (`internal/spec/lint.go:1290-1295`:
superseded / archived / rejected / completed). So 49% of the projected findings are outside the
shelter D2 leans on — "overwhelming majority" is false against the SPEC's own numbers, and whether
those 48 demote depends on era classification alone, which is **not measured anywhere in this SPEC**.

The consequence is concrete and I measured its headroom:
`jq '[.[] | select(.advisory != true)] | length' lint-baseline.json` → **0**. Every one of the 1096
current findings is advisory, and `.github/workflows/spec-lint.yml:36` runs
`go run ./cmd/moai spec lint --strict` on every push to `main` and `develop`. Today the strict gate
has exactly zero gating findings; the first non-advisory finding this rule emits reddens the
integration branch. Gap, stated as such: I did **not** measure how many of the 48
`completed → implemented` documents are modern-era, so I assert no count — that unmeasured
population is the defect.
*Required fix*: restate D2's rationale to match the census (say "roughly half", not "overwhelming
majority"), and add the missing gate — see D2 below. The **decision** (no emission-site `Advisory`)
is sound and I am not asking for it to be reversed; its stated reason is what fails.

**D2 — plan.md §F risk row 3 names a mitigation that no AC binds.** Severity: **major**.
Class: **blocking**.
plan.md §F: *"D2's premise is measurable — M2 reports how many findings land on non-terminal,
modern-era SPECs, which is the only population that gates."* Correct, and it is the right
measurement. But AC-STV-013 requires only the **per-code** count against the §A.1 table; nothing in
the AC set requires the advisory/non-advisory split, and nothing requires a decision on the result.
Run-phase can therefore satisfy every one of the 19 ACs while never reporting the gating population,
and the card can close leaving `spec-lint` red on `develop` with no recorded decision. A risk whose
mitigation lives only in the plan's prose is a risk with no gate.
*Required fix*: extend AC-STV-013 (or add an AC against REQ-STV-009) to require the post-change
findings be split by `advisory`, the non-advisory count reported, and — if it is non-zero — an
explicit recorded decision (accept the reddening, or gate it) before close.

**D3 — spec.md §A.6 — the `StatusTokenUnrecognized` projection is inconsistent with its own
enumeration.** Severity: **minor**. Class: **blocking**.
§A.6 projects *"**6** `StatusTokenUnrecognized` (`Completed` 3, `synced`, `approved`, `Superseded`,
`cancelled` …)"*. Its own enumeration sums to **7** (3 + 1 + 1 + 1 + 1). I confirmed each token is
outside the enum by reading `internal/spec/status.go:27-37` (`ValidStatuses` = draft, planned,
in-progress, implemented, completed, superseded, archived, rejected) and confirmed the counts in
`transition-census.log` (`Completed → completed` 3; `synced → completed` 1; `approved → completed`
1; `Superseded → superseded` 1; `cancelled → rejected` 1). AC-STV-013 makes a wide projection miss "a
finding to explain before the card closes", so an off-by-one baked into the projection buys
run-phase an investigation of the SPEC's own arithmetic.
*Required fix*: change ~6 to ~7 in §A.6, and in the AC-STV-013 note that repeats it.

**D4 — acceptance.md AC-STV-016 cannot fail, so REQ-STV-015 is unverified.** Severity: **major**.
Class: **blocking**.
REQ-STV-015 is a prohibition: the new code *"shall not duplicate what `StatusValueEnumRule` already
reports for the same document"*. Its only AC says: *"the set of documents reported under
`StatusTokenUnrecognized` and the set reported under `StatusValueInvalid` are compared explicitly,
and any document appearing in both is listed with the two messages side by side to show they
describe different facts"* — and closes with *"expected to be disjoint on this corpus — expected,
not established, until this AC runs."* There is no outcome that fails this AC. If the two sets
overlap, listing the overlap satisfies it. An AC that mandates an activity rather than an outcome
cannot detect the violation its REQ prohibits — and on this card specifically, an AC that cannot
fail is the exact defect class the card exists to close (§A.3 layers 1-2 are three checks that all
passed silently).
*Required fix*: give AC-STV-016 a failing condition. Either assert disjointness outright, or —
if genuine overlap is acceptable — state the bounded, enumerable set of documents allowed to appear
in both, so anything outside it fails.

**D5 — REQ-STV-015 mixes a system behaviour with a verification activity.** Severity: **minor**.
Class: **optional**.
*"…the two shall be shown, by measurement rather than by assumption, to describe disjoint facts."*
"Shall be shown by measurement" constrains how the team verifies, not how the system behaves. It
belongs in the AC (where D4 is asking for it anyway), not in the requirement.

**D6 — spec.md §B — REQ-STV-012 is placed out of sequence.** Severity: **minor**.
Class: **optional**.
It appears between REQ-STV-009 and REQ-STV-010. Numbering is intact (MP-1 passes); only reading
order and any hand-built cross-reference is affected.

**D7 — acceptance.md AC-STV-003 is not satisfiable as worded.** Severity: **minor**.
Class: **blocking**.
*"…the two findings differ in no field other than the commit SHA."* The Given constructs **two
repositories**, so the findings necessarily differ in `File` as well (each finding carries the
document's path — see the `Finding` shape in `lint-baseline.json`, whose entries carry
`file`/`line`/`severity`/`code`/`message`/`advisory`). An implementer following this literally either
fails a correct implementation or silently relaxes the assertion. The property being pinned —
trailer-independence, the measured reason a matrix-only fix would be vacuous
(`lint_ownership.go:414-416` returns nil when the trailer is absent, and the probe row
`implemented_to_completed_no_trailer` shows the silence) — is exactly right and is the AC I most
wanted to find. Only its wording is wrong.
*Required fix*: "differ in no field other than the commit SHA and the file path".

**D8 — two cited line ranges have imprecise boundaries.** Severity: **minor**. Class: **optional**.
spec.md §A.3 layer 1 cites `lint_ownership.go:78-94` for `expectedOwnerForTransition`, whose
function body spans **64-95** — the cited range opens mid-`case ""` arm and omits the signature.
plan.md §B cites `:412-415` for the `rec.AuthoredByAgent == ""` guard, which is at **414-416** — the
range opens on a comment and stops before the closing brace. Both substantive claims are true;
only the boundaries are loose. Every other coordinate in both documents is exact.

**D9 — `AC-STV-007a` uses a suffix form the rest of the set does not.** Severity: **minor**.
Class: **optional**.
Nothing mechanical depends on it: the lint engine's AC id pattern is a different four-segment
scheme entirely (`parser.go:218`, `^(AC-[A-Z0-9]+-[0-9]+-[0-9]+(?:\.[a-z]…)?)`), which matches none
of this SPEC's `AC-STV-NNN` ids, and the corpus lint is clean. But `007a` is the AC covering the two
largest census edges (217 + 50 = 267 records), so it is the worst one to lose to a hand-rolled
`AC-STV-\d{3}` extractor later. The codebase's own idiom for a sub-AC is a `.a` suffix.
*Suggested fix*: `AC-STV-007.a`, or renumber to `AC-STV-019`.

---

## Answers to the specific attacks commissioned

**Vacuous-AC hunt — survived, with one exception.** AC-STV-010 requires a must-fire case
(AC-STV-001/002) in the **same execution** as the silence assertions, and REQ-STV-011 backs it.
That is the correct shape, and it is the single most important thing in this AC set. The exception is
D4: AC-STV-016 is the one AC that cannot fail.

**Bidirectionality — satisfied.** Firing: AC-STV-001, 002, 014, 015, 018. Silent: AC-STV-004, 005,
006, 007, 007a, 008, 009, 017. acceptance.md's preamble binds them to one execution, and the DoD
repeats the binding.

**Projection vs observation — clean.** §A.6 is titled "(derived from A.4, **NOT measured**)" and its
body says "projections computed from a table, not observations of the rule running." AC-STV-013
states the projections "are the comparison target, not the pass condition." I found no AC that
treats a projection as a pass condition. The projection arithmetic itself has one error — D3.

**Trailer independence — pinned directly.** AC-STV-003 targets exactly the measured property, and
cites the probe row that establishes it. Wording defect only (D7).

**Check-order — load-bearing, and not ceremony, but its extra test has no AC.** The order is
genuinely decision-changing: with `(none)`-skip first, the real census row
`(none) → "in-progress"` is skipped entirely; with token-recognition first it would emit
`StatusTokenUnrecognized`. Both order boundaries already have externally-observable ACs — AC-STV-017
(neither code fires on `(none)`) and AC-STV-015's second half (no `StatusTransitionInvalid` for a
token-flagged pair). So plan.md M1's *additional* "assert the order in a test" is redundant with
those two, and no AC requires it — a plan deliverable with no acceptance criterion. Recommend either
dropping it in favour of the two behavioural ACs, or accepting it as a white-box convenience test
that closes no AC. Not a defect worth blocking on; recorded here rather than in the defect list
because the decision it protects is sound.

**Counts — both figures are right; they count different things.** acceptance.md contains **19**
`### AC-STV-…` blocks and **18** distinct numeric ids (001-018, contiguous, no gaps, no duplicates);
the nineteenth is `AC-STV-007a`. The DoD's arithmetic is internally consistent
("AC-STV-001..013 incl. 007a, plus AC-STV-014..018" = 14 + 5 = 19), and `progress.md`
`ac_count: 19` agrees with it. Nothing mechanical depends on either number (D9). The handoff's "19"
is not an error.

**Scope — layers 3 and 4 are stated clearly enough, and layer 4's limitation is properly recorded
as debt.** §C carries a dedicated headed subsection per non-goal. The layer-4 subsection does the
thing that distinguishes accepted debt from an omission: it names the consequence in the concrete
("`draft → completed` — one of the two edges this card exists to catch — is reported but does not
gate, because the very status it transitioned into shelters it"), says where the finding remains
visible, and says plainly that it is not a gate. The t371 subsection additionally pre-empts the
cross-card illusion ("t371 landing `fetch-depth: 0` does not close this card"), which is the right
instinct. A reader cannot mistake this card for closing layers 3 or 4. This part of the SPEC is
better than the norm.

**§A.5 honesty about D4/D5 — yes, explicitly.** The block after D5 reads: *"D4 and D5 were **not** in
the decision set handed to this SPEC; they were forced by the census and are recorded here for the
same reason D1's debt is."* Their rationales are sound on their own terms: D4 derives
`in-progress → completed` from matrix row 3's single-sync-commit note rather than inventing it, and
D5's argument — that `(none)` is an extractor artifact that cannot distinguish creation from
truncated history — is the correct reading and is consistent with the shape a squash merge leaves.
D1's SSOT divergence is likewise disclosed rather than papered over. Per the commission I did not
re-litigate D4/D5 as decisions.

---

## Recommendation

**PASS-WITH-DEBT.** No must-pass criterion fails, and 0.90 clears the Tier M 0.80 threshold. The
survey work behind this SPEC is unusually strong: every coordinate I re-read was accurate, every
count I re-derived matched, the probe and census are real executions rather than assertions, and the
projections are quarantined from the pass conditions.

Four blocking-class defects should be closed before the card closes. Priority order:

1. **D4** (AC-STV-016 cannot fail) — Priority High. Give it a failing condition; otherwise
   REQ-STV-015 ships unverified, which is this card's own failure mode.
2. **D2** (no AC binds the gating-population measurement) — Priority High. Extend AC-STV-013 to
   require the advisory split and a recorded decision on a non-zero non-advisory count. The strict
   gate currently has zero headroom (measured: 0 non-advisory findings in the 1096-finding
   baseline), and `spec-lint.yml` runs `--strict` on every `develop` push.
3. **D1** (§A.5 D2's rationale) — Priority Medium. Restate to match the census; the decision itself
   stands.
4. **D3** (~6 → ~7) and **D7** (AC-STV-003 wording) — Priority Medium, both one-line edits.

D5, D6, D8, D9 are optional-class: surface them to the orchestrator, do not route them into a
revision on their own.

Iteration ceiling: Tier M = 2. A confirming iteration 2, if run, should be scoped to this enumerated
defect delta rather than a from-scratch re-audit.
