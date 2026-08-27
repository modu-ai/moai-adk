# SPEC Review Report: SPEC-ASTGREP-LANG16-001

Iteration: 4 (delta-scoped; **past the Tier L ceiling of 3** — see § Retry-contract note)
Verdict: **FAIL**
Overall Score: **0.84** (iter-1 0.68 → iter-2 0.69 → iter-3 0.71 → iter-4 **0.84**)
Delta: **+0.13** vs iter-3, **+0.15** vs iter-2, **+0.16** vs iter-1. Tier L PASS threshold: **0.85**.
STOP signal: **not raised** (no regression — this is the largest single-round gain of the four).

Auditor: plan-auditor. Bias-prevention M1-M6 active.

**M1 Context Isolation statement.** Reasoning context ignored per M1 Context Isolation. The
invocation prompt's verification sweep was treated as a **claim to falsify**, exactly as instructed,
not as evidence. Every line below rests on my own measurement against the tree. Where the prompt's
sweep and mine differ, mine governs — and on two points they do differ, both recorded.

---

## Measurement attribution

| field | value |
|---|---|
| worktree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t228` |
| branch | `WT-astgrep-16-langs` |
| HEAD | `294b4b6ab3e9797d7a928e98a6241e9c44e5b445` (`git rev-parse HEAD`) |
| **`origin/main`** | **`f7eec06c7d250af680f50f1fe6e841cf8dc65da1` — NOT equal to HEAD** (`git rev-list --count 294b4b6ab..origin/main` → **10**) |
| working tree | `git status --porcelain` → three `??` entries only; `git diff --stat HEAD -- internal/hook/` → empty. No tracked file modified. |
| tool | `ast-grep 0.40.5` (`sg --version`) |
| cwd | worktree root for every command, per the SPEC's own §A.0 hygiene rule |

**The prompt's premise "HEAD 294b4b6ab (== origin/main)" is falsified.** `origin/main` has advanced
ten commits past the SPEC's pinned baseline. This is not a bookkeeping nit — it is the root of the
one critical defect this round produces (N1 below), and it was invisible to a sweep that assumed the
two refs were the same.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** `grep -oE '^\*\*REQ-A16-[0-9]{3}' spec.md | sort | uniq -d`
  → empty. Sequence measured: `001 002 … 023`, contiguous, three-digit zero-padded, 23 entries.
- **[PASS] MP-2 GEARS format compliance — judged against the `REQ-XXX` requirement layer in
  `spec.md` §C only.** The `AC-A16-*` entries are Given-When-Then by design and are graded under
  Group 4, not here. Four requirements were amended this round; I re-read each: `REQ-A16-001`
  (Ubiquitous — "The ruleset **shall** treat … No file … **shall** label"), `REQ-A16-011`
  (Ubiquitous — "Every security-family rule **shall** carry … **shall** instantiate … **shall** be
  anchored"), `REQ-A16-018` (When — "The source of truth **shall** be … and when it changes,
  `make build` **shall** be run … that build **shall** leave … unchanged"), `REQ-A16-021`
  (Ubiquitous — "No file under … **shall** contain … **shall** be English"). All four carry a named
  `<subject>`, a `shall`, and a modality matching their label. The remaining 19 are unchanged from
  the iteration-2 verification.
- **[PASS] MP-3 YAML frontmatter validity.** `spec.md:1-15` — all 12 canonical fields present with
  correct types: `id`, `title` (quoted), `version: "0.6.0"` (quoted semver), `status: draft`,
  `created: 2026-08-24`, `updated: 2026-08-25` (both ISO), `author`, `priority: P2`, `phase`,
  `module`, `lifecycle: spec-anchored`, `tags` (comma-separated string), plus optional `tier: L`.
  No rejected snake_case alias. The iteration-3 defect D9 (no version bump, no HISTORY row) is
  closed: `0.4.0` → `0.6.0` with a v0.6.0 HISTORY row at `spec.md:28`.
- **[PASS] MP-4 language neutrality.** Executed: `grep -rniE 'primary|planned|unsupported' $T/` →
  exactly one line, `sgconfig.yml:10` ("future addition, never an unsupported one") — the sanctioned
  equal-opportunity paragraph. `grep -rlE 'SPEC-[A-Z][A-Z0-9]+-[0-9]{3}' $T/ | wc -l` → **0**. All
  16 languages enumerated with equal weight at `spec.md` §C / `REQ-A16-001`.
- **[PASS] MP-5 D7 cross-SPEC reconciliation — no BLOCKING finding.** Extraction →
  `SPEC-ASTG-UPGRADE-001` (`archived`), `SPEC-ASTGREP-BREADTH-001` (`draft`),
  `SPEC-ASTGREP-DOGFOOD-CLEANUP-001` (`completed`), `SPEC-ASTGREP-MULTILANG-001` (`completed`), all
  four measured by reading their own `status:` frontmatter. The one archived reference is reconciled
  at three live sites (`spec.md:511`, `:537`, `:557`), each stating it owns no active work and that
  R6's re-probe is owned here.
- **[PASS] MP-6 D8 cross-platform discipline.** `grep -rn 'syscall' .moai/specs/SPEC-ASTGREP-LANG16-001/`
  → rc=1. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate.** `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-ASTGREP-LANG16-001/`
  → rc=1 across all six artifacts.

**All seven must-pass criteria pass.** The FAIL is score-driven, as in iterations 2 and 3.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.80 | 0.75 band | Every inter-artifact contradiction I have raised across three iterations is gone, and I checked each by reading both sides rather than one. The `catalog.yaml` requirement and its criterion now assert the same thing (`spec.md:400-404` vs `acceptance.md:223`); the PRESERVE tree and the pinning test no longer collide (`plan.md:73-80` carve-out vs `plan.md:141-152` M1 item 6); `design.md`'s D13 placement, wrong in three consecutive iterations, now reads the same way in all three places (`:119-122`, `:166`, `:171-179`). One real ambiguity remains and it is new in weight: the SPEC pins base SHA `294b4b6ab` in five places (`spec.md:43`, `plan.md:14`, `acceptance.md:10`, `:272`, `research.md:4`) and then uses the **moving** ref `origin/main` as the untouchedness baseline in four (`plan.md:78`, `:147`, `:284`, `acceptance.md:120-121`), never saying whether they are the same. Measured today they differ by ten commits (N1). |
| Completeness | 0.88 | 0.75+, approaching 1.0 | All required sections present; §D carries seven `### Out of Scope — <topic>` H3 sub-headings with specific bullets; frontmatter complete; Tier L 5-artifact set present and substantive. `progress.md` §E.1 is now current and I verified each field against the tree: `requirements: 23` (measured 23), `acceptance_criteria: 19` (measured 19, last id `AC-A16-019`), `spec_version: 0.6.0` (matches frontmatter), `audit_iteration: 3 revisions (FAIL 0.68 -> 0.69 -> 0.71 -> propagation pass)`, budget sentence corrected to "headroom of 2 and 6 respectively". D8 and D9 both closed. Deduction: `design.md:48`'s example evidence cell still writes a rule-test path as bare `rule-tests/rust/credentials/`, which reads as `$T`-relative in a document that now says everywhere else that the root is outside `$T` (N3). |
| Testability | 0.82 | 0.75+ | I re-executed the §H rows the prompt named plus three more, on the tree, after the amendments. Row 016: `grep -rh 'cwe:' $T/security/*.yml \| wc -l` → **14**, security rule count → **14**, and the citation/probe scan → **0** — the recorded "CWE half already passes, anchor half absent at 0 of 14" is exact. Row 019: `find $T -name '*test*'` → no output; `grep -rlE 'SPEC-…' $T/` → **0** — exact. Row 006: `grep -c testConfigs $T/sgconfig.yml` → `0`, rc=1 — exact. Row 001: `sg test --config $T/sgconfig.yml` → `Running 0 tests` / `ok. 0 passed; 0 failed;` EXIT=0 — exact, and `grep -rh '^id:' $T --include='*.yml' \| wc -l` → **26**, so 0 ≠ 26 still holds. The catalog measurement underpinning the D1 amendment reproduces: `grep -c astgrep internal/template/catalog.yaml` → **0** (rc=1), `grep -c 'astgrep-rules'` → **0**. Every GREEN PATH now resolves to a real work item — I checked `M1 item 6` (`plan.md:141`) and `M3 item 4` (`plan.md:209`) by name, the two that named nothing in iteration 3. AC-A16-001's open count clause is bound to a derivable number (`acceptance.md:51-52`). Deduction: **AC-A16-008 is currently non-attributable** — its baseline is a moving ref that has already moved (N1) — and §H row 019's "Why RED" column does not disclose the newly-added `catalog.yaml` clause as a guard-that-holds-today alongside the other four (N2). |
| Traceability | 0.86 | 0.75+, approaching 1.0 | The renumbering damage that dominated iteration 3 is fully repaired and I verified it against `acceptance.md`'s own section boundaries rather than against the fix note: §A = 001…008 / §B = 009…014 / §C = 015…018 / §D = 019, and `plan.md:155` / `:188` / `:228` now read exactly those ranges. `grep -rnoE 'AC-A16-0(2[0-9])'` over all six artifacts returns **one** hit, `progress.md:61`, which is self-annotated in place as the audit trail of what iteration 1 asked for — the same acceptable class I allowed in iteration 3. The two dangling ids are gone; `design.md:70` now cites `AC-A16-011` and `research.md:33` now cites `AC-A16-018`, both correct; `plan.md:335` reads 19. All four §I rows I rejected in iteration 3 are re-derived and each now holds against the criterion body — including the `REQ-A16-014` row, which drops `AC-A16-003` and states in the cell **why** it did not hold. §I gained a `[HARD]` scope-gap clause (`acceptance.md:317-322`). The successor's inherited-contract table (`SPEC-ASTGREP-BREADTH-001/spec.md:104-112`) cites only live requirement ids and carries all three amendments verbatim — I checked rows 106, 111, 112 against the amended requirement text. Deduction: N1 and N3. |

