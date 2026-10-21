auditor-model: glm-5.3-flash

# SPEC Review Report: SPEC-FACTORY-LEASE-BUSY-001
Iteration: 1/3
Verdict: FAIL
Overall Score: 0.69
Plan Artifact Hash: computed over spec.md + plan.md + acceptance.md (Tier M set) at HEAD 61b5af4b14505926b3435bc9daf6d3e38eaafecc
Auditor Version: plan-auditor/v3.2 (served by glm-5.3-flash; independent of manager-spec)

verdict: FAIL
audited_sha: 61b5af4b14505926b3435bc9daf6d3e38eaafecc
overall_score: 0.69
must_pass_failed: 0
blocking_count: 3
scope: full
defect_class: ac-wording (D1, D5, D7, D10); design (D2, D3, D4, D8); other (D6, D9)
fix_scope: spec.md#REQ-FLB-004, spec.md#REQ-FLB-005, spec.md#REQ-FLB-008, spec.md#REQ-FLB-009, acceptance.md#AC-FLB-004, acceptance.md#AC-FLB-006, acceptance.md#AC-FLB-007, acceptance.md#AC-FLB-009, acceptance.md#AC-FLB-010, plan.md#§F-M1, plan.md#§F-M3, plan.md#§F-M4
reread_hunks: spec.md#Requirements, spec.md#Constraints, acceptance.md#§A, acceptance.md#§B, plan.md#§F
convergence_overall: fail
required_backend: codex fail

Reasoning context from the SPEC author was not present in the spawn input; the audit read only the SPEC artifacts, the card's file-map survey, and the tree. M1 Context Isolation holds. Tier M input contract: spec.md + plan.md + acceptance.md read (+ progress.md, decision-index.md present in the directory); no design.md/research.md required at Tier M.

