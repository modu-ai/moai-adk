# SPEC Review Report: SPEC-VERSION-STAMP-PREDICATE-001

Iteration: 2/2 (Tier M ceiling — `harness.plan_audit_tier_ceilings`)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.84** (Tier M PASS threshold 0.80) — delta from iter-1 **+0.05** (0.79 → 0.84)

Audit tree: worktree `.claude/worktrees/t392`, HEAD `9a3e2dabe`, branch `WT-version-stamp-predicate`
(`git rev-parse HEAD` / `git branch --show-current`, run in this session).

Reasoning context ignored per M1 Context Isolation. The orchestrator's framing was read only to
locate artifacts and to know which iter-1 defects the author claims to have answered; every
judgment below rests on the artifact files and on commands re-run here. Where I accepted an
author claim, I say so and name the command that made me accept it.

Scope: this re-audit is scoped to the iter-1 defect delta (D1-D8) plus a hunt for defects the
repair itself planted. The iter-1 "What I re-measured and found sound" list was not re-opened;
the author left it alone.

**Monotonicity:** the score rose. No STOP signal — the LEAN score-regression clause does not fire.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -n '^\*\*REQ-VSP-' spec.md` returns exactly
  REQ-VSP-001…014 at L275, 279, 287, 290, 294, 302, 306, 311, 316, 320, 325, 332, 336, 339.
  Sequential, no gap, no duplicate, uniform 3-digit padding. The three new requirements
  (012/013/014) extend the run rather than interrupting it.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX`
  in `spec.md` §4), not the verification layer; stated explicitly per MP-2's layer clause. All 14
  match a GEARS pattern. The three added this round: REQ-VSP-012 (L332) ubiquitous + canonical
  negative ("shall read its file population from…", "It shall not enumerate that population by
  walking the filesystem"); REQ-VSP-013 (L336) ubiquitous + event-driven ("When an entry does not
  resolve, the check shall fail…"); REQ-VSP-014 (L339) ubiquitous ("shall state who edits the
  registry…"). The rewritten REQ-VSP-005 (L294) and REQ-VSP-008 (L311) both remain ubiquitous.
  The `Given/When/Then` entries in `acceptance.md` are AC-layer and are graded under Group 4.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md` L2-14) plus optional `tier: M`. `version: "0.2.0"` quoted semver, bumped from 0.1.0;
  `status: draft`; `created`/`updated` both ISO `2026-09-01`; `priority: Medium`;
  `lifecycle: spec-anchored`; `tags` comma-separated string. No rejected snake_case alias.
  Independently re-run here: `moai spec lint` → rc=0, `0 error(s), 1096 warning(s)`, and
  `grep -c 'VERSION-STAMP-PREDICATE'` over all 1100 output lines → **0**. No lint finding of any
  severity against this SPEC.
- **[N/A] MP-4 language neutrality** — scoped to this repository's own Go code, its docs-site and
  its release stamps. No multi-language tooling claim. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -oE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md`
  → itself + `SPEC-VERSION-STAMP-GUARD-001`. `grep -n '^status:'` on that SPEC → `5:status:
  completed`. Not in {retired, superseded, archived}. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -c 'NEEDS CLARIFICATION' -r` over the SPEC directory
  → 0 in all four artifacts, rc=1. `research.md` absent, correct for Tier M.

The firewall is clean. The verdict rests on the rubric and on the defect list.

---

## Category Scores (rubric-anchored)

| Dimension | Score | iter-1 | Rubric Band | Evidence |
|---|---|---|---|---|
| Clarity | 0.75 | 0.75 | 0.75 band — minor ambiguity a reasonable engineer resolves | iter-1's two clarity defects are gone: §2.0 (L100-146) states the population as an explicit decision with the not-taken branch measured; §3 (L225-245) is a 6-row enumeration over three named axes with its population stated. What replaces them is narrower but real: the pinned replacement text (`plan.md` §E) and seven AC judgment greps are written in Korean against a target document that contains **zero** Korean characters (`grep -cP '[가-힣]' .moai/docs/version-management.md` → 0 over all 104 lines) — N3. |
| Completeness | 0.90 | 0.70 | 1.0 band on the mechanical criteria, discounted for the coverage-map gap | All required sections present; frontmatter complete; §7 now carries **six** `### Out of Scope — <topic>` H3 sub-headings each with `-` bullets, including the new "추적되지 않는 파일". The maintenance contract iter-1 called absent is now §6.1 + REQ/AC-VSP-014, with the numstat measurement that iter-1 said was one command away actually taken. §8 gains R-8 (index/worktree skew) and rewrites R-4/R-5. Discount: §3's coverage grid — the document's central map, with a "누가 잡는가" column — never names REQ-VSP-013 (`grep -n 'REQ-VSP-013' spec.md` → L22, 336, 359, 393, 456, 587; **nothing in L225-246**) — N4. |
| Testability | 0.70 | 0.70 | between 0.50 and 0.75 — every AC is binary and pre-pins its failure string, but four of the fourteen judgment mechanisms are weaker than the AC claims, and one region has no AC at all | The central iter-1 defect is fixed and I verified the fix by measurement (see § What I verified by measurement). Against it: AC-VSP-014's (b) and (c) carry no judgment command at all (`acceptance.md` L331-350); AC-VSP-011's (d) regex evades 6 of 7 plausible closed-count phrasings, measured (L258); AC-VSP-010's (b) denylist omits `ls-tree`/`show-ref`/`describe`/`for-each-ref` (L219-224); AC-VSP-012's both-direction test prescribes mutating the git index (L285-288), contradicting the pure-core design §6 establishes for exactly this purpose; and no AC asserts the sweep's reach — N1. |
| Traceability | 1.00 | 1.00 | 1.0 band | `spec.md` §4.1 (L344-359) maps REQ-VSP-001…014 to AC-VSP-001…014 one-to-one; `grep -n '^### AC-'` on `acceptance.md` returns exactly 14 headings in the same order; `moai spec lint` emits **no** `CoverageIncomplete` for this SPEC ID (0 lines matching, over 1100). No orphan AC, no uncovered REQ. The three added REQs each got an AC. |

Arithmetic mean = (0.75 + 0.90 + 0.70 + 1.00) / 4 = **0.8375 → 0.84**, above the Tier M
threshold of 0.80.

---

## What I verified by measurement (not by reading)

The task asked for measurement rather than acceptance on three load-bearing claims. All three
verify. Recording them because a verdict that does not separate the verified from the defective
is not usable.

**1. The D1 attribution claim is TRUE — all four constants came from the `git grep` family, and
`git grep` reads exactly the tracked set.** Re-run at `9a3e2dabe`:

```
git grep -lF v3.1.3 -- .                                       → 155
git grep -lF v3.1.3 -- . <six-group deny list>                 → 34
git grep -lF v3.1.3 -- .moai/reports                           → 57
git grep -lF v3.1.3 -- .moai/specs                             → 58
```

Same commands, same values as the SPEC records. 28 is derived arithmetically (34 − 6 goldens),
not measured separately, which is stated in `acceptance.md` L52-57. So the repair is a genuine
**attribution**, not a re-derivation dressed as one: the numbers did not move because the
population they were always measured against is the one now declared. The author's claim holds.

**2. `git ls-files` is consistent with the narrowed REQ-VSP-010 and with t388's precedent.**
REQ-VSP-010 (L320) now forbids history, any ref other than the checked-out one, and network.
`git ls-files` reads `.git/index` — not the object store, not a ref, no network. t388's property
was "queries no repository object"; the index is not an object, so the precedent survives, and
the SPEC states the narrowing as a cost rather than hiding it (§2.0 L138-146). Accepted.

**3. All three replacement assertions are bump-invariant — measured on both bump commits.**

```
git show --name-status --format= eba919e44   → 6 lines, all M   (no A, no D)
git show --name-status --format= 61921f1ba   → 7 lines, all M   (no A, no D)
```

A bump adds and deletes no file, so `registry == 28` and `stamp == 7` cannot move. For the third:

```
git grep -lF v3.1.4 61921f1ba -- <the 7 stamp paths>   → all 7 present   (clean bump: green)
git grep -lF v3.1.3 eba919e44 -- <the 7 stamp paths>   → 6; docs-site/hugo.toml ABSENT
```

So `sweep ⊇ stamps` stays green across a correct bump and fires naming `hugo.toml` at the exact
historical commit this whole card lineage exists to catch. The assertion is invariant to the
event and sensitive to the defect. This is the strongest thing in the repair.

**4. The D2 diagnosis — that the removed constant was wrong from the start — is TRUE.**

```
git grep -lF v3.1.4 61921f1ba -- . <six-group deny list> | wc -l   → 7
```

At the tree immediately after a bump the sweep matches exactly **7**, not 28. The iter-1 design
would have failed with `sweep matched=7 expected=28` at every release. Subtraction was the right
repair, not enlargement. (This same measurement is what makes N1 below concrete.)

**5. AC-VSP-011(d) works in the two directions the author names.** Run here:

```
old wording  「여전히 잡지 못하는 것이 셋이다.」    → 1 match, rc=0
new wording  「…것들이 있다. 적어도 다음이 남으며…」 → 0 matches, rc=1
```

The both-directions claim is true as stated. N2 is about what else the regex lets through.

**6. D5 genuinely dissolves rather than being papered over.** `grep -n 'v3\.1\.3'
.moai/docs/version-management.md` → exactly one hit, L90, and L90 is verbatim the sentence
`plan.md` §E replaces. So M5 does drop the doc from the sweep. Under the repaired design that
violates nothing: registered `prose`, exempt from REQ-VSP-004, and REQ-VSP-002 defines the
registry as a superset. The file survives, so REQ-VSP-013 passes too. `plan.md` §D.0 records this
as reasoning rather than as an ordering rule, which is the correct home for it.

**7. The author's two corrections to my iter-1 report are both right.**

```
grep -rlF --exclude-dir=.git v3.1.3 . | wc -l   → 162   (iter-1 said 161)
comm -23 <fs> <tracked>                          → exactly 7 paths, all this card's artifacts
find .../.claude/worktrees -maxdepth 1 -mindepth 1 -type d | wc -l   → 183   (iter-1 said 181)
```

Both corrections accepted. Neither changes an iter-1 conclusion; the walk-vs-tracked argument
holds at 7 as it did at 6.

**8. Tier M derivation survives the repair.** SSOT re-read at `spec-workflow.md:138-141` and
`:145-150`: Tier M = 5-15 files, 300-1000 LOC, ceilings 16/16, threshold 0.80. The §5 table
enumerates 1 + 2 + 6 + 1 = **10** files and I confirmed the three new requirements add no
eleventh: 012 changes how the driver obtains its population (same file), 013 adds an `os.Stat`
per entry (same file), 014 adds a paragraph to a document already in the table. 14 REQ / 14 AC are
under both ceilings. Note the SPEC understates its own case — it calls ~350 LOC "S/M 경계"
(L395), but 350 is inside Tier M's 300-1000 band, so **both** axes say M, not just the file count.
Conclusion sound, derivation sound.

---

## Defects Found

Ordered by severity. Each is marked **resolved** / **carried-over** / **new-this-round**.

### N1 — the repair removed the only assertion on the sweep's reach, and post-bump the surviving assertion has zero slack
**Artifact:** `spec.md:294-301` (REQ-VSP-005), `spec.md:449-467` (§6.1);
`acceptance.md:98-133` (AC-VSP-005) — **Severity: major — Class: blocking — new-this-round**

This is the defect the repair planted, and it is the direct answer to the task's question
"can the check now pass while sweeping nothing?"

Not *nothing*: an empty sweep is caught, because `sweep ⊇ stamps` then names all 7 missing paths.
But the sweep can lose everything **except** the 7 stamps and no assertion fires. Enumerating
what constrains the sweep from below after the repair: `registry == 28` and `stamp == 7` are
properties of the registry literal and say nothing about whether the sweep ran; `sweep ⊇ stamps`
bounds it over 7 paths; REQ-VSP-003 and AC-VSP-002's superset clause bound it from *above*.
Nothing bounds the region between 7 and 34.

That region is not incidental — it is exactly the region REQ-VSP-003 exists to police. Today it is
27 of the 34 files the SPEC measured, plus the entire not-yet-existing unregistered surface. A
plausible bug reaches it: REQ-VSP-007 makes the exclusion set a literal enumeration a human edits,
and one over-broad group (say someone adds `docs-site/`) removes 20 prose pages and every future
unregistered docs-site candidate from the sweep while all four assertions stay green.

And the failure mode is not merely reachable, it is the **steady state after every release**.
Measured above: `git grep -lF v3.1.4 61921f1ba -- <deny list>` → **7**. Immediately after each
bump the legitimate sweep equals the stamp set exactly, so at that moment the check cannot
distinguish "the sweep reached the whole repository and found 7" from "the sweep reached almost
nothing and found 7". The author's own `acceptance.md:355` states the stake: "하나라도 공허하면
이 SPEC이 아무것도 세우지 않는다". No AC covers this vacuity.

I am not asking for the removed constant back — §6.1's argument against it verifies (measurement 4
above). A bump-invariant lower bound is available and I measured it: `git ls-files | wc -l` →
**10,048**, and it cannot move on a bump because both bump commits are all-`M` (measurement 3).

