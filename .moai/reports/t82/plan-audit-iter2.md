# SPEC Review Report: SPEC-AGENTS-MD-CANON-001

Iteration: 2/3
Verdict: **FAIL** (marginal — 0.83 against a 0.85 Tier L threshold)
Overall Score: **0.83** (harmonic mean)

Audited tree: `.claude/worktrees/t82`, commit **`e5699d4fc`**
("feat(spec-agents-md-canon-001): cite the existing negative test instead of duplicating it"),
SPEC version `0.3.1`.

Reasoning context ignored per M1 Context Isolation.

**Provenance correction — this report first named the wrong commit, and gave a wrong reason for it.**

The original text of this section claimed the audited tree was `cd6e12459`, and explained the ` M`
that `git status --porcelain` reported on three artifacts as a "stale stat cache, not a content
difference". **Both statements were wrong**, and the second was the more serious error: it was a
diagnosis asserted from two observations that did not support it.

What actually happened, measured after the fact: the author's v0.3.1 edits were already saved to
disk when this audit began. `git status` at 02:32:37 reported ` M` because the files genuinely
differed from the then-`HEAD` `cd6e12459`. The author committed them as `e5699d4fc` at 02:33:10 —
**during** the audit. My later `git diff -- <specdir>` returned empty not because the content
matched `cd6e12459`, but because `HEAD` had moved out from under the comparison. I read "empty
diff" as "no content difference" without checking whether the baseline was still the same baseline.

The audit itself is unaffected, and the evidence for that is the pin I took at the start. All six
SHA-256 hashes recorded at 02:32:37 match the blobs at `e5699d4fc` exactly — verified per file with
`git show e5699d4fc:<path> | shasum -a 256`, e.g. `acceptance.md` → `b1c20d58ed37…`, which is the
pinned value, while the same file at `cd6e12459` is `3e361301e852…`, which is not. The end-of-audit
re-hash was identical to the start. **Every artifact I read was the `e5699d4fc` content, for the
entire audit window, and the content was stable throughout.** The audited state and the state the
dispatch asks to be judged are the same state; only my label for it was wrong.

Recorded rather than quietly fixed, because the mechanism is the one this SPEC's own §D.4
measurement-provenance constraint exists to prevent: a comparison whose baseline moved silently,
reported as though the baseline had held. It is the same failure shape as the D6 defect this audit
graded — a conclusion reached from a real observation whose premise had expired.

