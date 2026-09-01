# SPEC Review Report: SPEC-IGNORED-EVIDENCE-CITATION-001

Iteration: 2/2 — **Tier M ceiling reached** (`harness.plan_audit_tier_ceilings`: S=1, M=2, L=3)
Verdict: **FAIL**
Overall Score: **0.78** (harmonic mean) — Tier M PASS threshold **0.80**
Score movement: iter1 **0.74** → iter2 **0.78**. **Monotonic improvement, no regression** — no STOP
signal fires.

Reasoning context ignored per M1 Context Isolation. The coordinator's iter2 claims were treated as
hypotheses and re-executed; where they did not reproduce, that is recorded in the final section.

Audit tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t381`, branch `WT-ignored-evidence-cite`,
HEAD `3f03d9c36` (re-read at audit time, unchanged from iter1). Scope: the enumerated iter1 defect
delta plus a regression check, per the iteration-2 contract.

---

## The headline: D1 is genuinely fixed

I executed **all ten MUST commands, both structural checks, and both companion commands** as written,
in this worktree. **Every one ran.** None was refused. This was the whole of iter1's critical defect
and it is closed by measurement, not by assertion:

| Criterion | Command ran? | Observed |
|---|---|---|
| AC-IEC-001 primary | YES | 4 files listed — **RED**, exactly the four claimed |
| AC-IEC-001 companion | YES | `1` for each of five files |
| AC-IEC-002 | YES | `1`, `1` — **RED** on the first |
| AC-IEC-003 | YES | `5` |
| AC-IEC-004 | YES | `…zeroexec_test.go:1`, `…20260710.html:1` |
| AC-IEC-005 | YES | `1`, `0` — **RED** |
| AC-IEC-006 | YES | empty, `exit=0`; positive control sums to **12** |
| AC-IEC-007 | YES | empty, `exit=0`; `ls` → No such file |
| AC-IEC-010 | YES | `exit=1`, `?? .moai/reports/t381/verify/`, state absent |
| AC-IEC-011 | YES | `go build …` → `exit=0` |
| AC-IEC-012 | YES | `1`, `1` — **RED** |
| S-1 / S-2 | YES | `1`,`6` / `5`,`1`,`1` |

The author's stated guard vocabulary is correct as measured: multi-path arguments to one call,
multi-file `grep -c`, pipes, and `echo "exit=$?"` all execute; `for`, `$(…)`, and subshells are
refused. Nothing was moved into a script file.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-IEC-001` … `REQ-IEC-011` at spec.md:140, 148, 156,
  162, 169, 177, 186, 194, 200, 205, 223. Sequential, no gap, no duplicate, uniform zero-padding.
  The two new ids extend the run correctly.
- **[PASS] MP-2 GEARS format compliance** — requirement layer only. The two new entries both match:
  REQ-IEC-010 (spec.md:207-208) "A tracked citation **shall not** name a line number without naming
  the tree that line number was resolved in" — Unwanted; REQ-IEC-011 (spec.md:225-227) "The repair
  **shall not** alter runtime behavior" — Unwanted. The nine iter1 requirements are unchanged and
  were verified compliant at iter1.
- **[PASS] MP-3 YAML frontmatter validity** — spec.md:1-16 still carries all 12 canonical fields with
  correct types and no rejected snake_case alias. (`version` is unchanged at `"0.1.0"`, which is a
  valid quoted semver and therefore passes MP-3 — the staleness is graded under Completeness as N4,
  not here.)
- **[N/A] MP-4 Section 22 language neutrality** — single-language scope, unchanged.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — verb re-executed. `SPEC-HIERARCHICAL-TEAM-001` →
  `status: completed`. `SPEC-EVIDENCE-CITATION-CANON-001` absent from `.moai/specs/` (D7-5 SHOULD),
  disclosed and re-confirmed unmerged. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall` = 0 on all three artifacts.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' …` → rc=1, no match.

All seven pass. The FAIL is driven by the rubric, specifically by two defective MUST criteria.

---

## Category Scores