**Required fix:** add to REQ-VSP-005 an assertion on the **population the driver returned**, not on
the sweep result — a floor on `len(population)`, held as its own constant and not derived from the
sweep or the registry, failing with the observed size. That closes the collapsed- and
truncated-driver cases (including the `git ls-files`-run-from-`internal/cli` cwd-scoping case,
which is otherwise unaddressed anywhere in the artifacts) while staying invariant to a bump. Add
the matching AC with a synthetic short-population RED. Optionally also state in §8 that the
exclusion literal is now the check's single unguarded blast radius.

### N2 — AC-VSP-011(d)'s regex is calibrated to the one mutant it was shown, and 6 of 7 plausible closed-count phrasings evade it
**Artifact:** `acceptance.md:258` (the (d) judgment command), `acceptance.md:259-262`
(the both-directions clause) — **Severity: major — Class: blocking — new-this-round**

The regex is
`(잡지 못하는 것이|catch:).*(둘|셋|넷|다섯|여섯)이다`.
The mutant the AC proposes feeding it is the **original wording it was written against**, so the
both-directions test is guaranteed to succeed by construction. Proving a guard is non-vacuous is
not the same as choosing the mutant that shows it. I ran the regex against seven plausible ways an
M5 author could write a closed count:

```
1 잡지 못하는 것은 셋이다.                  → no match
2 다음 세 가지를 잡지 못한다.               → no match
3 여전히 잡지 못하는 것이 셋 있다.          → no match
4 아래 세 경우가 남는다.                    → no match
5 Three cases remain uncaught.              → no match
6 It does not catch: three cases.           → no match
7 여전히 잡지 못하는 것이 여섯이다.         → MATCH
```