Aggregate = harmonic mean of (0.80, 0.88, 0.82, 0.86) = **0.8388 → 0.84**. Tier L threshold 0.85.

---

## Audit question 1 — re-verification of the six claimed closures

Each measured against the tree. My result governs.

| # | Claim | My verdict | Evidence I observed |
|---|---|---|---|
| D1 | `catalog.yaml` — build leaves it unchanged, `plan.md` agrees | **FIXED** | `spec.md:400-404`: *"that build shall leave `internal/template/catalog.yaml` **unchanged**. No regenerated embed artifact accompanies a ruleset edit."* `plan.md:254-260` states the same inverted form and calls a non-empty catalog diff *"the defect, not the deliverable"*. `grep -rn 'REGENERATE'` over the SPEC dir → rc=1, so the falsified `plan.md:58` line is gone. I re-measured the underlying fact myself: `grep -c astgrep internal/template/catalog.yaml` → **0**. Requirement, plan, criterion and §I row (`acceptance.md:343`) all now say the same thing. |
| D2 | `REQ-A16-011` carries citation-or-probe; `plan.md` M3 has the work | **FIXED** | `spec.md:319-325` third clause: *"**and** the symbol its pattern matches at the head shall be anchored by at least one of: a **citation** … or a **probe** …"*. `plan.md:209-217` is a numbered M3 item 4 doing exactly that, and `acceptance.md:291` (§H row 016) now names `M3 item 4` rather than the non-existent `M3.3`. The criterion no longer asserts more than the requirement and the plan together deliver. |
| D3 | `REQ-A16-021` re-scoped to ruleset dir + rule-test root | **FIXED** | `spec.md:434-438` scopes to `internal/template/templates/.moai/config/astgrep-rules/**` plus the repo-side rule-test root. `REQ-A16-001` (`spec.md:249-253`) carries the same narrowing with a cross-reference. `grep -rn 'templates/\*\*'` returns five hits, **all** of them narrative records of the correction (`spec.md:255`, `:440`, `acceptance.md:225`, `:326`, `:346`) plus two audit-trail lines — no live obligation is stated over the wide tree. |
| D4 | PRESERVE carve-out for one new file; §A.4 otherwise byte-unchanged | **FIXED** | `plan.md:73-80` carves out exactly `internal/hook/astgrep_corpus_pin_test.go` and says it is *"the only permitted addition under this tree"*; `plan.md:141-152` is M1 item 6 authoring it; `plan.md:284` restates the standing command with `':(exclude)internal/hook/astgrep_corpus_pin_test.go'`. The collision iteration 3 found is genuinely resolved. **But the restated command is where N1 lives** — see below. |
| D5 | `design.md:119` and `:166` place `testConfigs` outside `$T` | **FIXED** | `design.md:119-122`: *"points at the **repo-side rule-test root, outside the distributed template tree** … never under `$T/rule-tests/`"*. `design.md:166`: *"testConfigs testDir points OUTSIDE $T"*. `design.md:171`: *"`<repo-side rule-tests root>`  # NOT under the template tree"*. All three now agree with the §4 paragraph at `:176-180`. **The three-iteration stagnation is broken.** |
| — | Zero AC references above `AC-A16-019` | **PARTIAL — the sweep's claim is inaccurate** | `grep -rnoE 'AC-A16-0(2[0-9])'` → **one** hit: `progress.md:61`, `AC-A16-020`. It is immediately self-annotated (*"that criterion was folded into AC-A16-019 at v0.5.0; the id is retained here as the audit trail of what iteration 1 asked for"*), which is the acceptable historical class I allowed in iteration 3 — so it is **not a defect**. But the sweep asserted zero and there is one, and an inaccurate absence claim is worth naming even when the residue turns out to be sanctioned. |
| — | Counts 23 REQ / 19 AC | **CONFIRMED** | Measured 23 and 19 independently. |