## Must-Pass Results
- [PASS] MP-1 REQ number consistency: `grep -oE 'REQ-FLB-[0-9]+' spec.md | sort -u` → exactly REQ-FLB-001..012, 12 unique, sequential, zero-padding consistent (spec.md:83-128).
- [PASS] MP-2 EARS/GEARS format compliance (requirement layer, spec.md § Requirements — ACs are the verification layer and are graded under Group 4, not here): REQ-FLB-001/006/007/010/011 are Ubiquitous ("The <subject> shall …", spec.md:83, 100, 102, 112, 119); REQ-FLB-002/003/004/008/009/012 are Event-driven ("When …, the <subject> shall …", spec.md:87, 90, 94, 105, 109, 124). REQ-FLB-005 (spec.md:98) is the legacy EARS negative form ("shall not") — counted as a legacy equivalent under the backward-compatibility window (active through 2026-11-22; today 2026-10-21). No IF/THEN forms anywhere; no informal "should/may" in normative text.
- [PASS] MP-3 YAML frontmatter validity: all 12 canonical fields present with correct types (spec.md:2-13 — id, quoted title, quoted "0.1.0", status: draft, created/updated 2026-10-21, author, priority: P1, phase: "v3.2.0 target" (a release target, not a prohibited stage name), module, lifecycle: spec-anchored, comma-string tags). No snake_case aliases. Optional tier: M / card / depends_on / related_specs well-formed.
- [N/A] MP-4 Section 22 language neutrality: single-language (Go) SPEC — auto-passes per LN-3.
- [PASS] MP-5 D7 cross-SPEC reconciliation: D7 verb output — SPEC-FACTORY-ATOMIC-LEASE-001 (in-progress), SPEC-TODO-AUTO-PICK-001 (completed), SPEC-FACTORY-LEASE-RENEW-001 (completed) all exist; zero REVIEW lines (no retired/superseded/archived reference); no SHOULD lines (the fourth grep hit was the SPEC's own id).
- [PASS] MP-6 D8 cross-platform discipline: `grep -c syscall` over spec.md / plan.md / acceptance.md → 0 / 0 / 0. Auto-pass per D8-4 (observed absence, measured).
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no match in plan.md; research.md absent, which is the correct Tier M artifact set (N/A branch stated). decision-index.md routes open items through the decision-gate mechanism, not [NEEDS CLARIFICATION] markers.
- [N/A] MP-8 RED-now cell re-execution: no acceptance criterion in acceptance.md is classified release-blocking, so the four-element RED-now cell obligation does not bind (reason stated per the MP-4 precedent). Evidence gathered anyway: AC-FLB-001's RED-now premise was re-executed indirectly — both pinned tests pass on this tree and their verbatim stderr shows the misclassification the SPEC fixes (see Claim/Evidence table R1).
- [PASS] MP-9 cross-artifact ordering consistency: the CN-4 awk verb was refused by the worktree isolation guard ("awk … construct too complex to verify"); re-run in plain-command form per the refusal's instruction: `grep -inE 'before|after|first|prior to|pre-change' acceptance.md` → 5 candidates (L5, L38, L48, L93, L98), each hand-read against plan §F order M0→M1→M2→M3→M4→M5→M6 (no `Exit:` bindings exist in the plan, so no CONFLICT class is reachable). AC-FLB-008's "baseline … precedes the first commit modifying the read path" (acceptance.md:48) is jointly satisfiable with M4-before-M5; DoD #1's "M0 gate recorded as open before any implementation commit" matches M0-first. Zero conflicts; MP-9 passes on the hand-read of all surfaced candidates.

## Category Scores (0.0-1.0, rubric-anchored)
| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | one-or-two-requirement ambiguity | Two-axis problem statement and class definitions are precise (spec.md:46-59); defects: REQ-FLB-004 mandates a causal narrative the measured evidence contradicts (D4), REQ-FLB-009's trigger names no locatable event (D8), AC-FLB-001 defers the (b)/(c) token strings to M1 (recorded in plan §F) |
| Completeness | 0.50 | multiple substantive gaps | The `--wait` retry surface the fix provably changes is absent from REQ/plan/AC (D3); M4 baseline recipe lacks the executable form (D6); text preservation unmeasured for 3 of 4 preserved classes (D7); M1's deliverable lands in a file its owner may not edit (D9) |
| Testability | 0.50 | several ACs require judgment or admit mutants | AC-FLB-006 unsatisfiable on one arm as written (D1); AC-FLB-009 passes a large-regression mutant (D5); AC-FLB-010's pin check needs name-level anchors (D10). Strong points: AC-FLB-010's empty-sweep count guard, AC-FLB-005's byte-identity pin, AC-FLB-008's commit-graph witness |
| Traceability | 1.00 | full coverage | acceptance.md §B table maps all 12 REQs (001→001, 002→001, 003→001, 004→001+002, 005→002, 006→003, 007→004, 008→006, 009→007, 010→005, 011→008, 012→009); AC-FLB-010 declared regression-only; no orphan AC. (Traceability awk verb refused by the worktree guard; mapping hand-verified line-by-line against §B — the hand-read is the measurement.) Partial-coverage note on REQ-FLB-007→AC-FLB-004 is graded under Testability (D7), not here |

Aggregate 0.69 < Tier M skip threshold 0.80. Not skip-eligible.

## Defects Found (structured defect-list)

D1. FLB-D1 — acceptance.md:37-41 (AC-FLB-006) — The Then clause requires "the promoted queue item is restored to `queued`" for all three claim classes, but the assertion it cites (factory_lease_bound_test.go:302-307) asserts the opposite for the mid-claim-stall arm: card stays `assigned` to the lane and the queue item stays `picked`, and that test passes in that state today. Measured this run (R1): `record=assigned holder= queue=picked`. As written AC-FLB-006 is unsatisfiable without violating REQ-FLB-008 ("the same compensation it runs today") — an impossible-criterion direction. — Severity: BLOCKING — Class: blocking — Required fix: split AC-FLB-006 by failure point: pre-record-write arms restore to `queued`; mid/post-write arms preserve `assigned`/`picked` exactly as the cited assertions pin today.

D2. FLB-D2 — spec.md:98-99 (REQ-FLB-005) vs spec.md:132-137 (Out of Scope, queue-lock) and acceptance.md:60-61 (AC-FLB-010 pins 443/464) — REQ-FLB-005 forbids surfacing `raced` for any lease-attempt failure that is not a race lost to another claimant, while the preserved queue-lock refusal (timeout with no winner; TestFactoryLeaseQueueLockStallBounded at factory_lease_bound_test.go:447, assertions at 464) keeps token `raced` by explicit Out-of-Scope disposition (decision-index Q4 DEFAULT-APPLIED). The requirement layer and the exclusion layer contradict as written. — Severity: BLOCKING — Class: blocking — Required fix: scope REQ-FLB-005 to the claim-write failure classification this SPEC performs ("a lease-attempt failure classified by this SPEC's cause table"), or add the queue-lock site as a named exception in the requirement text.

D3. FLB-D3 — spec.md:105-108 (REQ-FLB-008) / plan.md:99-102 (M2) — The reclassification is not "refusal surface only": `waitable()` (factory_card.go:1270-1276) keys retry-on-`--wait` on the token (`raced`, `serial-slot`, `quota-hold` are waitable; all else permanent), and the `--wait` loop gates on `refusal.waitable()` (factory_card.go:1768). Re-tokening record-contention and claim-deadline makes those refusals permanent, so `--wait` stops waiting through transient record contention — a behavior change on a third surface that no REQ, plan clause, or AC addresses. — Severity: BLOCKING — Class: blocking — Required fix: state the waitable disposition for the two new tokens in a requirement (and in M2), and add an AC covering the lease-success-after-transient-contention path under `--wait`.

D4. FLB-D4 — spec.md:94-97 (REQ-FLB-004) and spec.md:156-157 (Constraints, deadline precedence) — The mandated detail narrative "the host is under load, not a lock race" is contradicted by this run's verbatim stderr: `context deadline exceeded; database is locked (5) (SQLITE_BUSY)` joined with the deadline signal — a competing-writer lock error co-occurring with expiry. The SPEC mandates a potentially false causal claim in the very surface it exists to make cause-accurate. — Severity: SHOULD-FIX — Class: optional (does not block implementation; degrades the SPEC's own purpose) — Required fix: have the claim-deadline detail name the observed signals (deadline expired; busy error present/absent) rather than assert the load cause unconditionally.

