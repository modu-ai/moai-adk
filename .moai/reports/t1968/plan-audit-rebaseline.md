auditor-model: glm-5.3

# SPEC Review Report: SPEC-FETCH-WEB-001 — plan-audit re-baseline (card t1968)

Round: re-baseline (2026-10-27) — a mechanical re-binding of the plan-audit verdict to the current artifact set for the retroactive factory stage chain (T7 kickoff evidence), executed as a full-scope re-read of the three plan artifacts, not a delta round. Prior trail: iter1 FAIL 0.80 → iter2 FAIL 0.82 → iter3 FAIL 0.87 → iter4 PASS 0.92 (receipt `rcpt-e08ce4d1c1c9e3843a82299f`), all in `.moai/reports/t1968/plan-audit.md`.

```
verdict: FAIL
overall_score: 0.86
blocking_count: 1
must_pass_failed: 0
plan_artifact_hash: 46521ef520a58e472eb35ea815c90df4c4676d81d4d2eeed5cac15b482390241
convergence_overall: fail
required_backend: codex fail
audited_sha: 62583ecf57e4b4f872ad9c36c023ff2996fbf3b0
scope: full
fix_scope: .moai/specs/SPEC-FETCH-WEB-001/acceptance.md#AC-FW-002 (first-block exception wording); .moai/specs/SPEC-FETCH-WEB-001/acceptance.md#AC-FW-003 (non-empty Given qualifier); .moai/specs/SPEC-FETCH-WEB-001/acceptance.md#AC-FW-006 (CLI/MCP output-parity method limb); .moai/specs/SPEC-FETCH-WEB-001/acceptance.md#G-R8-R10 (RED-basis limitation relabel)
defect_class: ac-wording (RB-2)
reread_hunks: .moai/specs/SPEC-FETCH-WEB-001/acceptance.md#AC-FW-002
```

Fresh cross-model receipt (minted this run): `rcpt-62a7d8b3f5d37f23d63a7e41` — `mcp__moai__audit_multi`, project_root = this worktree, target = `.moai/specs/SPEC-FETCH-WEB-001` (the plan-artifact set; the implementation is corroborating context only — the known `internal/jev/cache.go` priorModelEntry under-restriction is already a carried finding in the sync stream, origin receipt `rcpt-ef44873aaf70911dc9776577`, and is not re-litigated here). Result: codex **fail**, 4 findings (all P2); claude/glm gate-off in this tree's plan (`participant_count: 1`, `disagreement_flag: null`); digest written to `.moai/state/audit-plan-result.json` and `moai verify audit-plan --result-file …` returned `convergence_check: {ok: true, unmet: []}` — the required backend answered; its verdict is fail.

## Claim

1. The re-binding precondition holds: since the iter4 audited head `f57bbdd789` and since the stage-recorded evidence commit `62583ecf57`, the ONLY plan-artifact delta is the spec.md frontmatter lifecycle pair (`status: draft → completed`, `updated: 2026-10-26 → 2026-10-27`); plan.md and acceptance.md are byte-identical; progress.md changed but is not a plan-artifact hash subject. No requirement, plan, or AC content moved — the "stop and report" trigger for content drift did NOT fire.
2. All nine must-pass criteria pass on the current artifact set.
3. The full-scope re-adjudication (mine + the codex leg over the same subject) surfaces one blocking AC wording defect (RB-2) that the delta-scoped iter2-4 rounds never re-traced, plus three non-blocking quality defects. The re-baseline therefore earns **FAIL**, diverging from iter4's PASS: iter4 was delta-scoped to the N5 anchors; this round re-opened the whole criterion set and AC-FW-002's internal contradiction became visible.

## Evidence (commands run in this audit, this tree)

