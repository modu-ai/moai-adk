---
id: SPEC-CADENCE-BRIDGE-001
title: "AUTOMATE Bridge — Sanctioned Cadence Recipes — Acceptance Criteria"
version: "0.1.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "cadence, automate, read-only, workflow-reflex, acceptance"
---

# SPEC-CADENCE-BRIDGE-001 — Acceptance Criteria

> Observable, testable assertions derived from spec.md §Requirements (GEARS). Each AC traces to REQs and audit finding L1.

## §D AC Matrix

| AC ID | REQ trace | Finding | Severity | Description |
|-------|-----------|---------|----------|-------------|
| AC-CDB-001 | REQ-CDB-001 | L1 | MUST-PASS | The cadence-bridge rule exists (placement per D1) and defines ≥3 named recipes including all three mandated ones: `/loop 30m /moai gate` drift watcher, nightly `/moai review --lean`, loop-verdict backlog re-discovery |
| AC-CDB-002 | REQ-CDB-002 | L1 | MUST-PASS | Catalog-level HARD invariant present: no scheduled commit or push (Level-1 uncommitted working-tree edits are the sole permitted exception); no scheduled run-phase entry; Implementation Kickoff Approval named as human-only and cadence-unsatisfiable |
| AC-CDB-003 | REQ-CDB-003 | L1 | MUST-PASS | Discovery-to-queue contract present: found work persists to TaskList or the D2 backlog record, surfaces at next interactive session, never auto-executes remediation |
| AC-CDB-004 | REQ-CDB-004 | L1 | MUST-PASS | goal-directive.md distinctness note intact in meaning ("not interchangeable" preserved) + cross-reference to the cadence-bridge rule added |
| AC-CDB-005 | REQ-CDB-001 | L1 | MUST-PASS | Every recipe/eligibility-table entry point is verifiably read-only or advisory (gate = validation-only; review --lean = "applies no fixes, modifies no files"; backlog re-discovery = prose reader) — no write-capable subcommand appears |
| AC-CDB-006 | REQ-CDB-005 | — | MUST-PASS | New rule created template-first + goal-directive.md mirrored; `make build` green; template copy carries no internal SPEC IDs (neutrality guard clean) |

## §D.1 Severity Classification

All 6 ACs are MUST-PASS. AC-CDB-005 is the safety-critical criterion: a single write-capable entry point in the catalog fails the SPEC regardless of all other ACs (the HARD invariant is the deliverable's entire value).

## §D.2 Given-When-Then Scenarios

### AC-CDB-001 — recipe catalog

**Given** the post-M1 cadence-bridge rule
**When** reading the recipe catalog
**Then** it names at least: drift watcher (`/loop 30m /moai gate`), lean review (nightly `/moai review --lean`), and backlog re-discovery (periodic read of `.moai/state/loop-verdict-*.json` leftovers), each with interval guidance and the entry point's read-only rationale.

### AC-CDB-002 — HARD invariant

**Given** the same rule
**When** grepping for the safety constraint
**Then** a [HARD]-marked catalog-level clause states: scheduled invocations are read-only or Level-1-fix-only with no commit and no push; no scheduled run-phase entry; Implementation Kickoff Approval remains human — phrased as binding ALL recipes, present and future.

### AC-CDB-003 — discovery-to-queue

**Given** the same rule
**When** reading the "when a cadence run finds work" section
**Then** the contract specifies: persist (TaskList when live, else the D2 backlog record), surface to the user next session, and an explicit never-auto-execute prohibition.

### AC-CDB-004 — bridge cross-reference

**Given** the post-M2 goal-directive.md
**When** reading the § Comparing Autonomous-Continuation Approaches region
**Then** the "not interchangeable" note is present with unchanged meaning AND a cross-reference points to the cadence-bridge rule as the sanctioned composition surface.

## §D.3 Edge Cases

- **EC-1**: cadence run fires while another session holds the checkout (multi-session race) → the rule directs read-only recipes are race-safe by construction; the backlog WRITE (queue record) follows the multi-session pre-spawn check discipline or degrades to report-only output.
- **EC-2**: `.moai/state/loop-verdict-*.json` absent (SPEC 3 not yet landed or no ceiling exits) → recipe 3 completes as a no-op with an informational note, not an error.
- **EC-3**: user pastes a recipe with a write-capable substitution (`/loop 30m /moai fix`) → the rule's invariant text explicitly names this as unsanctioned so reviewers/orchestrators can reject it; mechanical blocking is out of scope.
- **EC-4**: Cron tools unavailable in the runtime version → recipes degrade to native `/loop` only; the rule states the fallback.

## §D.4 Verification Commands (indicative)

```bash
ls .claude/rules/moai/workflow/cadence-bridge.md internal/template/templates/.claude/rules/moai/workflow/cadence-bridge.md  # AC-CDB-001/006
grep -n "loop 30m /moai gate\|review --lean\|loop-verdict" .claude/rules/moai/workflow/cadence-bridge.md   # AC-CDB-001
grep -n "HARD" .claude/rules/moai/workflow/cadence-bridge.md                                               # AC-CDB-002
grep -n "commit\|push\|run-phase\|Kickoff" .claude/rules/moai/workflow/cadence-bridge.md                   # AC-CDB-002
grep -n "TaskList\|never auto\|auto-execute" .claude/rules/moai/workflow/cadence-bridge.md                 # AC-CDB-003
grep -n "not interchangeable\|cadence-bridge" .claude/rules/moai/workflow/goal-directive.md                # AC-CDB-004
grep -n "moai run\|moai sync\|moai loop" .claude/rules/moai/workflow/cadence-bridge.md                     # AC-CDB-005 (only as negative examples/prohibitions)
grep -rn "SPEC-LOOP-VERDICT\|SPEC-CADENCE" internal/template/templates/.claude/rules/moai/workflow/cadence-bridge.md  # AC-CDB-006 (expect 0)
make build                                                                                                  # AC-CDB-006
```

## §D.5 Quality Gate Criteria

- TRUST 5: Tested (grep-verifiable assertions + mirror diff), Readable (recipe table + one invariant clause), Unified (rules-file conventions: footer, cross-references), Secured (the HARD invariant IS the security posture — no unattended writes), Trackable (Conventional Commits; D1/D2 decisions recorded).
- Template neutrality: template copy of the new rule carries zero internal SPEC IDs / internal dates (CI guard `template-neutrality-check.yaml` + `internal_content_leak_test.go`).

## §D.6 Definition of Done

1. All 6 ACs PASS with verbatim evidence in run-phase §E self-verification.
2. D1 (placement) and D2 (backlog record surface) decisions recorded in progress.md.
3. `make build` green; neutrality guards clean.
4. Recipe 3's forward/backward link to SPEC-LOOP-VERDICT-CONTRACT-001 state (landed vs pending) recorded accurately at close time.
