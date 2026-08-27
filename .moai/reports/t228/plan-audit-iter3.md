# SPEC Review Report: SPEC-ASTGREP-LANG16-001

Iteration: 3/3 (ceiling reached — this is the final escalation report)
Verdict: **FAIL**
Overall Score: **0.71** (iter-1 0.68 → iter-2 0.69 → iter-3 **0.71**; delta **+0.02** vs iter-2,
**+0.03** vs iter-1). Tier L PASS threshold: 0.85.
STOP signal: **not raised** (no regression). Iteration ceiling: **reached**.

Auditor: plan-auditor. Bias-prevention M1-M6 active.

**M1 Context Isolation statement.** Reasoning context ignored per M1 Context Isolation. The
invocation prompt's account of what changed was treated as a list of **claims to falsify**, not as
evidence. Every judgement below rests on the six SPEC A artifacts and my own measurements. Where the
prompt's account and the tree disagree, the tree governs, and I say so.

## Measurement attribution

| field | value |
|---|---|
| worktree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t228` |
| branch | `WT-astgrep-16-langs` |
| HEAD | `294b4b6ab` (`git rev-parse --short HEAD`) |
| tracked-file state | `git status --porcelain` → three `??` entries only (`.moai/reports/t228/`, both SPEC dirs). No tracked file modified, so every measurement is against `origin/main` content. |
| tool | `ast-grep 0.40.5` (`sg --version`) |
| scratch | `/tmp/t228i3b/` — outside the repo. No in-tree probe edit was made this round. |

All commands run with cwd = worktree root, per the SPEC's own §A.0 hygiene rule.

---

## The decisive structural finding, stated first

**This revision edited exactly one artifact.** File mtimes: `acceptance.md` 00:43; `spec.md` 00:10,
`plan.md` 00:12, `design.md` 00:12, `research.md` 00:13, `progress.md` 00:15 — the latter five are
unchanged from the iteration-2 revision round. mtime alone is soft evidence, so I corroborated it
textually: **every iteration-2 finding whose fix lived outside `acceptance.md` is verbatim
unchanged**, and I quote each below (E2's `spec.md`/`plan.md` half, E3's `spec.md` half, E4's
`design.md` half, all five E6 citations, E7's `plan.md` half).

This single fact explains the entire defect list. It also produces a defect class that did not exist
in iteration 2: the criteria layer moved and the requirement layer did not, so `spec.md` and
`acceptance.md` now **contradict each other** on two load-bearing points where iteration 2 merely
had them wrong-but-consistent.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** `grep -c '^\*\*REQ-A16-'` → **23**; `grep -o … | sort |
  uniq -d` → empty (no duplicate). `REQ-A16-022a` appears in the id sweep but only in the `spec.md`
  HISTORY row recording its removal — an audit trail, not a live reference. Contiguous `001`…`023`,
  three-digit zero-padded.
- **[PASS] MP-2 GEARS format compliance — judged against the `REQ-XXX` requirement layer in
  `spec.md` §C only.** The `AC-A16-*` entries are Given-When-Then by design and are graded under
  Group 4, not here. `spec.md` §C is unchanged from iteration 2, where I verified all 23 entries
  carry a named `<subject>`, a `shall`, and a modality matching their label (`REQ-A16-012` at
  `spec.md:89-90` in `Where … shall` form). Note for the record: `REQ-A16-021` is well-**formed** and
  factually **false** (D3 below). GEARS binds form, not truth, so this is a PASS here and a defect
  there.
- **[PASS] MP-3 YAML frontmatter validity.** `spec.md:1-15` carries all 12 canonical fields with
  correct types: `id` / `title` (quoted) / `version: "0.4.0"` (quoted semver) / `status: draft` /
  `created: 2026-08-24` / `updated: 2026-08-25` (ISO) / `author` / `priority: P2` / `phase` /
  `module` / `lifecycle: spec-anchored` / `tags` (comma-separated string), plus optional `tier: L`.
  No rejected snake_case alias. (That `version` and HISTORY were not advanced for this revision is a
  defect — D9 — but not a frontmatter-validity failure.)
- **[PASS] MP-4 language neutrality.** All 16 enumerated with equal weight (`spec.md:35-50`,
  `REQ-A16-001`). Measured: `grep -rniE 'primary|planned|unsupported'` over the ruleset directory
  returns exactly one line — `sgconfig.yml:10`, *"future addition, never an unsupported one"* — the
  sanctioned equal-opportunity paragraph. `grep -rlE 'SPEC-[A-Z][A-Z0-9]+-[0-9]{3}'` over the same
  directory → **0**. The SPEC creates no neutrality violation.
- **[PASS] MP-5 D7 cross-SPEC reconciliation — no BLOCKING finding.** Extraction returns
  `SPEC-ASTG-UPGRADE-001`, `SPEC-ASTGREP-BREADTH-001`, `SPEC-ASTGREP-DOGFOOD-CLEANUP-001`,
  `SPEC-ASTGREP-MULTILANG-001`. Measured statuses: `archived`, `draft`, `completed`, `completed`.
  `spec.md` is unchanged, so the six reconciliation sites verified in iteration 2 (`spec.md:470-473`,
  `:496`, `:516-517`, `research.md:106-108`, `plan.md:141`, `plan.md:259`) still stand.
- **[PASS] MP-6 D8 cross-platform discipline.** `grep -rn 'syscall' .moai/specs/SPEC-ASTGREP-LANG16-001/`
  → rc=1. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate.** `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-ASTGREP-LANG16-001/`
  → rc=1 across all six artifacts.

**All seven must-pass criteria pass.** The FAIL is score-driven, exactly as in iteration 2.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.68 | between 0.50 and 0.75 | `acceptance.md` §0 is the clearest prose in the document set: the vacuous/impossible mutant table (`acceptance.md:19-22`) names the failure modes precisely and ties them to the SPEC's own existing standards. Against that, the single-artifact revision created **two direct inter-artifact contradictions** where iteration 2 had none: `REQ-A16-018` (`spec.md:146-149`) mandates the commit carry a regenerated `catalog.yaml` while `AC-A16-019` (`acceptance.md:199-200`) requires `make build` to leave it **unchanged** (D1); and `AC-A16-007`/`AC-A16-008` require a new test pinning `internal/hook` behaviour while `plan.md:68-70` lists `internal/hook/**` as PRESERVE "in its entirety" and `plan.md:244-247` makes an empty `internal/hook` diff the standing proof (D4). `design.md:119`/`:154` still state the D13 placement decision the opposite way from `design.md:163-169` — unchanged for a third iteration (D5). |
| Completeness | 0.75 | 0.75 band | All required sections present; §D carries seven `### Out of Scope — <topic>` H3 sub-headings with specific bullets; frontmatter complete; Tier L 5-artifact set present and substantive. Real additions this round: §G (relocation register), §H (19-row base-red table), §I (traceability matrix). Deductions: `progress.md` §E.1 still declares `acceptance_criteria: 23` / `requirements: 23` / *"headroom of 2 in each"* / `audit_iteration: 1 revision` — the machine-readable audit-ready signal is stale by a full iteration (D8); no HISTORY row and no `version` bump record this revision (D9); `REQ-A16-011` lacks the citation-or-probe half its own criterion `AC-A16-016` asserts (D3). |
| Testability | 0.78 | 0.75+ | **The largest real gain in three iterations, and I verified it by execution rather than by reading.** I ran 10 of the 19 §H RED-NOW commands spanning §A/§B/§C/§D and **all 10 reproduced the recorded output exactly** (table below). The central iteration-2 defect is genuinely closed: I built a scratch config with two rules where only one carried a test file and measured `Running 1 tests` — so `sg test`'s reported count equals the number of rules carrying cases, which makes `AC-A16-001`'s count-equals-rule-count assertion non-vacuous **and** genuinely discharges `REQ-A16-010`. The `[Wrong] No probe-a baseline found` output confirms "zero missing snapshots" is a real failure class too. Against that: `AC-A16-016`'s mechanical half and `AC-A16-007`/`AC-A16-008`'s pinning test have **green paths naming milestone work that does not exist in `plan.md`** (D3, D4), and `AC-A16-004`'s *"or the count the M1 id-keying decision yields"* is an escape hatch an implementer can fill with any number (D10). |
| Traceability | 0.66 | between 0.50 and 0.75 | §I is a genuine gain in kind — 23 of 23 requirements now named, against 8 in iteration 2 — and I checked its rows against the criteria bodies rather than their titles, as instructed. But the renumbering damage outweighs it. **Eight AC references outside `acceptance.md` were not renumbered**: two are now **dangling** (`AC-A16-020` at `plan.md:279`; `AC-A16-021`…`023` at `plan.md:196` — the document stops at 019), five resolve to a **real but different** criterion, which is worse than dangling because it reads as correct, and one is a stale count. All three milestone "Verified by" ranges in `plan.md` are wrong against `acceptance.md`'s own milestone sections. Within §I itself, three rows map a requirement to a criterion that contradicts it, under-scopes it, or over-reaches it, and one row is a criterion that merely sounds related (D7). |

Aggregate = harmonic mean of (0.68, 0.75, 0.78, 0.66) = **0.71**. Tier L threshold 0.85.

---

## Audit question 1 — is the two-column table real, or was it reasoned?

**It is real.** This is the finding I most expected to falsify and could not. I executed 10 of the 19
RED-NOW rows, chosen to span all four sections, and every one reproduced the recorded output
verbatim. No entry differed.

| §H row | Command I ran | Recorded | I observed | Match |
|---|---|---|---|---|
| 001 / 015 / 018 | `sg test --config $T/sgconfig.yml` | `Running 0 tests` / `ok. 0 passed; 0 failed;` EXIT=0 | identical, EXIT=0 | ✅ |
| 001 (rule count) | `grep -rh '^id:' $T --include='*.yml' \| wc -l` | 26 | 26 | ✅ |
| 004 | `grep -c 'settled at M1\|not yet measured' design.md` | `2` | `2` (rc=0) | ✅ |
| 005 | `find $T -name 'rule-tests' -o -name '*test*'` | *(no output)* | *(no output)* | ✅ |
| 006 | `grep -c testConfigs $T/sgconfig.yml` | `0` (rc=1) | `0`, rc=1 | ✅ |
| 007 | `grep -rn 'func Test' internal/hook/pre_tool_scan_differential_test.go` | one function | one: `TestScanWriteContentDifferential` (line 183) | ✅ |
| 009-014 | `test -f $M` | `ABSENT` (rc=1) | rc=1 | ✅ |
| 016 | `grep -rh 'cwe:' $T/security/*.yml \| wc -l` | `14` of 14 | 14 cwe entries, 14 security rules | ✅ |
| 017 | `grep -A2 'sec-csrf-no-token-check' $T/security/web.yml \| grep severity` | `severity: warning` | `severity: warning` | ✅ |
| 019 | `find $T -name '*test*'` ; `grep -rlE 'SPEC-…' $T/` | *(none)* ; `0` | *(none)* ; `0` | ✅ |

I also independently reproduced both figures in the §D scope note — `grep -rlE 'SPEC-[A-Z][A-Z0-9]+-[0-9]{3}' internal/template/templates/`
→ **22**, `grep -rlE '20[0-9]{2}-[0-9]{2}-[0-9]{2}'` → **90** — and the ranking-scan claim (exactly
one match, `sgconfig.yml:10`).

**Credit where it is due, since two prior FAILs should not obscure this.** The two-column method is
sound and it was executed, not narrated. It closed the single worst defect of iteration 2 by
construction. The method is not the problem, and I say so plainly in the recommendation because the
standing agreement's escalation would discard it.

## Audit question 3 — is E1 really fixed?

**Yes, and it is stronger than the fix note claims.** I did not take the count semantics on trust; I
measured them:

```
# /tmp/t228i3b/proj — 2 rules (probe-a, probe-b), test file for probe-a only
$ sg test --config /tmp/t228i3b/proj/sgconfig.yml
Running 1 tests
[Wrong] No probe-a baseline found.
FAIL probe-a  .W
Error: test failed. 0 passed; 1 failed;
EXIT=4
```

Two facts follow, and both are load-bearing for `AC-A16-001`. `Running 1 tests` against two rules
establishes that **the reported count is the number of rules carrying cases**, so
count-equals-rule-count cannot be satisfied while any rule ships untested — which is exactly what
`REQ-A16-010` demands and what `sg test` is otherwise silent about. And `[Wrong] No … baseline
found` with EXIT=4 establishes that "zero missing snapshots" names a real, non-zero-exit failure
class rather than a decorative clause. The `AC-A16-001` note at `acceptance.md:48-51` is accurate on
both points.

One residue: the criterion's *"or the count the M1 id-keying decision yields"* leaves the target
number open. `AC-A16-004` forces the keying to be measured and recorded, which bounds the abuse, but
an implementer can still declare an inconvenient number to be "the count the decision yields". Minor
(D10).

## Audit question 2 — green paths that do not exist

Three of nineteen name milestone work `plan.md` does not contain. I checked each GREEN PATH cell
against `plan.md` §B item by item.

Sound (spot-checked, all resolve): `M1.2` = `testConfigs` wiring (`plan.md:108-115`); `M1.3` = 26
case pairs (`plan.md:116-117`); `M1.4` = id keying (`plan.md:118-123`); `M1.5` = both mutants
(`plan.md:124-126`); `M2.1` parser re-probe; `M2.2` matrix authoring; `M2.4` four classes; `M3.1`
predicate application; `M3.2` `sec-csrf-no-token-check`.

Broken:

- **§H row 016 → "M3.3 adds citation-or-probe per security rule".** `plan.md:183-185` (M3 item 3)
  reads *"Add `metadata.cwe` to every security rule lacking it, and verify each `invalid` case
  instantiates that weakness class in idiomatic code (REQ-A16-011)"*. There is no citation-or-probe
  work in it, and `REQ-A16-011` (`spec.md:71-74`) does not require any. The criterion's mechanical
  half is required by nothing and built by nothing.
- **§H rows 007 and 008 → "M1 adds the pinning test".** `plan.md` M1's five work items
  (`plan.md:107-126`) contain no test authoring at all.

That last one is not merely a missing item — it collides with the plan (D4).

## Audit question 4 — the three relocations (§G)

| Relocated obligation | Genuinely carried elsewhere? | Verdict |
|---|---|---|
| `make build` regenerates a committed artifact | **No.** §G says it "lives … as a build step in `plan.md` §D". `plan.md:222-224` is unchanged and states the **falsified** version: *"`make build` rides the same commit as any template edit, and the artifact is `internal/template/catalog.yaml`. A template change without it is a broken commit."* `plan.md:58` repeats it (`# REGENERATE via make build, same commit`), and `REQ-A16-018` still mandates it. I re-measured: `grep -c astgrep internal/template/catalog.yaml` → **0**; the catalog's entries are skills and agents only. So the relocation moved a **falsehood** into the destination while the corrected form lives only in `AC-A16-019`. | **FAILED relocation** (D1) |
| Never mirror the dogfood tree | **Yes.** Folded into `AC-A16-019` as a byte-identity clause, and `REQ-A16-020` (`spec.md:164-166`) plus the `plan.md:277-278` anti-pattern both survive. The dogfood tree exists (`.moai/astgrep-rules/` — `go/`, `security/`, `sgconfig.yml`), so the comparison has a subject. | **OK** |
| No language is ranked | **Yes, with a scope caveat.** Folded into `AC-A16-019`; `REQ-A16-001` keeps it normative. Caveat: `REQ-A16-001`'s prohibition covers *"no file under `internal/template/templates/**`"* while the criterion scans `$T`/`$R` only — the same under-scoping as D3. The obligation survives; its verification is narrower than its statement. | **OK, narrowed** |

So: two legitimate relocations, one relocation that carried the wrong version of the obligation into
its new home. A deletion wearing a relocation costume is what I was asked to look for; this is not
that — it is worse in one respect and better in another. The obligation was *kept*, and what was kept
is the measured-false form.

## Audit question 5 — defects from the renumbering 23 → 19

Complete sweep (`grep -rn 'AC-A16-0[0-9][0-9]'` excluding `acceptance.md`):

| Site | Text | Problem |
|---|---|---|
| `plan.md:132` | M1 "Verified by: AC-A16-001 … AC-A16-005" | `acceptance.md` §A (Milestone 1) is **001…008**. AC-006/007/008 are orphaned from their own milestone. |
| `plan.md:165` | M2 "Verified by: AC-A16-006 … AC-A16-011" | `acceptance.md` §B (Milestone 2) is **009…014**. The stated range names five criteria belonging to M1. |
| `plan.md:196` | M3 "AC-A16-012 … AC-A16-015, plus AC-A16-019 … AC-A16-023" | §C (M3) is **015…018**; §D is **019** alone. `AC-A16-020`…`023` **do not exist**. |
| `plan.md:229` | "`internal/hook/**` is read-only … **AC-A16-018** proves it with two empty diffs" | `AC-A16-018` is now the severity criterion. The corpus-preservation criterion is **AC-A16-008**, and it is a *test*, not two diffs. Resolves to a real but wrong criterion. |
| `plan.md:70` | "…it changes neither. **AC-A16-018** proves it." | Same wrong resolution. |
| `plan.md:279` | "Weakening **AC-A16-020**'s grep until it matches nothing" | **Dangling** — the ranking scan folded into `AC-A16-019`. |
| `plan.md:290` | "`acceptance.md` — the **23** criteria each milestone is verified against" | Stale count; 19. |
| `design.md:70` | "distinct failure class (**AC-A16-007**)" for the unevidenced-exemption class | Correct id is now **AC-A16-011**. This citation was already wrong in iteration 2 (E6) and the renumbering made `AC-A16-007` mean the corpus-gate pinning criterion — so it is now confidently wrong about a different thing. |
| `research.md:33` | "…and **AC-A16-011** checks the two agree at close" (severity record vs tree) | `AC-A16-011` is now exemption evidence. The severity criterion is **AC-A16-018**. |
| `progress.md:41` | "D9 (**AC-A16-020** names its sanctioned exemption)" | **Acceptable** — historical audit-trail sentence, same class as the `REQ-A16-022a` HISTORY row. Not a defect. |

## Audit question 6 — §I traceability, audited as adversarially as the rest

The instruction to treat this as orchestrator-authored and not neutral is honoured; I read each
cited criterion's body and asked whether it *asserts* the requirement.

**Rows that hold** (spot-verified against criterion bodies): REQ-002→AC-014, REQ-003→AC-009,
REQ-004→AC-010, REQ-005→AC-011, REQ-006→AC-013, REQ-008→AC-001/002/003, REQ-009→AC-001,
REQ-010→AC-001 (verified by my own count-semantics probe above), REQ-012→AC-015/018, REQ-013→AC-018,
REQ-015→AC-017, REQ-016→AC-007(a), REQ-017→AC-007(b), REQ-019→AC-005/006, REQ-020→AC-019,
REQ-022→AC-006, REQ-023→AC-006. That is 17 of 23 clean, which is a real gain over the 8 requirements
named anywhere in iteration 2.

**Rows that do not hold:**

1. **REQ-A16-018 → AC-019.** The criterion does not merely fail to assert the requirement — it
   asserts its **negation**. REQ: the commit *shall carry the regenerated* `catalog.yaml`. AC-019
   final clause: `make build` *leaves* `catalog.yaml` **unchanged**. The matrix's "what carries the
   assertion" cell reads *"scope of the close scan is `$T` and `$R`"*, which describes neither.
2. **REQ-A16-021 → AC-019.** REQ scope: *"No file under `internal/template/templates/**`"*. AC scope:
   `$T` and `$R`. The criterion asserts strictly less than the requirement, and the requirement as
   written is measured false on the tree (22 files) and unfixable by this SPEC. The matrix records
   the narrower scope in its own cell without flagging the gap.
3. **REQ-A16-011 → AC-016.** The criterion asserts **more** than the requirement: the citation-or-probe
   half exists in no requirement. The cell describes it as "the mechanical citation-or-probe half" as
   though REQ-011 carried it.
4. **REQ-A16-014 → AC-017, AC-003 — the "merely sounds related" row I was asked to hunt for.**
   REQ-014: a shape-matching rule *shall not carry `error`*. `AC-A16-003` asserts that the harness
   emits a `[Noisy]` failure when a rule matches its own `valid` case. That is a **harness** property;
   it asserts nothing about any rule's severity field. The cell's gloss — *"`[Noisy]` guards the
   direction"* — is an argument about why the harness helps, not evidence that the criterion asserts
   the requirement. `AC-A16-017` carries REQ-014 for the one named rule; the general prohibition rests
   on `AC-A16-018`, which is not cited in this row.
5. **REQ-A16-001 → AC-014, AC-019** inherits defect 2's scope narrowing (prohibition stated over the
   whole template tree, verified over `$T`/`$R`).

**The deliberate non-back-fill of AC-A16-004 is correct and I would have flagged the opposite.**
Inventing a requirement to give a precondition-criterion a row is the failure this table exists to
prevent, and recording it as an exception is the right disposition.

## Audit question 7 — anything still passing on an untouched tree

**No criterion is now vacuous.** This is the round's real achievement. Every one of the 19 carries a
RED-NOW row, I executed 10 of them, and all 10 were red as recorded. `AC-A16-016` and `AC-A16-017`
are honest about carrying a half that already passes (`acceptance.md` §H rows 016/017 say so
explicitly), and both ride a genuinely red half. `AC-A16-019` is a composite whose guard clauses hold
today and rides one red clause (`$R` does not exist) — the §H row states this rather than hiding it,
which is the correct disposition.

---

## Defects Found (structured defect-list)

**D1. `catalog.yaml` obligation still factually wrong at the requirement layer, and now
self-contradictory** — `spec.md:146-149` (REQ-A16-018), `plan.md:58`, `plan.md:222-224`,
`acceptance.md:199-200` (AC-A16-019), `acceptance.md:240` (§G row 1) — iteration-2 E2 required
`REQ-A16-018` to be restated as what is true. `spec.md` was not edited, so the requirement still
reads *"when it changes, the same commit shall carry the regenerated embed artifact
`internal/template/catalog.yaml`"*. Re-measured this round: `grep -c astgrep
internal/template/catalog.yaml` → **0**, `grep -c 'astgrep-rules'` → **0**; the catalog holds skill
and agent entries with content hashes and tracks no config path. `AC-A16-019` now asserts the
opposite (unchanged), so requirement and criterion conflict directly, and `plan.md` §D still calls a
template change without the regeneration *"a broken commit"* — instructing run-phase to do the thing
the criterion forbids. §G's claim that the obligation "lives … as a build step in `plan.md` §D" is
true only in the sense that the *falsified* version lives there.
  — Severity: **critical** — Class: **blocking** — Required fix: rewrite `REQ-A16-018` to the
  measured truth (`$T` is the source of truth; it is compiled in by `//go:embed all:templates` with no
  committed artifact; `make build` must leave `internal/template/catalog.yaml` unchanged), delete the
  `# REGENERATE via make build, same commit` line at `plan.md:58`, and rewrite `plan.md:222-224` to
  the same inverted form. Keep the §A.5-style note explaining why a reader expects an artifact.

**D2. AC-A16-016 asserts a mechanical obligation no requirement states and no milestone builds** —
`acceptance.md:162-168`, `spec.md:71-74` (REQ-A16-011), `plan.md:183-185` (M3 item 3),
`acceptance.md:268` (§H green path) — the criterion requires each security rule's matched head symbol
to be anchored by *"a named language / stdlib / framework reference … or a recorded probe"*.
`REQ-A16-011` requires only `metadata.cwe` plus an idiomatic `invalid` case. `plan.md` M3 item 3 does
only those two things. The §H green path *"**M3.3** adds citation-or-probe per security rule"* names
work that milestone does not contain. This is the impossible-criterion mutant §0 defines, wearing a
milestone number — the exact shape I was asked to hunt.
  — Severity: **critical** — Class: **blocking** — Required fix: add the citation-or-probe half to
  `REQ-A16-011` (headroom exists at 23/25) **and** add the corresponding work item to `plan.md` M3, or
  drop the half from `AC-A16-016`. Do not leave the criterion asserting more than the requirement and
  the plan together deliver.

**D3. REQ-A16-021's scope is measured false and unfixable by this SPEC — the impossible clause moved
up a layer instead of being removed** — `spec.md:168-170`, `acceptance.md:191-209` (AC-A16-019),
`acceptance.md:316` (§I row) — iteration-2 E3 required both the criterion **and** `REQ-A16-021` to be
narrowed to the ruleset directory. Only the criterion was. The requirement still prohibits a SPEC ID,
internal date, or commit SHA in *"No file under `internal/template/templates/**`"*. Re-measured: 22
files under that tree carry a SPEC ID and 90 carry an ISO date, none of them owned by this SPEC. A
requirement that is false on arrival and that no work in this SPEC can make true is the **impossible**
mutant `acceptance.md` §0 defines — relocated from the criteria layer into the requirement layer,
where §0's discipline does not reach it. `REQ-A16-001`'s ranking prohibition carries the same
over-scope.
  — Severity: **critical** — Class: **blocking** — Required fix: narrow `REQ-A16-021`'s and
  `REQ-A16-001`'s scope to `internal/template/templates/.moai/config/astgrep-rules/**` plus the
  repo-side rule-test root, matching the criterion. Then extend §0's two-column discipline to the
  requirement layer in one sentence, so a requirement that is red-forever is caught where this one was
  not.

**D4. AC-A16-007 / AC-A16-008 require a new test under a tree the plan declares PRESERVE, and prove
untouchedness with a diff that test would make non-empty** — `acceptance.md:89-108`,
`acceptance.md:259-260` (§H green paths), `plan.md:64-70` (§A.4 PRESERVE), `plan.md:107-126` (M1
work), `plan.md:229`, `plan.md:244-247` (§E standing command) — both criteria require *"the new
pinning test"* asserting (a) the corpus validity gate calls `t.Skip`, (b) the differential run did not
skip, and (c) the twelve `scan-corpus/` fixtures and every `wantDeny` are byte-unchanged. Those are
assertions about `internal/hook`. `plan.md` §A.4 lists *"`internal/hook/**` in its entirety —
including `pre_tool_scan_differential_test.go` and every file under
`internal/hook/security/testdata/scan-corpus/`"* as do-not-modify, §D repeats *"`internal/hook/**` is
read-only for this SPEC"*, and §E makes `git diff --stat origin/main -- internal/hook/` *"the standing
proof that this SPEC's PRESERVE list held"*. M1's five work items author no test, and the criteria
name no home for one. Adding it under `internal/hook/` makes the standing proof fail; putting it
elsewhere is unstated. The two cannot both hold as written.
  — Severity: **critical** — Class: **blocking** — Required fix: name the pinning test's file path in
  `acceptance.md` §A and add it as an M1 work item in `plan.md`; then either carve the new test file
  out of the PRESERVE clause explicitly (§A.4 and §D) and re-word §E's standing command to
  `git diff --stat origin/main -- internal/hook/ ':(exclude)<new test path>'`, or place the test
  outside `internal/hook/` and say so.

**D5. `design.md` still states the D13 placement decision both ways — unchanged across all three
iterations** — `design.md:119`, `design.md:154` vs `design.md:163-169`, `spec.md:155-156`
(REQ-A16-019), `acceptance.md:73-77` (AC-A16-005) — §3.1 reads *"`testConfigs: [{testDir: rule-tests}]`
in the template `sgconfig.yml`, with fixtures under `$T/rule-tests/`"* and the §4 layout diagram
repeats it on the `$T/sgconfig.yml` line, 44 lines above the paragraph stating the opposite and citing
`REQ-A16-019`. Verbatim identical to what I quoted in iteration 2 (E4) and raised in iteration 1
(D13). **Stagnation flagged**: this defect appears in all three iterations unchanged. Per the retry
contract that indicates a propagation failure rather than a missed fix — and here the cause is
identifiable: `design.md` was not opened this round.
  — Severity: **major** — Class: **blocking** — Required fix: rewrite `design.md` §3.1's wiring
  sentence and §4's layout comment to name the repo-side root. Record in §3.1 the measured fact
  (iteration 2) that `testDir` resolves relative to `sgconfig.yml` and traverses upward at exit 0, so
  the placement is known-feasible.

**D6. Renumbering 23 → 19 not propagated: two dangling references, five confidently-wrong ones, one
stale count** — `plan.md:132`, `:165`, `:196`, `:229`, `:70`, `:279`, `:290`, `design.md:70`,
`research.md:33` — enumerated in the audit-question-5 table above. `AC-A16-020` and
`AC-A16-021`…`023` no longer exist. All three milestone "Verified by" ranges disagree with
`acceptance.md`'s own milestone sections, so no milestone currently owns AC-006/007/008 or AC-016…019
correctly. `plan.md:229` and `:70` cite `AC-A16-018` (severity) as the proof that `internal/hook` was
untouched — a real criterion that asserts something else entirely.
  — Severity: **major** — Class: **blocking** — Required fix: renumber all nine sites; set M1 =
  001…008, M2 = 009…014, M3 = 015…018, close = 019; replace the two `AC-A16-018` PRESERVE cites with
  `AC-A16-008`; replace `AC-A16-020` with `AC-A16-019`; `design.md:70` → `AC-A16-011`;
  `research.md:33` → `AC-A16-018`; `plan.md:290` → 19.

**D7. §I traceability — four rows do not hold** — `acceptance.md:294-318` — (a) `REQ-A16-018` → AC-019
maps a requirement to a criterion asserting its negation (see D1); (b) `REQ-A16-021` → AC-019 and
`REQ-A16-001` → AC-019 map a template-tree-wide prohibition to a `$T`/`$R`-scoped scan, so the
criterion asserts strictly less than the requirement (see D3); (c) `REQ-A16-011` → AC-016 records a
half the requirement does not carry (see D2); (d) `REQ-A16-014` → `AC-A16-003` is the "sounds related"
row — `AC-A16-003` asserts a harness property (`[Noisy]` fires) and says nothing about any rule's
severity, while the requirement it is cited for prohibits `error` on shape matchers. The remaining 17
rows I checked hold against the criteria bodies.
  — Severity: **major** — Class: **blocking** — Required fix: repair (a)-(c) at the requirement layer
  per D1/D2/D3 and re-derive those rows; for (d) replace `AC-A16-003` with `AC-A16-018` in the
  `REQ-A16-014` row, or state in the cell that the general prohibition is carried by AC-018 and AC-017
  covers only the named rule. Add a scope column, or a note, wherever a criterion's scope is narrower
  than its requirement's — an unflagged narrowing is the same class of silence this table exists to
  end.

**D8. `progress.md` audit-ready signal is stale by a full iteration** — `progress.md:8-19`, `:34-35`,
`:37-45` — declares `requirements: 23`, `acceptance_criteria: 23`, `audit_iteration: 1 revision (FAIL
0.68 -> revised)`, and *"Budget after revision: 23 requirements, 23 acceptance criteria — headroom of
2 in each"*. Measured: 23 requirements, **19** criteria, headroom **6**, and two audit iterations have
completed. The blocking-findings paragraph lists only iteration-1's D1…D16. This is the machine-read
plan-phase signal block, so a consumer reading it gets the iteration-1 picture.
  — Severity: **minor** — Class: **blocking** — Required fix: set `acceptance_criteria: 19`,
  `audit_iteration: 2 revisions (0.68 → 0.69 → revised)`, correct the budget sentence, and add a
  paragraph naming the iteration-2 findings addressed.

**D9. This revision has no HISTORY row and no version bump** — `spec.md:20-26` (HISTORY table),
`spec.md:4` (`version: "0.4.0"`) — the substantive changes of this round (23 → 19 criteria, three
relocations, §0 / §G / §H / §I added) are recorded nowhere in the SPEC's own change log; the last row
is v0.4.0 describing the iteration-1 revision. The SPEC's `updated:` is `2026-08-25`, which is also
v0.4.0's date, so nothing distinguishes the two rounds.
  — Severity: **minor** — Class: **blocking** — Required fix: bump to `0.5.0` and add a HISTORY row
  naming the iteration-2 findings and the two-column method.

**D10. AC-A16-001's target count carries an open escape clause** — `acceptance.md:44-46` — *"equals
the number of rules in the shipped ruleset — 26 at M1, or the count the M1 id-keying decision
yields"*. The alternative is unbounded: an implementer can declare any observed number to be what the
keying decision yields. `AC-A16-004` forces the decision to be measured and recorded, which limits the
damage, but the criterion does not require the two to be reconciled.
  — Severity: **minor** — Class: **optional** — Required fix: bind the alternative — *"…or, if M1
  settles keying by id alone, the number of **distinct rule ids** in the shipped ruleset, as recorded
  in `design.md` §3.3"* — so the target is derivable rather than declarable.

**D11. `design.md` §5 still says "three classes" above four, and three stale requirement citations
remain** — `design.md:192` (*"any of the three classes below"* above four enumerated classes),
`design.md:110` (cites `REQ-A16-014` for the matrix-annotation obligation; correct is `REQ-A16-015`),
`design.md:180` (cites `REQ-A16-021` for the `ruleDirs` obligation; correct is `REQ-A16-023`),
`design.md:274` (cites `spec.md` §A.4 for the `ShouldAlert` → `DecisionDeny` wiring; correct is §A.5).
All four are verbatim unchanged from iteration-2 E6; each resolves to a real but wrong target.
  — Severity: **minor** — Class: **blocking** — Required fix: mechanical; correct the count and the
  three citations.

---

## Regression Check (iteration 2 → 3)

| iter-2 finding | Class | Status | Evidence |
|---|---|---|---|
| **E1** central criterion passes vacuously | critical/blocking | **FIXED** | `AC-A16-001` now asserts count-equals-rule-count. I measured the semantics myself (`Running 1 tests` with 2 rules, 1 tested), so the assertion is non-vacuous and also discharges `REQ-A16-010`. |
| **E2** `catalog.yaml` obligation factually wrong | critical/blocking | **HALF-FIXED, now self-contradictory** | Criterion inverted; `REQ-A16-018` and `plan.md` §A.3/§D unchanged and still assert the falsified form. → **D1** |
| **E3** AC-019 red on untouched content, misstates CI guard | major/blocking | **HALF-FIXED** | Criterion re-scoped to `$T`/`$R` and the CI guard's class set now stated accurately (`acceptance.md:202-209`). `REQ-A16-021` unchanged, so the impossibility moved up a layer. → **D3** |
| **E4** `design.md` contradicts the D13 placement | major/blocking | **UNFIXED — third iteration unchanged** | `design.md:119`/`:154` verbatim identical. → **D5**, stagnation flagged |
| **E5** `REQ-A16-010` has no criterion | major/blocking | **FIXED** | Discharged by `AC-A16-001`'s count comparison; I verified the mechanism rather than the wording. |
| **E6** four stale cites + "three classes" in `design.md` | minor/blocking | **UNFIXED** | All five verbatim unchanged; `design.md:70` is now wrong about a *different* criterion after renumbering. → **D11**, **D6** |
| **E7** AC-016/017 owned by no milestone | minor/blocking | **NOT FIXED, made worse** | `plan.md` "Verified by" lines unchanged, so after renumbering all three ranges are wrong and two cite non-existent ids. → **D6** |
| **E8** AC-A16-018 had no requirement | minor/optional | **RESOLVED by renumbering** | The corpus-baseline criterion is now `AC-A16-008` and §I maps it to `REQ-A16-017`; the §H row gives it a real red-now. |
| **E9** presence/absence evidence asymmetry | major/optional | **PARTIALLY APPLIED, unbacked** | `AC-A16-016` now carries the citation-or-probe half — but no requirement states it and no milestone builds it. → **D2** |

**Stagnation: FLAGGED for one defect.** `design.md`'s D13 contradiction appears in iteration 1 (D13),
iteration 2 (E4), and iteration 3 (D5) unchanged. Per the retry contract this indicates a
misunderstanding rather than a missed fix — and the misunderstanding is now identifiable: the revision
scope was `acceptance.md` alone, so no `design.md` finding could have been fixed in this round
regardless of its merit.

---

## Is the score movement real improvement, or a third flat round?

**Both, and the distinction matters more than the aggregate.** I am answering plainly because the
standing agreement turns on this answer.

The aggregate moved 0.68 → 0.69 → 0.71, which reads flat. The composition did not:

| Dimension | iter-1 | iter-2 | iter-3 | Movement |
|---|---|---|---|---|
| Clarity | — | 0.72 | 0.68 | ↓ new inter-artifact contradictions |
| Completeness | — | 0.78 | 0.75 | ↓ stale `progress.md`, no HISTORY row |
| **Testability** | — | **0.60** | **0.78** | **↑ +0.18 — the vacuous-criterion class is closed** |
| Traceability | — | 0.70 | 0.66 | ↓ renumbering damage exceeds §I's gain |

The one dimension the revision actually targeted moved **+0.18**, verified by execution rather than
by reading: I ran 10 of 19 RED-NOW rows and all 10 reproduced exactly, and I independently measured
the `sg test` count semantics that make the central fix work. The other three dimensions **fell**, and
every point of that fall traces to one cause — five of six artifacts were not opened.

**So I do not recommend re-designing the authoring approach, and I want to be unambiguous about
why.** The escalation trigger exists to catch a method that cannot converge. This method converged
sharply where it was applied. The blocker is not the two-column rule, the SPEC's structure, or the
author's understanding of the failure modes — `acceptance.md` §0 states those failure modes more
precisely than my own iteration-2 report did. **The blocker is revision scope: a criteria-layer
rewrite was performed without propagating its consequences into the requirement layer
(`spec.md`), the work layer (`plan.md`), the design layer (`design.md`), and the signal layer
(`progress.md`).** Discarding the method now would throw away the only thing that has worked in three
rounds.

Two of this round's four critical defects (D1, D3) are literally *the same fix I asked for in
iteration 2, applied to one of the two files it named*. That is not a method failure; it is a
propagation failure, and it is mechanically enumerable — I have enumerated it.

---

## Recommendation

**FAIL at 0.71 against a Tier L threshold of 0.85. Iteration ceiling (3) reached.** Escalating to the
operator with three options, per the retry contract. My recommendation is option 3.

1. **PASS-with-debt — I advise against this.** Two defects would actively misdirect run-phase rather
   than merely leave it under-specified: `plan.md` §D instructs the implementer to commit a
   regenerated `catalog.yaml` that `AC-A16-019` forbids (D1), and `AC-A16-007`/`008` require a test
   under a tree `plan.md` declares read-only and proves untouched with a diff that test would break
   (D4). Debt is acceptable when it is inert; this debt gives contradictory instructions.
2. **Scope reduction — not warranted.** The SPEC is already the contract half of a split I confirmed
   in iteration 2 is genuine sequencing. Splitting further would fragment a coherent three-milestone
   contract to solve a problem that is not scope.
3. **Bounded override — one propagation-only pass, recommended.** Not a fourth open revision: a
   closed, fully-enumerated edit list touching the five artifacts this round did not open, adding **no
   new content** beyond two requirement amendments. Every item is named above with its file, line, and
   exact change. In priority order:
   - **D1** — `spec.md:146-149`, `plan.md:58`, `plan.md:222-224`: invert the `catalog.yaml`
     obligation to match `AC-A16-019`.
   - **D3** — `spec.md:168-170` and `spec.md:7-9`: narrow the neutrality and ranking scopes to the
     ruleset + rule-test trees.
   - **D4** — `acceptance.md` §A and `plan.md` M1/§A.4/§D/§E: name the pinning test's path, add the M1
     work item, and reconcile it with PRESERVE.
   - **D2** — `spec.md:71-74` and `plan.md:183-185`: add the citation-or-probe half to `REQ-A16-011`
     and to M3, or drop it from `AC-A16-016`.
   - **D6** — nine renumbering sites across `plan.md` / `design.md` / `research.md`.
   - **D5, D11** — `design.md` §3.1, §4, §5 and three citations.
   - **D7** — re-derive the four §I rows once D1/D2/D3 land; add a scope note.
   - **D8, D9** — `progress.md` §E.1; `spec.md` HISTORY + version bump.
   - **D10** — optional, one clause in `AC-A16-001`.

   A confirming re-audit of this list would be scoped to the enumerated delta, not a fresh full audit.

**What this revision got right.** The two-column rule is the first thing in three iterations that
closed a defect class rather than an instance, and it was *executed* — I sampled it hard, expecting to
find reasoned-not-run entries, and found none. The `AC-A16-001` fix is stronger than its own note
claims: I measured that `sg test`'s count tracks rules-carrying-cases, so one criterion now genuinely
discharges two requirements. §G is the right disposition for an obligation with no greenable
criterion, and two of its three relocations are clean. §I moved requirement coverage from 8 named to
23 named, and 17 of its 23 rows survive an adversarial read of the criteria bodies — including its
decision **not** to invent a requirement for `AC-A16-004`, which is exactly right. The defects that
remain are, for the third time, not defects of honesty; they are the consequences of a fix applied to
one file when it named two.