- Artifact-drift check: `git diff 62583ecf57..HEAD -- .moai/specs/SPEC-FETCH-WEB-001/spec.md` → exactly the two frontmatter lines (`-status: draft +status: completed`, `-updated: 2026-10-26 +updated: 2026-10-27`); `git diff --quiet 62583ecf57..HEAD -- …/acceptance.md …/plan.md` → unchanged. Read-time HEAD `ffe0f5135c`.
- Plan-artifact hash (runtime algorithm, read from source): `internal/runtime/audit_cache.go:97-105` (subject list, sorted) + `:135-160` (`%s:%d\x00` + exact bytes + NUL per file); computed over the current dir with subjects present acceptance.md / plan.md / spec.md → `46521ef520a58e472eb35ea815c90df4c4676d81d4d2eeed5cac15b482390241`.
- MP-6 / MP-7: `grep -c syscall` → 0/0/0 (plan/acceptance/spec); `grep -c "NEEDS CLARIFICATION" plan.md` → 0; research.md absent (Tier M — expected).
- MP-5: SPEC-JEV-CORE-001 / SPEC-JEV-OPTIN-MEASURE-001 / SPEC-JEV-CONSUMERS-001 all `status: completed`; SPEC-JEV-LEDGER-001 absent from this tree (in-flight t1762 — SHOULD note, commit-pinned citation `954adab25e` in plan §A.2, unchanged from iter1-4).
- MP-1 / MP-2 / MP-3: full reads — REQ-FW-001…016 (16 sequential unique), AC-FW-001…014 (14); GEARS forms valid on the requirement layer (REQ-FW-003/008 While; REQ-FW-006 compound; REQ-FW-012 When; REQ-FW-015 Where; REQ-FW-010/016 legacy negative equivalents, compat window through 2026-11-22 — today 2026-10-27, inside it); frontmatter 12/12 canonical fields, `status: completed` (valid enum), `version: "0.2.2"` quoted, `phase: "v3.3.0 target"` release label, no snake_case aliases.
- MP-8: RED ledger R1-R10 read, not re-executed (four elements present per row; document-level pin `4f0e58197d`; right-reason quality of R8-R10 is graded as RB-4, not as an MP-8 completeness failure, per the MP-8 boundary clause).
- Cross-model: as above — codex fail, 4 findings, receipt `rcpt-62a7d8b3f5d37f23d63a7e41`, convergence check ok/unmet-empty.

## Baseline-attribution

All measurements in this run, worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1968`, read-time HEAD `ffe0f5135cdcc95efb1f7d54f86d151e451ad974`. The bound commit `62583ecf57e4b4f872ad9c36c023ff2996fbf3b0` is the stage-recorded plan-audit evidence commit ("docs(t1968): plan-audit iter4 PASS 0.92 — kickoff gate met"); between it and read-time HEAD the plan artifacts differ only by the disclosed frontmatter lifecycle pair, so the audit subject is byte-stable across the binding. `moai verify audit-plan` ran on installed `moai` (server build `b1806498fe`); the hash algorithm was read from this tree's own `internal/runtime/audit_cache.go`, unchanged between that build's commit and this HEAD for the hashing path.

## Gaps

- The Group-4 traceability verb and the Group-6 CN-4 verb were again REFUSED by the worktree-session guard (trivial `awk 'BEGIN{print "awk-ok"}'` executed; both full verbs refused as "construct too complex to verify" — consistent with plan-auditor memory, observed iter2-4). Both adjudications were performed by hand over full artifact reads this run: traceability 16/16 REQ covered via the §F.1 table (every REQ-FW-001…016 mapped; every AC row maps to existing REQs; 0 uncovered, 0 orphan); CN-4 milestone order M1→M7 (`### M1`…`### M7` headings), ordering keywords in acceptance.md are content words or in-pipeline orderings ("first block" AC-FW-002, "first output line" AC-FW-004, "ambiguous after delivery" AC-FW-010, "preceding passthrough" in-request-lifecycle) binding nothing across milestones — no CONFLICT. Recorded as refused-tool degradation per verification-claim-integrity §3.1.
- MP-8 boundary: reading the RED cells confirms what they record, not that the commands measure the premise their author claims (RB-4 records where the framing oversells).
- Codex's RB-4 empirical demonstration (adding the search strings to a temp copy of pin `4f0e58197d` flipped R8-R10 exit 1→0) was not re-executed by this auditor; RB-4 is confirmed on textual analysis alone (a grep for a symbol string is flipped by any no-op mention of the string — no execution needed to establish that).

## Residual-risk

