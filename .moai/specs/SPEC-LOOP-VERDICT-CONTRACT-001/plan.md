---
id: SPEC-LOOP-VERDICT-CONTRACT-001
title: "Mechanical Loop Termination Predicate and Ceiling-Exit Verdict Contract — Implementation Plan"
version: "0.1.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai/workflows"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "loop, ralph, termination-predicate, verdict-contract, ceiling-exit, max-iterations-precedence, workflow-reflex, plan"
---

# SPEC-LOOP-VERDICT-CONTRACT-001 — Plan

> plan.md is the derived execution plan. WHAT/WHY SSOT is spec.md. This document carries the HOW skeleton; exact wording of skill-doc rewrites is run-phase discretion within the REQ boundaries.

## §A Context

### §A.1 Problem summary

`/moai loop` success-exit string-matches a self-emitted prose sentinel (builder performs AND declares verification); ceiling exits have no verdict contract (loop.md Step 9 = "Display remaining issues and options"; moai.md termination cause 2 lacks the protocol causes 3/4 have); iteration ceilings are fragmented across four surfaces (100 / 100 / 10 / 10 / memory-safe 50) with no precedence rule; remaining issues evaporate at exit; and the workflow.yaml agentic_loop comment stale-claims "no Go-side loader field yet" against the implemented AgenticLoopConfig.

### §A.2 Evidence baselines (measured 2026-07-09 by this agent via Bash/Read, vci §2 attribution)

```
loop.md § Supported Flags        → "--max N (alias --max-iterations): Maximum iteration count (default 100)"
loop.md Step 1                   → sentence string-match exit: "All loop completion conditions satisfied; exiting loop."
loop.md Step 4                   → emits the sentence when zero errors AND tests pass AND coverage met
loop.md Step 2 area              → "If memory-safe limit reached (50 iterations): Exit with checkpoint"
loop.md Step 9                   → "If max iterations reached: Display remaining issues and options"
ralph.yaml:20-29                 → loop.completion {coverage_threshold: 85, tests_pass: true, zero_errors: true, zero_warnings: false}; max_iterations: 10
workflow.yaml:2-7                → stale comment "no Go-side loader field yet"; agentic_loop.max_iterations: 10
workflow.yaml:15-18              → loop_prevention.max_iterations: 100
internal/config/types.go (~326)  → AgenticLoop AgenticLoopConfig `yaml:"agentic_loop"` (SPEC-V3R6-AGENTIC-LOOP-CONFIG-001)
internal/config/defaults.go (~441-444) → AgenticLoop: AgenticLoopConfig{MaxIterations: DefaultAgenticLoopMaxIterations}
moai.md § Agentic Completion Loop → Termination four causes; cause 2 = agentic_loop ceiling (default 10) with NO protocol; causes 3/4 have protocols
goal-directive.md                → "/goal ... a small fast model (Haiku by default) checks whether the condition holds"
harness.yaml levels.minimal      → evaluator: false (evaluator presence is an explicit switch)
fix.md Phase 4                   → "Re-run affected diagnostics on modified files" (L5 provenance — OUT of scope)
```

Template mirrors verified present (2026-07-09): loop.md, moai.md, run.md, workflow.yaml, ralph.yaml, harness.yaml under `internal/template/templates/`. Template-First applies to all edits.

### §A.3 Approach — three milestones, doctrine/skill-doc primary

- **M1 — exit predicate + independent final pass (loop.md rewrite)**: Step 1 rewritten from sentence-detection to predicate re-evaluation over the previous iteration's PARSED diagnostics; Step 4 keeps the mechanical condition check and demotes the sentence to display-only; a new pre-success step inserts the independent final pass (fresh `/moai gate` re-run or read-only verifier spawn) with a divergence-continues rule; § Completion Conditions rewritten to name the predicate + independent confirmation, not the sentence.
- **M2 — ceiling verdict contract + persistence schema**: loop.md Step 9 ceiling branch rewritten to emit the vci §3 5-section report and write `.moai/state/loop-verdict-<id>.json`; moai.md Termination cause 2 gains the same protocol reference; lesson-capture proposal step added for unsuccessful exits; verdict-file JSON schema documented inline (doctrine-defined; orchestrator-written).
- **M3 — precedence unification + comment fixes + template sync**: the CLI-flag > ralph.yaml > workflow.yaml loop_prevention precedence rule stated identically in loop.md / ralph.yaml comments / workflow.yaml comments; loop.md flag-default prose reconciled (default derives from ralph.yaml value 10 when flag absent — see D1); memory-safe 50 documented as orthogonal memory-pressure checkpoint; agentic_loop stale loader comment fixed; `make build` template sync.