**Nothing escaped sideways this time.** The specific shape I was asked to hunt — an impossible clause
relocating out of the criteria layer into `plan.md` prose or `design.md`, where neither §0's discipline
nor a requirement review reaches — did not recur for these six. I checked by scanning both destination
documents for wide-scope obligations and for the pre-amendment `catalog.yaml` wording, and found only
narrative records of the corrections. The escape route was additionally **closed structurally**:
`acceptance.md:27-33` now extends §0's two-column discipline to the requirement layer as a `[HARD]`
clause, and `spec.md:440-452` records the blind spot that made it necessary. That is the right fix and
it is the reason I am not scoring the escape risk again.

## Audit question 2 — new contradictions from the amendments

I swept every citation of the four amended requirements (`grep -rnE 'REQ-A16-(001|011|018|021)\b'`
across all six artifacts plus the successor) and read each site against the amended text.

| Site | Cites | Agrees with amended text? |
|---|---|---|
| `spec.md:221` | REQ-011 | yes — "adds that anchor" |
| `spec.md:255` | REQ-021 | yes — records the narrowing |
| `spec.md:432` | REQ-021 | yes — dogfood tree "would violate REQ-A16-021 on contact" (about mirroring *into* the scoped tree — consistent) |
| `spec.md:533` (R2 risk row) | REQ-011 | yes — "necessary-not-sufficient", matches the amended third clause's framing |
| `acceptance.md:31` | REQ-021 | yes — the §0 extension's worked example |
| `acceptance.md:263` (§G row) | REQ-018 | yes — states the inverted form and that the requirement now matches |
| `acceptance.md:291` (§H row 016) | REQ-011 | yes — names `M3 item 4`, which exists |
| `acceptance.md:326/336/343/346` (§I) | 001/011/018/021 | yes — all four re-derived, each flagging that the scopes now match |
| `plan.md:62` | REQ-021 | yes — matrix lives under `.moai/specs/`, outside both scoped trees |
| `plan.md:207`, `:209` | REQ-011 | yes — both halves, as separate numbered items |
| `plan.md:324` | REQ-001 | yes — the `sgconfig.yml` paragraph as named sanctioned exemption |
| `research.md:187` | REQ-011 | yes — necessary-not-sufficient |
| `SPEC-ASTGREP-BREADTH-001/spec.md:106, :111, :112` | REQ-011/018/021 | **yes, all three** — the successor's inherited-contract table carries the citation-or-probe clause, the `catalog.yaml`-unchanged form, and the narrowed neutrality scope with an explicit *"**not** to `internal/template/templates/**`"*. |

