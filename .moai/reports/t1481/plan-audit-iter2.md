auditor-model: claude-opus-5-5

# SPEC Review Report: SPEC-FACTORY-DECISION-AUTO-001
Iteration: 2/3 (delta-scoped re-audit; CN-4 re-run in full)
Verdict: FAIL
Overall Score: 0.80
Plan Artifact Hash: not recomputed this iteration (Gap — see §4)
Auditor Version: plan-auditor (agent definition as loaded in this session)

verdict: FAIL
audited_sha: 5d094991f

Card: t1481 · Tier L (threshold 0.85) · Tree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a6b97f013ea3df7de · branch WT-decision-automation · HEAD 5d094991f (iter1 audited ba2033d22, FAIL 0.74)

Reasoning context ignored per M1 Context Isolation. Known bootstrap gap noted, not failed on alone: decision-index.md labels Q1-Q16 `LEADER-DECIDED` (decision-index.md:5-8), outside manager-spec's four-label vocabulary until REQ-FDA-005 lands.

Cross-model: `audit_multi` (target baseBranch, project_root = card tree): claude FAIL (in-session anchor, required), codex FAIL (required, 2×P1 + 2×P2), GLM inconclusive (advisory, "z.ai response carried no text content"). overall_verdict=fail, disagreement_flag=false, participant_count=2. No `audit_receipt` id in the result. build_lag: MCP binary 45600e4ee is an ancestor of HEAD 5d094991f.

---

## 1. Claim

FAIL at 0.80 (< 0.85). All ten iter1 blocking defects D1-D10 are closed in substance, and MP-1..MP-9 all pass. The revision introduces or exposes seven new blocking defects. Two of them are critical: the `audit` decider's "own lease" premise contradicts the store, because a kickoff card holds no lease; and the decision-index that gates the decider is outside the plan hash, so a post-audit reclassification bypasses the product-level keep-set check. The new REQ-FDA-013 exception (focus item) is not fully mechanical. Its condition 1 names no relation kind or contract field. Its two verdict fields are bound by no REQ-006 or AC-006. Its exit verdict is undefined.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: REQ-FDA-001..025 in sequence (spec.md:89-245), no gaps or duplicates. Trace verb: `COLLECTED: 25 REQ definitions (acceptance input: read)`.
- [PASS] MP-2 GEARS (requirement layer, spec.md §C only): Ubiquitous 001, 002, 006, 009, 010, 024, 025; Event-driven 003, 004, 007, 008, 011-016, 018, 020-023; Where 005, 017; While 019. REQ-FDA-024 is recast as "shall refuse" (spec.md:239-242), which closes D12. The ACs are Given-When-Then in acceptance.md and were not graded here.
- [PASS] MP-3 frontmatter: id, title, version "0.2.0" (quoted), status draft, created/updated 2026-10-03, author, priority P1, phase, module, lifecycle spec-anchored, tags (comma string) (spec.md:2-13). No rejected aliases.
- [N/A→PASS] MP-4: not multi-language tooling. Template neutrality is stated at spec.md:252.
- [PASS] MP-5 D7: 6 referenced SPECs, all `status: completed`; no REVIEW lines, no BLOCKING. (See N1 for a conflict with an existing test-asserted invariant; it is not a status-based D7 finding.)
- [PASS] MP-6 D8: `grep -c syscall spec.md` → `0`.
- [PASS] MP-7: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no output, `exit=1`.
- [PASS] MP-8 RED-now: acceptance.md:10-16 pins tree `ba2033d22…` at document level. research.md §2 carries the command, verbatim stdout and exit for P1-P24, plus control P1c. `git diff --stat ba2033d22 5d094991f` shows only `.moai/specs/SPEC-FACTORY-DECISION-AUTO-001/*` and `.moai/reports/t1481/plan-audit-iter1.md` changed, so the probed tree is unchanged. Re-executed at HEAD 5d094991f (see §2): every probe reproduces its recorded exit code and its deciding lines. The 22 release-blocking ACs all cite a reproducing probe.
- [PASS] MP-9 ordering: CN-4 again collected 0 milestone headings (milestones are in a table, plan.md:37-47), which is a GAP, so I read the order by hand: M0→M8. Candidates checked: acceptance.md:12 and :112 ("M0 baseline committed before the first implementation commit") match M0 being first; :30 ("existing T7/T13 tests characterized before the change") is internal to M2; :25 ("step ② before step ③") is doctrine order, not milestone order. The plan.md:51-53 dependencies agree with the table. No conflict.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.70 | 0.50-0.75 | REQ-FDA-013 cond. 1 is not operationalized (spec.md:167-168); "sync thresholds" undefined (spec.md:140-141 vs acceptance.md:30); REQ-017/018 product-level rule tension (spec.md:198-199 vs 206-208) |
| Completeness | 0.85 | 0.75-1.0 | All sections present (HISTORY :20, §A :28, §B :63, §C :85, §F :277 with 5 Out-of-Scope H3s). Exception exit verdict missing; unrankable implementation-level row unhandled |
| Testability | 0.80 | 0.75 | Matrix is binary; AC-019/020 classification is incoherent with the RG definition (acceptance.md:14-16, 40-41); no fixture for a post-audit decision-index edit or an unrankable row |
| Traceability | 0.85 | 0.75-1.0 | 25/25, no UNCOVERED or ORPHAN; but `defect_class`/`reread_hunks` (consumed by REQ-013) are emitted by no REQ and verified by no AC; P13 greps `release_blocking`, a token the SPEC never defines |