### §A.4 Tier evidence (M)

- Files affected: ~10 (loop.md, moai.md, ralph.yaml, workflow.yaml — each × template mirror, + possible small run.md cross-ref) — within Tier M's 5-15 band.
- LOC: doc-heavy, 300-700 lines of skill/config prose changes; zero-to-minimal Go — Tier M by file count and cross-surface consistency risk, not code volume.
- No constitutional surface; no new Go subsystem → not Tier L. Cross-surface consistency (4 surfaces × 2 trees) + safety-rule preservation → not Tier S.

### §A.5 PRESERVE / EXTEND map

| Surface | Disposition |
|---------|-------------|
| loop.md Steps 1/4/9, § Completion Conditions, § Supported Flags | REWRITE (predicate + verdict + precedence) |
| loop.md safety machinery (memory checkpoint, snapshots, resume, fix levels) | PRESERVE |
| moai.md § Agentic Completion Loop — cause 2 | EXTEND (protocol reference) |
| moai.md loop safety rules (escalation / dark-flow / semantic-failure) | PRESERVE |
| ralph.yaml values; workflow.yaml loop_prevention values | PRESERVE (comments EXTENDED) |
| workflow.yaml agentic_loop comment | FIX (stale loader claim) |
| internal/config Go loaders | PRESERVE (referenced only) |
| fix.md | PRESERVE (L5 out of scope) |

## §B Known Issues (filtered, Tier M — doc-primary)

- **B2 Cross-SPEC conflicts**: SPEC-V3R6-AGENTIC-LOOP-CONFIG-001 (closed) owns the agentic_loop/loop_prevention distinctness doctrine — the comment fix must cite, not contradict, its §A.4 distinctness. SPEC-MOAI-AGENTIC-LOOP-001 (closed) owns moai.md's agentic completion loop — cause-2 protocol addition must preserve its four-cause taxonomy.
- **B6 spec-lint heading conventions**: not applicable to skill files, but the SPEC's own artifacts follow OutOfScopeRule.
- **B8 Working-tree hygiene**: workflow.yaml + loop.md may be concurrently touched by sibling SPEC 2 (workflow.yaml comment areas differ: model_routing vs agentic_loop — hunks are disjoint; coordinate commit ordering).
- **B10 Scope discipline**: do not touch fix.md (L5 deliberately out of scope), goal-directive.md, or harness.yaml — they are cited precedents only.
- **B12-adjacent**: template mirrors must receive byte-identical hunks; template-neutrality guard forbids leaking this SPEC's ID into template files — write mechanism prose, not SPEC citations, in templates.

## §C Pre-flight checklist

```bash
git branch --show-current && git rev-parse HEAD
grep -n "All loop completion conditions satisfied" .claude/skills/moai/workflows/loop.md internal/template/templates/.claude/skills/moai/workflows/loop.md   # sentinel anchor re-verification (expect: Step 1 + Step 4 + Completion Conditions)
grep -n "max_iterations" .moai/config/sections/{ralph,workflow}.yaml               # ceiling fragmentation re-verification
grep -n "no Go-side loader" .moai/config/sections/workflow.yaml                    # stale comment anchor
grep -n "AgenticLoopConfig" internal/config/types.go internal/config/defaults.go   # loader existence re-verification
diff .claude/skills/moai/workflows/loop.md internal/template/templates/.claude/skills/moai/workflows/loop.md | head  # pre-existing live/template divergence check
go test ./internal/config/... 2>&1 | tail -3                                       # config loader baseline (must stay green — comments only)
```

## §D Constraints + open decisions

Constraints: see spec.md §Constraints (loop identity + safety rules preserved; vci verbatim section names; subagent boundary; doctrine-primary classification).

Open decisions (run-phase discretion unless marked):

