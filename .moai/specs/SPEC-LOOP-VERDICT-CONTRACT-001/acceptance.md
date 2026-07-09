---
id: SPEC-LOOP-VERDICT-CONTRACT-001
title: "Mechanical Loop Termination Predicate and Ceiling-Exit Verdict Contract — Acceptance Criteria"
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
tags: "loop, ralph, termination-predicate, verdict-contract, ceiling-exit, max-iterations-precedence, workflow-reflex, acceptance"
---

# SPEC-LOOP-VERDICT-CONTRACT-001 — Acceptance Criteria

> Observable, testable assertions derived from spec.md §Requirements (GEARS). Each AC traces to REQs and an audit finding (L3/L4/L6/L8). Deliverables are doc-primary, so most ACs verify by grep against the edited surfaces (live + template mirror).

## §D AC Matrix

| AC ID | REQ trace | Finding | Severity | Description |
|-------|-----------|---------|----------|-------------|
| AC-LVC-001 | REQ-LVC-001 | L3 | MUST-PASS | loop.md exit decision references parsed Step-3 diagnostics evaluated against ralph.yaml `loop.completion` |
| AC-LVC-002 | REQ-LVC-002 | L3 | MUST-PASS | No instruction remains to exit on detecting the completion sentence; sentence is display-only |
| AC-LVC-003 | REQ-LVC-003 | L3 | MUST-PASS | Independent final pass step present before success-exit (gate re-run / read-only verifier; not the loop executor) |
| AC-LVC-004 | REQ-LVC-003 | L3 | MUST-PASS | Divergence rule present: builder-vs-independent diagnostic mismatch continues/escalates, never exits success |
| AC-LVC-005 | REQ-LVC-004 | L4 | MUST-PASS | Ceiling exit emits vci §3 5-section report (Claim/Evidence/Baseline-attribution/Gaps/Residual-risk named verbatim) |
| AC-LVC-006 | REQ-LVC-004 | L4 | MUST-PASS | moai.md Agentic Completion Loop termination cause 2 carries the verdict protocol (parity with causes 3/4) |
| AC-LVC-007 | REQ-LVC-005 | L8 | MUST-PASS | Remaining issues persist to `.moai/state/loop-verdict-<id>.json` (schema documented) or TaskList; transcript-only residue prohibited |
| AC-LVC-008 | REQ-LVC-006 | L8 | MUST-PASS | Unsuccessful exit requires a lesson-capture proposal step |
| AC-LVC-009 | REQ-LVC-007 | L6 | MUST-PASS | Precedence rule (CLI flag > ralph.yaml > workflow.yaml loop_prevention) stated IDENTICALLY on all three surfaces |
| AC-LVC-010 | REQ-LVC-007 | L6 | MUST-PASS | Defaults reconciled: no freestanding "default 100" flag claim contradicting the Go-loaded ralph.yaml 10; memory-safe 50 documented as orthogonal checkpoint |
| AC-LVC-011 | REQ-LVC-008 | L8 | MUST-PASS | workflow.yaml agentic_loop comment references the implemented Go loader; "no Go-side loader field yet" removed |
| AC-LVC-012 | REQ-LVC-009 | — | MUST-PASS | All edited surfaces mirrored template-first; `make build` green; no SPEC-ID leak into templates |
| AC-LVC-013 | (constraint) | — | MUST-PASS | Loop safety machinery preserved: no-progress escalation, dark-flow guard, semantic-failure escalation, memory checkpoint, alias contract all still present verbatim-or-equivalent |

## §D.1 Severity Classification

All 13 ACs are MUST-PASS. AC-LVC-013 is the regression guard — a rewrite that improves termination but drops a safety rule is a net loss and fails the SPEC.

## §D.2 Given-When-Then Scenarios

### AC-LVC-001/002 — mechanical predicate replaces sentinel

**Given** the post-M1 loop.md (live and template copies)
**When** grepping the per-iteration cycle for the exit decision
**Then** Step 1 (or its successor) instructs re-evaluating the previous iteration's parsed diagnostics (exit codes / error count / test result / coverage) against ralph.yaml `loop.completion`, AND no text instructs exiting because the completion sentence string was detected; the sentence appears only in a display/reporting role.

### AC-LVC-003/004 — independent final pass

**Given** the post-M1 loop.md
**When** reading the success-exit path
**Then** a step requires an independent verification pass (documented primary vehicle: `/moai gate` re-run; fallback: read-only verifier spawn) executed outside the loop executor's own claim, and a divergence rule states that mismatch between builder-observed and independently-observed results continues the loop or escalates — success-exit is only reachable after independent confirmation.

### AC-LVC-005/007 — ceiling verdict with persistence