Aggregate ≈ 0.80 < 0.85 (Tier L). Not a regression (iter1 0.74), so no STOP.

## Defects Found (structured defect-list)

N1. LEASE-AT-KICKOFF — spec.md:185-191 (REQ-FDA-015, REQ-FDA-016), acceptance.md:36-37, 75-78, design.md:126-135 — The REQs presuppose that a card in `kickoff` carries the lane's lease: "keeping its current lease", "the card its own lease holds", "leased by worker-3". The store clears the lease on T7. `internal/homestate/fr_lease_test.go:148-162` (`TestFR_AC013_DecisionPendingHoldsNoLease`) asserts `k.LeaseHolder == ""` and `k.LeaseExpiresAt == ""` after T7 → kickoff, by design ("kickoff and needs-decision hold no lease"). Codex reports running it PASS at 5d094991f. As written, the lane admission of REQ-016 can never match, and REQ-015 contradicts an existing tested invariant. — Severity: critical — Class: blocking — Required fix: key the lane admission and the carry-over on the registered owner (`OwnerLabel`), which survives T7, rather than on the lease; or explicitly change T7's lease policy and reconcile the originating AC-013 invariant (name it, and amend it with a row). Add a fixture for the real chain plan-audit→kickoff→run.

N2. DECISION-INDEX-UNHASHED — spec.md:177-184 (REQ-FDA-014), spec.md:155-157 (REQ-FDA-011), design.md:126-131 — The `audit` decider reads the *current* decision-index for open product-level rows. Codex measured that the plan-artifact hash does not cover decision-index.md (a `product-level` row with an empty verdict changed to `implementation-level` + `DEFAULT-APPLIED` leaves the `ComputeHash` value unchanged), and REQ-011 exempts decision-index.md from the delta-eligibility diff. After the audit, a lane can therefore reclassify a product-level row and self-approve Kickoff, which bypasses a keep-set item (REQ-FDA-024). (Measured by codex, not re-run by me; see Gaps.) — Severity: critical — Class: blocking — Required fix: bind the audited decision-index into the admission check (in the hash, or by comparing it against the decision-index at `audited_sha`). After the audit, permit only `Operator verdict:` fills of the `DEFAULT-APPLIED` form on rows that were `implementation-level` with a `Default:` at `audited_sha`; any Class, Label, anchor or Default change refuses. Remove decision-index.md from REQ-011's exemption, or limit the exemption to verdict-line fills. Add an AC-014 refusal fixture: a product-level row reclassified after the audit.

N3. UNRANKABLE-IMPL-ROW — spec.md:181-182 (REQ-014), spec.md:202-205 (REQ-018), design.md:153-154 — design.md says "A row the rule cannot rank carries no Default and blocks". REQ-014, however, refuses only on an empty *product-level* row, and REQ-018 handles only implementation-level rows *with* a Default. An implementation-level FOUNDER row with no Default and an empty verdict is therefore admitted by the decider. Codex flagged this as P2. — Severity: major — Class: blocking — Required fix: REQ-014 refuses on any FOUNDER row whose verdict is empty and is not a valid DEFAULT-APPLIED fill. Add that fixture to AC-014 and AC-018.

N4. EXCEPTION-COND1-NOT-MECHANICAL — spec.md:166-168 (REQ-FDA-013), design.md:112-113, acceptance.md:34 — "lies on the dependency path, per queue relation records, of a card inside a release scope recorded as operator-approved in a mission contract or a standing board record" names no relation kinds and no fields. The tree has two relation vocabularies: `internal/kanban/backlog_store.go:141-163` (`contains|absorbs|replaces|conflicts|blocks|depends`) and `internal/kanban/gtd_relation.go:12-20` (`depends_on|part_of|…`). `internal/mission/contract.go:32-46` carries generic `Scope []string` + `Approved bool` and no release-scope field. A standing board record's body is free text. Focus item: "all three conditions mechanically decidable" is not met for condition 1. — Severity: major — Class: blocking — Required fix: name the store and the relation kinds that form a dependency edge, and their direction. Define "release scope operator-approved" as a machine predicate, for example a sealed `MissionContract` with `Approved == true` whose `Scope` contains the target card id, or a standing board record with a structured `release_scope` field. Board free text does not count. Add an AC-013 fixture for each source.