1. **D1 — flag-default reconciliation direction** (REQ-LVC-007): loop.md currently documents `--max` default 100; the Go-loaded ralph.yaml says 10. Recommendation: document the effective default as "ralph.yaml `loop.max_iterations` (shipped 10) when `--max` is absent" and remove the freestanding "default 100" claim — the flag overrides upward/downward explicitly. Alternative (raise ralph.yaml to 100) touches config values (spec.md declares value retuning out of scope) → rejected.
2. **D2 — verdict persistence mechanism** (REQ-LVC-005): doctrine-defined JSON schema written by the orchestrator via Write/Bash at exit time (RECOMMENDED — zero Go, matches sibling state files like context-usage.json) vs a Go helper/CLI writer (rejected here; declared out of scope). Schema minimum: `spec_or_scope`, `exit_kind` (ceiling|manual-residue), `iterations_used`, `ceiling_applied` + its source (flag|ralph|loop_prevention), `conditions` final state, `remaining_issues[]` ({severity, description, file, suggested_action}), `vci_report_ref`, `created_at`.
3. **D3 — independent final pass vehicle** (REQ-LVC-003): `/moai gate` re-run (RECOMMENDED — existing mechanical gate, no new spawn cost) vs read-only verifier `Agent()` spawn (fallback where gate unavailable). Document both; primary = gate re-run.
4. **D4 — TaskList vs state-file when ledger active** (REQ-LVC-005): when a team task ledger is active, mirror remaining issues into TaskList AND write the state file (state file is the always-on floor; TaskList is additive).

## §E Self-Verification (run-phase deliverables)

Per manager-develop-prompt-template.md §E, vci 5-section format each:
- E1: AC matrix (acceptance.md §D) with verbatim grep/diff outputs.
- E2: `make build` green (template edits embedded); `go build ./...` unaffected.
- E3: n/a for coverage (doc-primary); config tests stay green (`go test ./internal/config/...`).
- E4: no AskUserQuestion additions to skill bodies (grep 0 new instances in edited files instructing subagent prompting).
- E5: template-neutrality: no `SPEC-LOOP-VERDICT-CONTRACT` token inside `internal/template/templates/**` (CI guard).
- E6: commit SHAs + push state (Route A main-direct).
- E7: blocker report if live/template pre-existing divergence found on an edited file (§C last diff check).

## §F Milestones (priority-ordered; no time estimates)

| Milestone | Scope | REQs | Exit criterion |
|-----------|-------|------|----------------|
| M1 — exit predicate + independent final pass | loop.md Steps 1/4 + § Completion Conditions rewrite; independent-pass step | REQ-LVC-001, REQ-LVC-002, REQ-LVC-003 | AC-LVC-001..004 PASS |
| M2 — ceiling verdict contract + persistence | loop.md Step 9 + moai.md cause 2 protocol; verdict JSON schema; lesson-capture step | REQ-LVC-004, REQ-LVC-005, REQ-LVC-006 | AC-LVC-005..008 PASS |
| M3 — precedence unification + comment fixes | 3-surface identical precedence rule; default reconciliation (D1); agentic_loop comment fix; template sync + make build | REQ-LVC-007, REQ-LVC-008, REQ-LVC-009 | AC-LVC-009..012 PASS |

Ordering rationale: M1 defines the predicate M2's verdict report references ("conditions final state"); M3 is textual consistency work last so it can state the finished contract.

## §G Anti-Patterns (do NOT)

- Keeping the sentence string-match as a fallback exit path ("belt and suspenders") — that re-opens the self-declared-success hole this SPEC closes.
- Letting the builder run the "independent" final pass in the same context (independence means fresh context or a distinct mechanical gate, per goal-directive precedent).
- Adding a FOURTH ceiling or renaming existing config keys — the fix is precedence + documentation, not new knobs.
- Editing ralph.yaml/workflow.yaml VALUES (only comments are in scope; values are PRESERVE).
- Writing this SPEC's ID into template-tree files (neutrality guard).
- Touching fix.md Phase 4 (L5 is explicitly out of scope — record, don't remediate).

## §H Cross-References

- spec.md (SSOT), acceptance.md (AC matrix), progress.md (§E skeleton).
- `verification-claim-integrity.md` §1.1 + §3 (verdict format source).
- SPEC-V3R6-AGENTIC-LOOP-CONFIG-001 (agentic_loop loader + distinctness doctrine — cite in the fixed comment).
- SPEC-MOAI-AGENTIC-LOOP-001 (moai.md agentic completion loop owner — cause-2 addition preserves its taxonomy).
- goal-directive.md (independent-evaluator precedent), harness.yaml minimal (evaluator switch precedent).