| Dimension | iter1 | iter2 | Band | Why it moved (or did not) |
|---|---|---|---|---|
| Clarity | 0.85 | **0.85** | 0.75 | The §A/AC-IEC-001 contradiction is gone — §A now states the honest contract and the execution-form constraint (acceptance.md:13-31). Offset by N1: acceptance.md:101-102 asserts a resolution that is factually false about what its own command does. |
| Completeness | 0.95 | **0.90** | 1.0 | All sections intact; `grep -c '^### Out of Scope'` = 5. Lowered by N4: spec.md gained two requirements and four corrected sections while `version` stayed `"0.1.0"` and HISTORY gained no row. |
| Testability | 0.55 | **0.65** | 0.50→0.75 | The largest real gain: 12/12 commands verified executing (D1 closed), a RED-now column added and mostly accurate, honest demotion of the two unfalsifiable criteria. Held down because 2 of 10 MUST criteria are defective in kind — one unsatisfiable (N1), one green-today-but-recorded-RED (N2). |
| Traceability | 0.75 | **0.75** | 0.75 | Net flat. Gain: AC-IEC-012 now has REQ-IEC-010, which is non-circular (verified below). Loss: REQ-IEC-008's only coverage, the former AC-IEC-008, was demoted to non-gating S-1, orphaning it (N3). |

Harmonic mean of {0.85, 0.90, 0.65, 0.75} = **0.7753 → 0.78**. Threshold 0.80. **FAIL**, but up from
0.74 with no dimension regressing.

---

## Defects

### Surviving-as-new (introduced by the iter2 fixes)

**N1 — AC-IEC-001 is unsatisfiable under the SPEC's own chosen treatment, and its stated resolution
is false.** `acceptance.md`:L89 (command), L101-102 (the claim).
`grep -L` lists files that do **not** contain a marker; PASS requires the list to be empty, i.e. **all
five files must contain a marker word**. spec.md §C.3 selects treatment **(a)** for
`internal/cli/mcp_glm.go` — delete the path, keep the `t225` token — which adds no marker. Measured:

```
$ grep -ciE 'gitignored|does not resolve|not exported|machine-local|scratch' internal/cli/mcp_glm.go
0
```

So after treatment (a), `mcp_glm.go` still has zero markers, `grep -L` still lists it, and
AC-IEC-001 can never print empty. acceptance.md:101-102 states "`mcp_glm.go` reaches empty by
treatment (a) — deleting the citation removes the obligation." That is not what the command does:
`grep -L` has no knowledge of whether a citation is present, so deletion removes nothing from its
input. As written, AC-IEC-001 **forces treatment (b) onto `mcp_glm.go`**, contradicting §C.3.

Blast radius is not one milestone: `plan.md`:L107, L111, L118, L124 gate **four of the five repairs**
on AC-IEC-001. Severity: **critical**. Class: **blocking**.
*Required fix* — make the obligation conditional on a citation still being present, as two plain
invocations whose outputs are compared:

```bash
grep -l '\.moai/state/verify' internal/cli/mcp_glm.go internal/cli/audit_pin_live_test.go internal/hook/evidence_writer_zeroexec_test.go .moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt .moai/reports/template-skill-improvement-plan-20260710.html
grep -liE 'gitignored|does not resolve|not exported|machine-local|scratch' internal/cli/mcp_glm.go internal/cli/audit_pin_live_test.go internal/hook/evidence_writer_zeroexec_test.go .moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt .moai/reports/template-skill-improvement-plan-20260710.html
```

PASS when every path in the first output also appears in the second. Both forms run in this tree —
I verified `grep -l` and `grep -L` are equally accepted. Then correct the false sentence at
acceptance.md:101-102.

**N2 — AC-IEC-010 is GREEN today; the re-scope did not make it RED-capable, and the matrix records
the opposite.** `acceptance.md`:L58-67 matrix row ("RED today? **YES** (dir not yet populated)"),
L234-247 (the criterion), L245-247 (the observation). All three sub-checks pass **now**, before M5:

```
$ git check-ignore -v .moai/reports/t381/verify ; echo "exit=$?"      → exit=1   (PASS condition)
$ git status --short .moai/reports/t381/verify                        → ?? .moai/reports/t381/verify/   (non-empty)
$ ls .moai/state/verify                                               → No such file or directory  (PASS condition)
$ ls -la .moai/reports/t381/verify/                                   → 3 files already present
```