**No citation site was left disagreeing with an amended requirement.** The iteration-3 lesson —
editing one layer without sweeping its citations manufactures a contradiction pair — was applied.
This is the single most convincing evidence that the propagation was done rather than narrated.

## Audit question 3 — the §H evidence table after the amendments

Rows re-run, chosen per instruction to cover `catalog.yaml`, neutrality scope, and the CWE anchor,
plus two controls.

| §H row | Command | Recorded | I observed | Match |
|---|---|---|---|---|
| 016 (CWE anchor) | `grep -rh 'cwe:' $T/security/*.yml \| wc -l` | `14` of 14 | `14`; security rule count `14`; citation/probe scan over the same files → **0** | ✅ still accurate on **both** halves |
| 019 (neutrality) | `find $T -name '*test*'` ; `grep -rlE 'SPEC-…' $T/` | *(no output)* ; `0` | *(no output)* ; `0` | ✅ |
| 019 ancillary (ranking) | `grep -rniE 'primary\|planned\|unsupported' $T/` | one sanctioned line | one line, `sgconfig.yml:10` | ✅ |
| 019 ancillary (catalog) | `grep -c astgrep internal/template/catalog.yaml` | `0` | `0` (rc=1); `astgrep-rules` also `0` | ✅ |
| 001 / 015 / 018 | `sg test --config $T/sgconfig.yml` | `Running 0 tests` / `ok. 0 passed; 0 failed;` EXIT=0 | identical, EXIT=0 | ✅ |
| 001 (rule count) | `grep -rh '^id:' $T --include='*.yml' \| wc -l` | 26 | 26 | ✅ |
| 006 | `grep -c testConfigs $T/sgconfig.yml` | `0` (rc=1) | `0`, rc=1 | ✅ |