Only the original phrasing with a different numeral is caught. Lines 5 and 6 matter most: the
target document is English (N3), so the English forms are the *likely* ones and the regex is blind
to both — including line 6, which contains the literal `catch:` the alternation was meant to cover,
because nothing after it is a Korean numeral.

The guard is therefore a check that passes on almost every way the defect could recur. Since
AC-VSP-011 is the AC that carries the entire D4 repair, this is the D4 repair being asserted by a
mechanism that mostly does not assert it.

**Required fix:** replace the phrase-shaped regex with a structure-shaped one, or drop (d)'s regex
and make (b) do the work: require the enumeration to be preceded by an explicit open-set marker
whose presence is asserted positively (a grep expecting ≥1, not 0), so a file-not-found or a
rewritten paragraph fails rather than passes. Then choose mutants the author did not write —
at minimum the two English forms above.

### N3 — the pinned replacement text and four judgment greps are Korean; the target document contains zero Korean
**Artifact:** `plan.md:243-263` (§E replacement text, entirely Korean);
`acceptance.md:249` (`미등록`), `acceptance.md:254` (`추적`), `acceptance.md:258` ((d)),
`acceptance.md:342` (AC-VSP-014's `범프` grep) — **Severity: major — Class: blocking —
new-this-round**

Measured: `grep -cP '[가-힣]' .moai/docs/version-management.md` → **0**, over all 104 lines. The
file is English throughout, including the paragraph §E replaces (L90).

This creates a fork with no stated resolution. `plan.md:264` states
"교체 문안은 `plan.md` §E에 미리 못박혀 있다. run-phase가 문안을 새로 짓지 않는다" — the run-phase
is forbidden from re-authoring the text. But the text as pinned is Korean prose destined for an
English document. The implementer must either paste Korean into an English document (a register
break, and one that leaves the rest of the section in English), or translate — which is
re-authoring, which the plan forbids. Four of AC-VSP-011/014's judgment greps additionally become
unsatisfiable unless Korean is written in.

Scoping the defect precisely: of AC-VSP-011's six (b) greps, four search language-neutral
identifiers (`prose`, `_test.go`, `releaseDate`, `system.yaml.tmpl`) and survive either way. The
defect is the two Korean-word greps in (b), plus (d), plus AC-VSP-014's `범프`, plus the pinned
prose itself.

I did not check this axis in iter-1, so I do not claim it entered this round; the D4 repair
expanded the grep list from two items to six, which is what brought it into view.

**Required fix:** re-author §E's replacement text in English, matching the surrounding document's
register, and re-pin it in `plan.md` §E so the no-re-authoring rule stays enforceable. Rewrite the
four Korean-word greps against the English tokens the new text actually uses. State in §E that the
target document is English so the next editor does not re-introduce the mismatch.

### N4 — the 6-row coverage grid has a seventh state it does not carry, and row 4 reads as absolving it
**Artifact:** `spec.md:225-246` (§3 grid), specifically `spec.md:243` (row 4);
`spec.md:336` (REQ-VSP-013) — **Severity: minor — Class: blocking — new-this-round**

The task asked whether the six-row enumeration leaves a seventh case. It does, and the reason is
structural: the three axes are properties of a **file**, while REQ-VSP-013 is an assertion about a
registry **entry**. §3 states its population as "tracked files minus the six exclusion groups"
(L233-236), so a registry entry naming a file that does not exist has no file and falls outside
the grid entirely — while remaining inside the registry the grid is meant to map.

Read literally, row 4 (`registered / prose / no token → (정상 — 면제)`) covers the ghost-prose
state and marks it normal. It is not normal; it is the exact state REQ-VSP-013 was promoted this
round to catch, and §6.1 L456 says so. `grep -n 'REQ-VSP-013' spec.md` returns L22, 336, 359, 393,
456, 587 — **nothing between L225 and L246**. The grid's "누가 잡는가" column, which is the
document's coverage map, omits the assertion the repair just added.

A second latent state sits in the same seam: a registered `stamp` whose path falls inside an
exclusion group would never enter the sweep, so `sweep ⊇ stamps` would fail permanently. No such
entry exists today (I checked the 7 stamp paths against the six groups), so this is latent, not
live — but the grid does not predict it either.

**Required fix:** add a row or a note to §3 for the entry-without-file state, attributing it to
REQ-VSP-013, and qualify row 4 so it covers only entries whose file exists. One sentence stating
that the grid's rows are file-states and that entry-states are covered separately would close both.

### N5 — AC-VSP-010(b)'s forbidden-subcommand list is a denylist with holes, guarding a requirement stated as an allowlist
**Artifact:** `acceptance.md:219-224` — **Severity: minor — Class: blocking — new-this-round**

REQ-VSP-010 (L320-324) is an allowlist: the checked-out index and working tree only, one permitted
external invocation. AC-VSP-010(b) verifies it with a denylist of ten subcommands —
`show|log|rev-list|rev-parse|fetch|merge-base|cat-file|diff|branch|ls-remote`. Omitted, and each
of them reachable: `ls-tree` (reads a tree object from any ref), `show-ref`, `for-each-ref`,
`describe`, `worktree`, `status`. A driver written `git ls-tree -r HEAD --name-only` would read
history, violate REQ-VSP-010 in its own words, and pass (b) cleanly. A denylist fails silently on
what it omits; an allowlist fails only on what it does not enumerate.

(c) already does the allowlist's positive half — `grep -nc '"ls-files"'` → 1. The negative half is
the one to invert.

**Required fix:** replace (b) with an allowlist assertion: extract every quoted git subcommand
argument in the file and assert the resulting set is exactly `{"ls-files"}`. That is the same one
grep, and it has no omissions to enumerate.

### N6 — AC-VSP-012's both-direction test mutates the git index, contradicting the pure-core design §6 establishes for exactly this purpose
**Artifact:** `acceptance.md:285-288`; against `spec.md:471-476` (§6's [HARD] pure-core clause)
— **Severity: minor — Class: blocking — new-this-round**

The AC prescribes: create an untracked file carrying the authoritative token, run the check
(must pass), then `git add` it and run again (must fail naming it). `spec.md` §6 [HARD] separates
a pure judgment core from a thin population driver precisely so that REDs come from synthetic
inputs rather than from tree manipulation, and `plan.md` M2 builds five synthetic REDs on that
basis. AC-VSP-012 then reaches for the one mechanism the design exists to avoid — and reaches for
`git add` specifically, in a checkout the project's own doctrine treats as shared.

The design already affords the cheaper form: feed the pure core two synthetic populations, one
containing the path and one not, and assert the finding appears in exactly one. Same two
directions, no index mutation, no cleanup contract needed.

**Required fix:** restate the both-direction test as two synthetic populations against the pure
core. If a real-index observation is genuinely wanted as a one-off integration check, make it a
manual run-phase step with an explicit `git restore --staged` cleanup contract, and say so.

### D6 — the pre-declared delta is already falsified by the audit round it was written for
**Artifact:** `spec.md:170-175` (§2.3's 델타 bullet); `acceptance.md:373-375` (DoD) —
**Severity: minor — Class: optional — carried-over, partially resolved**

The iter-1 fix landed: the six counts are pinned to `9a3e2dabe`, labelled point-in-time, and the
table's staleness is now covered by a declared delta plus a re-derivation trigger. That is the
right shape and I credit it.

What did not land is immunity from the thing it declares. The delta names **7** artifacts and
predicts `.moai/reports` 57 → 60, sum 121 → 128, tree 155 → 162. This report is an eighth
artifact under `.moai/reports/t392/` and it contains the literal `v3.1.3`, so if it is committed
the true values are 61 / 129 / 163. The re-derivation obligation in §2.3 and the DoD line at
`acceptance.md:373` both cover this, so nothing breaks — but a number declared as a prediction and
falsified before run-phase begins is worth naming rather than leaving for the merge-tree
re-measure to absorb silently.

**Required fix:** optional. Either drop the specific post-commit figures and keep only the
re-measure obligation, or state the delta as "one row per artifact this card commits, count at
merge time" rather than as fixed integers.

### D3-residual — §3.1's unmeasured hole is honestly scoped, but one cheap adjacent measurement was available
**Artifact:** `spec.md:255-262` (§3.1 나), `spec.md:576-583` (§8 R-2) —
**Severity: minor — Class: optional — carried-over, substantially resolved**

The task asked whether the admission is honest scoping or an evasion. **It is honest scoping.**
The hole's size as the SPEC defines it — the number of genuine stamp sites currently classified
`prose` — requires applying REQ-VSP-002's human-judgment predicate to each entry and then
simulating a bump across the full suite. The SPEC says so and puts the cost out of scope with a
reason. That is not avoidance.

But one adjacent measurement was available and would have bounded the *silence* of the hole over
today's registry. I ran it:

```
# the 21 prose registry paths, derived from the 34-file list
git grep -lF version-management.md -- '*.go'   → internal/cli/version_sync_list_test.go  (reads the
                                                  Version Stamps list, not a version token)
git grep -lF 'docs-site/content' -- '*.go'     → scripts/convert-nextra-to-hextra/main.go,
                                                  scripts/docs-version-snapshot/main.go
                                                  (neither is a test; both are release-time tools)
```

So **zero of the 21 prose entries carry a consumer test that a missed bump would break**. By
§3.1's own framing — the hole equals the count of stamp sites with no consumer test — a
misclassification anywhere in today's prose set would be maximally silent. That sharpens the
admission rather than contradicting it, and it costs two greps.

**Required fix:** optional. Add one line to §3.1 (나) or §8 R-2 recording the measurement above and
stating that it bounds today's silence, not the future hole.

---

## Regression Check — iter-1 defects D1-D8

| iter-1 | Status | Evidence |
|---|---|---|
| **D1** — sweep reads a population the constants were never measured against | **RESOLVED** | §2.0 (L100-146) decides the population as the tracked set, adds REQ/AC-VSP-012, narrows REQ-VSP-010 with the cost stated, and measures the not-taken branch (162 fs / 155 tracked / 183 sibling worktrees / 144 token files in one of them). I re-ran all four constants' commands and confirmed the attribution claim (measurement 1). `git ls-files` is consistent with the narrowed prohibition and with t388's precedent (measurement 2). AC-VSP-010 was split into (a)/(b)/(c) so permitting `ls-files` cannot readmit `exec.` to the core — the split is correct in intent; see N5 for (b)'s coverage. |
| **D2** — a bump breaks the check wholesale, no maintenance contract | **RESOLVED**, with N1 planted | The diagnosis verified: post-bump sweep = 7 (measurement 4). All three replacement assertions verified bump-invariant by measurement (measurement 3). §6.1 supplies who / when / what-the-check-says-between, and REQ/AC-VSP-014 binds the documentation. §8 R-4 rewritten to the measured numstat result. The removal is right; what it stopped asserting is N1. |
| **D3** — 2×2 overstates; classification predicate undefined | **RESOLVED**, with N4 residual | §3 redrawn as a 6-row enumeration over three axes with its population stated (L233-236) and with the explicit statement that exclusion-group silence means "not looked at". REQ-VSP-002 (L279-286) now carries the predicate: "classified `stamp` when a version bump must rewrite that file, or a check that reads it breaks when the bump does not". §3.1 names both uncovered cases and bounds the misclassification one. Residual: the grid omits the entry-without-file state (N4). |
| **D4** — closed "셋" enumeration | **RESOLVED**, with N2/N3 on the guard | `plan.md` §E is now an open form ("적어도 다음이 남으며, 이 열거가 전부라는 뜻은 아니다") with six items including the `_test.go`-inline site and the misclassification site. AC-VSP-011 grew to six (b) greps plus a (d) closed-count check, and (d) works in the two directions the author names (measurement 5). The wording repair is complete; the guard on it is weak (N2) and in the wrong language (N3). |
| **D5** — M5 deterministically breaks M4's GREEN | **RESOLVED — dissolved** | Verified independently (measurement 6): the doc's single `v3.1.3` is at L90, exactly the replaced sentence, and with the sweep-count constant gone the doc's exit from the sweep violates no assertion. `plan.md` §D.0 records this as reasoning rather than as an ordering rule, which is correct — fixing it by ordering would have left the same defect to reappear at the next documentation edit. |
| **D6** — six volatile counts, no staleness contract | **PARTIALLY RESOLVED** | Pinned to `9a3e2dabe`, labelled point-in-time, delta pre-declared, re-derivation trigger written for the `changelog-pages` clause. The trigger and the pin are the substance and they landed. See D6 above for the falsified figures. |
| **D7** — no ghost-entry guard | **RESOLVED** | Promoted to REQ-VSP-013 (L336) + AC-VSP-013 (L316-330) with a named failure string. **I accept the author's severity dispute.** Under the repaired superset design a deleted `prose` path violates nothing else — the sweep-count assertion that would have caught it as a bare `27` no longer exists — so 013 is now the only assertion covering it. Promoting it from optional to a requirement was correct, and my iter-1 "minor / optional" grading was made against a design that no longer exists. |
| **D8** — requirement names an implementation identifier | **RESOLVED** | REQ-VSP-008 (L311-315) is now behavioural: "shall render a fixed version token independent of the build-time value, restoring the original value when the test ends". `grep -n 'version\.Version' spec.md` → L36, 37, 38 only — all in §1 narrative, none in §4. The identifier moved to `plan.md` M4 and AC-VSP-008 cites the `version_test.go:180-186` precedent. |

**Stagnation:** none. No defect appears unchanged across both iterations. Every iter-1 blocking
defect (D1-D5) is resolved, three of them with a measurement the author had to take rather than
argue.

---

## Recommendation

**PASS-WITH-DEBT at 0.84.** The must-pass firewall is clean, the score clears the Tier M threshold
of 0.80, the score moved upward, and all five iter-1 blocking defects are genuinely resolved — I
verified the three load-bearing ones by measurement rather than by reading the repair's account of
itself. The D2 repair in particular is the right kind: the author found the constant was wrong
rather than fragile and subtracted it, and the measurement backing that (post-bump sweep = 7) is
reproducible.

The debt is six blocking defects, none critical, none a must-pass failure. Three are major:

1. **N1** — one requirement clause and one AC. Add a bump-invariant floor on the **population** the
   driver returned (measured basis: `git ls-files | wc -l` → 10,048), held as its own constant.
   This is the only defect that touches the check's design, and it is the one to fix first.
2. **N3** — re-author `plan.md` §E's replacement text in English and re-pin it, then rewrite the
   four Korean-word judgment greps. Artifact-only; no design change.
3. **N2** — replace AC-VSP-011(d)'s phrase-shaped regex with a structure-shaped assertion, and pick
   mutants the author did not write (the two English forms above are the ones that matter).

Three are minor: **N4** (one row or one sentence in §3), **N5** (invert (b) from denylist to
allowlist — same single grep), **N6** (restate AC-VSP-012's both-direction test as two synthetic
populations, which the pure-core design already affords).

Two optional findings (**D6** figures, **D3-residual** measurement) are surfaced for the
orchestrator's discretion and need not gate anything.

**Ceiling reached.** Tier M allows 2 plan-audit iterations and this is iteration 2. A third
iteration is not available without an explicit operator override. Per the LEAN max-iterations
clause the orchestrator should present the operator with: (1) accept PASS-with-debt and carry N1-N6
into run-phase as enumerated debt — my recommendation, since N1 is one clause and the remaining
five are artifact edits; (2) reduce scope; (3) override the cap for an iteration 3. There is no
score regression, so no STOP escalation is required.

If option (1) is taken, N1 should be fixed **before** M1 rather than deferred, because M1 fixes the
constants into the registry literal and adding a population-floor constant afterwards means
revisiting the most-expensive-to-reverse milestone.

---

**Self-demonstrating note (D6).** This report contains the literal `v3.1.3`. It is an eighth
artifact under `.moai/reports/t392/`, so if it is committed the §2.3 delta becomes
`.moai/reports` 57 → 61, sum 121 → 129, tree 155 → 163. Run-phase should expect that, not read it
as drift in the SPEC's original arithmetic.