The directory is not "not yet populated" — it holds `spec-lint-full.txt`,
`spec-lint-targeted.txt`, `spec-lint-verification.md`. Worse, the middle check cannot do the job
asked of it: for an **untracked** directory `git status --short` collapses to a single `??` line
regardless of contents, so it cannot distinguish "M5 wrote twelve evidence files" from "one stale
file sits here". D3's stated purpose — make the criterion assert a property of what M5 produces —
is not achieved. Severity: **major**. Class: **blocking**.
*Required fix*: count the evidence files by name against the criteria that must produce them, e.g.
`ls .moai/reports/t381/verify/ac-iec-001.txt .moai/reports/t381/verify/ac-iec-002.txt …` with
PASS on exit 0, which is RED today and green only once M5 has actually written them. Correct the
matrix's RED-now cell either way.

**N3 — REQ-IEC-008 is now orphaned from the gating layer.** `spec.md`:L194-198 (the requirement),
`acceptance.md`:L299-306 (S-1, its only remaining coverage, explicitly "not MUST — recorded, not
gating"). The iter2 demotion correctly removed a vacuous criterion, but REQ-IEC-008 was that
criterion's parent, so a requirement now has no MUST AC. The other ten map cleanly (001→001,
002→002, 003→003, 004→004, 005→005, 006→006, 007→010, 009→007, 010→012, 011→011). Severity:
**major**. Class: **blocking**.
*Required fix*: demote REQ-IEC-008 to a §D constraint alongside its check — it states a
documentation property of this SPEC, not a property of the repair, so it is a constraint in the same
sense §D's other entries are. Do **not** re-promote S-1 to MUST: it would re-introduce exactly the
vacuity D3 removed.

**N4 — spec.md and plan.md were substantively revised without a version bump or a HISTORY row.**
`spec.md`:L4 (`version: "0.1.0"`), L22-24 (HISTORY carries only the 0.1.0 row) — yet spec.md gained
REQ-IEC-010 and REQ-IEC-011 and corrected §A.4, §C.1, §C.5, §C.7. `plan.md`:L4 likewise unchanged at
`"0.1.0"`. Only `acceptance.md` was bumped (L4 `version: "0.2.0"`, L6 `updated:`). On a card whose
subject is provenance integrity, a revision the document does not record is in-domain. Severity:
**minor**. Class: **blocking**. *Required fix*: bump both to `0.2.0`, add the iter2 HISTORY row.

**N5 — AC-IEC-004's recorded observation does not reproduce.** `acceptance.md`:L152 records
"`template-skill-improvement-plan-20260710.html:2`". Measured:

```
$ grep -cE '\.moai/reports/|\.moai/state/verify' internal/hook/evidence_writer_zeroexec_test.go .moai/reports/template-skill-improvement-plan-20260710.html
internal/hook/evidence_writer_zeroexec_test.go:1
.moai/reports/template-skill-improvement-plan-20260710.html:1
```

The value is `1`, not `2`: `grep -c` counts matching **lines**, and both matches sit on line 684.
The gate itself is unaffected (PASS needs `≥1`), but the recorded number is wrong. Severity:
**minor**. Class: **blocking** — the card's own §A promises each criterion's today-value is measured
in this tree, and this one was not.

**N6 — AC-IEC-007's recorded `rc` does not reproduce on this platform.** `acceptance.md`:L214 records
`No such file or directory (rc=2)`; measured `exit=1` (BSD `ls`). The PASS condition is worded as
"the file absent" rather than a numeric rc, so the gate holds. Severity: **minor**. Class:
**optional**. *Fix*: drop the parenthetical rc, or use `test ! -f … ; echo "exit=$?"`, which is
platform-stable.

**N7 — REQ-IEC-010 lacks the scope limiter its siblings carry.** `spec.md`:L207-208 binds "a tracked
citation" with no scope clause, whereas REQ-IEC-001 explicitly says "The in-scope lines are exactly
the five files enumerated in §C.1" (spec.md:146). As written REQ-IEC-010 reaches the whole
repository, which this card cannot deliver and AC-IEC-012 does not test (it checks one file).
Severity: **minor**. Class: **optional**. *Fix*: append the same §C.1 limiter.