**No amendment silently turned a red row green, and no green path was left pointing at moved work.**
I additionally confirmed the ruleset tree is untouched by main's ten new commits
(`git diff --stat 294b4b6ab origin/main -- $T/` → empty), which is why the whole `$T`-scoped half of
§H survives the drift intact. Row 016's green path moved from the non-existent `M3.3` to the real
`M3 item 4`, and rows 007/008 moved from an unnamed "M1 adds the pinning test" to the real
`M1 item 6` — both verified by reading `plan.md`.

One honesty gap, small but real: row 019's "Why that is RED" column names four guards that hold
today (SPEC-ID, date, ranking, dogfood) and rides the one red clause (`$R` absent). The D1 amendment
added a **fifth** guard-that-holds-today to that row's GREEN PATH cell — *"`make build` leaves
`catalog.yaml` unchanged"* — without adding it to the guard list. The row is the SPEC's own model of
disclosing which clauses are decorative today; a newly-added one that skips the disclosure is a
regression in that discipline, however minor (N2).

## Audit question 4 — scope creep

**None of substance.** Measured: 23 requirements before and after, 19 criteria before and after — the
propagation added no requirement and no criterion, exactly as `progress.md:39-41` claims and as I
verified by counting. Three observations, none rising to creep:

1. **Four requirements were amended, not two.** `REQ-A16-001`, `-011`, `-018`, `-021`. The prompt
   authorized "at most two". All four trace directly to an iteration-3 *Required fix* line — D3's fix
   named `REQ-A16-021` **and** `REQ-A16-001` explicitly, D1 named `-018`, D2 named `-011`. This is
   the audit's own prescription being executed, not scope expansion. I record the count discrepancy
   because the authorization and the work disagree on their face, and a reader comparing them without
   the iteration-3 text would misread it.
2. **The §0 requirement-layer extension is a paragraph, not the "one sentence" iteration 3 asked
   for** (`acceptance.md:27-33`, six lines plus a worked example). Over-delivery against a fix
   instruction. It is the right content and I would not remove it. Optional.