- Iter1-4 report headers described the plan-artifact hash as "sha256 over spec+plan+acceptance+progress concatenated" — that prose label does not match the runtime algorithm (progress.md is not a subject; the format carries name/length separators). This round's hash is computed per the runtime source, which is the binding definition for the evidence reader; the older prose labels in the trail are descriptive only and their values should not be compared against runtime-computed ones.
- Context correction for the dispatcher: the sync stream is NOT currently PASS-WITH-DEBT 0.93 — `.moai/reports/t1968/sync-audit.md` carries iter6 verdict FAIL (audited `b62aadf26e`), with the cache.go finding among its carried items. The implementation-side defect chain is owned there; nothing in this report substitutes for it.
- RB-2's contradiction is invisible in per-limbo sub-test staging (each selection rule stages its own fixture); it surfaces only when one fixture exercises both limbs — which is exactly what a full-scope reviewer did.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: 16 sequential unique REQ definitions / 14 AC entries (full-read enumeration).
- [PASS] MP-2 EARS/GEARS compliance (requirement layer only): all 16 forms valid; REQ-FW-010/016 legacy negative equivalents within the compat window (through 2026-11-22).
- [PASS] MP-3 YAML frontmatter: 12/12 canonical fields, correct types, `status: completed` valid enum value, no rejected aliases.
- [N/A] MP-4 language neutrality: single-language (Go) SPEC scope.
- [PASS] MP-5 D7 cross-SPEC reconciliation: 3 completed + 1 in-flight absent (SHOULD note in Gaps); no retired/superseded/archived reference → no REVIEW, no BLOCKING.
- [PASS] MP-6 D8 cross-platform discipline: `syscall` 0/0/0 (auto-pass per D8-4).
- [PASS] MP-7 clarification gate: plan.md clean; research.md absent (Tier M — stated reason).
- [PASS] MP-8 RED-now cell completeness: 10-row ledger complete (command / verbatim output / exit code / pin `4f0e58197d`); read, not re-executed.
- [PASS] MP-9 cross-artifact ordering: hand-run full re-check (verbs refused — Gaps); M1→M7 order; no CONFLICT; no ordering clause binds a criterion against the plan's milestone order.

## Per-finding adjudication (codex rcpt-62a7d8b3f5d37f23d63a7e41, 4 findings)

| # | Codex finding | This auditor's verdict |
|---|---|---|
| F1 | AC-FW-003 Given lacks a non-empty qualifier (0-block doc satisfies ≤2,000 but `empty` precedes passthrough per REQ-FW-003) | **CONFIRMED — minor/optional (RB-1).** Mirror image of iter3-N3 (there the REQ lacked the scope, graded MINOR); here REQ-FW-003 (spec.md:56) carries the disambiguation ("this guarantee applies only to non-empty documents") and AC-FW-004 stages the empty branch explicitly. One-word Given fix. |
| F2 | AC-FW-002 "kept only when a same-heading neighbor scores ≥ 0.6" contradicts "the first block is kept whenever at least one stage-1 candidate exists" | **CONFIRMED — BLOCKING (RB-2).** Hand-trace: first block is a stage-1 candidate (p=0.7), stage-2 standalone+context 0.2, no same-heading neighbor ≥ 0.6 → limb A (only-when) forbids the keep, limb B requires it. Reachable input; the criterion is not binary-testable on it. REQ-FW-006 (spec.md:62) carries the exception structure ("the selection stage shall preserve these guarantees: the first block of the document is always kept…") — the AC flattens the guarantees into a flat conjunction without the exception marker. Same defect class as iter2-N1 (BLOCKING then). Not visible to iter2-4: the limb was added by the N1 repair and delta scopes never re-traced the interaction. |
| F3 | AC-FW-006 passes an MCP mutant that calls the shared pipeline but mangles status/blocks/exit mapping | **CONFIRMED — should-fix/optional (RB-3).** The Then asserts presence + shared dispatch, not output parity; REQ-FW-009's "same output contract … identical semantics" limb is unexercised (mutant-writable per verification-completeness §2). Same class as iter1-D4 (SHOULD-FIX precedent). Fix: add a CLI-vs-MCP parity comparison on one shared fixture to the Method. |
| F4 | R8-R10 labeled "behavior probes" but are symbol-string greps (flip on a no-op mention) | **CONFIRMED — should-fix/optional (RB-4).** The ledger's framing (acceptance.md:68) oversells: "a mutant adding only a path constant leaves them red" is true but a mutant adding the symbol string itself (comment, unused constant) flips all three. MP-8 completeness unaffected (four elements present); right-reason quality is the D4 family again. Relabel with the limitation stated, as R4 already does. |

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.85 | 0.75-0.9 | RB-2's contradictory limbs + RB-1's unscoped Given form one interpretive cluster resolvable only by consulting REQ-FW-003/006; remainder crisp and numeric |
| Completeness | 0.90 | 0.75-0.95 | all sections/mechanisms pinned (as iter4); deduction: AC-FW-006 does not carry its REQ's output-parity limb (RB-3) |
| Testability | 0.75 | 0.75 band | RB-2 criterion not binary-testable on a reachable input; RB-3 mutant-writable; RB-4 shallow RED basis; rest binary-testable |
| Traceability | 0.95 | near 1.0 | 16/16 clean via §F.1 (hand-read); zero orphan/uncovered |