**N8 — a soft modal inside normative text.** `spec.md`:L215-216 — "the repair **shall** prefer
removing it over updating it". "Shall prefer" is not binary-testable; nothing decides whether a
preference was honoured. Severity: **minor**. Class: **optional**. *Fix*: state it as guidance
outside the requirement body, or make it a hard clause ("shall remove the coordinate where it is not
load-bearing").

### Iter1 defects — regression check

| iter1 | Status | Evidence |
|---|---|---|
| D1 (critical) — 7 commands unrunnable | **RESOLVED** | All 12 commands executed in this tree; table above |
| D2 (major) — prose verdict on AC-IEC-001 | **RESOLVED in form, superseded by N1** | The verdict is now mechanical (`grep -L`), but the mechanism is unsatisfiable under treatment (a) |
| D3 (minor) — three vacuous MUSTs | **PARTIALLY RESOLVED** | S-1/S-2 honestly demoted with the reason stated (acceptance.md:293-297); AC-IEC-010's re-scope failed (N2) |
| D4 (minor) — false `status: draft` | **RESOLVED** | spec.md:68 and :477 now read `status: in-progress` **as read on 2026-08-31**; verified against t375 spec.md:5 |
| D5 (minor) — phantom `.codex` path | **RESOLVED** | Corrected at spec.md:370 and acceptance.md:204; `git ls-files '*manager-lead.toml'` → exactly one path, the template one |
| D6 (major) — three ACs with no REQ | **RESOLVED for AC-IEC-012, regressed for REQ-IEC-008** | REQ-IEC-010 added and binds AC-IEC-012; old AC-IEC-009 demoted (its parent was a section, so no loss); but see N3 |
| D7 (minor) — `467` underived | **RESOLVED** | spec.md:403 now shows `git grep -n '\.moai/state' -- . ':!*.md'` = 492; 492 − 25 = 467 |
| D8 (minor) — lane-8 claim unattributed | **RESOLVED** | spec.md:126-133 labels the decision relayed with its source and date, and separates the part verified directly (`manager-lead.md:150`) |
| D9 (minor) — brace-glob argued against 004 alone | **RESOLVED** | spec.md:318-327 now argues against the 004+005 pair; I verified t375 `spec.md:171` carries that pairing note verbatim |

**Stagnation check**: no defect persisted unchanged across both iterations. Every iter1 defect was
engaged; the two that regressed did so by introducing a *different* defect, not by being ignored.

---

## Is REQ-IEC-010 a real requirement or a circular one?

The coordinator asked this directly. **It is real, on two independent tests.**

1. **It constrains more than its AC tests.** REQ-IEC-010 binds any tracked citation naming a line
   number without its tree; AC-IEC-012 tests one instance (`gitignore:284` in `extract.txt`). A
   requirement that is *broader* than its criterion is under-covered, not circular — circularity
   would be the reverse, a requirement that merely restates the one check that exists. (The
   over-breadth is N7.)
2. **It has independent purchase on this SPEC's own text.** The card already obeys it where iter1
   showed it must: spec.md §A.4 cites t375 line numbers and names the tree
   (`.claude/worktrees/t375`, read 2026-08-31); §A.1 cites `.gitignore:298` and says "**in this
   tree** (measured)". Those are the requirement being satisfied, not a restatement of AC-IEC-012.

REQ-IEC-011 is a promotion of an existing §D constraint into a requirement so AC-IEC-011 has a
parent. That is legitimate — the constraint binds the repair independently and `go build` tests it —
though §D:L463-464 still carries the same clause with no cross-reference, which is harmless
duplication.

## Did the §D demotions remove vacuity or hide it?

**Removed, honestly.** acceptance.md:293-297 states the reason in the document itself: "no milestone
M1-M5 can falsify them, because no milestone edits `spec.md`. A criterion that cannot go RED at any
point in the run measures nothing." Both checks are retained with their commands, so nothing was
deleted to flatter a count — and I verified both still run and still print their expected values
(S-1 → `1`, `6`; S-2 → `5`, `1`, `1`). The one thing the demotion did not account for is the
requirement left orphaned behind it (N3).

## The pattern narrowing: correct, and it loses nothing

The coordinator's over-breadth finding holds, and my iter1 D2 remediation advice was wrong because
of it. Measured:

```
$ grep -c '\.moai/state'        .moai/reports/template-skill-improvement-plan-20260710.html   → 2
$ grep -c '\.moai/state/verify' .moai/reports/template-skill-improvement-plan-20260710.html   → 1
$ sed -n '529p' …  → "…with .moai/state/loop-verdict-<id>.json persistence…  Retain; this schema is a load-bearing cross-file co[ntract]"
```

Line 529 is mechanism prose about a schema the report recommends **retaining** — census class B2,
outside `/verify`, out of scope. A `.moai/state/` pattern sweeps it in and demands an edit §C.8
forbids. My iter1 fix proposal (compare `.moai/state/` occurrence counts against marker windows)
inherited exactly that over-breadth and would have produced a false RED. The narrowing corrects it.

**And it leaves no in-scope citation unreachable.** I checked all five: every one contains
`/verify` — `verify/t225` (×2), `verify/t341`, `verify/t299-sync`, `verify/eb01063e`. The
AC-IEC-001 companion confirms it, printing `1` for each of the five files under the narrowed
pattern. The narrowing is exactly as wide as the scope and no wider.

---

## Recommendation

**FAIL at 0.78 against a 0.80 threshold — but this is a near-miss on a card that improved on every
axis it was asked to improve, and the Tier M iteration ceiling is now reached.**

Three blocking defects remain, and all three are small, surgical, and precisely located:

1. **N1** — rewrite AC-IEC-001 as the two-invocation `grep -l` comparison above and correct the false
   sentence at acceptance.md:101-102. This is the only one that would actually stop run-phase: four
   of five repairs are gated on it, and as written it cannot reach PASS.
2. **N2** — replace AC-IEC-010's `git status --short` check with a named-file existence check, and
   correct its RED-now cell.
3. **N3** — demote REQ-IEC-008 to a §D constraint.

Plus **N4** (version bump + HISTORY row on spec.md and plan.md) and **N5** (correct the `2` to `1`).
The four optional findings (N6-N8 and the §D duplication) should be surfaced and left to the
orchestrator — routing them all would add ceremony to a SPEC that is already unusually disciplined.

**Because the Tier M ceiling (2) is reached, this verdict cannot be re-issued without an operator
decision.** The three options, per the retry-loop contract:

- **Scope-reduction** — not warranted. The card is 5 lines across 5 files and its scope discipline
  has been sound across both iterations.
- **PASS-with-debt** — **not recommended.** N1 is not debt: AC-IEC-001 cannot reach PASS under the
  plan's own treatment (a), so accepting it as debt schedules a contradiction at M2 rather than
  deferring a cost.
- **Explicit override for a delta-scoped iter3** — **recommended.** The five fixes are mechanically
  specified above and touch `acceptance.md` (three edits), `spec.md` (two), and `plan.md` (one). A
  confirming re-audit scoped to that enumerated delta is cheap, and Testability is the only
  dimension below its band — closing N1 and N2 moves it to the 0.75-1.0 band and carries the
  aggregate over 0.80 without touching any other dimension.

---

## Coordinator claims that did not hold when I ran them

Three, all small, all in the same direction — recorded observations that were not re-measured after
the last edit:

1. **"AC-IEC-010 re-scoped so it is RED until M5."** It is **green today**. All three sub-checks pass
   before M5 runs, and the evidence directory already holds three files, so the matrix's "(dir not
   yet populated)" is false on its face. This is N2, and it is the one iter1 defect the iter2 fix
   claimed to close but did not.
2. **"AC-IEC-004 observed `…html:2`."** Measured `1`. Both pattern matches are on line 684 and
   `grep -c` counts lines. N5.
3. **"the `ls` reports … (rc=2)."** Measured `exit=1` on this platform (BSD `ls`). N6.

Everything else in the iter2 report reproduced exactly: the guard vocabulary, the four files listed
by AC-IEC-001, the `1`/`1` on AC-IEC-002, the `5` on AC-IEC-003, the `1`/`0` on AC-IEC-005, the
`exit=0` and 12-line positive control on AC-IEC-006, the single `manager-lead.toml` path, the
`status: in-progress` re-read, the `492 − 25 = 467` derivation, and the pairing note at t375
`spec.md:171`. The over-breadth finding the coordinator verified independently is correct, and it
corrected a wrong remediation I gave at iter1 — recorded above.
