# Plan-Phase Audit — SPEC-CI-VERDICT-PRODUCER-001 (iteration 1)

Auditor: plan-auditor (independent, adversarial). Tree: worktree `.claude/worktrees/t1268`, branch `WT-ci-verdict-producer`, HEAD `bf3d5144f` (develop). Tier M, PASS threshold 0.80, iteration limit 2. Reasoning context from the SPEC author: none passed; M1 Context Isolation holds trivially.

## Verdict

**PASS-WITH-DEBT** — score **0.86** (harmonic mean). Tier M threshold 0.80 met; no must-pass failure. Two blocking-class defects (D1, D5) must close before Kickoff; four optional findings left to the orchestrator's discretion.

## Claim

1. AC-AE-012(c) re-judgeability AS WRITTEN is faithfully served: the SPEC's trip predicate (REQ-CV-007, spec.md:51) matches the original limb-(c) text — "a local verification pass and a recorded CI failure on the same head" (.moai/specs/SPEC-AUTONOMY-ESCALATION-001/acceptance.md:94) — exactly: same head on both sides (record `head_sha` == checkpoint head, local pass at that same head per REQ-CV-006).
2. Limb (e) survives the change: REQ-CV-008 (spec.md:52) preserves the no-record → listed-not-observed behavior "byte-for-byte" (plan.md §D constraint), and the existing test assertions (operational_m5_test.go:190-192, :204) remain satisfiable because every existing fixture has no CI record.
3. All 7 must-pass criteria pass (MP-4 N/A: single-language Go project).
4. Two blocking-class defects exist: an internal inconsistency on success-conclusion labeling (§G Q4 vs REQ-CV-008/AC-CV-006(b)) with the `neutral` schema value never given a detector-side disposition; and a go.mod guard measured against the literal `develop` tip, which produces a false violation after any develop absorption.
5. The mission-expected `depends_on`-omission rationale is NOT stated in any artifact — but the omission itself is verified correct (the strict fulfillment rule would hard-block on the dependency's `status: implemented`).

## Evidence

Verbatim commands and outputs, this run, this tree (HEAD `bf3d5144f`):

**Must-pass results**

- **MP-1 REQ numbering — PASS.** `grep -n 'REQ-CV-00' spec.md` → `42:REQ-CV-001, 43:REQ-CV-002, 44:REQ-CV-003, 45:REQ-CV-004, 46:REQ-CV-005, 50:REQ-CV-006, 51:REQ-CV-007, 52:REQ-CV-008, 53:REQ-CV-009`. Sequential 001–009, no gaps, no duplicates, consistent zero-padding.
- **MP-2 GEARS format — PASS (judged against the requirement layer, spec.md §C).** All nine REQs match GEARS patterns: Event-driven ("When …, the producer shall …" :42-44; "When a commit checkpoint runs …, the detector shall …" :51-53), Ubiquitous ("The producer shall write records whose schema is …" :45; "The local verification pass evidence source … shall be …" :50), Unwanted ("The producer shall not invoke the escalation detector …" :46). Note: the labels "(Event-detected)" on :44 and :52 are non-canonical pattern names, but the sentence structure is Event-driven — cosmetic, not a format defect. ACs (Given-When-Then) are the verification layer and are correctly NOT graded here.
- **MP-3 frontmatter — PASS.** Field-by-field against spec-frontmatter-schema.md: id:2, title:3 (quoted), version "0.1.0":4 (quoted semver), status draft:5 (valid enum), created/updated 2026-09-26:6-7, author:8, priority P1:9 (valid), phase "v3.2.0 target":10 (whole-value comparison — not a prohibited stage name), module:11, lifecycle spec-anchored:12, tags comma-string:13; optional tier M:14, related_specs:15. No rejected snake_case aliases (grep `depends_on` across all five artifacts → exit 1; no `created_at`/`updated_at`/`labels`/`spec_id` anywhere). Sibling artifacts carry no `status:` field (statelessness rule holds).
- **MP-4 language neutrality — N/A** (single-language Go project; auto-pass).
- **MP-5 D7 — PASS.** `grep -hoE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' {spec,plan,acceptance,research,progress}.md | sort | uniq -c` → `14 SPEC-AUTONOMY-ESCALATION-001`, `10 SPEC-CI-VERDICT-PRODUCER-001` (self). `grep -n '^status:' .moai/specs/SPEC-AUTOMY-ESCALATION-001/spec.md` → `5:status: implemented`. Not retired/superseded/archived → no BLOCKING; the referenced SPEC exists → no SHOULD.
- **MP-6 D8 — PASS.** `grep -rn 'syscall'` over the SPEC dir → exactly three hits, all negative mentions: plan.md:22 "no syscall surface is introduced; verify with GOOS=windows…", acceptance.md:56 "no new syscall surface; Windows cross-build green", research.md:43 "no new syscall surface; temp+rename already proven on Windows". No syscall usage is introduced, so no //go:build obligation arises; the Windows cross-build is mandated in spec.md:57, plan §C, acceptance §E. No BLOCKING.
- **MP-7 clarification gate — PASS.** `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-CI-VERDICT-PRODUCER-001/` → no output, exit 1.

**Cross-reference anchors verified on disk**

- `notObservedCIVerdict`: internal/escalation/operational.go:25-28 (const) and :325 (unconditional `notObs := []string{notObservedCIVerdict}`) — exactly the two sites plan §C targets.
- `convergenceFile` parse precedent operational.go:311-317; `classContradictoryEvidence` :319-354; `freshEvidence` :363-371; `contradiction` :394-410 — all as research.md P2/P3 claims state.
- `TestContradictoryEvidenceTrips` operational_m5_test.go:162; limb (c) absent with the missing-producer comment at :156-161; existing cases (a),(b),(d) + subtest (e) at :168-207. The plan's RED-now framing (§C: the M2-written limb-(c) test is red because `classContradictoryEvidence()` never reads `.moai/state/ci-verdicts/`) is a correct right-reason RED per verification-completeness §2.
- Exit-code-0 predicate derivable: `CheckEntry.ExitCode` internal/verify/schema.go:37, `Snapshot.Checks` schema.go:55-59, `SnapshotDir` store.go:15, atomic temp+rename `Save` store.go:61-94. A head's local pass is derivable from a snapshot whose key's head portion matches and which holds ≥1 entry with `ExitCode == 0`.
- gh fail-open precedent: internal/cli/doctor.go:444-452 — `exec.LookPath("gh")` → warn-and-continue. REQ-CV-003 (message + exit 0 + no file) matches the house discipline; `os/exec` + `encoding/json` are stdlib, so the no-new-dependency claim is feasible.
- `newStateCmd()` house style: internal/cli/state.go:24 — the plan M3 claim is anchored.
- SPEC-AUTONOMY-ESCALATION-001 anchors cited by this SPEC all exist: REQ-AE-010 (spec.md:201), REQ-AE-020 (:283, State-driven — the freshEvidence gate), REQ-AE-022 (:291), AC-AE-023 (acceptance.md:136).
- Debt provenance: primary checkout `/Users/goos/MoAI/moai-adk-go/.moai/reports/t1235/verdict.md` — "Debt — no CI verdict producer: AC-AE-012(c) is left not-observed … follow-up card **t1268** (the lead issued it)" (:56) and sync-audit "AC-AE-012(c) is UNVERIFIED … follow-up card t1268" (:68). Corroborated by this tree's code (constant + seed + test comment all present).
- Mission's FAILURE-snapshot question: covered by composition — a snapshot with no exit-0 entry fails REQ-CV-006's local-pass definition, so a CI failure + local FAILURE falls to REQ-CV-008's "a missing local pass at the checkpoint head … named in `not_observed`". No trip. (Untested as an explicit AC-CV-006 case — see Residual-risk.)
- Mission's primary-vs-worktree question: explicitly answered — spec.md:89 (§G Q2: "a detector in any tree consumes a record only when that record's pinned head equals the detector's own checkpoint head") and research.md:51 (P7). Records for heads not checked out anywhere remain consumable by head match; a stale worktree fails the match and lists it — specified, not accidental.
- Mission's M1-deferral question: bounded HOW — plan.md:59 and research.md:47 pin "one location, one import direction: cli → escalation, never the reverse" while the WHAT (five-field schema, atomicity, idempotence) is fully specified in REQ-CV-004. Not an unresolved WHAT.

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.75 | minor ambiguity in two requirements | Trip predicate and schema unambiguous (spec.md:42-53, :87-93). Defects: Q4 (:93) groups "CI success/neutral … listed under `not_observed`" while REQ-CV-008 (:52) and AC-CV-006(b) (acceptance.md:30) say success is NOT listed; `neutral` (a REQ-CV-004 enum value) has no detector-side labeling disposition; REQ-CV-004's "byte-equivalent rewrite" over-generalizes fetch-mode re-observation (:45). |
| Completeness | 1.00 | all required sections present | §A History:20-24, §B Problem:26-36, §C Requirements:38-53, §D Success Criteria:55-57, §E Out of Scope with four `### Out of Scope — <topic>` H3s each carrying specific bullets (:59-75), §F:77-83, §G:85-93, §H:95-99. Frontmatter complete per MP-3. |
| Testability | 1.00 | every AC binary-testable | AC-CV-001..008 (acceptance.md:17-38) all Given-When-Then, offline (injected runner, `t.TempDir()`, fabricated fixtures — :13), no weasel words; AC-CV-008's predicates are mechanical (`go test` ok, `git diff` empty, GOOS=windows exit 0). The limb-(c) extension is concretely specified (AC-CV-005 :28, fixture shape + test name). |
| Traceability | 0.75 | one AC not REQ-anchored | Every REQ-CV-001..009 has ≥1 AC and every AC-CV-001..007 maps to valid REQs (acceptance.md §D table :44-51). Exception: AC-CV-008 maps to "§D Success Criteria" (:37, :51), not a REQ — indirect anchoring. |

Harmonic mean: 4 / (1/0.75 + 1/1.00 + 1/1.00 + 1/0.75) = **0.857 → 0.86** ≥ 0.80 (Tier M).

## Defects Found

- **D1** — spec.md:93 vs spec.md:52 vs acceptance.md:30 — Internal inconsistency + unpinned edge: §G Q4 says "CI success/neutral … → no trip with the incomplete limb(s) listed under `not_observed`" while REQ-CV-008 and AC-CV-006(b) say a `success` conclusion is NOT listed (observation completed and agreed); and `neutral` — a conclusion value REQ-CV-004 itself introduces — is never given a labeling disposition in the requirement or AC layer (AC-CV-006 tests neither neutral-at-head nor success-without-local-pass). An implementer trusting Q4 over REQ-CV-008 implements the wrong labeling. — Severity: major — Class: **blocking** — Required fix: amend Q4 to separate the cases ("CI neutral → no trip, listed under `not_observed`; CI success at the checkpoint head → completes the observation, not listed"), and add one REQ-CV-008 clause pinning the neutral-at-head disposition (listed, per the "could not complete" general clause) plus an AC-CV-006 sub-case (or an explicit statement that neutral is listed).
- **D2** — spec.md:45/§E — Record lifetime/pruning unstated: `.moai/state/ci-verdicts/` accumulates one file per judged head, the detector globs `*.json` at every commit checkpoint, and no artifact states unbounded growth is acceptable or defines retention. — Severity: minor — Class: optional — Required fix: one sentence in §E or REQ-CV-004 declaring unbounded growth accepted runtime-state debt (the verify-snapshot store precedent) or a bounded retention rule.
- **D3** — acceptance.md:37, :51 — AC-CV-008 is anchored to "§D Success Criteria" rather than any REQ-XXX; it is the only AC without a REQ mapping (Traceability 0.75). — Severity: minor — Class: optional — Required fix: additionally map AC-CV-008 to the REQs its test exercises (REQ-CV-006/007 via the limb-(c) case) and carry the no-new-dependency / Windows-build constraints as explicit REQs, or record the §D anchoring as deliberate.
- **D4** — all five artifacts — The `depends_on`-omission rationale is nowhere stated, though the omission is correct: the strict fulfillment rule (spec-workflow.md § Depends_on Pre-flight Check — fulfillment = `status: completed` only) would hard-block run entry because SPEC-AUTONOMY-ESCALATION-001 has `status: implemented` (verified: its spec.md:5). — Severity: minor — Class: optional — Required fix: one line in spec.md §F or §H stating why the dependency is referenced via `related_specs` and not `depends_on`.
- **D5** — plan.md:37, plan.md:53, acceptance.md:37 — The go.mod guard diffs against the literal `develop` tip (`git diff (—name-only) develop..HEAD -- go.mod`). Per the on-record measurement discipline (gitflow-lane-protocol §8; the t543 measurement: literal base yielded 51 phantom files vs merge-base's 0), any sibling card's go.mod change absorbed via a develop absorption lands in this range and reads as this card's violation — a stated criterion mis-measurable by construction. — Severity: minor — Class: **blocking** — Required fix: use the merge-base form at all three sites: `git diff --name-only "$(git merge-base develop HEAD)"..HEAD -- go.mod`.
- **D6** — spec.md:45 — REQ-CV-004's "re-recording the same head with the same conclusion is an idempotent byte-equivalent rewrite" over-generalizes: a fetch-mode re-observation writes a fresh `observed_at`, so same head + same conclusion yields different bytes; only the parenthetical "(last-writer-wins on differing content)" resolves it. AC-CV-001 tests only the offline path where `observed_at` comes from the input file. — Severity: minor — Class: optional — Required fix: qualify as "byte-equivalent when the recorded fields (including `observed_at`) are identical; last-writer-wins otherwise".
- **D7** — acceptance.md (whole file) — No per-AC RED-now cells, unlike the sibling SPEC-AUTONOMY-ESCALATION-001 acceptance.md format; the RED obligation lives only in plan.md §C and §E item 6. — Severity: minor — Class: optional — Required fix: add a RED-now line per new-test AC (all are red-now by construction: the producer verb and the limb-(c) test do not exist), or record plan-level coverage as the deliberate carrier.

## Regression Check

Iteration 1 — no prior defects.

## Baseline-attribution

All readings and greps executed in this run against worktree `.claude/worktrees/t1268` @ `bf3d5144f` (branch `WT-ci-verdict-producer`; SPEC dir untracked: `?? .moai/specs/SPEC-CI-VERDICT-PRODUCER-001/`). Line citations are to this tree's files. The debt-provenance file (`/Users/goos/MoAI/moai-adk-go/.moai/reports/t1235/verdict.md`) was read read-only from the primary checkout and used as provenance only; its AC-AE-012(c)-UNVERIFIED claim was cross-checked against this tree's code, which corroborates it.

## Gaps

- No test execution, lint, or Windows cross-build was run: the code this SPEC plans does not exist yet; those commands belong to the run-phase pre-flight (plan §C) and §E evidence.
- internal/cli/verify.go was not read in full; the plan M3 "house style" claim was verified via internal/cli/state.go:24 (`newStateCmd`) instead.
- How the extended limb-(c) test fabricates the local-pass fixture (a verify snapshot with an exit-0 CheckEntry at H) is unpinned — `verify.RecordCheck` or a hand-written JSON fixture both work; the escalation test harness helpers observed in m5_test.go (`armedWith`, `isolateStore`, `postBash`) do not yet include a snapshot helper. Bounded HOW; a small unplanned M2 fixture cost.
- No cross-model audit (`audit_multi`) was run — this iteration is the assigned single-auditor plan-phase review.
- The FAILURE-local-snapshot case is verified covered only by composition (REQ-CV-006's "≥1 exit-code-0 entry" + REQ-CV-008's missing-local-pass labeling); no AC exercises it explicitly.

## Residual-risk

- D1's resolution direction changes AC-CV-006's fixture table: if the team instead rules `neutral` completes the observation (not listed), the REQ and the AC row must move together — either way one clause must be pinned before M2 test authoring, or the test and requirement will drift.
- t1235's carried debt N6 (`ConsumedEvidence` unbounded) gains new writers from this card: every CI-record trip appends a content-hash key, and AC-CV-007's flip-and-back churn accumulates keys. Not absorbed by this SPEC (per §E), but this card compounds it.
- The limb's real-world trip frequency depends on lead discipline (the producer runs post-push in the lead's tree); a lane checkpoint at an older head lists the head mismatch under `not_observed` — specified behavior, but the criterion's re-judgement in production (as opposed to the test) rests on that operational path.

## Recommendation

Close D1 and D5 (both one-to-three-clause edits: Q4 wording + one REQ-CV-008 neutral clause + one AC-CV-006 sub-case; merge-base form at plan.md:37, plan.md:53, acceptance.md:37) before Implementation Kickoff Approval. D2/D3/D4/D6/D7 are at the orchestrator's discretion — D3 and D6 are the two cheapest score-recoverers if a re-audit is desired. Per the Retry Loop Contract, a confirming re-audit is scoped to this defect delta.