N5. EXCEPTION-FIELDS-AND-EXIT — spec.md:119-125 (REQ-006 field list) vs spec.md:169-173 (REQ-013); acceptance.md:27, 34; research.md:124-129 (P13) — (a) REQ-013 consumes `defect_class` and `reread_hunks`, but REQ-006 does not oblige the auditor to emit them (only plan.md:41 and design.md:77 mention them), and AC-006 does not verify them. P13 greps `release_blocking`, a token no requirement defines. (b) The exception's outcome is a "re-read confirmation", with no statement that it is a full machine verdict block (label, score, must_pass_failed, blocking_findings 0, new `audited_sha`, hash) admitted by the REQ-009 predicate. Without that, either REQ-014 refuses Kickoff (the exception leads nowhere) or the predicate is bypassed. Nothing states how the hold is resolved (`resolves`). (c) "a fix limited to the listed hunks" has no mechanical check, unlike REQ-011's diff test. Codex flagged (b) as P2. — Severity: major — Class: blocking — Required fix: add `defect_class` (closed enum including `ac-wording`) and `reread_hunks` to REQ-006 and AC-006, and fix P13 to grep the defined tokens. REQ-013 states that the auditor emits a full verdict block at the fix SHA, that the block must pass the plan-phase predicate, and that the leader's exception record or a following record `resolves` the hold wait id. The hunk limit is checked with the REQ-011 diff rule against `reread_hunks`. AC-013 asserts all of these.

N6. RG-MISCLASSIFICATION — acceptance.md:14-16 vs :40-41, :112 — The RG definition reads "assert a property that holds at their baseline and must still hold at close". AC-FDA-019 (wait ids, `resolves`, one-shot recheck) and AC-FDA-020 (zero Open on a cache hit) assert new behavior that does *not* hold at the M0 commit, and both carry valid RED cells (P19, P20, P21). Labeling them RG takes the D10 fix (focus: "wait ends only on a record whose resolves names it") and the D5 cache behavior out of the release-blocking set, and the DoD "re-asserted against their baselines at close" cannot be evaluated against them. — Severity: major — Class: blocking — Required fix: reclassify AC-019 and AC-020 as release-blocking, keeping their existing RED cells. Split the M0-value clauses ("shipped default equals M0(b)"; "M0(a) rate vs post-M7 re-measurement") into a measurement note or separate RG sub-assertions. Update the DoD count (22 → 24 release-blocking, RG = AC-024 plus the split clauses).

N7. SYNC-THRESHOLD-AMBIGUITY — spec.md:140-141 (REQ-009) vs acceptance.md:30, design.md:71 — REQ-009 has T13 check "the label and the sync thresholds". `card_transition.go:461-472` checks only the label today, and AC-009 says "PASS and PASS-WITH-DEBT admitted as today". The design's sync row adds "binding run conditions all disposed" as a predicate condition, which no REQ binds at T13 (REQ-008 binds the auditors). Whether T13 behavior changes is undecided. — Severity: minor — Class: blocking — Required fix: state either that T13 checks the label only (unchanged) or which named thresholds and fields it checks, and align design.md §3 and AC-009 with that.

O1. DEBT-FIELD-NAME — spec.md:123 "the phase (`run` or `sync`)", design.md:70/84 `dispose_in`, acceptance.md:27 "phase". — minor — optional — Use one field name.
O2. PRODUCT-LEVEL-DEFAULT — spec.md:198-199 ("no Class → product-level") vs :206-208 ("product-level exactly when …"). — minor — optional — State that the closed list governs Class assignment and that the missing-Class rule is a fail-closed fallback.
O3. ANCHOR-SECTION-FOR-TABLE-ROWS — design.md:101-105: an AC id anchor in the acceptance.md matrix is a table row, not a section. — minor — optional — Define the hunk range of a REQ/AC id anchor (the row, plus its Scenario heading section).
O4. NON-FACTORY-SPLIT — spec.md:162-165: outside a factory, "card creation … stay with the leader" has no leader. — minor — optional — State that the split proposal is informational for the user.

## Regression Check (iter1 defects)