D5. FLB-D5 — acceptance.md:50-54 (AC-FLB-009) — The "defect not reproduced" arm conditions only on the baseline (baseline p95 < 200 ms); post-change latency is unbounded in that arm, so a 10x regression passes when the baseline is fast (mutant probe fails, verification-completeness §2). — Severity: SHOULD-FIX — Class: blocking (a criterion this document states admits the failure it guards against) — Required fix: add a post-change ceiling to the second arm (e.g., post p95 ≤ max(2× baseline, 200 ms)), or require the not-reproduced arm to also record post-change p95 < 200 ms.

D6. FLB-D6 — plan.md:106-111 (M4) — The baseline recipe fixes N, load condition, and percentiles but not the executable form: which binary (the installed moai lags this tree — audit_multi's build_lag notice names 8249bff8c as an ancestor of HEAD 61b5af4b1), which factory root (the default root reaches the primary checkout's live DB), what fixture data, and the exact invocation. Two runs cannot be attributed as "the same recipe" without these. — Severity: SHOULD-FIX — Class: blocking — Required fix: pin per-revision binary build, isolated temp project root, fixed fixture, and the exact command in the M4 recipe (and mirror the command into AC-FLB-008/009's recipe reference).

D7. FLB-D7 — acceptance.md:28-32 (AC-FLB-004) vs spec.md:102-104 (REQ-FLB-007) — REQ-FLB-007 requires unchanged refusal texts for serial-slot, quota-hold, backend-skip, foreign-worktree; AC-FLB-004 checks exit codes only, and byte-identity is pinned for serial-slot alone (AC-FLB-005). The text half of the requirement is unmeasured for three of the four classes. — Severity: SHOULD-FIX — Class: blocking — Required fix: extend AC-FLB-004 with verbatim-text assertions (or recorded-prior-text comparison) for quota-hold, backend-skip, and foreign-worktree.

D8. FLB-D8 — spec.md:109-111 (REQ-FLB-009), plan.md:103-105 (M3), acceptance.md:42-44 (AC-FLB-007) — The trigger "the lease path's serial-slot read fails as busy" names no existing refusal-producing event: the serial-slot refusal fires at factory_card.go:1368 from `factorySerialInFlightExcluding` (factory_card.go:330), a pure function over an in-memory card list returning bool; DB read failures in the peek path fail open (factory_card.go:1868-1878) and produce no refusal. M3's "attribute a busy serial-slot read … via the slot seam" has no locatable seam as described. — Severity: SHOULD-FIX — Class: blocking — Required fix: name the actual code site and an induction method for the busy serial read, or re-scope/drop REQ-FLB-009 (note the adjacent hazard codex flagged: reclassifying a genuine serial-slot outcome would collide with REQ-FLB-010's byte-identical preservation).

