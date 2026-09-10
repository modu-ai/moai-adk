# SPEC Review Report: SPEC-VACUOUS-FLOOR-GUARD-001 (card t378)

Iteration: 1/1 (Tier S ceiling — `harness.plan_audit_tier_ceilings`, S=1)
Verdict: **FAIL**
Overall Score: **0.80** (harmonic mean of the four dimensions)
Tier threshold applied: **0.75** (Tier S, `spec-workflow.md` § SPEC Complexity Tier)

Tree audited: `.claude/worktrees/t378`, branch `WT-vacuous-floor-guard`, HEAD `b2e3af7fb`, base `3f03d9c36`.
Reasoning context from the SPEC author was not supplied and would have been ignored per M1 Context Isolation. All judgements below are made against `spec.md`, `plan.md`, `progress.md`, and the Go source at this tree.

**The score clears the threshold and no must-pass criterion fails. The FAIL is driven entirely by the three blocking defects in `## Defects Found` (D1-D3), each of which is a correctness defect in a stated criterion or in the SPEC's load-bearing argument.** The corrective for all three is small and confined to `spec.md`; a re-audit is scoped to the D1-D3 delta.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-VFG-001` … `REQ-VFG-007`, sequential, no gaps, no duplicates, uniform 3-digit padding. `grep -c '^\*\*REQ-VFG-' spec.md` = `7`. AC side: `AC-VFG-001` … `AC-VFG-008`, `grep -c '^\*\*AC-VFG-' spec.md` = `8`.
- **[PASS] MP-2 GEARS format compliance (requirement layer only)** — judgement made against the seven `REQ-VFG-*` entries in `spec.md` §B, **not** against the `AC-VFG-*` verification layer (which is Given-When-Then by design). REQ-001 (spec.md:L127) Ubiquitous "shall carry no assertion…"; REQ-002 (L133) Ubiquitous; REQ-003 (L138) Unwanted "shall not introduce"; REQ-004 (L144) Ubiquitous; REQ-005 (L149) Ubiquitous with generalized subject "Each retained assertion"; REQ-006 (L158) Unwanted "shall not modify"; REQ-007 (L165) Event-driven "**When** any verification for this SPEC is run, it shall be a single serial `go test` invocation". All seven match a GEARS pattern.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (spec.md:L2-L14): `id`, `title`, `version: "0.1.0"` (quoted semver), `status: draft` (enum), `created`/`updated: 2026-08-31` (ISO), `author`, `priority: P2` (enum), `phase: "v3.1.5 target"` (release label, not a prohibited lifecycle token), `module: "internal/kanban"`, `lifecycle: spec-anchored` (enum), `tags` (comma-separated string). No rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`) present. `tier: S` optional field present.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go, `internal/kanban`). Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — verb executable. Two referenced SPECs extracted: `SPEC-BACKLOG-LOCK-BUDGET-001`, `SPEC-STRESS-INVARIANT-VERDICT-001`. Both exist at `.moai/specs/<ID>/spec.md` **in this worktree** and both read `status: implemented` — outside `{retired, superseded, archived}`. No BLOCKING finding. (Note: neither exists in the primary checkout; the `project_root` / worktree-scoping instruction was load-bearing here, and an audit run against the primary tree would have reported them missing.)
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` = `0`. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-VACUOUS-FLOOR-GUARD-001/` returns no match (exit 1). `research.md` is absent by Tier S design; `plan.md` exists, so the verb was executable and this is a PASS, not an N/A.

No must-pass failure. The FAIL comes from the defect list.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | 1.0 / 0.75 boundary | Every requirement carries a single interpretation; measurable thresholds throughout. One residual ambiguity: REQ-VFG-001 (spec.md:L131) mandates removing "lines 35-41", which includes the comment at board_lock_wait_test.go:35-36, while REQ-VFG-004 (L144) requires "the guard's comment shall state…". Resolved only in plan.md:L106-110 (delete, then rewrite). |
| Completeness | 0.80 | 0.75 | All sections present (HISTORY L20, WHY §A, approach §A.4, REQUIREMENTS §B, AC §C, Exclusions §D, Constraints §E, Tier §F, xrefs §G). `grep -c '^### Out of Scope — '` = `6`, each with specific `-` bullets. Deduction: the §A.4 argument is incomplete (D2) and no Out-of-Scope entry covers the constant family it overlooks. |
| Testability | 0.75 | 0.75 | Seven of eight ACs are binary and cite verbatim failure messages plus exact commands. One AC (AC-VFG-007, L260-274) is observable but its stated significance is arithmetically false (D3), so what it establishes requires interpretation. AC-VFG-006 (L246) omits the pre-plant prediction its three siblings carry (D4). |
| Traceability | 0.75 | 0.75 | Every AC references a REQ that exists. **One requirement is uncovered**: REQ-VFG-007 appears in no row of the §C mapping table (L176-185) and in no AC's Given/When/Then (D1). Rubric band "one REQ uncovered" = 0.75. |

Harmonic mean of {0.90, 0.80, 0.75, 0.75} = **0.7956 → 0.80**.

---

## Verification of the enforcer table (§A.3) — line by line against source

This is the claim the entire deletion argument rests on. Each row was checked against `internal/kanban/board_lock_wait_test.go` at HEAD `b2e3af7fb`, not carried from the SPEC.

| §A.3 claim | Claimed line | Source at this tree | Holds? |
|---|---|---|---|
| ten-lane contender count — `boardLockSupportedWriters != 10` | 45 | `board_lock_wait_test.go:45` `if boardLockSupportedWriters != 10 {` | **YES** — and it is an equality, i.e. stronger than the `>= 10` the clause needs |
| CI-class per-mutation cost — `boardLockCIMutationCost < 33*time.Millisecond` | 54 | `board_lock_wait_test.go:54` `if boardLockCIMutationCost < 33*time.Millisecond {` | **YES** |
| stated headroom factor — `boardLockHeadroom < 2` | 59 | `board_lock_wait_test.go:59` `if boardLockHeadroom < 2 {` | **YES** |
| budget IS that product — `boardLockWaitBudget != recomputed`, `t.Fatalf` | 28 | `board_lock_wait_test.go:28-33`, `t.Fatalf` | **YES** — hard stop, so line 39 is reached only where equality already held |

All four rows verify. Three are `t.Errorf` rather than `t.Fatalf`, which does not weaken them: the test still fails.

**Composed floor, arithmetic checked.** `budget == 10 × cost × headroom` ∧ `cost >= 33ms` ∧ `headroom >= 2` ⟹ `budget >= 10 × 33ms × 2 = 660ms`. The SPEC's figure is correct, and each conjunct is falsifiable by a change to exactly one constant.

**Vacuity of the deleted branch, confirmed independently.** `board_store.go:126` declares `boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom` as a `const`. `recomputed` (test L26-27) and `floor` (test L37-38) are therefore both textually and semantically the same expression as the budget itself, and the L28 `t.Fatalf` pins the budget to it. `budget < floor` is unreachable for every assignment of the three constants. Deleting it removes no coverage. **The chosen direction — deletion — is correct.**

**Baselines for AC-VFG-001 verified at this tree**: `grep -c 'boardLockWaitBudget <' internal/kanban/board_lock_wait_test.go` = `2` (L39, L102); `grep -c 'floor :=' …` = `2` (L37, L100). The SPEC's stated pre-repair baseline of 2/2 is accurate, and L198's `2 * boardLockWaitBudget; elapsed > bound` correctly does not match the first pattern.

**§A.2's dynamic evidence verified**: `.moai/reports/t372/mutant-headroom4-orchestrator.md` exists at this tree and records `389` `=== RUN` lines, exactly one `--- FAIL` (`TestBoardLockWaitBudgetCoversSerializedMutations`), and `--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs`. The attribution in §A.2 is accurate.

**Cross-SPEC premises verified**: `SPEC-STRESS-INVARIANT-VERDICT-001` REQ-SIV-010 (spec.md:L217-231) does name this branch verbatim as the self-comparison precedent; its §E (L359-363) does defer the repair; `AC-SIV-013` exists (acceptance.md:L25, progress.md:L75/L164) and is **OPEN at merge**, requiring ≥5 non-cancelled post-landing develop heads. §D's untouchability premise holds.

---

## Defects Found (structured defect-list)

**D1. REQ-VFG-007 has no acceptance criterion — uncovered requirement**
`.moai/specs/SPEC-VACUOUS-FLOOR-GUARD-001/spec.md`:L165-170 (requirement), L176-185 (mapping table)
The §C mapping table's eight rows cover REQ-VFG-001 (×2), 002, 003, 004, 005 (×5), 006. REQ-VFG-007 — the verification-load discipline (single serial `go test` scoped to `./internal/kanban/`, no `-race`, no background process, no `go test ./...`) — appears in no row and in no AC's Given/When/Then. The constraint is *embedded* in the other ACs' `When` commands, which is not the same as being *verified*: nothing in §C establishes that no background process was spawned, that no `-race` run occurred, or that `go test ./...` was not run. This is a HARD load-discipline requirement whose violation has a recorded incident (load 413, 2026-08-15) and whose breach leaves no trace in the artifacts the other ACs collect.
— Severity: **major** — Class: **blocking** — Required fix: add an acceptance criterion for REQ-VFG-007, or extend AC-VFG-008 with a REQ-VFG-007 row. A workable form: *Given the recorded evidence under `.moai/reports/t378/`, When the run log for every observation is read, Then every `go test` invocation recorded is `go test -timeout 600s -count=1 [-v -run …] ./internal/kanban/` — scoped to the package, carrying no `-race`, and issued one at a time — and no invocation of `go test ./...` and no backgrounded process (`&`, `nohup`, `run_in_background`) appears in any recorded command.* Adding a 9th AC exceeds the Tier S ceiling of 8, so prefer extending AC-VFG-008 (which already reads the committed change) with this row rather than adding AC-VFG-009.

**D2. §A.4's exhaustiveness claim is false — a third source of outside terms exists**
`spec.md`:L100-106 (the "Repair the inequality" bullet)
The SPEC states: *"Any floor built from terms outside the budget expression would have to reach for the stress constants — which is precisely the relation t372's guard already pins."* This is a universal impossibility claim and it does not hold at this tree. `internal/kanban/board_store.go` declares, in the same `const` block, `boardLockWaitMin = 5 * time.Millisecond` (L131), `boardLockWaitMax = 50 * time.Millisecond` (L132), and `boardLockWaitStep` (L136-ish, consumed at L152-156) — independently-authored figures that are neither inputs to the budget expression nor the stress constants. A floor of the shape `boardLockWaitBudget >= <named minimum attempt count> * boardLockWaitMax` would be a real comparison between independently-authored terms, exactly the property REQ-SIV-010 demands, and it would catch a failure mode nothing presently guards: raising `boardLockWaitMax` toward the budget leaves the policy with effectively no retries while every existing assertion stays green.
This does **not** overturn the choice of deletion — deleting a provably unreachable branch is right whatever other guard could exist — but the SPEC reaches that choice through a premise it has not verified, which is the shape `verification-claim-integrity.md` §1.1 surface 4 names (recommendation-premise claim).
— Severity: **major** — Class: **blocking** — Required fix: (a) narrow the claim to what is true — that no floor built from the three budget inputs can add information, and that the stress-constant relation is already pinned next door; and (b) add an `### Out of Scope — the retry-bound coherence relation` entry to §D naming `boardLockWaitMin` / `boardLockWaitMax` / `boardLockWaitStep` as a distinct, unguarded relation deliberately left to a separate card, so the omission is recorded rather than implied to be impossible.

**D3. AC-VFG-007's negative evidence uses a mutant its own composed floor would not catch — the stated significance is false**
`spec.md`:L260-274 (AC-VFG-007), L115-123 (§A.5), L94-96 (§A.4)
AC-VFG-007 reinstates the deleted branch, plants mutant M1 (`boardLockCIMutationCost` 33ms → 20ms), and reads the branch staying silent as the deletion's evidence. §A.5 states this *"shows the removed branch could not be made RED even under a mutant that a real floor would have caught."*
Both halves fail on the SPEC's own arithmetic. Under M1 the budget becomes `10 × 20ms × 5 = 1.0s`, which is **above** the `660ms` composed floor §A.3 derives — so no "real floor" as REQ-BLB-002 defines it would have caught M1. What actually catches M1 is the cost-floor assertion at L54, which is a different obligation. And a single observation of silence does not license *"could not be made RED"*; the universal claim rests on the static proof (same expression, pinned by the L28 equality), not on this criterion.
As written, the run-phase would produce a record under `.moai/reports/t378/` whose stated conclusion does not follow from its observation — a criterion that reads as evidence and is not. That is precisely the t241-class failure this card exists to repair, reproduced inside the card, so §A.5's discharge is presently a gesture rather than a discharge.
— Severity: **critical** — Class: **blocking** — Required fix: switch AC-VFG-007's planted mutant from M1 to **M2** (`boardLockHeadroom` 5 → 1). Under M2 the budget falls to `330ms`, **below** the 660ms composed floor, so a genuine REQ-BLB-002 floor *would* fire — while the reinstated branch's own floor moves with it (`10 × 33ms × 1 = 330ms`) and stays silent. That construction makes the criterion say something the static argument does not already trivially give. Then restate §A.5 and AC-VFG-007's closing limit as: *the branch was not RED under a mutant that breaches the composed 660ms floor; the universal unreachability claim rests on the L28 equality pinning both operands, not on this observation.*

**D4. AC-VFG-006 omits the pre-plant prediction its three sibling mutant criteria carry**
`spec.md`:L246-258
AC-VFG-003 (t372's guard cost-independent, expected GREEN), AC-VFG-004 (t372's guard reddens, `10 × 1 = 10 < 48`; `TestConcurrencyStress` may redden), and AC-VFG-005 (t372's guard reddens, `8 × 5 = 40 < 48`) each state their census prediction in advance, which is what REQ-VFG-005's census clause exists to produce. AC-VFG-006 records the census command but states no prediction — yet the 1400ms mutant **will** redden t372's guard (`floor = 48 × 33ms = 1.584s > 1.400s`, `t.Fatalf` at board_lock_wait_test.go:102-112), and may perturb the budget-derived deadlines in `integration_lock_mutation.go:97`, `backlog_store.go:730`, and `integration_lock_cross_test.go:351`. Without the prediction, that second RED arrives as a surprise indistinguishable from a real one.
— Severity: **minor** — Class: **blocking** (REQ-VFG-005 states the census-before-plant obligation; this criterion under-implements it) — Required fix: add to AC-VFG-006's Given clause the prediction that `TestBoardLockWaitBudgetCoversSerializedMutations` reddens under the 1400ms form (`1400ms < 48 × 33ms = 1584ms`) and that any budget-consuming deadline sites named by the census are to be attributed individually.

**D5. REQ-VFG-001 mandates removing the comment that REQ-VFG-004 mandates writing**
`spec.md`:L127-131 vs L144-147
REQ-VFG-001 requires removal of "lines 35-41", which spans the explanatory comment at `board_lock_wait_test.go:35-36` as well as the branch at L37-41. REQ-VFG-004 requires "the guard's comment shall state where REQ-BLB-002's floor obligation is enforced". Read together the two requirements are satisfiable only by knowing that the removed comment and the written comment are different comments — a fact stated only in plan.md:L106-110, not in `spec.md`.
— Severity: **minor** — Class: **optional** — Required fix: reword REQ-VFG-001 to "the `budget < floor` branch and its preceding comment at lines 35-41 shall be removed and replaced by the comment REQ-VFG-004 specifies."

**D6. `related_specs` is not a field in the canonical frontmatter schema**
`spec.md`:L15
`related_specs: [SPEC-BACKLOG-LOCK-BUDGET-001, SPEC-STRESS-INVARIANT-VERDICT-001]` is neither one of the 12 required fields nor one of the six documented optional fields (`issue_number`, `depends_on`, `lint.skip`, `bc_id`, `amendment_of`, `tier`) in `.claude/rules/moai/development/spec-frontmatter-schema.md`. It produces no lint finding (`FrontmatterSchemaRule` checks required-field presence only), and the relationship it records is genuinely carried in §G, so nothing is lost by removing it.
— Severity: **minor** — Class: **optional** — Required fix: drop the field (the §G cross-references already carry the information), or propose it to the schema SSOT in a separate card.

**D7. Tier S sits exactly at the 8/8 ceiling for a ~10-line deletion**
`spec.md`:L172-185, L337-344
7 REQ / 8 AC against the Tier S ceilings of 8 / 8 — verified against `spec-workflow.md` § SPEC Complexity Tier, and the Tier S PASS threshold of 0.75 is likewise correct. The Tier classification itself is honest: one file, ~10 LOC, far inside `< 300 LOC` / `< 5 files`. Five of the eight ACs are mutant-evidence criteria mandated by the §A.5 self-application discipline, so the count is earned rather than padding. The cost is that there is no headroom left — D1's corrective cannot be a 9th AC without breaching the ceiling.
— Severity: **minor** — Class: **optional** — Required fix: none mandated. If the artifact is revised, consider collapsing AC-VFG-003..006 into one criterion with four mutant rows, which loses no evidence and reopens three AC slots.

---

## Recommendation

The SPEC's central judgement is **correct and independently verified**: the `budget < floor` branch is unreachable for every input assignment, the four retained assertions do compose into a real `budget >= 660ms` floor, and deletion loses no coverage. The §A.3 enforcer table holds row by row against source, the arithmetic is right, the dynamic evidence from t372 is real and accurately quoted, and the cross-SPEC untouchability premise for `TestBoardLockWaitBudgetCoversSerializedMutations` is verified (AC-SIV-013 open). Scope discipline in REQ-VFG-006 / §D / AC-VFG-008 is tight and mechanically checkable.

It fails on three things, in priority order:

1. **Fix D3 first.** The negative-evidence criterion the whole §A.5 self-application discharge rests on uses a mutant that the SPEC's own composed floor would not have caught, and then claims a real floor would have caught it. Switch the planted mutant to M2 (headroom 5 → 1, budget 330ms, below the 660ms floor) and restate the limit. Without this, the card ships a false-evidence artifact of exactly the kind it exists to eliminate — and, as this audit's own reliance on `.moai/reports/t372/` shows, such artifacts get cited by later cards.
2. **Fix D1.** Give REQ-VFG-007 a verification. Extend AC-VFG-008 rather than adding a 9th criterion, so the Tier S ceiling holds.
3. **Fix D2.** Narrow the exhaustiveness claim in §A.4 to what is true, and record the `boardLockWaitMin` / `boardLockWaitMax` / `boardLockWaitStep` coherence relation as an explicit Out-of-Scope item rather than leaving it implied impossible.

D4 is a one-line addition to AC-VFG-006's Given clause and should ride along. D5-D7 are optional and left to the orchestrator's discretion.

The score (0.80) clears the Tier S threshold (0.75) and no must-pass criterion fails; the FAIL is issued on the blocking defect list, not on the aggregate. The corrective is confined to `spec.md` and touches no code, so the confirming re-audit is scoped to the D1-D4 delta rather than a full re-audit.