- D1 MP8-CELLS — RESOLVED: research.md:24-199 holds P1-P24 + P1c with command, verbatim stdout and `exit=`; doc-level pin at acceptance.md:11-12; AC-025 now cites P10 (acceptance.md:46); M0 ACs no longer use M0 as a RED cell. All reproduce at 5d094991f.
- D2 PRED-PHASE — RESOLVED: REQ-009 is phase-parameterized (spec.md:136-143); AC-009 carries T13 fixtures (acceptance.md:30). Residual wording issue → N7.
- D3 PRED-UNDERSPEC — RESOLVED: `must_pass_failed` and `blocking_findings` are in REQ-006 (spec.md:120-121); the field list is in REQ-009; AC-009 has score<threshold and must_pass_failed=1 PASS-label fixtures.
- D4 AUDIT-DECIDER-BLOCKERS — RESOLVED as written: REQ-014 (spec.md:177-184) re-checks audit-ready, blocker, hold, product-level row and `audited_sha`; AC-014 has a fixture for each. Two new holes in the same guard → N2, N3.
- D5 BINDCACHE-STALE — RESOLVED: REQ-020 runs the probe on every hit (spec.md:220-224); REQ-021 invalidates and rebinds on retired or different run, and surfaces a probe failure (spec.md:225-228); §D constraint at :258-259; AC-021 has release-blocking fixtures. AC-020 classification → N6.
- D6 CEILING-DECIDABILITY — RESOLVED: REQ-011 uses a post-delta `audited_sha` diff ⊆ `fix_scope` plus identical REQ/AC id sets (spec.md:152-159); AC-011 has Out-of-Scope-edit and AC-added fixtures. Residual: decision-index exemption → N2.
- D7 CEILING-NONLANE — RESOLVED: REQ-010 binds every session (spec.md:150-151); final hit = ceiling + `auto_delta_rounds` (spec.md:160-161); non-lane notice plus override-only question (spec.md:163-164); AC-012 has a non-lane fixture.
- D8 SYNC-REREAD-PATH — RESOLVED: REQ-008 names `sync-audit-4dim` and sync-auditor; undisposed = must-pass FAIL (spec.md:131-135); AC-008 has a fixture for both owners.
- D9 RED-PROBE-RELEVANCE — RESOLVED: P5, P6, P8, P11, P13 and P18 now grep for the asserted absent token (P13's token choice → N5).
- D10 WAIT-RESOLUTION-LINK — RESOLVED in requirement: REQ-002 `resolves` (spec.md:94-98); REQ-019 says an unrelated same-card record leaves the wait open (spec.md:212-216); AC-019 scenario at acceptance.md:85-88. Classification → N6.
- D11 FORGED-VERDICT (optional) — RESOLVED: `audited_sha` binding in REQ-014.
- D12 LEGACY-UNWANTED (optional) — RESOLVED: REQ-024 is "shall refuse".
- D13 MAY (optional) — RESOLVED: REQ-017 has no "may".

## fix_scope (machine list for the next delta round)

- spec.md#REQ-FDA-006
- spec.md#REQ-FDA-009
- spec.md#REQ-FDA-011
- spec.md#REQ-FDA-013
- spec.md#REQ-FDA-014
- spec.md#REQ-FDA-015
- spec.md#REQ-FDA-016
- spec.md#REQ-FDA-018
- acceptance.md#verification-discipline (lines 10-16)
- acceptance.md#AC-FDA-006
- acceptance.md#AC-FDA-009
- acceptance.md#AC-FDA-011
- acceptance.md#AC-FDA-013
- acceptance.md#AC-FDA-014
- acceptance.md#AC-FDA-015
- acceptance.md#AC-FDA-016
- acceptance.md#AC-FDA-018
- acceptance.md#AC-FDA-019
- acceptance.md#AC-FDA-020
- acceptance.md#definition-of-done
- research.md#P13
- design.md#3-admission-predicate
- design.md#4-ceiling-policy
- design.md#5-factory-audit-decider
- plan.md#F-milestones (M2, M3 rows)

## 2. Evidence (commands run in this audit, observed output)

- `git log --oneline ba2033d22..5d094991f` → `5d094991f docs(SPEC-FACTORY-DECISION-AUTO-001): revise for plan-audit iter1 FAIL 0.74 (card t1481)`. `git diff --stat ba2033d22 5d094991f` → 8 files, all under `.moai/specs/SPEC-FACTORY-DECISION-AUTO-001/` plus `.moai/reports/t1481/plan-audit-iter1.md`.
- Probe re-execution (absolute paths into the card tree):
  - P1 → no output, `P1 exit=1`; P1c → `internal/cli/contract.go:449:		Use:   "contract",` `exit=0`
  - P2 `exit=1`; P3 `exit=1`; P4 → all four files `:0`, `exit=1`; P5 `exit=1`; P6 `exit=1`
  - P7 (`-c`) → local `:1`, template `:1`, `exit=0`; P8 `exit=1`
  - P9 → `rules.go:25`, `decide.go:393`, `card_transition.go:470` lines as recorded, `exit=0`; P9b → `108: {"T7", …guardVerdictPass}`, `114: {"T13", …}` `exit=0`
  - P10 `exit=1`; P11 `exit=1`; P12 (`-c`) → `:1` both, `exit=0`; P13 `exit=1`; P14 `exit=1`
  - P15 → `109: {"T8", CardKickoff, CardAssigned, guardKickoffDecision},` `exit=0`; P16 (`-c`) → `4` `exit=0`
  - P17-P20 `exit=1` each; P21 → `124: runState, _, probeErr := factoryHookProbeRun(ctx, dbPath, runID)` `exit=0`
  - P22 → `56: slog.Warn("factory hook: message broker close failed", "error", err)` `exit=0`; P23 `exit=1`; P24 → `1987:` / `2029:` `exit=0`
- Trace verb (scratchpad t1481-trace.sh, literal paths): `COLLECTED: 25 REQ definitions (acceptance input: read)`, no UNCOVERED, no ORPHAN.
- CN-4: milestone-heading count in plan.md → `0` (GAP; order read by hand). Ordering-word records in acceptance.md: lines 12, 25, 30, 35, 36, 83, 88, 112, all judged above.
- MP-7 `exit=1` (no markers). D8 `0`. D7 related statuses: 6× `status: completed`.
- N4: `internal/cli/todo_relate.go:41` `--relation <contains|absorbs|replaces|conflicts|blocks|depends>`; `internal/kanban/gtd_relation.go:12-20` kinds; `internal/mission/contract.go:32-46` `Scope []string`, `Approved bool`; `grep -rln -i 'release.scope|releaseScope|release_scope' internal/` → no output.
- N1: Read `internal/homestate/fr_lease_test.go:146-162` (assertion `k.LeaseHolder != "" || k.LeaseExpiresAt != ""` → Fatalf).
- N7: Read `internal/homestate/card_transition.go:461-472` (T7/T13 label-only check).
- audit_multi: overall_verdict fail; codex findings: lease cleared at kickoff (P1, ran the test PASS), decision-index not in hash (P1, isolated `ComputeHash` test), unrankable implementation-level row (P2), exception exit verdict (P2).

## 3. Baseline-attribution

Every grep and file read above ran in this audit against the card tree at HEAD 5d094991f. That tree differs from ba2033d22 only in SPEC and report files, per the git diff above. Backend results are attributed to MCP build 45600e4ee (an ancestor of HEAD; build_lag reported by the tool).

## 4. Gaps

- Refused commands: the worktree guard refused `git -C <card tree> …`, a `sh script > out 2>&1` run, and a `sed` using a shell variable. Fallbacks: git read-only commands run from my own worktree (shared object store, explicit SHAs); greps with absolute paths; the Read tool. These outputs are equivalent for the claims made.
- I did not run Go tests. The N1 PASS of `TestFR_AC013_DecisionPendingHoldsNoLease` and the N2 `ComputeHash` behavior are codex-measured. I independently read the N1 test assertion; I did not verify N2 (`ComputeHash`'s input set) myself, so N2 rests on codex's measurement plus the SPEC text (REQ-011's decision-index exemption).
- Plan Artifact Hash was not recomputed.
- GLM advisory was inconclusive; `audit_multi` returned no `audit_receipt`.

## 5. Residual-risk

- residual_risk_note (verbatim): "required-backend FAIL: claude, codex".
- Even after N1-N7, lane Kickoff self-approval remains self-attested beyond the verdict file (spec.md §E.2).
- Run M0 values may move defaults that the ACs rely on.

## Recommendation

Fix N1 and N2 first (keep-set and reachability), then N3, N5, N4, N6 and N7, within the fix_scope list above. This is iteration 2 of a Tier L ceiling of 3. The next round is the last before the ceiling policy (or the current Max-3 contract) applies.

## Operational Notes (unverified)

- assumption: measure N2 with `grep -n 'func ComputeHash' -A30 internal/contract/kickoff/*.go` to confirm which artifacts enter the hash.
- inferred (rule: OwnerLabel is set at assignment, card_transition.go:440-442): measure whether OwnerLabel survives T7 with a fixture read of `k.OwnerLabel` after the T7 transition.