D9. FLB-D9 — plan.md:96-98 (M1 deliverable) — "the class table recorded in this file's §F amendment note" assigns the run-phase agent a plan.md body edit — a forbidden ownership crossing (spec-frontmatter-schema § Forbidden ownership crossings: manager-develop MUST NOT modify plan.md body content; the route is a blocker report and manager-spec re-delegation). — Severity: SHOULD-FIX — Class: blocking — Required fix: author the token/class table in plan-phase now (manager-spec amends plan §F or spec § Constraints), or change M1's deliverable to record the table in progress.md §E.2 / a report artifact.

D10. FLB-D10 — acceptance.md:60-64 (AC-FLB-010) — "the 443/464 queue-lock pins appear in the matched list untouched and passing": `-v` output lists test function names, not line numbers; the AC never names the functions those lines belong to (TestFactoryLeaseQueueLockStallBounded, factory_lease_bound_test.go:447; anchors 109/443 are doc-comment lines, 151/298/464 are assertions). — Severity: MINOR — Class: optional — Required fix: name the test functions (and subtests) the swept list must contain.

Regression Check (Iteration 2+ only): N/A — iteration 1.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

| # | Claim | Evidence (command → verbatim output) | Baseline-attribution |
|---|-------|--------------------------------------|----------------------|
| R1 | The misclassification premise and AC-FLB-001's RED-now reason are real on this tree; the pinned tests pass asserting `raced` | `go test ./internal/cli/ -run 'TestFactoryLeaseRecordStallBounded\|TestFactoryLeaseMidClaimStallBounded' -count=1 -v` → `--- PASS: TestFactoryLeaseRecordStallBounded (7.26s)`, `--- PASS: TestFactoryLeaseMidClaimStallBounded (5.82s)`; stderr verbatim: `factory next: refused raced: t1 was not leased: the factory record stayed busy — the claim deadline (2.2s) passed while a record write was still running: the host is under load, not a lock race: context deadline exceeded; database is locked (5) (SQLITE_BUSY); retry when the host is quieter`; mid-claim log: `record=assigned holder= queue=picked` | This run, this tree @ 61b5af4b1 |
| R2 | Tier-M cap 16/16 respected; P1 consistent | `grep -cE '^- \*\*REQ-FLB-[0-9]+\*\*' spec.md` → 12; AC count → 10; frontmatter `priority: P1` (spec.md:9) = plan §A "P1" = card contract | This run |
| R3 | waitable() keying (D3) | `grep -n 'waitable' internal/cli/factory_card.go` → 1270 (func), 1272 (`case factoryRefuseRaced, factoryRefuseSerialSlot, factoryRefuseQuotaHold: return true`), 1768 (`if wait && refusal.waitable() && …`) | This run |
| R4 | serial-slot refusal is bool-keyed, read failures fail open (D8) | Read factory_card.go:330, 1368, 1866-1879 | This run |
| R5 | Convergence | `moai verify audit-plan --project-root <toplevel>` → config_status ok, codex gate required, cross_model_active true; `mcp__moai__audit_multi` → codex verdict fail (8 findings, receipt rcpt-9dcfbcfff0470950105681e3); digest + `--result-file` re-check → `convergence_check: {ok: true, unmet: []}` | This run |
| R6 | M0 is a sequencing precondition, correctly encoded | plan §A.3 + §F M0 ("Hold gate (precondition, no code)… Nothing else starts before M0 closes"), acceptance preamble ("verified only after the hold lifts"), DoD #1; baseline-first ordering per verification-claim-integrity §2.3 in AC-FLB-008 ("witnessed by the commit graph, not by a commit message assertion", acceptance.md:99) and plan §E | Document read |

