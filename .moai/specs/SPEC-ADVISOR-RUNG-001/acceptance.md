---
id: SPEC-ADVISOR-RUNG-001
title: "Executor-Advisor Escalation Rung + GLM Judgment Carve-Out — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai/workflows"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "advisor-rung, escalation, glm-carve-out, workflow-reflex, acceptance"
---

# SPEC-ADVISOR-RUNG-001 — Acceptance Criteria

> Observable, testable assertions derived from spec.md §Requirements (GEARS). Each AC traces to REQs and an audit finding (R3/R6).

## §D AC Matrix

| AC ID | REQ trace | Finding | Severity | Description |
|-------|-----------|---------|----------|-------------|
| AC-ADV-001 | REQ-ADV-001 | R3 | MUST-PASS | loop.md carries the advisor rung instruction: N=2 consecutive same-diagnostic failure trigger, failure-evidence payload, diagnosis re-seed, user escalation only after the rung fails |
| AC-ADV-002 | REQ-ADV-001 | R3 | MUST-PASS | fix.md carries the advisor rung on BOTH repeat-failure paths (Level-3-class re-failure AND CI-loop patch failure), positioned before the respective AskUserQuestion escalations |
| AC-ADV-003 | REQ-ADV-002 | R3 | MUST-PASS | Advisor spawn contract present on the edited surfaces: read-only whitelist (no Write/Edit), per-spawn model/effort runtime args (never frontmatter pin), citing the SPEC-MODEL-ROUTING-WIRE-001 / model-policy `[1m]`-safe channel |
| AC-ADV-004 | REQ-ADV-003 | R3 | MUST-PASS | Escalation contracts intact: ci-autofix frozen text ("at most 3 iterations", "MANDATORY BLOCKING AskUserQuestion") unchanged verbatim; advisor documented as within-budget; no first-failure trigger anywhere |
| AC-ADV-005 | REQ-ADV-004 | R6 | MUST-PASS | glm.md carries the all-GLM judgment carve-out: plan-auditor/sync-auditor + planning named reduced-assurance, SHOULD-run-in-Claude-or-flag directive, CLAUDE.md §15 Avoid-list items mirrored |
| AC-ADV-006 | REQ-ADV-004, REQ-ADV-005 | R6 | MUST-PASS | run.md §CG cross-references the carve-out AND names leader review as the on-demand advisor (blocker report → leader advisory diagnosis → re-delegation), extending the blocker-report boundary |
| AC-ADV-007 | REQ-ADV-006 | — | MUST-PASS | All 5 edited surfaces mirrored template-first; `make build` green; template-neutrality guard unaffected |

## §D.1 Severity Classification

All 7 ACs are MUST-PASS. AC-ADV-001 carries a **sequencing precondition**: the loop.md hunk is evaluated only after SPEC-LOOP-VERDICT-CONTRACT-001's run-phase has landed on loop.md (or explicit orchestrator coordination is recorded); editing loop.md into a conflicting state is itself a FAIL condition.

## §D.2 Given-When-Then Scenarios

### AC-ADV-001 — loop advisor rung

**Given** the post-M1 loop.md
**When** grepping the iteration-cycle/escalation region for the advisor instruction
**Then** the text specifies all four elements: (1) trigger = 2 consecutive failed iterations on the same diagnostic, (2) read-only strong-model advisor spawn with the failure evidence, (3) executor re-seed with the advisor diagnosis, (4) user escalation only after the advisor rung also fails.

### AC-ADV-004 — frozen contract preservation

**Given** the post-M1 ci-autofix-protocol.md
**When** diffing against the pre-edit baseline
**Then** every [ZONE:Frozen] clause is byte-identical; the only change is an additive Evolvable subsection stating the advisor consult happens within the 3-iteration budget and never delays the iteration-4+ blocking AskUserQuestion.

### AC-ADV-005 — GLM carve-out

**Given** the post-M2 glm.md
**When** grepping for the carve-out
**Then** a section exists naming all-GLM (`team_mode: glm`) judgment gates (plan-auditor/sync-auditor) and planning as reduced-assurance, directing they SHOULD run in a Claude session or be flagged, and listing the mirrored Avoid items (planning/architecture, security reviews, complex debugging).

### AC-ADV-006 — CG advisor naming

**Given** the post-M2 run.md § CG Mode
**When** reading the teammate-stuck guidance
**Then** the leader-review path is explicitly named as the on-demand advisor rung: teammate returns a blocker report → the Claude leader produces an advisory diagnosis → re-delegation carries the diagnosis; the underlying blocker-report format is cited, not redefined.

## §D.3 Edge Cases

- **EC-1**: advisor itself fails or returns no usable diagnosis → the workflow proceeds to the existing user escalation (the rung never becomes an infinite loop; one advisor consult per stuck-window).
- **EC-2**: diagnostic changes between iterations (not "same diagnostic") → counter resets; no advisor spawn.
- **EC-3**: user already answered a blocking AskUserQuestion for this diagnostic → advisor rung does not retroactively fire; the user decision governs.
- **EC-4**: all-GLM session runs `/moai loop` — advisor spawn under all-GLM resolves to a GLM model; the carve-out's reduced-assurance flag applies to the advisor's diagnosis too (documented, not blocked).

## §D.4 Verification Commands (indicative)

```bash
grep -n "advisor" .claude/skills/moai/workflows/loop.md                        # AC-ADV-001
grep -n "advisor" .claude/skills/moai/workflows/fix.md                         # AC-ADV-002 (expect both paths)
grep -n "per-spawn\|read-only" .claude/skills/moai/workflows/loop.md .claude/skills/moai/workflows/fix.md  # AC-ADV-003
grep -n "MANDATORY BLOCKING AskUserQuestion" .claude/rules/moai/workflow/ci-autofix-protocol.md  # AC-ADV-004 (verbatim intact)
grep -n "reduced-assurance\|Avoid" .claude/skills/moai/team/glm.md             # AC-ADV-005
grep -n "advisor\|carve-out" .claude/skills/moai/team/run.md                   # AC-ADV-006
for f in workflows/fix.md workflows/loop.md team/glm.md team/run.md; do diff .claude/skills/moai/$f internal/template/templates/.claude/skills/moai/$f; done  # AC-ADV-007
diff .claude/rules/moai/workflow/ci-autofix-protocol.md internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md  # AC-ADV-007
make build                                                                      # AC-ADV-007
```

## §D.5 Quality Gate Criteria

- TRUST 5: Tested (grep-verifiable text assertions + template diff parity), Readable (advisor instruction states trigger/spawn/re-seed/escalate in ≤1 section per surface), Unified (identical advisor vocabulary across loop.md/fix.md), Secured (no secrets; advisor read-only), Trackable (Conventional Commits; D1/D2 decisions recorded).
- Template neutrality: no SPEC IDs / internal dates leaked into `internal/template/templates/` (CI guard).

## §D.6 Definition of Done

1. All 7 ACs PASS with verbatim evidence in run-phase §E self-verification.
2. D1 (placement), D2 (ci-autofix touch scope), D3 (same-diagnostic identity), D4 (advisor model default) decisions recorded in progress.md before the respective hunks commit.
3. Frozen-anchor diff evidence archived (AC-ADV-004).
4. loop.md sequencing precondition satisfied and recorded (§D.1).
