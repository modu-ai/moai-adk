# SPEC Review Report: SPEC-AGENTS-MD-CANON-001

Iteration: 5/6
Audited document: **version `0.3.5`** (frontmatter `spec.md:4`) at commit **`f10c8691f`**
Verdict: **FAIL**
Overall Score: **0.90** (above the Tier L threshold — the FAIL is carried by one blocking finding, not by the score)

Reasoning context ignored per M1 Context Isolation. The dispatcher's own verification was re-executed
rather than repeated; every figure below is attributed to a command run in
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t82`.

---

## §0. The tree moved mid-audit — reconciled, not restarted

The document moved twice relative to what I was told. This iteration's dispatch pinned `0.3.4` at
`707f06ef6` and stated `git status --short .moai/specs` was empty at freeze. At audit time it was
not:

```
$ git status --short .moai/specs
 M .moai/specs/SPEC-AGENTS-MD-CANON-001/acceptance.md
 M .moai/specs/SPEC-AGENTS-MD-CANON-001/plan.md
 M .moai/specs/SPEC-AGENTS-MD-CANON-001/spec.md
```

Rather than stop, I audited the pinned `0.3.4` and read the uncommitted delta in full. That delta
has since been committed as **`f10c8691f`, version `0.3.5`**, and it is the state judged here. I did
**not** restart, because the `0.3.5` content is the same text I had already read — verified, not
assumed:

```
$ git diff --stat 707f06ef6 f10c8691f
 acceptance.md | 10 ++++++---
 plan.md       |  8 +++++++
 progress.md   | 10 +++++++++
 spec.md       | 26 ++++++++++++++++------
 4 files changed, 44 insertions(+), 10 deletions(-)
$ git status --short .moai/specs        # now clean, no output
```

Four hunks, each one the dispatcher named: §C.4's duplication-not-relocation paragraphs, M1's block
quote, `AC-AMC-007`'s framing note, `progress.md` (record only). The `spec.md` hunk is byte-identical
to the uncommitted text I had already characterised, plus the HISTORY row.

**The "no requirement and no criterion changed" claim — checked rather than accepted, and it
holds**, with one precision the phrasing understates:

```
$ diff <(grep '^\*\*REQ-AMC-' 0.3.4) <(grep '^\*\*REQ-AMC-' 0.3.5)   →  identical, 18 lines
$ grep -o '^\*\*AC-AMC-[0-9]*' 0.3.5/acceptance.md | sort -u | wc -l  →  24
```

REQ 18, AC 24, and no `REQ-AMC-*` text moved — the sole `spec.md` hunk sits in the §C.4 preamble,
between the section heading and `REQ-AMC-013`, touching no requirement body. `AC-AMC-007` **did**
change textually, but only its trailing gloss: I diffed the criterion itself and the
`Given … When … Then …` deciding condition is byte-identical, the old parenthetical
("Negative-path criterion: …") replaced by two sentences carrying the same content. No *deciding
condition* changed anywhere, which is what the claim needs to mean. Not a finding against you.

Two process notes, neither a finding against the document. The freeze claim was an **unobserved
verification claim** (`verification-claim-integrity.md` §1.1 surface 1) — asserted, not observed —
and this is its third occurrence. The diagnosis offered for it is correct and has doctrinal backing:
`cross-session-messaging.md` § An idle notice is a scheduling hint states that an idle notice
establishes *when to look* and nothing about what the evidence says, and that treating it as
completion evidence is itself an unobserved completion claim. The fourth condition being added is
the right repair. Nothing further is owed here, and the offer to restart as iteration 6 is declined
— reconciliation cost one commit-range diff and three greps, and it preserves the last iteration
under the ceiling for a revision that fixes F1.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -c '^\*\*REQ-AMC-' spec.md` → 18;
  `grep -o 'REQ-AMC-[0-9]*' | sort -u | wc -l` → 18. Sequential 001-018, no gap, no duplicate,
  consistent 3-digit padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-AMC-*` in
  `spec.md` §C); the `AC-AMC-*` entries are Given-When-Then, which is the correct verification-layer
  form and is graded under Group 4. All 18 REQs carry `shall` / `shall not` with a GEARS pattern
  label and body. Ubiquitous ×8 (001, 006, 008, 009, 011, 013, 015, 017), Unwanted ×6 (002, 004,
  005, 010, 016, 018 — all `shall not`), Event-driven ×4 (003, 007, 012, 014 — all
  `When …, the … shall …`).
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:1-15`): `id`, `title`, `version` (quoted `"0.3.4"`), `status` (`draft`), `created` /
  `updated` (`2026-08-22`), `author`, `priority` (`P1`), `phase`, `module`, `lifecycle`
  (`spec-anchored`), `tags` (comma-separated string). Plus `tier: L`. No rejected snake_case alias.