**Gaps**: (1) the traceability and CN-4 awk verbs were refused by the worktree isolation guard; replaced by a line-by-line hand mapping (traceability) and plain-grep + hand-read of all 5 ordering candidates (CN-4) — the substitutions are named here per VCI §3.1. (2) audit_multi's build_lag notice: the installed moai binary (8249bff8c) predates this tree's HEAD; no moai CLI measurement is cited in this verdict, so no claim rests on the stale build. (3) AC-FLB-001's (b)/(c) class-specific assertions do not exist yet (M1 authors them), so the RED-now cell's failure could only be verified indirectly (R1); the stated reason — the misclassification — is measured, not vacuous or wrong-reason. (4) codex's transient-verification claims (its "임시 검증" runs) were not re-executed by me; each codex finding was adjudicated against my own reads of the cited code/artifact lines, and D1-D4/D8 rest on my measurements above.

**Residual-risk**: line-number anchors (109/151/298/443/464) drift on any rebase of this branch after #1859 lands — D10's name-level fix removes the exposure. The Q1 default (exit 4 for every class) is encoded in the requirements; if the operator later answers Q1 with the alternate (distinct exit code), REQ-FLB-002/003/004/007 and AC-FLB-004 need revision — a known, deliberate coupling recorded in decision-index Q1 and spec § Constraints.

## decision-index routing assessment (audit item 5)
Q1 (FOUNDER, product-level — refusal-surface default) and Q2 (EVIDENCE-NEEDED — latency budget) do NOT block run-phase entry: Q1's default (exit 4 for every class, token-only discrimination) is already encoded in the requirement layer (REQ-FLB-002/003/004/007) and measured by AC-FLB-004, so the implementable contract is fully specified under the published default; Q2 is by design resolved in-card — M4 produces the baseline and AC-FLB-009 is baseline-relative, with DoD #3 requiring both dispositioned by close. Q4 is DEFAULT-APPLIED (decision-index.md:56). The run-phase entry blocker is M0 (PR #1859 merged + leader resume signal), which is correctly a sequencing precondition, not a scope cut (R6).

## Pinned-test plan testability (audit item 7)
The 109/151/298-reclassify / 443/464-unchanged split is stated in plan §B and enforced by AC-FLB-010's swept-list + count guard + 443/464-presence clause, with AC-FLB-001's RED-now naming 151/298 as the pins that flip. Testable in intent; D10 records the name-vs-line-number gap as the one repair needed to make the 443/464 clause mechanically checkable.

## Recommendation (numbered fixes for manager-spec)
1. D1: rewrite AC-FLB-006's Then clause split by failure point (queued-restore arms vs assigned/picked-preserve arms), keeping the 155-157/302-307 citations on the arms they actually pin.
2. D2: scope REQ-FLB-005 to this SPEC's claim-write classification or add the queue-lock exception inline.
3. D3: add a waitable-disposition requirement + M2 clause + lease-success-after-contention AC.
4. D4: soften REQ-FLB-004's detail mandate to signal-naming.
5. D5: add the post-change regression ceiling to AC-FLB-009's second arm.
6. D6: make M4's recipe executable (binary revision, isolated root, fixture, command).
7. D7: extend AC-FLB-004 with text assertions for the three remaining preserved classes.
8. D8: re-anchor REQ-FLB-009/M3/AC-FLB-007 to a real, inducible site — or drop the requirement.
9. D9: move M1's class-table deliverable out of plan.md into manager-spec's plan-phase scope or progress.md §E.2.
10. D10: name the queue-lock pin test functions in AC-FLB-010.

Fit to enter run phase after M0: NOT in current form — this FAIL blocks (score 0.69 < 0.80, plus 3 blocking defects). After the D1-D3 repairs (and the SHOULD-FIX batch) and a delta re-audit of the listed reread_hunks, the SPEC's contract is otherwise sound — sequencing (M0), decision routing, traceability, and the M4-before-M5 baseline ordering are all correctly structured — and it would be fit to enter run phase once M0 opens (PR #1859 merged, leader go).