**Given** the post-M2 loop.md Step 9 (ceiling branch)
**When** reading the ceiling-exit instructions
**Then** the branch instructs emitting a 5-section report naming Claim, Evidence, Baseline-attribution, Gaps, Residual-risk (vci §3), and writing the remaining issues to `.moai/state/loop-verdict-<id>.json` per the documented schema (minimum fields: exit_kind, iterations_used, ceiling_applied + source, conditions final state, remaining_issues[], created_at) — with the TaskList mirror rule when a ledger is active.

### AC-LVC-009 — identical precedence rule on three surfaces

**Given** the post-M3 loop.md, ralph.yaml, workflow.yaml (live + mirrors)
**When** grepping each for the precedence statement
**Then** all three carry the same rule text (CLI `--max` flag > ralph.yaml `loop.max_iterations` > workflow.yaml `loop_prevention.max_iterations`), and cross-surface diff of the rule sentence shows no semantic divergence.

### AC-LVC-011 — stale loader comment fixed

**Given** the post-M3 workflow.yaml (live + mirror)
**When** reading the agentic_loop comment
**Then** it references the implemented typed loader (AgenticLoopConfig / internal/config) and the string "no Go-side loader field yet" is absent from both trees.

## §D.3 Edge Cases

- **EC-1**: predicate satisfied but independent pass unavailable (gate not runnable in environment) → documented degradation: loop reports the gap explicitly in its exit report (Gaps section) rather than silently claiming full verification — never a silent success.
- **EC-2**: ceiling reached on iteration 1 (pathological small ceiling) → verdict contract still applies; report notes zero-progress baseline.
- **EC-3**: `.moai/state/` unwritable at exit → verdict content surfaced in-conversation AND the write failure named in Residual-risk; exit proceeds (fail-open on persistence, not on honesty).
- **EC-4**: sentence string appears inside quoted historical text (e.g. this SPEC's own citations) → no exit implication; the predicate path never scans transcript for the string.
- **EC-5**: memory-safe 50-iteration checkpoint fires before the configured ceiling → existing checkpoint behavior preserved; documented as orthogonal to the precedence chain.

## §D.4 Verification Commands (indicative)

```bash
grep -n "loop.completion\|parsed diagnostics\|exit code" .claude/skills/moai/workflows/loop.md            # AC-LVC-001
grep -n "completion sentence" .claude/skills/moai/workflows/loop.md                                        # AC-LVC-002 (display-only context; no exit instruction)
grep -n "independent\|/moai gate\|fresh" .claude/skills/moai/workflows/loop.md                             # AC-LVC-003/004
grep -n "Claim\|Baseline-attribution\|Residual-risk" .claude/skills/moai/workflows/loop.md                 # AC-LVC-005
grep -n "loop-verdict" .claude/skills/moai/workflows/loop.md .claude/skills/moai/workflows/moai.md         # AC-LVC-006/007
grep -n "lesson" .claude/skills/moai/workflows/loop.md                                                     # AC-LVC-008
grep -n "precedence\|--max" .claude/skills/moai/workflows/loop.md .moai/config/sections/ralph.yaml .moai/config/sections/workflow.yaml  # AC-LVC-009/010
grep -n "no Go-side loader" .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml  # AC-LVC-011 (expect 0)
grep -rn "SPEC-LOOP-VERDICT-CONTRACT" internal/template/templates/ | wc -l                                 # AC-LVC-012 (expect 0)
grep -n "no-progress\|dark-flow\|semantic-failure\|memory-safe\|--mode loop" .claude/skills/moai/workflows/{loop,moai}.md  # AC-LVC-013
make build && go test ./internal/config/... 2>&1 | tail -3                                                 # AC-LVC-012 + config green
diff <(sed -n '/precedence/p' .claude/skills/moai/workflows/loop.md) <(sed -n '/precedence/p' internal/template/templates/.claude/skills/moai/workflows/loop.md)  # mirror parity spot-check
```

## §D.5 Quality Gate Criteria

- TRUST 5: Tested (grep-AC matrix + config tests green), Readable (rewritten steps keep the numbered-step structure), Unified (live/template byte-parity on edited hunks), Secured (no new executable surface), Trackable (Conventional Commits per milestone).
- vci compliance: the SPEC's own run-phase §E report uses the same 5-section format it mandates for the loop (self-consistency).

## §D.6 Definition of Done

1. All 13 ACs PASS with verbatim grep/diff evidence in run-phase §E self-verification.
2. Verdict-file schema documented in loop.md (doctrine-defined; no Go loader added).
3. Both trees (live + template) carry identical hunks; `make build` green; template-neutrality guard clean.
4. L5 (fix.md) remains untouched and is recorded as open audit debt in the sync-phase CHANGELOG note.