- **[N/A] MP-4 Section 22 language neutrality** — the SPEC targets this repository's own harness
  surface; it enumerates no per-language tooling. `REQ-AMC-016`'s template-neutrality obligation is
  a content-class rule, not a language enumeration. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — one external reference,
  `SPEC-ALWAYS-LOADED-DIET-001`. `grep '^status:' .moai/specs/SPEC-ALWAYS-LOADED-DIET-001/spec.md`
  → `status: completed` — not in {retired, superseded, archived}, so no reconciliation is required.
  §A.5 reconciles it anyway ("closed … does not reopen it … inherits the guard and the pattern").
  No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → 0. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -c 'NEEDS CLARIFICATION'` → 0 in both `plan.md` and
  `research.md`.

`moai spec lint` → `0 error(s), 64 warning(s)`, exit 0; **zero findings against
`SPEC-AGENTS-MD-CANON-001`** (all 64 warnings are grandfathered-era `MissingExclusions` on other
SPECs). The dispatcher's lint claim is confirmed.

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|------:|-------------|----------|
| Clarity | 0.90 | 1.0 / 0.75 boundary | §C.4 now *states* the net-additive mechanism (`spec.md:231-238`) with the governing formula at `:240`; one input remains under-restated (F2, Arm B's baseline tree) |
| Completeness | 0.95 | 1.0 | HISTORY, §A Context, §B Goals, §C Requirements, §E Scope, §F Acceptance, four `### Out of Scope — <topic>` H3s each with `-` bullets (`spec.md:497-528`); §D.5 numbering gap is cosmetic |
| Testability | 0.75 | 0.75 | Nearly every AC names its deciding command; but `AC-AMC-015` and `AC-AMC-017` are **mutually unsatisfiable** (F1) — one is guaranteed to fail regardless of implementation quality |
| Traceability | 1.00 | 1.0 | All 18 REQs covered by ≥1 AC; all 24 ACs map to an existing REQ. Verified pairwise; no orphan, no uncovered REQ |

Arithmetic mean **0.90**. Stated for the record — the FAIL is carried by F1 under M6, not by the
score, exactly as in iteration 4.

---

## Answers to the five judgment questions

### 1. Is E1 closed, or has the hazard moved again? — **CLOSED.**

I re-derived every figure independently rather than reading the SPEC's:

```
$ python3 -c "print(24576//4, 71212+6144-66371, 75282+6144-66371, round((75000+1000)/1.13,1))"
6144 10985 15055 67256.6
```

All four match the document. `1.13 × 66,371 = 74,999.23 ≤ 75,000` — the bound is tight and correct.

I then checked each input the dispatch named as a candidate free variable:

- **The formula's inputs.** `stated surface + |AGENTS.md| − 66,371` (`spec.md:240`). `stated
  surface` is pinned by the table column header "Unextended surface" and by §A.1's measured 71,212;
  `66,371` is derived at `:227-229` from the 13 % floor and the 75,000 cap. Neither is free.
- **What counts as `|AGENTS.md|`.** Both readings are present and both are *bounded*: the ceiling
  case (6,144 tokens) is the table, and the authored size governs via the formula. The floor is
  also derivable — §D.1's optimistic case is 18,183 B ≈ 4,546 tokens — so `|AGENTS.md|` ranges over
  4,546-6,144 and the required cut over **9,387-10,985** (worktree) / **13,457-15,055**
  (integration). A range, not a free variable.
- **Which enumeration each figure is measured over.** `spec.md:231-232` states it explicitly and
  disclaims the §A.1 rows as unextended. §A.1 itself carries the pointer at `:52-53`
  ("**The enumeration does not yet include `AGENTS.md`**"), so the four-row table no longer reads
  as contradicting `REQ-AMC-013`.
- **Any other figure still computed over the unextended surface.** `grep -c '4,841\|8,911'` returns
  **two hits across all six artifacts at `0.3.5`**, both in explicit correction context and neither
  load-bearing: `spec.md`'s HISTORY row, and the new `plan.md` block quote's "4,841 → 10,985".
  (At `0.3.4` this was one hit; the second arrived with the M1 block quote. The arrow form is the
  legitimate one — it names the wrong figure only to retire it.) I also checked the neighbouring
  derived figures (24,576 / 32,543 / 14,360 / 8,192): all are byte-domain contract-layer figures
  measured over the `[HARD]` proxy, not over the always-loaded enumeration, so the correction does
  not propagate to them.

The hazard has not moved. What remains in this neighbourhood is F2, and it is a different quantity
(which *tree*, not which *enumeration*).

### 2. Does Arm B bind? — **Yes, but with one under-restated input (F2).**

Re-derived against `0.3.5`'s text, not carried over from the `0.3.4` reading. Applying the
iteration-4 test: does an actor who does the wrong thing fail a criterion, or merely contradict
prose?

**The `0.3.5` framing additions strengthen Arm B rather than weakening it**, and the mechanism is
worth naming because it is not obvious. Under `0.3.4`, an M1 actor whose projection exceeded 66,371
faced a return that read as a milestone failure — an incentive to choose optimistic projection
inputs, and the inputs at M1 are projections rather than measurements, so that incentive had
somewhere to act. `0.3.5` removes it by making the blocker return a *pass* of `AC-AMC-007`. That is
a correct reading of the criterion (its `Then` clause has always required the blocker, not the
clearance), and it closes an incentive gap I would otherwise have had to raise.

I checked the obvious way the new framing could backfire — that "M1 passes by halting" might be read
as "M1 is done, M2 may start". It cannot: `plan.md` still heads the section "**Both must clear before
M2 starts**", `AC-AMC-007`'s `Then` still requires "no file has been moved", and the retained
sentence still states that clearing Arm A alone is insufficient. Three independent barriers.

Arm B is carried by `AC-AMC-007`, which names it as one of two trip conditions and states that
clearing Arm A alone is insufficient. An actor who projects, exceeds 66,371, and proceeds anyway
fails `AC-AMC-007` outright — the criterion is a negative-path criterion with a named deliverable
(a blocker naming the shortfall in tokens, with no file moved). That binds.

The gap is the projection's **baseline tree**. Arm B says "project the post-diet always-loaded
surface … against 66,371 tokens" without naming which tree the pre-diet baseline comes from, and the
two candidates differ by 4,070 tokens (71,212 vs 75,282 — a 37 % larger cut). An actor projecting
from the card worktree can clear Arm B and still fail `AC-AMC-018` / `AC-AMC-019` at M5 — the
discovered-at-M5 shape Arm B exists to prevent.

`0.3.5` **slightly aggravates this** without changing its class. The new M1 block quote is now the
only place at the point of use that quotes a cut figure, and the figure it quotes is the **worktree**
one: "roughly doubled the minimum cut (4,841 → **10,985** tokens)". §C.4 names 15,055 as the figure
"against the state the ratchet is actually measured on", so an actor reading M1 top-to-bottom now
meets 10,985 at exactly the moment Arm B asks for a projection. The fix note below is updated
accordingly.

I classify this **optional, not blocking**, and the reason is mechanical rather than charitable:
`plan.md` §C **pre-flight** already requires "Integration-branch surface measured (`release/vX.Y.Z`,
branch + ahead-count recorded), so the ratchet has a real baseline rather than a worktree-local one"
— a gate that lands *before* M1, so the correct input is in hand when Arm B runs. §B known-issue 2
states "A ratchet proposed from the worktree figure alone would be meaningless", and §C.4 names
15,055 as the figure "against the state the ratchet is actually measured on (`REQ-AMC-014`)". The
plan supplies the pinned input one section earlier; Arm B simply does not restate it.

### 3. Is the "ceiling case versus governing formula" framing sound? — **Yes.**

It does not let an actor choose a convenient cut, because **the cut is not a criterion**. Nothing is
graded against 10,985 or 15,055. The binding check is `REQ-AMC-013` / `AC-AMC-018`: the achieved
figure is *measured* over an enumeration that includes the real `AGENTS.md` (`AC-AMC-017` asserts
the enumeration contains it). An actor who authors a smaller `AGENTS.md` gets a smaller measured
surface automatically — the formula describes the measurement rather than substituting for it. The
ceiling case exists to size M1's work pessimistically, which is the correct direction.

One caveat, folded into F2: at **M1** the projection is not yet a measurement, so a convenient
`|AGENTS.md|` assumption is available there. Arm A constrains it from above (24,576 B) and
`REQ-AMC-001`'s completeness requirement from below, so the room is the 4,546-6,144 band, not
unbounded.

### 4. Does the SPEC state a reachable goal? — **Yes, and I computed it rather than assuming it.**

This is the question the iteration was convened for, so it gets the arithmetic. The `0.3.5` delta
touches the prose around these figures but **not the figures themselves** — §C.4's table, the
66,371 bound, and the governing formula are byte-identical across `707f06ef6` and `f10c8691f`
(the sole hunk sits above them, in the duplication-not-relocation paragraphs). The computation below
therefore stands unchanged, and I re-derived it against the `0.3.5` text rather than carrying it.

The output style is out of scope (§E.3, §H DoD "the output style was not edited by this SPEC"), so
the cuttable surface is rules + `CLAUDE.md`:

| Quantity | This worktree | Integration state |
|---|---:|---:|
| Total measured surface | 284,850 B (71,212 tok) | ≈ 301,128 B (75,282 tok) |
| Less output style (untouchable) | −61,706 B | −61,706 B |
| **Eligible surface** | **223,144 B** | **≈ 239,422 B** |
| Less `[HARD]` contract (must stay always-loaded, `REQ-AMC-002`) | −32,543 B | −32,543 B |
| **Eligible non-contract material** | **190,601 B** | **≈ 206,879 B** |
| Required cut (ceiling case) | 43,940 B (10,985 tok) | 60,220 B (15,055 tok) |
| **Cut as a share of eligible non-contract material** | **23.1 %** | **29.1 %** |

Against §D.4's stated precedent — `SPEC-ALWAYS-LOADED-DIET-001` reduced `goal-directive.md` by
**72 %** of its always-loaded footprint with no obligation moved off the surface — a 23-29 %
aggregate reduction of relocatable material is **well inside demonstrated range**. §B goal 2,
`REQ-AMC-004`'s 24,576 B ceiling, and M1's two arms are mutually satisfiable.

So the corrected arithmetic does **not** make the target unreachable. It makes it materially harder
than the card assumed — roughly double the original cut — and the honest reading is that Arm B has a
real chance of firing at the integration baseline, which is exactly why Arm B is the right addition
and why `0.3.5`'s "a blocker here is a correct outcome" framing is well judged.

### 5. Regression over iterations 1-4 — spot re-executed, not re-read.

| Origin | Finding | Status | Evidence (re-executed) |
|---|---|---|---|
| iter1 | `AGENTS.md` singleton check must use tracked files, not `find` | **RESOLVED** | `git ls-files --full-name ':(top)*CLAUDE.md' ':(exclude,top)internal/template/templates/*' \| wc -l` → **6**; `find . -name CLAUDE.md` → **7**. `AC-AMC-010`'s stated 7-vs-6 figures are exactly right |
| iter1 | Byte-guard criteria must be executable | **RESOLVED** | `TestAlwaysLoadedTokenBudget` emits only via `t.Errorf` (`token_budget_guard_test.go:67`) — no `t.Logf` anywhere in the file. The executability note's premise is true, and M5's first task addresses it |
| iter1 | `AC-AMC-016` must cite the existing over-budget fixture | **RESOLVED** | `TestAlwaysLoadedTokenBudget_OverBudgetFails` exists at `:77` with over-budget and under-budget subtests, as cited |
| iter1 | Integration branch needs a discriminator | **RESOLVED** | `REQ-AMC-014` + `AC-AMC-018` both require `git rev-parse --abbrev-ref HEAD` and `git rev-list --count main..HEAD` |
| iter2 N1 | Headroom ratio pinned | **RESOLVED** | `REQ-AMC-013` fixes 15 % ±2 pts → 13 %-17 % band (`spec.md:259-264`) |
| iter2 N2 | Stale `AC-AMC-012` cross-ref in `design.md` §4 | **RESOLVED** | `grep -n 'AC-AMC-01[23]' design.md` → one hit, `:173` "AC-AMC-013 tests exactly this". No residue |
| iter3 D1 | `AC-AMC-019` must read the band | **RESOLVED** | `acceptance.md` checks the ratio against the band **and** the ±1,000 agreement; both halves stated load-bearing |
| iter3 D2 | Enumeration must include `AGENTS.md` | **RESOLVED (and its premise re-verified)** | `alwaysLoadedSurface()` (`token_budget_guard.go:107-133`) carries exactly three fixed slots — `CLAUDE.md`, the output style, `MEMORY.md`; `AGENTS.md` absent. Bound into `REQ-AMC-008`, `REQ-AMC-013`, `AC-AMC-017`, `plan.md` M5 |
| iter3 D3 | 66,371 stated in §C.4 | **RESOLVED** | `spec.md:227-229` |
| iter3 D4 | Version / HISTORY current | **RESOLVED at `0.3.4`** | `version:` matches the last HISTORY row at `HEAD`. (The uncommitted `0.3.5` also matches its own row — the provenance rule is holding) |
| iter3 D5 | `AC-AMC-018`'s measured state defined | **RESOLVED** | "post-diet state, defined as one where `AC-AMC-021` passes … so the measurement follows M6 rather than landing mid-diet" |
| iter4 E1 | §C.4 cut computed over a surface excluding `AGENTS.md` | **RESOLVED** | See judgment 1 |
| iter4 E2 | No milestone tests the diet against the ratchet ceiling | **RESOLVED** (F2 is a residual on it, not a reopening) | `plan.md` M1 Arm B + `AC-AMC-007` |
| iter4 E3 | ±1,000 tolerance widens the ceiling by 885 | **RESOLVED** | `spec.md:254-257`, noted and deliberately not relaxed |

No regression. No stagnation: nothing carried unchanged across three iterations.

---

## Defects Found

**F1. `REQ-AMC-008` and `AC-AMC-015` are mutually unsatisfiable — the enumeration extension cannot
land without editing an expected-behavior assertion.**
`internal/config/token_budget_guard_test.go:130`, `:198` — Severity: **critical** — Class:
**blocking**.

*Unaffected by the `0.3.5` delta, checked rather than assumed:* `diff` of `AC-AMC-015` and of
`REQ-AMC-010` across `707f06ef6` and `f10c8691f` returns identical in both cases, and neither
Go test file is a SPEC artifact. The finding stands verbatim at the judged commit.

`plan.md` M5 requires adding `AGENTS.md` to `alwaysLoadedSurface()` **as a fourth fixed slot**, and
fixed slots are appended unconditionally, whether or not the file exists on disk
(`token_budget_guard.go:127-132`; the doc comment states it, and `measureAlwaysLoaded` handles
absence by contributing 0 tokens, not by dropping the entry). So `len(surface)` goes from
`rules + 3` to `rules + 4` the moment the slot is added — independent of whether `AGENTS.md` has
been authored yet.

Two existing assertions encode the 3:

```go
// token_budget_guard_test.go:130-132
wantTotal := wantRuleCount + 3 // + CLAUDE.md + moai.md + MEMORY.md fixed slots
if len(surface) != wantTotal {
    t.Errorf("surface has %d entries, want %d (= %d no-paths: rules + 3 fixed surfaces)", ...)

// token_budget_guard_test.go:197-199
// Enumeration: 1 no-paths: rule + 3 fixed slots = 4.
if len(surface) != 4 {
    t.Errorf("surface len = %d, want 4 (1 no-paths: rule + 3 fixed)", len(surface))
```

Both pass today, so the break is attributable to this SPEC's change, not pre-existing:

```
$ go test -v ./internal/config/ -run 'TestAlwaysLoadedSurfaceEnumeration|TestMeasureAlwaysLoaded'
--- PASS: TestAlwaysLoadedSurfaceEnumeration (0.01s)
--- PASS: TestMeasureAlwaysLoaded_WithMemory (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.431s
```

Both must become `+4` / `5`. But `AC-AMC-015` reads: "Then it is green, **and no expected-behavior
assertion was edited to accommodate the change**", and `REQ-AMC-010` forbids changing "any existing
test's expected behavior". So the run-phase actor faces two exits and **both fail a criterion**:

- extend the enumeration and update the two counts → fails `AC-AMC-015`;
- leave the enumeration alone → fails `AC-AMC-017`, and re-opens iteration 3's D2 defect entirely.

This is not a prose contradiction that an actor could route around by reading carefully — it is a
guaranteed criterion failure on every implementation path, which is a stronger failure than the
ordering hazard iteration 4 rejected. It also sits on the SPEC's single most load-bearing mechanism:
the enumeration extension is what makes the ratchet honest.

**Required fix** (one clause, in `acceptance.md` `AC-AMC-015`): carve out surface **cardinality**
from the no-edit rule, naming the two assertions. For example — "an assertion whose expected value
is the cardinality of the always-loaded surface is updated by `REQ-AMC-008`'s extension and is
exempt; the exemption covers the expected count only
(`token_budget_guard_test.go` `wantRuleCount + 3` → `+ 4`, and the temp-tree `want 4` → `want 5`),
never a behavioral expectation. Every other assertion remains under the no-edit rule." A matching
half-sentence in `REQ-AMC-010` keeps the requirement layer consistent.

**F2. M1's Arm B does not name the tree its projection is baselined on.**
`plan.md` §E M1 (Arm B), `acceptance.md` `AC-AMC-007` — Severity: **minor** — Class: **optional**.

The two candidate baselines differ by 4,070 tokens (71,212 worktree vs 75,282 integration), a 37 %
difference in the required cut. An actor projecting from the worktree can clear Arm B and still fail
`AC-AMC-018` at M5. Not blocking because `plan.md` §C pre-flight already mandates measuring the
integration-branch surface *before* M1, §B known-issue 2 states the worktree figure alone is
meaningless, and §C.4 names 15,055 as the operative figure — the input is supplied, just not
restated at the point of use.

**Suggested fix**, two parts, both small: eight words in Arm B — "…project the post-diet
always-loaded surface, **baselined on the integration-branch figure recorded at pre-flight**, …" —
and the same qualifier in `AC-AMC-007`; plus, in the new M1 block quote, quote **15,055** or both
figures rather than 10,985 alone, so the figure a reader meets at the point of use is the one the
ratchet is measured against.

**F3-F8. The six long-standing optional findings** (§5.2 ordering, the §D.5 numbering gap,
`probe-fixture.sh` printing rather than asserting, `AC-AMC-021`'s "make build has been run",
`AC-AMC-016`'s prose/GWT mix, the four "(Event-detected)" labels) — Severity: **minor** — Class:
**optional**, unchanged.

I re-checked each and **none has become blocking**; nothing changed to warrant routing them. Two
notes for the record: the §D.5 gap persists (`grep '^### D\.'` → D.1-D.4, D.6-D.9), and the
"(Event-detected)" label is a naming nit only — all four clause bodies are genuine GEARS
Event-driven form (`When …, the … shall …`), so MP-2 is unaffected.

---

## Recommendation

**FAIL**, on F1 alone. The aggregate 0.90 clears the Tier L threshold; the verdict is carried by a
blocking consistency defect, per M6.

**The smallest change reaching PASS** is the single `AC-AMC-015` carve-out quoted in F1 — one
sentence in `acceptance.md`, optionally mirrored as a half-sentence in `REQ-AMC-010`. Nothing else
in the document requires touching. F2's eight-word qualifier is worth folding in while editing,
but it is not what the verdict turns on.

**Is PASS-with-debt defensible on F1?** Defensible, but I do not recommend it. The argument for it is
real: updating a cardinality constant when the surface legitimately gains a slot is fixture
bookkeeping, not a relaxed expectation, and a competent run-phase actor would recognise that. The
argument against is what this SPEC's own history demonstrates — deferring it converts a one-sentence
plan-phase edit into a run-phase judgment call about whether breaching a stated criterion is
acceptable, made at M5, by whoever is holding the card, under pressure to close. That is the precise
class of decision this document has had to correct in four consecutive iterations. Spend the
sentence now.

**If a future iteration returns PASS, run-phase inherits this residual risk:**

1. **Arm B is likely to fire at the integration baseline** (judgment 4). That is the mechanism
   working, not a failure — but the card should expect a blocker return from M1 and should not read
   it as a stall.
2. **The `[HARD]` proxy is a bracket, not a point** (§A.4). Every downstream figure inherits its
   error in both directions; the 8,192 B reserve is what absorbs it, and trading against that
   reserve is a decision to state.
3. **The probe is single-platform** (§D.9): macOS, `codex-cli` 0.147.0. A smaller upstream default
   on another OS or version would be caught only by re-probing.
4. **The 66,371 bound is conservative by ~885 tokens** (§C.4, `(75,000+1,000)/1.13 = 67,256.6`) —
   deliberate, and it should not be spent as slack.
5. **F2 survives as a documentation-only residual** if it is not folded in: correct behavior then
   depends on the run-phase actor reading `plan.md` §C before M1 rather than on Arm B saying so.

---

## Verification appendix

All commands run in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t82` (`pwd` confirmed before the
first measurement and re-anchored per command; no `cd` into a nested `internal/*/` tree or into
`internal/template/templates/`). Pinned artifacts extracted with
`git show HEAD:.moai/specs/SPEC-AGENTS-MD-CANON-001/<f>.md`.

| Claim | Command | Observed |
|---|---|---|
| Judged tree | `git rev-parse --short HEAD` | `f10c8691f` (version `0.3.5`) |
| Tree now clean | `git status --short .moai/specs` | no output |
| Freeze had been violated | `git status --short .moai/specs` (before `f10c8691f`) | 3 modified files |
| Delta bounded to 4 files | `git diff --stat 707f06ef6 f10c8691f` | 44 insertions, 10 deletions across spec/plan/acceptance/progress |
| REQ text invariant | `diff` of `grep '^\*\*REQ-AMC-'` across both commits | identical, 18 lines |
| AC-AMC-007 deciding condition invariant | `diff` of the `Given…Then…` span across both commits | identical; only the trailing gloss differs |
| REQ count | `grep -c '^\*\*REQ-AMC-' spec.md` | 18 |
| AC count | `grep -o '^\*\*AC-AMC-[0-9]*' acceptance.md \| sort -u \| wc -l` | 24 |
| No stale figure load-bearing | `grep -c '4,841\|8,911'` over all artifacts | 2, both in `wrong → right` correction form |
| Arithmetic | `python3 -c "print(24576//4, 71212+6144-66371, 75282+6144-66371, round((75000+1000)/1.13,1))"` | `6144 10985 15055 67256.6` |
| Lint | `moai spec lint` | `0 error(s)`, exit 0, no finding for this SPEC |
| Clarification gate | `grep -c 'NEEDS CLARIFICATION' plan.md research.md` | 0, 0 |
| D8 | `grep -c syscall spec.md` | 0 |
| D7 | `grep '^status:' .moai/specs/SPEC-ALWAYS-LOADED-DIET-001/spec.md` | `completed` |
| F1 premise (slots) | `sed -n '107,133p' internal/config/token_budget_guard.go` | 3 fixed slots, appended unconditionally |
| F1 premise (assertions pass now) | `go test -v ./internal/config/ -run 'TestAlwaysLoadedSurfaceEnumeration\|TestMeasureAlwaysLoaded'` | both PASS, `ok … 0.431s` |
| `AC-AMC-010` figures | `git ls-files … \| wc -l` vs `find . -name CLAUDE.md \| wc -l` | 6 vs 7 |

**Gaps (not observed):** the full Go test suite was not run (excluded by dispatch); the integration
branch `release/vX.Y.Z` was not checked out, so its 75,282-token figure is carried from
`spec.md` §A.1 / `REQ-AMC-014` rather than re-measured, and the integration-state row of judgment
4's table is derived from it rather than measured directly; `design.md` and `research.md` were read
for the Tier L input contract and cross-reference regression but not audited clause-by-clause;
`probe-fixture.sh` was not executed.
