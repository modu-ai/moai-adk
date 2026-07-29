---
id: SPEC-CLIFIX-LINTER-STALE-001
title: "CLI Linter/Doctor Staleness Remediation — agentlint yaml.v3, writeHeavyAgents roster, LR-07 dedupe (P3)"
version: "0.2.0"
status: draft
created: 2026-07-10
updated: 2026-07-29
author: manager-spec
priority: P3
phase: "v3.0.0 target"
module: "internal/cli/agentlint"
lifecycle: spec-anchored
tags: "cli, audit-remediation, agentlint, doctor, staleness, p3"
era: V3R6
tier: M
depends_on: [SPEC-CLIFIX-CRITICAL-001, SPEC-CLIFIX-CONTRACT-001, SPEC-CLIFIX-CONCURRENCY-001]
---

# SPEC-CLIFIX-LINTER-STALE-001 — CLI Linter/Doctor Staleness Remediation (P3)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-10 | manager-spec | Initial draft from CLI audit 2026-07-10 §3 clusters 4/5 (staleness rows) + §5 P3 roadmap row |
| 0.2.0 | 2026-07-29 | manager-spec | Re-scope per plan-audit FAIL (0.62 → Tier M 0.80). Dropped REQ-LINT-001-004 (doctor_skills allowlist — already fixed by SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001 #1088; `knownCoreSkills()` at `doctor_skills.go:33-36` now derives from `template.EmbeddedMoaiSkillNames()`). Re-scoped REQ-LINT-001-002 from "regenerate 17-agent effort matrix" to "writeHeavyAgents slice only" — `canonicalEffortMatrix` at `agent_lint.go:630-641` is already derived from `template.DefaultProfileMatrix()`; the only live residual is the `writeHeavyAgents` slice at `agent_lint.go:774-777` naming 4 archived agents. Fixed REQ-LINT-001-006 anchor (`team_spawn.go` deleted → `taskledger/taskledger.go:67-139`). Re-derived all §B line anchors against current main (`origin/main` 5832f0671). |

## §A Context

The quality gates themselves have gone stale against the live system (audit §5 P3: "품질 게이트가 실제 시스템을 다시 감사하게 됨"). Four live defects remain after the staleness sweep:

1. agentlint's hand-rolled frontmatter parser (`agent_lint.go:1183-1235`) never populates the Hooks/Skills/Sandbox fields — the `skills` and `hooks` cases of `setField` (L1229-1234) are TODO stubs ("actual list parsing happens in the main parser" / "we'll handle it separately") that populate nothing — permanently disabling rule LR-04 (dead-hook detection).
2. The `writeHeavyAgents` slice at `agent_lint.go:774-777` still names four archived agents (`expert-backend`, `expert-frontend`, `expert-refactoring`, `researcher`) — dead prose that will never match a live agent file, leaving LR-05 (isolation-drift) unable to fire on the actual write-heavy retained roster. (The `canonicalEffortMatrix` at L630-641 is already derived from `template.DefaultProfileMatrix()` and is NOT a defect; only `writeHeavyAgents` remains stale.)
3. agentlint scans both live agent files and their template mirrors (`agent_lint.go:185-191`), and `checkDuplicateMandateBlocks` (L877-967) accumulates `mandateBlocks` across ALL files (L892, L939) and emits LR-07 for every block past the first (L951-963) with no live↔mirror path-pairing or content-hash dedupe — structurally guaranteeing false positives whenever a live agent and its mirror both carry the Skeptical-Evaluator Mandate block.
4. The root help text at `help.go:57` advertises `{"moai brain", "Ideation workflow"}` — a phantom command with no implementation. And `ClaimTask` at `taskledger/taskledger.go:67-139` assigns `targetTaskID = taskID` (L101) BEFORE the pending-search loop, so nonexistent or already-completed IDs fall through the post-loop `targetTaskID == ""` check (L126) and produce a CLAIMED ledger row (L132-134) — reporting success for a non-pending task.

Findings SSOT: audit §3 cluster 4 Major row (help.go) + cluster 5 Major rows (agentlint ×2 live residuals: parser + LR-07) + the writeHeavyAgents residual + cluster 2 Major row (ClaimTask validation). The cluster-4 doctor_skills row is OUT OF SCOPE — already fixed by SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001 (#1088).

## §B Requirements (GEARS)

- REQ-LINT-001-001: The agentlint frontmatter parser shall parse agent frontmatter via yaml.v3 (agent_lint.go:1183-1235), populating the Hooks, Skills, and Sandbox fields (whose `setField` cases at L1229-1234 are currently TODO stubs that populate nothing), so rule LR-04 (dead-hook detection) can produce findings.
- REQ-LINT-001-002: The agentlint `writeHeavyAgents` slice (agent_lint.go:774-777) shall name only currently-retained write-heavy agents and shall name no archived agent, so LR-05 (isolation-drift) can fire on the actual live roster instead of matching four dead names (`expert-backend`, `expert-frontend`, `expert-refactoring`, `researcher`).
- REQ-LINT-001-003: When agentlint scans both live agent files and their template mirrors (agent_lint.go:185-191), rule LR-07 duplicate detection (checkDuplicateMandateBlocks at agent_lint.go:877-967) shall deduplicate by live↔mirror path pairing or content hash, so a live/template mirror pair is not reported as a duplicate-agent finding while genuine same-name duplicates across unrelated paths still produce findings.
- REQ-LINT-001-005: The root help text (help.go:57) shall not advertise the nonexistent `moai brain` command, and every help command entry shall resolve to a actually-registered cobra command.
- REQ-LINT-001-006: When `ClaimTask` is invoked for a task that does not exist or is not pending (taskledger/taskledger.go:67-139), the CLI shall report a claim failure instead of success; the pre-assignment `targetTaskID = taskID` at L101 shall be overridden when the pending-match search fails, so only a pending task match produces a CLAIMED ledger row at L132-134.
- REQ-LINT-001-007: **When** the run-phase regression suite executes, **the** test suite **shall** demonstrate each previously-dead check now fires: (a) one LR-04 finding on a dead-hook fixture and one non-firing on a clean fixture; (b) one LR-05 finding when the `writeHeavyAgents` slice is violated by a current-roster fixture and one non-firing on a compliant fixture; (c) zero LR-07 findings on a live/mirror fixture pair and one finding on a genuine same-name duplicate pair; (d) one claim-failure on a nonexistent task ID and one claim-failure on a completed task ID, plus one claim-success on a pending task ID.

## §C Scope

### In Scope

- agentlint yaml.v3 parser (REQ-001), writeHeavyAgents roster clean-up (REQ-002), LR-07 live/mirror dedupe (REQ-003), help.go phantom command removal + registered-command gate (REQ-005), ClaimTask task-state validation (REQ-006), and the regression tests proving the gates fire (REQ-007).

### Out of Scope — doctor_skills allowlist

- The doctor_skills allowlist defect from the original audit is already fixed by SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001 (#1088): `internal/cli/doctor_skills.go:33-36` `knownCoreSkills()` calls `template.EmbeddedMoaiSkillNames()`, and `doctor_skills_test.go:21/60` asserts the derivation. No further work is in scope here.

### Out of Scope — Ledger write mechanics

- The ClaimTask O_APPEND ledger fix is SPEC-CLIFIX-CRITICAL-001 (REQ-CRIT-001-002); this SPEC only adds task-existence/pending-state validation on top of the corrected write path.

### Out of Scope — canonicalEffortMatrix regeneration

- The `canonicalEffortMatrix` at `agent_lint.go:630-641` is already derived from `template.DefaultProfileMatrix()` (the project's single source of truth for per-agent {model, effort}). The original v0.1.0 framing ("regenerate the 17-agent matrix") described a defect that no longer exists; only the `writeHeavyAgents` slice carries stale names.

### Out of Scope — New lint rules

- No new agentlint rule codes are introduced; only existing rules are un-broken. Extending the rule set (e.g., dead-wire detection proposed in audit §4) is future work.

### Out of Scope — agentlint minor parser hygiene

- The cluster-5 minor items (frontmatter `---` anchoring, code-fence column-0 limitation, "specialist" substring over-match, errors package shadowing, first-broken-file abort) are deferred to SPEC-CLIFIX-HYGIENE-001 or later follow-ups; only the parser / writeHeavyAgents / LR-07 Major defects are fixed here.

## §D Acceptance Criteria

- AC-LINT-001-001: Given an agent fixture whose frontmatter declares hooks that reference nonexistent scripts, When agentlint runs, Then an LR-04 finding is emitted (maps REQ-LINT-001-001)
- AC-LINT-001-002: Given a fixture agent named `manager-develop` (a retained write-heavy agent) without `isolation: worktree`, When agentlint runs, Then an LR-05 finding is emitted; and given a fixture agent named `expert-backend`, Then no LR-05 finding treats it as write-heavy (maps REQ-LINT-001-002)
- AC-LINT-001-003: Given a live agent file and its byte-identical template mirror, When agentlint scans both, Then no LR-07 duplicate finding is emitted for the pair while true duplicates (two distinct live agents with the same Skeptical-Evaluator Mandate block) still produce findings (maps REQ-LINT-001-003)
- AC-LINT-001-005: Given `moai --help` output, When searched for `brain`, Then no `moai brain` entry appears and every listed command resolves to a registered cobra command (maps REQ-LINT-001-005)
- AC-LINT-001-006: Given a claim request for a nonexistent task ID and for an already-completed task, When ClaimTask runs, Then both return a claim failure and the ledger gains no claim line; and a pending task ID returns success (maps REQ-LINT-001-006)
- AC-LINT-001-007: Given the regression test suite, When it runs, Then the four previously-dead checks (LR-04, LR-05 writeHeavyAgents, LR-07 dedupe, ClaimTask pending-validation) each demonstrate a firing and a non-firing case (maps REQ-LINT-001-007)

Machine-verifiable commands and expected outcomes per AC: see `acceptance.md` (§D AC Matrix).

## §E Non-Goals and Dependencies

- Dependencies: SPEC-CLIFIX-CRITICAL-001 (ClaimTask ledger write path — O_APPEND base precedes pending-validation), SPEC-CLIFIX-CONTRACT-001 (agentlint/workflow_lint.go os.Exit removal precedes parser work in the same package), SPEC-CLIFIX-CONCURRENCY-001 (series order P0→P1→P2→P3). All three are `completed` per orchestrator-verified dependency state — no re-audit.
- Non-goal: changing agent authoring conventions or the agent catalog itself — the linter is realigned to the existing 10-agent SSOT, not vice versa. The `writeHeavyAgents` roster is reconciled to the CLAUDE.md §4 retained-agent catalog; the catalog itself is not edited.