3. **The §I `[HARD]` scope-gap clause** (`acceptance.md:317-322`) is new prose beyond the enumerated
   fixes, added in service of D7. Same disposition.

## Audit question 5 — anything still passing on an untouched tree; green paths naming absent work

**No criterion is vacuous**, and this holds after the amendments — I re-ran the rows most likely to
have flipped and none did. **Every green path resolves**: the two that named absent milestone work in
iteration 3 (`M3.3`, "M1 adds the pinning test") now name `M3 item 4` and `M1 item 6`, both of which
I located in `plan.md` by number and read.

**But one criterion now fails on an untouched tree for a reason no work in this SPEC can prevent**,
which is the mirror-image defect and is worse. That is N1.

---

## Defects Found (structured defect-list)

**N1. The untouchedness proof is anchored to a moving ref, and it is already false** —
`plan.md:284` (§E standing command), `plan.md:78`, `plan.md:147`, `acceptance.md:120-121`
(AC-A16-008), `acceptance.md:259-260` (§H rows 007/008) — the §E standing command and AC-A16-008
both baseline against **`origin/main`**, while every other measurement in the SPEC pins the base SHA
`294b4b6ab` (`spec.md:43`, `plan.md:14`, `acceptance.md:10`, `:272`, `research.md:4`). Measured
today, on a tree where this SPEC has changed nothing:

```
$ git rev-list --count 294b4b6ab..origin/main
10
$ git diff --stat HEAD -- internal/hook/
(empty — the worktree is clean)
$ git diff --stat origin/main -- internal/hook/ ':(exclude)internal/hook/astgrep_corpus_pin_test.go'
 ... 18 files changed, 23 insertions(+), 2825 deletions(-)
rc=0
```

`plan.md:266-268` calls that command *"the proof that every pre-existing file is byte-unchanged"*.
It is not: it reports eighteen files and 2,825 deletions before a single line of this SPEC's work
exists, and the number will keep drifting as `main` advances. AC-A16-008's *"byte-unchanged from
`origin/main`"* inherits the same defect — the criterion can fail for reasons entirely outside this
SPEC, which is precisely the **impossible** mutant `acceptance.md` §0 defines. An implementer running
the standing command at M1 will see a large diff and either stop, or learn to disregard the command —
and a guard that is routinely disregarded provides no protection at all. I verified the fixtures
themselves are stable (`git ls-tree -r --name-only origin/main | grep scan-corpus | wc -l` → 12, same
at the pinned SHA), so the fix is purely the choice of baseline, not a scope change.

