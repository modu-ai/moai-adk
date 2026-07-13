# research.md — SPEC-WORKFLOW-CACHE-OPT-001

> Plan-phase findings with live file anchors (verified 2026-07-13 against the working tree; line numbers drift — content tokens are the durable anchors). Source: 3-lens parallel bottleneck analysis (55 findings, 12 HIGH) + this plan session's direct reads.

## §A — Duplicate-Verification Inventory (axis 1 evidence)

Who independently executes the same diagnostic classes per SPEC lifecycle:

| # | Layer | Surface (anchor) | What it runs |
|---|-------|------------------|--------------|
| 1 | `/moai gate` | `workflows/gate.md` Phase 2 | lint + format + type + full test, parallel |
| 2 | run Phase 2.5 | `workflows/run/task-decomposition.md` "Phase 2.5: Quality Validation" | TRUST 5 validation incl. tests |
| 3 | run Phase 2.75 | same file, "Phase 2.75: Pre-Review Quality Gate" | gate_report per check category |
| 4 | run Phase 2.8a/2.8b | same file | sync-auditor active evaluation + TRUST 5 static |
| 5 | sync Phase 0 | `workflows/sync/quality-gates-context.md` ("Runs full test suite", "Verify all tests pass") | full suite, pre-sync |
| 6 | sync Phase 0.5/0.7 | `workflows/sync/quality-gates-quality.md` | quality re-check + coverage measurement |
| 7 | loop Step 3 | `workflows/loop.md` Step 3 | LSP + AST-grep + tests + coverage, per iteration |
| 8 | loop Step 1.5 | `workflows/loop.md` Step 1.5 | independent re-run of the diagnostic gate |
| 9 | stop-goal evaluator | `internal/goal/evaluate.go` (`CmdRunner`, Tier-1 `ConditionMechanical`) | re-executes condition shell commands each turn-end |
| 10 | orchestrator verification batch | `agent-common-protocol.md` § Parallel Execution (7-item canonical batch) | tests + coverage + lint + smoke |
| 11 | sync-phase-quality-gate.sh | Stop hook (OUT OF SCOPE this SPEC) | lint + test + coverage delta |

Layers 1-10 are in-scope consumers/producers; layer 11 is deliberately excluded (shell-side key computation risk — see spec.md Out of Scope). Root cause confirmed: no layer can attributably trust a sibling's observation, so each re-observes. VCI §2 requires command+output attribution — the snapshot contract satisfies it by making the recorded evidence itself citable under a freshness rule.

### Existing infrastructure (EXTEND, not rebuild)

- `.moai/state/verify/<session>/` — ALREADY EXISTS, gitignored, ~29 per-session evidence dirs in active use (e.g., `61ed15d7/{1-build,2-vet,3-test,4-lint,5-cover}.log`). The snapshot joins this namespace; the evidence-persistence obligation (agent-common-protocol) is the doctrinal precedent.
- loop-verdict JSON (`loop.md` § Remaining-Issue Persistence): `conditions {zero_errors, error_count, tests_pass, coverage_threshold, coverage_actual, zero_warnings}` — the seed shape (REQ-SNAP-001 keeps read-compat).
- `internal/goal/state.go` — atomic per-session JSON persistence pattern (temp+rename) to reuse for the snapshot store.
- `internal/goal/evaluate.go` — `CmdRunner` abstraction makes the snapshot-aware path unit-testable with a fake runner (call-count assertions).
- `.moai/cache/loop-snapshots/` — explicitly best-effort with "no mechanical writer guarantee" (loop.md § Snapshot Management); REQ-SNAP-008 provides the mechanical writer in the shared schema without touching the resume-snapshot mechanism.

## §B — Gate-Stacking Inventory (axis 2 evidence)

Default pipeline (`workflows/moai.md` Execution Summary) blocking rounds: Step 8 (post-exploration confirmation) → Step 11 annotation cycle (1-6) → Step 11.3 Kickoff Approval → Step 11.5 Execution Mode Selection Gate → sync gate-sync-1 → gate-sync-2 → Phase 4 next-steps question = 7+ rounds minimum.