Aggregate: 0.8625 → **0.86** (floor convention per iter2/3). Tier M threshold 0.80 met on score; FAIL on 1 blocking finding + required-backend convergence fail (same shape as iter3).

## Defects Found

- **RB-2 [BLOCKING, class ac-wording]** acceptance.md:11 (AC-FW-002) — the "kept only when a same-heading neighbor scores ≥ 0.6" limb and the "first block is kept whenever at least one stage-1 candidate exists" limb contradict each other on the input {first block = stage-1 candidate, stage-2 < 0.8, no qualifying neighbor}. Severity: major. Required fix: one clause marking the selection rule as subordinate to the preserved guarantees — e.g. "a candidate below 0.8 is kept only when a same-heading neighbor scores ≥ 0.6, except where a preserved guarantee of REQ-FW-006 applies (first block, <300-token backfill, header-row accompaniment)".
- **RB-1 [MINOR, class optional]** acceptance.md:13 (AC-FW-003) — Given lacks the non-empty qualifier its requirement carries (REQ-FW-003's empty-first scope). Required fix: "Given a **non-empty** document whose split token count is ≤ 2,000…".
- **RB-3 [SHOULD-FIX, class optional]** acceptance.md:21 (AC-FW-006) — criterion shallower than REQ-FW-009: no output-parity assertion. Required fix: extend Method with a CLI-vs-MCP parity check on one shared fixture (status, heading paths, verbatim blocks, exit-status mapping).
- **RB-4 [SHOULD-FIX, class optional]** acceptance.md:68 (§G preamble, R8-R10) — "behavior probes" framing oversells symbol-greps. Required fix: relabel with the stated limitation (they discriminate on the symbol string, which a no-op mention defeats), matching R4's in-row limitation pattern; the deeper behavior-RED records are run/sync-phase items now that the implementation has landed.

## Regression Check (vs iter4 PASS 0.92)

- N5 (iter3's blocking finding, resolved at iter4): still resolved — B5's applicability condition (plan.md:34) and AC-FW-009's mirrored limb (acceptance.md:29) unchanged; re-traced this run, holds.
- iter1-3 findings D1-D9, N1-N4: all still resolved (artifact bodies unchanged since `f57bbdd789` except the frontmatter pair — verified by git diff).
- The FAIL is a full-scope re-adjudication outcome (RB-2 newly surfaced), not an artifact regression: nothing that passed at iter4 got worse; what changed is the scope of the re-read.

## Recommendation

1. One wording revision in acceptance.md closes the blocking set: RB-2's exception clause (plus RB-1's one-word Given fix; RB-3/RB-4 may ride along). REQ/plan bodies need no change; REQ-FW-006 already carries the exception structure the AC must mirror.
2. Re-audit scope after the fix: the acceptance.md anchors above; `reread_hunks` names AC-FW-002 — under the ceiling-policy hunk-limited exception (single ac-wording blocking finding) a `scope: reread` verdict may apply, at the leader's discretion; otherwise one delta round over the four anchors suffices.
3. The re-binding mechanics themselves are clean: the only post-iter4 plan-artifact movement is the legitimate lifecycle transition, the hash is computed per the runtime algorithm, and the receipt above is fresh from this session. The FAIL verdict is substantive (a real criterion contradiction), not a mechanical artifact — do not paper over it; fix the clause and re-bind.

---

*Verdict authority: plan-auditor of record, factory depth-1 leaf, card t1968. Receipt `rcpt-62a7d8b3f5d37f23d63a7e41` was minted during this run on this tree and is cited by the machine lines above.*