Note on provenance, stated plainly: **iteration 3's own required-fix line prescribed this exact
command form.** It was written when the two refs coincided, and the propagation implemented the
letter and inherited a latent defect. That does not make it less blocking — a criterion that fails on
an untouched tree is a defect regardless of who authored the wording — but the responsibility is
shared, and I record it rather than presenting this as a fresh authoring error.
  — Severity: **critical** — Class: **blocking** — Required fix: replace `origin/main` with the
  SPEC's pinned base SHA at all four sites — `git diff --stat 294b4b6ab -- internal/hook/
  ':(exclude)internal/hook/astgrep_corpus_pin_test.go'` — and reword AC-A16-008 to *"byte-unchanged
  from the base SHA recorded in §A.0"*. Add one sentence to `plan.md` §E stating why the pinned SHA
  and not `origin/main`: the proof asserts *this SPEC changed nothing*, and only a fixed baseline can
  assert that. Then re-run §H rows 007/008 and record the new observed output.

**N2. §H row 019 gained a fifth guard-that-holds-today without disclosing it as one** —
`acceptance.md:294` — the row's "Why that is RED" column enumerates the guards riding the single red
clause (`$R` absent) so a reader knows which parts of the composite observe nothing today. The D1
amendment added *"`make build` leaves `catalog.yaml` unchanged"* to the GREEN PATH cell; that clause
also holds today (nothing produces a catalog diff now), so it is another guard and the row does not
say so. The disclosure discipline is this table's own contribution; a new clause that skips it
erodes exactly the property that made §H credible.
  — Severity: **minor** — Class: **blocking** — Required fix: add the catalog clause to row 019's
  guard list in the "Why that is RED" cell, matching how the SPEC-ID / date / ranking / dogfood
  clauses are already handled.

**N3. `design.md`'s matrix-schema example writes a rule-test path as `$T`-relative** —
`design.md:48` — the example evidence cell reads `rule-tests/rust/credentials/`, a bare relative path
with no root. Everywhere else in the document now states the root is **outside** `$T` (`:119-122`,
`:166`, `:171`, `:176-180`), and this line is the residue of the placement wording that D5 corrected
elsewhere. It is an illustrative cell rather than a normative statement, which is why it is minor —
but it is the one surviving place where a reader can still infer the old placement.
  — Severity: **minor** — Class: **optional** — Required fix: write the example cell's path against
  the repo-side root (or as `$R/rust/credentials/`, using the shorthand `acceptance.md` §0 already
  defines).

**N4. The prompt's sweep asserted zero AC references above `AC-A16-019`; there is one** —
`progress.md:61` — the reference is sanctioned and self-annotated in place, so it is **not itself a
defect**, and I would have objected had it been silently renumbered. Recorded because the verification
claim was inaccurate, and because an absence claim that overstates by one is the class of claim that
gets trusted next time without measurement.
  — Severity: **minor** — Class: **optional** — Required fix: none to the SPEC. Correct the sweep
  claim to "one, sanctioned and annotated".

**No other defect found.** Iteration 3's D1 through D11 are all closed, each verified by my own
measurement rather than by reading the fix note.

---

## Regression Check (iteration 3 → 4)

| iter-3 defect | Severity/Class | Status | Evidence |
|---|---|---|---|
| **D1** `catalog.yaml` requirement factually wrong and self-contradictory | critical/blocking | **FIXED** | `spec.md:400-404`, `plan.md:254-260`, `acceptance.md:343`; `grep -rn 'REGENERATE'` → rc=1; catalog measurement reproduced (`0`). |
| **D2** AC-016 asserts a half no requirement states, no milestone builds | critical/blocking | **FIXED** | `spec.md:319-325` third clause; `plan.md:209-217` M3 item 4; §H row 016 green path corrected. |
| **D3** `REQ-A16-021` scope measured false, impossible clause escaped upward | critical/blocking | **FIXED** | `spec.md:434-438`, `:249-253`; no live wide-scope obligation survives; escape route closed structurally by `acceptance.md:27-33`. |
| **D4** pinning test collides with PRESERVE | critical/blocking | **FIXED** (fix introduces N1) | `plan.md:73-80` carve-out, `:141-152` M1 item 6, `:284` excluded pathspec. The collision is gone; the restated command's baseline is N1. |
| **D5** `design.md` D13 contradiction — **stagnant across three iterations** | major/blocking | **FIXED — stagnation broken** | `design.md:119-122`, `:166`, `:171` all state the outside-`$T` placement. |
| **D6** renumbering 23 → 19 unpropagated (2 dangling, 5 wrong, 1 stale count) | major/blocking | **FIXED** | All three "Verified by" ranges verified against `acceptance.md`'s own section boundaries; `design.md:70` → AC-011; `research.md:33` → AC-018; `plan.md:335` → 19; no dangling id remains. |
| **D7** four §I rows do not hold | major/blocking | **FIXED** | Rows 001/011/018/021 re-derived; the `REQ-A16-014` row drops `AC-A16-003` and states why; `[HARD]` scope-gap clause added. |
| **D8** `progress.md` signal stale by a full iteration | minor/blocking | **FIXED** | Every field re-measured against the tree; all match. |
| **D9** no HISTORY row, no version bump | minor/blocking | **FIXED** | `version: "0.6.0"`; HISTORY row at `spec.md:28` enumerating D1-D11. |
| **D10** AC-001's open count clause | minor/optional | **FIXED** | `acceptance.md:51-52` binds the alternative to distinct rule ids as recorded in `design.md` §3.3. |
| **D11** "three classes" over four + three stale citations | minor/blocking | **FIXED** | `design.md:204`, `:219` read "four classes"; REQ-015 (`:110`), REQ-023 (`:192`), §A.5 (`:286`) all corrected. |

**Eleven of eleven closed. Stagnation: cleared** — the `design.md` D13 contradiction that persisted
through iterations 1, 2 and 3 is resolved, so the retry contract's stagnation flag is withdrawn.

---

## Recommendation

**This is real movement, not another flat round.** The three prior iterations moved 0.68 → 0.69 →
0.71; this one moves 0.71 → **0.84**, and the gain is not a scoring artifact. Iteration 3's core
finding was that a revision had touched one artifact of six and left every cross-artifact obligation
half-stated; this round closed all eleven enumerated defects, and I verified each against the tree
rather than against the fix note. Two things convince me in particular, and I would say so even if
the score had not moved: the citation sweep of the four amended requirements turned up **no**
disagreeing site anywhere in six artifacts plus the successor — the exact failure the previous round
demonstrated — and the `design.md` contradiction that survived three consecutive audits is gone.

**The verdict is FAIL at 0.84 against a 0.85 threshold.** I considered rounding toward PASS and
rejected it. One critical, blocking defect stands (N1): the SPEC's own untouchedness proof reports
eighteen changed files on a tree where nothing has been done, and AC-A16-008 can fail for reasons
outside this SPEC entirely. That is the impossible-criterion class §0 exists to prevent, and passing
a SPEC whose PRESERVE guard is non-functional on arrival would hand run-phase a check its
implementer is going to learn to ignore. The margin is one hundredth of a point and I am not
pretending otherwise — but the defect is not marginal, and the score landing where it did is a
coincidence rather than the reason.

**What remains is reachable in one bounded pass. Nothing structural is blocking.** The distance to
0.85 is three edits, all mechanical, none requiring re-authoring or a new requirement:

1. **N1 (critical)** — replace `origin/main` with the pinned base SHA `294b4b6ab` at four sites
   (`plan.md:78`, `:147`, `:284`, `acceptance.md:120-121`), reword AC-A16-008 to name §A.0's base
   SHA, add one sentence to `plan.md` §E on why a fixed baseline is required, and re-run §H rows
   007/008 recording the new output. Expected result: the standing command returns empty, and the
   criterion becomes attributable to this SPEC's work alone.
2. **N2 (minor)** — add the `catalog.yaml` clause to §H row 019's guard list.
3. **N3 (minor, optional)** — write `design.md:48`'s example path against the repo-side root.

With N1 closed I would expect Testability ≈ 0.90 and Traceability ≈ 0.90, taking the aggregate to
roughly 0.87 — clear of the threshold with margin, not on it.

**Retry-contract note (must not be silently absorbed).** The Tier L ceiling is **3** iterations and
this is **iteration 4**, conducted as an operator-authorized bounded delta pass. Under the Max-3
cap the orchestrator escalates to the user with three options — PASS-with-debt, scope-reduction, or
an explicit override extending the cap. My input to that choice: **scope-reduction is not warranted**
(the SPEC is coherent, its scope was already halved at v0.4.0, and the split is holding), and
**PASS-with-debt is defensible only if N1 is fixed first** — carrying a non-functional PRESERVE guard
into run-phase as documented debt would mean documenting that the guard does not work while leaving
run-phase to rely on it. My recommendation is a fifth bounded pass limited to the three edits above,
or, if the cap is to be respected strictly, fixing N1 and accepting the result without a further
audit round, since N2 and N3 are minor and N1's fix is mechanically verifiable by the operator in one
command (`git diff --stat 294b4b6ab -- internal/hook/ ':(exclude)…'` returning empty).

Credit where it is due: the two-column §H method held up under re-execution for a second consecutive
round, the requirement-layer extension of §0 closed the structural escape route that produced
iteration 3's worst defect, and the propagation was done rather than narrated. This SPEC is one
mechanical fix from passing.