- **Kickoff (11.3) + Mode Gate (11.5) adjacency**: two consecutive blocking rounds at the same boundary; AskUserQuestion supports ≤4 questions per call → mergeable into one round without touching gate semantics. The Kickoff mandatory/score-independent doctrine (orchestration-mode-selection.md header, run.md § Run-phase Autonomy #1, CLAUDE.local.md §19.1) is PRESERVE-class.
- **project Stage B**: `project/codebase-analysis.md` — "Present each remaining (un-inferred or ambiguous) axis as a separate AskUserQuestion" → up to 4 calls for 4 axes. Same file already batches elsewhere (`questions_per_round: 3`), so single-call multi-question is convention-consistent.
- **harness-build-entry.md**: final-round proposal (Phase 1.5 closing question) then Phase 3 approval gate — two rounds on an already-100%-clarity profile.
- **feedback.md**: type (L52) → title (L64) → description (L66) = 3 sequential rounds, all free-text-compatible via "Other".
- **Completion-report close**: askuser-protocol § Completion-Report Next-Step Discipline already permits "close with NO question"; moai.md/sync Phase 4 currently manufacture a next-step round even on full-pipeline success.
- **Cache coupling (Phase 1 motivation)**: each blocking wait >5 min expires the prompt cache over the full prefix (cache-aware-execution.md directive 1: batch unavoidable late gates into consecutive — ideally single — rounds).

## §C — Spawn-Count Inventory (axis 3 evidence)

- **codemaps.md**: Phase 1 Explore + Phase 2 manager-docs (analysis) + Phase 3 manager-docs (generation) = 3 spawns for what `go list -deps -json` + `go doc` (deterministic baseline) + one read-only exploration can feed orchestrator-direct generation.
- **clean.md** Agent Chain Summary: Phase 1 refactoring specialist + Phase 2 refactoring specialist + Phase 4 refactoring specialist + Phase 5 manager-develop = 4 spawns (+ optional Phase 5.5). Phases 1+2 share the identical whitelist/role → 1 combined spawn; Phases 4+5 (remove + verify) → 1 combined spawn.
- **fix.md/loop.md**: "[HARD] ... ALL fix tasks MUST be delegated ... NEVER execute fixes directly" applies even to Level 1 (import sort/whitespace) — a formatter command run (gofmt/ruff/prettier) needs no agent. Constraint: fix.md Phase 3 @MX:WARN pins the Level→executor mapping as a static lookup (guarded by `TestAgentlessUtilityNoLLMControlFlow`); the relaxation keeps a static mapping, changing only the Level-1 executor.
- **mx.md** Pass 3: batch edit is one-Edit-per-file; <5-item runs don't amortize a spawn.

## §D — Audit-Repetition Inventory (axis 4 evidence)

- plan-audit iterations ≤3 (full re-audit each) + annotation cycle ≤6 (moai.md Phase 1.5) + run Phase 0.5 re-audit ("모든 harness level에서 SKIP 불가", run.md Quick Reference).
- Skip-eligible ≥0.90 today applies only to Phase 0.5 verdict re-execution (CLAUDE.local.md §19.1); Tier S SPECs (≤300 LOC, <5 files) pay the same full-re-audit loop as Tier L.
- project doc-generation Phase 3.1: FAIL → regenerate → re-audit, up to iteration 3 before AskUserQuestion escalation (`doc-generation.md` "If FAIL and iteration = 3: Escalate").
- Defect-list precedent: fix.md Phase 4 already demonstrates the claim/evidence + enumerated-set (Resolved/Persisting/Regression) delta pattern this axis generalizes to auditor verdicts.

## §E — Bookkeeping Fixed-Cost Inventory (axis 5 evidence)

- loop.md Steps 5/6/7: TaskCreate per issue + TaskUpdate in_progress per fix + TaskUpdate completed per fix = 3 calls/issue, serialized.
- loop.md Step 7.5 + § MX Tag Report: MX_TAG_REPORT "after each iteration".
- review.md Phase 2: "review from all 4 perspectives sequentially" (single sync-auditor pass); the `sync-audit-4dim` dynamic workflow (parallel read-only judges + in-script synthesis) already exists as an execution-vehicle precedent for Mode-4 parallelization; binding verdict stays sync-auditor.
- review.md secrets scan: `git log -p --all -G '...'` — full history every review; no checkpoint exists. `.moai/state/` is the established checkpoint namespace.

## §F — Contradiction Ledger (cross-lens; must be resolved consciously)

1. **Step 1.5 independence vs snapshot reuse**: loop.md Step 1.5 exists to make success-exit evidence non-self-referential. Naive snapshot injection there would dissolve the loop's only independent check. Resolution: REQ-SNAP-009 carve-out (Step 1.5 never consumes same-run snapshots; may produce for downstream).
2. **run.md "Phase 0.5 SKIP 불가" vs Tier S inversion**: resolved by precision — the audit still ALWAYS runs once for every tier; only the iterative re-execution loop is tiered (REQ-AUDIT-001 + plan.md Custom-3). The sentence must be amended, not deleted.
3. **[HARD] delegation mandate vs Level-1 direct execution**: the mandate's rationale (specialization, quality gates) doesn't apply to deterministic formatter runs; the static-dispatch guard test constrains the edit shape (plan.md Custom-2).
4. **VCI "no unobserved claim" vs reuse-without-re-running**: NOT a contradiction under §2 attribution — the snapshot IS the observed evidence (command + verbatim result), and the freshness key makes the attribution valid for the current tree. Stale reuse would be the violation; REQ-SNAP-003/011 close it.
5. **Advisory-check discipline vs stop-goal key computation**: Stop-hook path must stay constant-cost/time-boxed; fallback is re-execution (plan.md Custom-1). The optimization may be skipped; correctness never is.

## §G — Template-Mirror Verification (live measurement, 2026-07-13)

All 16 target doc files verified **MIRROR-BYTE-EQ** against `internal/template/templates/.claude/`: gate.md, moai.md, loop.md, fix.md, clean.md, codemaps.md, review.md, mx.md, feedback.md, harness-build-entry.md, project/codebase-analysis.md, project/doc-generation.md, run/task-decomposition.md, sync/quality-gates-context.md, SKILL.md, rules/moai/workflow/cache-aware-execution.md. Every doc edit in M1-M5 is therefore a local+template pair with `make build` re-verification (classification is time-varying — re-measure at run-phase entry, per the template-subset lesson).

## §H — Open Questions (mirrored in plan.md §D)

1. [NEEDS CLARIFICATION: snapshot key — untracked-file participation] — see plan.md §D #1.
2. [NEEDS CLARIFICATION: per-layer freshness TTL] — see plan.md §D #2.
3. [NEEDS CLARIFICATION: stop-goal command-matching granularity] — see plan.md §D #3.