**Scope.** Re-audit of the enumerated D1-D12 delta plus a regression pass over it, per iteration 1's
own scoping instruction. Not a from-scratch review; iteration 1's must-pass findings were
re-executed because they are cheap and because two of them (MP-1, MP-3) are directly perturbed by
the delta's renumbering.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-AMC-001` … `REQ-AMC-018`, sequential, no gaps, no
  duplicates, uniform 3-digit padding (18 matches of `^\*\*REQ-AMC`). AC side `AC-AMC-001` …
  `AC-AMC-024`, contiguous, 24 criteria, verified by extracting the ordered heading list. The delta
  added `REQ-AMC-018` and `AC-AMC-019` … `AC-AMC-024` without disturbing the sequence.
  Tier L ceilings (25 / 25, `spec-workflow.md` § SPEC Complexity Tier): 18 ≤ 25 and 24 ≤ 25 — but
  **the AC budget now has exactly one slot left**, which constrains iteration 3's fix options.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md` §C
  `REQ-AMC-*`), never against an AC; `acceptance.md`'s Given-When-Then entries are the correct
  verification-layer form and are graded under Group 4. All 18 match a GEARS pattern. The new
  `REQ-AMC-018` ("The design shall not treat a raised `project_doc_max_bytes` as a substitute for
  the diet") is canonical Unwanted form. `REQ-AMC-004`'s label was corrected to `(Unwanted)` per
  D10, matching the form it labels.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:1-15`), plus optional `tier: L`; `version: "0.3.0"` quoted, `created`/`updated` ISO,
  no rejected snake_case alias. Mechanically confirmed:
  `moai spec lint .moai/specs/SPEC-AGENTS-MD-CANON-001/spec.md` → `✓ No findings`, exit 0.
- **[N/A] MP-4 language neutrality** — the SPEC governs harness rule files and names no
  language-specific tooling. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the only external SPEC reference is
  `SPEC-ALWAYS-LOADED-DIET-001`; `grep '^status:'` on its `spec.md` → `status: completed`. Not in
  {retired, superseded, archived}. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -rn 'syscall'` over the SPEC directory → no
  match, exit 1. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → no
  match, exit 1.

No must-pass failure. The FAIL is the aggregate falling below the Tier L threshold, and it is
carried by a single blocking finding.

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.85 | 0.75→1.0 | Both iteration-1 clarity docks closed: "integration branch" is now defined with a discriminator (`spec.md:214-221`), and `AC-AMC-009` names which copies the ceiling binds (`acceptance.md:56-60`). Residual: one stale cross-reference (N2) and one unscoped criterion (N7). |
| Completeness | 0.93 | 1.0 | All required sections present; five `### Out of Scope — …` h3 sub-headings with `-` bullets (`spec.md:417/426/431/440/445`); Tier L 5-artifact set complete plus `progress.md`. D5/D6/D7 all closed with measured content in three artifacts each. Docked only for structural hygiene (N3, N4, N5). |
| Testability | 0.70 | 0.50→0.75 | Derivation below the table — this is the dimension that decides the verdict. |
| Traceability | 0.90 | 0.75→1.0 | All 18 REQs carry ≥1 AC (`REQ-AMC-006` now covered by `AC-AMC-012`); no AC references a non-existent REQ. Docked for one broken trace link (N2, `design.md:173` cites a renumbered AC) and one AC testing a property no REQ states (N5). |

**Testability derivation (shown because the verdict turns on it).** Three criteria are decidable but
require interpretation the criterion does not supply — `AC-AMC-002` ("reports each recorded run
against its expected result", where the script prints rather than asserts), `AC-AMC-021` ("`make
build` has been run", no observable named), `AC-AMC-024` ("a bare assertion of it **anywhere**", no
scan scope and no command). Three such criteria sit between the rubric's 0.75 band ("one AC … minor
interpretation") and its 0.50 band ("several ACs … require judgment calls"); none carries a weasel
word, and each names its artifact, so 0.75 with pressure is the honest read of that group alone.
A further deduction applies for `AC-AMC-019`: it is binary-testable, but it does not verify what its
own parenthetical claims it verifies (N1), and the property it fails to verify is the SPEC's
headline goal. The rubric has no band for "testable but ineffective on the load-bearing criterion";
it is worse than "measurable with minor interpretation", so 0.75 − 0.05 = **0.70**.

Harmonic mean of (0.85, 0.93, 0.70, 0.90) = **0.8348**. Tier L threshold **0.85**
(`spec-workflow.md` § SPEC Complexity Tier, verified in this tree) → **FAIL by 0.015**.

**The verdict is robust to how N1 is allocated but not to its existence.** Charging N1 to Clarity
instead of Testability (0.75 / 0.93 / 0.78 / 0.90) gives 0.8330 — the same verdict. Removing N1
entirely (Testability → 0.80) gives **0.867 → PASS**. One blocking fix stands between this SPEC and
a PASS.

---

## Per-Finding Resolution Status

| # | Iteration-1 finding | Status | Evidence |
|---|---|---|---|
| D1 | ratchet has no working enforcement criterion | **PARTIALLY RESOLVED** | (a) and (b) done; (c) added but ineffective — see N1 |
| D2 | `AC-AMC-010` fails on a correct tree | **RESOLVED** | command re-executed, both failure modes closed |
| D3 | "the integration branch" undefined | **RESOLVED** | `REQ-AMC-014`, `AC-AMC-018`, `plan.md:143-146`, `design.md:225-232` |
| D4 | `AC-AMC-002` cites an unreproducible fixture | **RESOLVED** | `probe-fixture.sh` committed and tracked; residual N6 |
| D5 | line-grep proxy undisclosed | **RESOLVED** | `spec.md` §A.4, `design.md` §1, `research.md:45-49`, `AC-AMC-003`, `plan.md` M1 |
| D6 | cap-raise rationale contradicted by measurement | **RESOLVED** | `design.md:92-108`, `design.md` §6 row, `spec.md` §D.8, `REQ-AMC-018`, `AC-AMC-024` |
| D7 | nested-`CLAUDE.md` asymmetry unaddressed | **RESOLVED** | `design.md` §3.1 table, `spec.md` §D.6 table |
| D8 | `REQ-AMC-006` has no AC | **RESOLVED** | `AC-AMC-012` — see ruling 3 |
| D9 | `AC-AMC-012` names no mechanism | **RESOLVED** | now `AC-AMC-013`, command re-executed |
| D10 | `REQ-AMC-004` mislabelled | **RESOLVED** | `spec.md:170` reads `(Unwanted)` |
| D11 | `REQ-AMC-006` leads with `MAY` | **RESOLVED** | `spec.md:179` now "This SPEC **shall** record" |
| D12 | inline rationale in two clauses | **RESOLVED** | `REQ-AMC-005` → "§D.6"; `REQ-AMC-009` → "Rationale: §D.7" |

**11 of 12 fully resolved. No fix was cosmetic, and no previously-passing property regressed.**

### Verification I executed (not carried from the dispatch)

- **D2 command.** `git ls-files --full-name ':(top)*AGENTS.md' ':(exclude,top)internal/template/templates/*'`
  → empty (no `AGENTS.md` yet). Against the `CLAUDE.md` analogue it returns the 6 live-tree files
  with the mirror excluded, versus **7** from `find . -name 'CLAUDE.md' -not -path './.git/*'`. Run
  again from `internal/spec`: **6, identical list** — `:(top)` cwd-independence holds. I confirmed
  git's glob crosses `/` (nested `internal/*/CLAUDE.md` matched `*CLAUDE.md`), so the pathspec
  genuinely scans the whole tree rather than the root only.
- **D9 command.** `cat CLAUDE.md AGENTS.md | grep -n '\[HARD\]' | sort -k2 | uniq -d -f1` — against
  `cat CLAUDE.md CLAUDE.md` it reports **4** duplicates, against a single copy **0**, and the root
  file carries exactly **4** `[HARD]` lines. Decides in both directions. One precision note the
  criterion does not need to fix: `-f1` skips `grep -n`'s line number *together with* the line's
  first token, so two lines differing only in a leading bullet compare equal — an error in the
  strict direction, which is the safe one for this criterion.
- **D1 premise still holds.** `go test -v ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$'` →
  `--- PASS`, `ok`, and **no token figure**. The executability note and the hoisted `t.Logf` are
  addressing a live condition, not a historical one.
- **D6 sweep.** `grep -rn 'cannot ship\|cannot help distributed\|per-user config'` over the SPEC
  directory returns 4 hits — `spec.md:354`, `design.md:96`, `research.md:24`, `acceptance.md:164` —
  and **every one is inside an explicit correction**. The retired premise survives nowhere as a
  bare assertion.
- **D4 artifact.** `probe-fixture.sh` is tracked (`git ls-files` returns it), builds the git-rooted
  three-level fixture, the 110 B-interval byte ruler past 40,000 B, and the fake `CODEX_HOME` — the
  construction iteration 1 said was missing.
- **Cited surfaces exist.** `.github/workflows/template-neutrality-check.yaml` (AC-AMC-022),
  `TestAlwaysLoadedTokenBudget_OverBudgetFails` ×2 in `token_budget_guard_test.go` (AC-AMC-016(a)),
  `AlwaysLoadedTokenBudget = 76000` at `token_budget_guard.go:32` with the "temporary raise" comment
  AC-AMC-020 retires.

---

## Rulings on the five questions put to me

**1. `AC-AMC-019`'s ±1,000 tolerance — wide enough to re-admit the parking hazard?**
The tolerance is not the problem; **the ratio is**. ±1,000 against ~75,000 is 1.3 %, and since
`estimateTokens` is deterministic the check needs no slack at all — the tolerance is tight.
But `AC-AMC-019` checks `constant ≈ N × (1 + ratio)` where **`ratio` is whatever the constant's own
comment says**, chosen by the same actor, in the same edit, and pinned by nothing. `REQ-AMC-013`
says "a stated headroom allowance" — unbounded. `design.md:232` ("The original constant carried
~15 %") is advisory prose, not a requirement, and `AC-AMC-020` requires only that *a* ratio be
stated. Worked counterexample: achieved N = 60,000, stated ratio 25 %, constant 75,000. `AC-AMC-018`
passes (≤ 75,000, figure quoted); `AC-AMC-019` passes exactly (|75,000 − 75,000| = 0); `AC-AMC-020`
passes (ratio stated). The constant sits at the ceiling with 15,000 tokens of silent regrowth
tolerance — **the exact scenario `design.md:235-237` says `AC-AMC-019` closes**. The hazard was
relocated from the constant to a free variable, not eliminated. This is N1.
(Partial mitigation, stated in fairness: `REQ-AMC-013`'s second sentence — "A constant set …
at the ceiling … does not satisfy this requirement" — forbids the outcome in prose. A human reviewer
would catch it. The mechanical criterion would not, and mechanical enforcement is what D1 asked for.)

**2. `AC-AMC-024` scoping — trivially passing, always failing, or actually discriminating?**
**It discriminates.** Not trivially passing: it demands three specific facts (project scope works
only under `trust_level = "trusted"`; the untrusted first session at 32,768 B is binding;
non-application is silent), and I verified all three are present in `spec.md` §D.8 and
`REQ-AMC-018`. Not always failing: all four occurrences of the retired premise are framed as
corrections, so the criterion passes today, and I confirmed that by sweep rather than by reading
§D.8 alone.
Two real scope defects remain (N7, minor). Its `Given` scopes to §D.8 + `REQ-AMC-018` while its
`Then` says "a bare assertion of it **anywhere** fails" — anywhere in the SPEC? the repo? the
shipped docs? No scan scope, and no command, so the sweep I ran is my construction rather than the
criterion's. And by a literal reading, `AC-AMC-024`'s own sentence is an occurrence of the retired
premise, making the criterion a candidate violator of itself.

**3. `REQ-AMC-006` circularity — honest documentation or self-referential nothing?**
**Honest, and I rule it acceptable.** Three reasons. It is falsifiable (delete or hollow §D.6 and it
fails), it is decidable by reading a *different* section than the one that states it — so it is not
viciously self-referential — and the property it binds is real: a decision recorded with its revival
condition is revisable on stated grounds, one recorded without it gets re-litigated. `AC-AMC-012`'s
parenthetical correctly disclaims the future SPEC's compliance as out of scope.
The honest criticism, which I record as optional rather than blocking: its truth value is fixed at
authoring time and never changes during run phase, so it contributes nothing to the Definition of
Done as a gate. That is a documentation invariant sitting in the requirement layer. **But iteration
1's D8 explicitly offered this as remedy (b)** — "add a criterion checking that *this* SPEC's body
records the revival condition" — and the author took the option the verdict sanctioned. Ruling
against it now would be moving the goalposts. Renumbering-avoidance is a legitimate secondary
reason; it is not the primary one, and the fix stands on its own merits.

**4. `AC-AMC-009` binding both copies — does anything assume they can diverge?**
**Nothing assumes divergence, and I checked directly**: `grep -rn 'byte-identical\|verbatim
cp\|identical'` over the SPEC directory returns nothing. The one place the two copies *are* required
to differ — `REQ-AMC-016` neutrality stripping, `plan.md` M6 "The mirror is not a verbatim `cp`" —
only removes bytes, so it cannot push the mirror over a ceiling the live file clears. Compatible.
The defect is the opposite of what was asked about (N5, minor): the both-copies binding is asserted
in `AC-AMC-009`, `plan.md` M5 ("It binds both the live root `AGENTS.md` and its template mirror"),
and §H's Definition of Done — but **not** in the requirement layer. `REQ-AMC-004` says "The root
`AGENTS.md` shall not exceed 24,576 B" and `REQ-AMC-007` names only "the root `AGENTS.md`"; and
`design.md:182`'s guard formula reads `codex chain guard : len(AGENTS.md) ≤ agentsMDCeilingBytes` —
one file. So an AC enforces a property no requirement states, and the design's own guard sketch
covers half of it. `plan.md` M5 carries the instruction, so run phase will not miss it; this is a
consistency polish, not an enforceability hole.

**5. Did a new untestable criterion enter with AC-019…024 or REQ-AMC-018?**
No criterion is *un*testable. Three are decidable-but-soft and are counted in the Testability
derivation above: `AC-AMC-021` ("`make build` has been run" names no observable — the natural check,
run it and assert `git diff --exit-code` on the mirror, is not stated), `AC-AMC-024` (N7), and
`AC-AMC-002` (N6, from the D4 fix). `AC-AMC-019` is testable but ineffective (N1). `REQ-AMC-018` is
sound Unwanted form and is covered by `AC-AMC-024`.
One form slip: `AC-AMC-016` opens "The negative path has two dimensions, and only one of them needs
new coverage" — a prose statement, not Given-When-Then; the (a)/(b) sub-parts underneath are
well-formed. Its (b) also carries implementation direction ("extends the existing table-driven test
in the same file"), which is HOW inside a criterion. Both optional.

---

## Defects Found

**N1 — `AC-AMC-019` does not close the vacuity it claims to close; the headroom ratio is a free
variable.** `acceptance.md:138-141`, against `spec.md:209-213` (`REQ-AMC-013`, "a stated headroom
allowance"), `acceptance.md:143-145` (`AC-AMC-020`, ratio merely "stated"), and `design.md:232`
(the ~15 % figure, advisory prose only). Nothing pins the ratio, so `constant = N × (1 + ratio)`
is satisfiable at any constant ≤ 75,000 by choosing the ratio to fit — worked counterexample in
ruling 1. This is the surviving half of D1 and the single finding carrying the FAIL.
Severity: major — Class: blocking — Required fix: pin the ratio rather than the tolerance. Either
(a) fix it in `REQ-AMC-013` — "plus a headroom allowance of 15 % (± 2 pp), stated in the constant's
comment" — so `AC-AMC-019`'s right-hand side is determined by N alone; or (b) add a second bound to
`AC-AMC-019`: the stated ratio must itself be ≤ 20 %, so no ratio choice reaches the ceiling from a
substantially lower achieved figure. **(a) is preferable** — it needs no new AC, which matters
because the Tier L acceptance-criterion budget has exactly one slot left (24 of 25).

**N2 — `design.md:173` cites a renumbered criterion.** "The failure mode to watch is **duplicate
injection** … AC-AMC-012 tests exactly this." The delta repurposed `AC-AMC-012` to cover
`REQ-AMC-006`'s §D.6 record; duplicate injection is now `AC-AMC-013` (`acceptance.md:93-96`). A
reader following the design's own pointer lands on the wrong criterion. Introduced by the D8/D9
fixes; it is the only stale link the renumber left — I checked every `AC-AMC-\d{3}` reference in
`spec.md`, `plan.md`, and `design.md` (7 references; the other 6 resolve correctly).
Severity: minor — Class: blocking — Required fix: `design.md:173` → `AC-AMC-013`.

**N3 — `design.md` §5.2 precedes §5.1, and carries a paragraph belonging to neither.** §5.2 "Why the
singleton check is `git ls-files`…" sits at `design.md:197`, §5.1 "Ratchet" at `design.md:221`. The
D2 fix was inserted above the existing sub-section instead of after it. Compounding: the paragraph
"**The guard is blocking, not advisory** …" (`design.md:215-219`) sits under the singleton-check
heading while discussing the byte guard's failure mode — it belongs to §5 proper or a §5.3.
Severity: minor — Class: optional — Required fix: move §5.2 below §5.1 (or renumber), and lift the
blocking-vs-advisory paragraph out of the singleton sub-section.

**N4 — `spec.md` §D has no §D.5.** Headings run D.1, D.2, D.3, D.4, **D.6**, D.7, D.8, D.9. The
delta renamed the former §D.5 ("Residual unmeasured items") to §D.9 and inserted D.6-D.8 without
closing the hole. No dangling reference results — `grep -rn 'D\.5'` over the SPEC directory returns
nothing, so this is cosmetic rather than broken.
Severity: minor — Class: optional — Required fix: renumber D.6-D.9 down by one, or note the gap.

**N5 — `AC-AMC-009` enforces a mirror ceiling that no requirement states.** `acceptance.md:56-60`
and §H bind both copies to 24,576 B; `REQ-AMC-004` (`spec.md:170-172`) and `REQ-AMC-007`
(`spec.md:184-185`) name only "the root `AGENTS.md`", and `design.md:182`'s guard formula checks one
file. `plan.md:133-134` does carry the both-copies instruction, so run phase is not misdirected.
Severity: minor — Class: optional — Required fix: extend `REQ-AMC-004` to "the root `AGENTS.md` and
its template mirror", and make `design.md:182`'s formula name both operands.

**N6 — `AC-AMC-002`'s verb overstates what the script does.** "Then it … **reports each recorded run
against its expected result**" (`acceptance.md:19-23`). `probe-fixture.sh` builds the fixture, then
`cat`s a heredoc **listing** the eight probe invocations with their expected markers; it neither
executes them nor compares. Running the script yields no verdict — the operator copies and eyeballs.
The criterion is satisfiable under "lists alongside" and not under "checks against". D4's stated
remedy (record the construction, cite it instead of the scratchpad) is fully met; this is residual.
Severity: minor — Class: optional — Required fix: either soften to "prints the fixture path and each
recorded run with its expected result", or have the script run the probes and print PASS/FAIL per
row — the latter also makes `plan.md` §C's third pre-flight checkbox self-deciding.

**N7 — `AC-AMC-024` has no scan scope and no command, and its own text is an occurrence.**
`acceptance.md:161-165`: `Given` scopes to §D.8 + `REQ-AMC-018`, `Then` says "a bare assertion of it
**anywhere** fails this criterion". Detail in ruling 2.
Severity: minor — Class: optional — Required fix: name the scope and the command — e.g. "`grep -rn
'cannot ship' .moai/specs/SPEC-AGENTS-MD-CANON-001/` returns only occurrences inside an explicit
correction (this criterion's own statement excepted)".

**N8 — `AC-AMC-016` mixes forms and carries implementation direction.** `acceptance.md:106` opens
with a prose statement rather than Given-When-Then; (b) at `acceptance.md:119-121` prescribes "extends
the existing table-driven test in the same file … does not introduce a parallel harness" — HOW,
inside a criterion. The intent (don't duplicate the fixture) is right and is already stated in
`plan.md` M5 and `REQ-AMC-008`, where it belongs.
Severity: minor — Class: optional — Required fix: none required for PASS.

**N9 — four requirements are labelled `(Event-detected)`.** `REQ-AMC-003`, `-007`, `-012`, `-014`.
The GEARS pattern name is **Event-driven**. The clause forms are correct ("When …, the … shall …"),
so MP-2 is unaffected; only the label is non-canonical. Not introduced by this delta and not raised
at iteration 1 — recorded now because D10 established that a wrong pattern label is worth fixing.
Severity: minor — Class: optional — Required fix: relabel to `(Event-driven)`.

---

## Regression Check

Every iteration-1 defect was re-verified against the `e5699d4fc` tree; results in the
per-finding table above. **No defect persisted unchanged**, so the stagnation clause does not fire,
and no fix reintroduced a defect it was meant to remove.

Two regression checks worth naming because they were the plausible failure modes of *this
particular* delta:

- **Renumbering damage.** Repurposing `AC-AMC-012` and appending 019-024 was the highest-risk edit.
  I re-extracted the full ordered AC heading list (001-024, contiguous) and every `AC-AMC-\d{3}`
  reference in the other three artifacts. Exactly one reference went stale (N2) out of seven.
- **Iteration-1's named recurring shape** — "a measurement landing after the artifact that depends
  on it needs a sweep of the dependents, not just of the section that names it." The D6 fix was the
  test of that lesson, and **it passed**: the P4 measurement was propagated to `design.md` §2,
  `design.md` §6's rejected-alternatives row, `spec.md` §D.8, `REQ-AMC-018`, `AC-AMC-024`, and
  `research.md`'s premise table — six surfaces, with no bare survival of the retired premise
  anywhere in the directory. The lesson held.

**Score movement: 0.69 → 0.83.** An improvement, so the STOP-on-regression clause does not fire and
iteration 3 is available within the Tier L ceiling of 3.

### The v0.3.1 delta (`cd6e12459` → `e5699d4fc`) — in scope and already judged

Because that delta landed mid-audit and I mislabelled the baseline, I state explicitly which
findings rest on it rather than leaving it inferable:

- **`AC-AMC-016`'s (a)/(b) split** — I read the split form, not the pre-split one. No finding was
  formed against the old `AC-AMC-016` and then carried; **N8** was derived against the split text
  (`acceptance.md:106` opening prose, `:119-121` implementation direction) and stands.
  Substantively the split is an **improvement**: it names `TestAlwaysLoadedTokenBudget_OverBudgetFails`
  by file, quotes its output, and states that a fresh fixture there would be the second measurement
  path `REQ-AMC-008` forbids. I verified the cited test independently —
  `go test -v ./internal/config/ -run 'OverBudgetFails'` → `--- PASS` on both `over-budget` and
  `under-budget` subtests, so (a)'s quoted result is real, not asserted.
  One property worth naming without grading it as a defect: (a) asserts that an already-green test
  stays green, so it passes on the entry tree and verifies nothing this SPEC changes. That is a
  legitimate regression-plus-anti-duplication role — the same role `AC-AMC-015` plays — and the note
  under the criterion says so itself ("proves the guard *fires* … proves nothing about whether the
  constant was derived or parked"). Consistent with my ruling on `REQ-AMC-006`: fixed-at-authoring
  criteria are acceptable when they bind something real and say what they do not cover.
- **`design.md` §5.2** — the three-way `find` / path-scoped `find` / `git ls-files` comparison is
  the new section, and **N3** (§5.2 sits above §5.1, and carries the blocking-vs-advisory paragraph)
  is a finding *against this delta*, not a pre-existing one. Its content is sound: I independently
  reproduced the subdirectory-scoping hazard it names — a bare `git ls-files` from `internal/spec`
  scopes to that subtree, which is exactly the silent-PASS it warns about, and the `:(top)` form
  returns the identical 6-file list from either directory.
- **`plan.md` M5's "do not author a new fixture" instruction** — read at `plan.md:139-142`, cited in
  the per-finding table, consistent with `AC-AMC-016(b)` and `REQ-AMC-008`. No finding.
- **`progress.md`** — record only; no criterion depends on it.

**`spec.md` is unchanged across the delta**, which I verified by hash rather than by trusting the
statement: `de2f0cce8b56…` at `e5699d4fc` equals my pinned value. So every `spec.md:NNN` citation in
this report — the whole must-pass sweep, `REQ-AMC-001`…`018`, §A.4, §D.1, §D.6-§D.9 — resolves
against the judged state.

**Net effect of the delta on the verdict: none.** No score moved, no finding was added or withdrawn.

---

## Recommendation

**FAIL, marginally, on one blocking finding.** This is a materially different SPEC from iteration 1:
eleven of twelve findings fully resolved, every fix verified by re-executing the mechanism rather
than by reading the claim, and the three criteria that could not be executed at all (`AC-AMC-010`,
`AC-AMC-002`, the old `AC-AMC-012`) now run and decide. The one surviving gap is that the ratchet's
new derivation criterion checks an equation whose right-hand side the author of the constant also
chooses.

Iteration 3, in order:

1. **N1 — pin the headroom ratio in `REQ-AMC-013`.** Prefer amending the requirement over adding a
   criterion: the AC budget has one slot left (24 of 25), and a fixed ratio makes `AC-AMC-019`
   determined by N alone, which is what D1 asked for.
2. **N2 — `design.md:173`: `AC-AMC-012` → `AC-AMC-013`.** One token.

That is the whole blocking set. Resolving both lifts Testability to ≥ 0.80 and the harmonic mean to
**0.867 → PASS**; N2 alone additionally clears the Traceability dock.

Optional, surfaced and left to the orchestrator's discretion — routing these into a revision is
**not** required to reach PASS, and they should not be bundled into the blocking pass by default:
N3, N4, N5, N6, N7, N8, N9. If any are taken, N5 and N7 are the two that repay the edit (one closes
a REQ/AC scope mismatch, the other makes a criterion self-deciding); the rest are hygiene.

---

## Evidence and Gaps

**Gaps (what I did not observe).** I did not run the full Go test suite (excluded by the dispatch);
I ran only `TestAlwaysLoadedTokenBudget`, to confirm D1's premise. I did not execute
`probe-fixture.sh` end-to-end — it requires `codex` on PATH and writes a fixture tree — so its
*correctness* is judged by reading, and only its existence, tracked status, and structure are
measured. I did not independently re-measure the ≈ 71,212-token worktree figure or the byte figures
in §A.1/§A.4; iteration 1 verified all of them against this same content and none was touched by
the delta. I did not re-audit sections outside the enumerated delta beyond the must-pass sweep.

**The tree moved again after this verdict was formed — `51fec2f25` is NOT judged here.** While this
report was being finalised, a third commit landed ("pin the headroom ratio in the requirement",
v0.3.2, 02:45:14), touching `spec.md` and `design.md`. **This report's verdict is pinned to
`e5699d4fc` and does not rule on it.** A bounded spot-check, offered as information for scheduling
iteration 3 and explicitly not a verdict:

- **N2 is fixed.** `design.md:173` now reads "AC-AMC-013 tests exactly this."
- **N1 is fixed at the requirement layer.** `REQ-AMC-013` now fixes the allowance at 15 % ± 2 pp
  (band 13 %-17 %) and states that a ratio outside the band fails the requirement even below
  75,000. My counterexample no longer survives: a declared 25 % is out of band, and an achieved
  60,000 admits only 67,800-70,200. The remedy taken is the one recommended — amend the
  requirement, add no criterion, leaving the AC budget at 24 of 25.
- **A residual of the same shape remains, and it is one line.** `acceptance.md` was **not** touched
  by that commit. `AC-AMC-019` still reads "the headroom ratio stated in the constant's comment"
  with no band check, and `AC-AMC-020` still requires only that *a* ratio be stated. So the
  requirement forbids an out-of-band ratio while no criterion detects one — the verification layer
  does not mirror the binding text. This is the N1 pattern one level down and in the permissive
  direction (unlike N5, which errs strict). Suggested amendment, in place, no new AC: have
  `AC-AMC-019` check that the stated ratio falls in the 13 %-17 % band **and** that the constant
  agrees with `N × (1 + ratio)` within ±1,000.

Ruling on `51fec2f25` requires an iteration-3 read against a tree that is actually frozen; three
commits landed here inside thirteen minutes.

**Residual risk.** `AC-AMC-019`'s counterexample is constructed, not observed — no run phase has
occurred, so the parking outcome is a demonstrated possibility rather than a demonstrated fact.
`REQ-AMC-013`'s prose forbids it, so a diligent run-phase actor would not hit it; the finding is
that the mechanical check does not enforce what the prose says. Separately, the freeze verification
covers the audit window only: all six artifacts were byte-identical to `e5699d4fc` at both start and
end, and I cannot speak to edits after this report was written. And the provenance error corrected
at the head of this report is a reminder about my own evidence rather than the SPEC's — a hash pin
survived it, a `git diff` did not.
