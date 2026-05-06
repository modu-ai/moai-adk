# Acceptance: SPEC-AUTO-DELEG-001 (Auto-Delegation Rules Strengthening)

> Companion to `spec.md`. Detailed Given-When-Then scenarios, edge cases, quality gates.

---

## Scenario 1 — CLAUDE.md §1 Contains 5 New [HARD] Triggers (REQ-DEL-001, AC-1)

**Given** the SPEC has rolled out and CLAUDE.md is updated
**When** a `grep '\[HARD\]' CLAUDE.md` is executed
**Then** the count of [HARD] entries SHALL be exactly 16 (11 pre-existing + 5 new delegation triggers).
**And** the 5 new entries SHALL appear under a sub-heading "Delegation Triggers" immediately after the "Multi-File Decomposition" rule.
**And** each new entry SHALL contain a quantitative threshold or observable trigger (file count, work-unit count, freshness signal, pipeline shape, write count).

**Edge case 1a**: An existing [HARD] rule (e.g., "Multi-File Decomposition") superficially resembles a delegation trigger but is actually about TodoWrite-style decomposition. The new triggers SHALL not duplicate this — they explicitly target Agent() spawn, distinguished by phrasing.

**Edge case 1b**: A future SPEC adds another [HARD] rule. The audit count expectation (16) is captured in `.claude/rules/moai/core/zone-registry.md` for traceability; future SPECs increment the registry.

---

## Scenario 2 — All 24 Agent Descriptions Conform to Trigger Format (REQ-DEL-009, REQ-DEL-010, AC-2, AC-4)

**Given** the SPEC has rolled out and all agent files in `.claude/agents/moai/*.md` are updated
**When** the audit script `internal/template/agent_description_audit_test.go` runs
**Then** for each of the 24 (or 23, per OQ-3 resolution) agent files, the test SHALL assert:
  - The frontmatter `description` field contains the literal substring `Use PROACTIVELY when:`.
  - The frontmatter `description` field does NOT contain the literal substring `Use PROACTIVELY for ` (legacy form).
  - The description block contains at least 2 trigger conditions (counted by `- ` bullets following `Use PROACTIVELY when:`).
  - The description block contains a `NOT for:` clause.
**And** all assertions SHALL pass.

**Edge case 2a**: An agent's description is so simple that 2 trigger conditions feel forced (e.g., `manager-git` is narrow). The 2-trigger minimum is enforced; the rewriter writes triggers like "Git operations needed (commit, branch, PR creation)" and "Release tag automation" — even narrow agents have multiple natural triggers.

**Edge case 2b**: An agent has a typo in `PROACTIVELY` (e.g., `PROACITVELY`). The audit SHALL fail with a clear error message identifying the file and the typo.

**Edge case 2c**: An agent description includes `Use PROACTIVELY when:` but the bullets are missing (e.g., "Use PROACTIVELY when: API design or auth flow needs to be implemented" all on one line). The audit SHALL fail because the 2-bullet count is not met. The rewriter restructures to bulleted form.

---

## Scenario 3 — NOT-for Graph is Acyclic (REQ-DEL-011, AC-8)

**Given** all 24 descriptions have NOT-for clauses with specific agent name references
**When** the graph audit script `internal/template/agent_routing_graph_test.go` runs
**Then** the script SHALL parse each description, extract `(agent, target_agent)` edges from NOT-for clauses, build a directed graph, and run a cycle-detection algorithm.
**And** the assertion SHALL be: zero cycles AND every target_agent name resolves to an existing file in `.claude/agents/moai/`.

**Edge case 3a**: Two agents reference each other (e.g., expert-backend NOT for X, use expert-frontend; expert-frontend NOT for Y, use expert-backend). This is a 2-cycle. The audit fails. The rewriter ensures one direction or uses generic phrasing for the second leg ("use a different specialist").

**Edge case 3b**: An agent's NOT-for clause references a non-existent agent (e.g., "use expert-cloud" but expert-cloud doesn't exist). The audit fails. Resolution: either create the referenced agent (out of scope per EX-2) or change the NOT-for clause to reference an existing agent.

**Edge case 3c**: An agent's NOT-for clause uses generic phrasing ("use a different specialist") without naming. The graph extractor records this as a no-op edge (no graph contribution); the audit does not penalize. But the routing-hint chain is weaker for this agent. Reviewer flag for follow-up SPEC.

---

## Scenario 4 — Auto-Delegation Rate Improves by 30 Percentage Points (REQ-DEL-013, REQ-DEL-014, AC-3, AC-9)

**Given** the baseline measurement artifact `.moai/reports/auto-deleg-baseline-{DATE}/sample-50-turns.md` exists with computed `baseline_rate = B`
**And** the post-rollout window is at least 14 days
**When** the pilot measurement is performed on a new 50-turn sample
**Then** the `pilot_rate = P` SHALL be computed using identical methodology (same definition of "auto-delegated" — no user phrasing "Use the X subagent").
**And** `P - B >= 0.30` SHALL hold.
**And** the pilot artifact SHALL be filed at `.moai/reports/auto-deleg-pilot-{DATE}/sample-50-turns.md`.

**Edge case 4a**: B is already high (e.g., B = 0.65). Then P >= 0.95 is the target. If P caps at 0.95 - 1.00, room for absolute improvement is constrained but the SPEC still adds value (consistency, not just rate). Document: if `B >= 0.65`, the differential target relaxes to `P >= 0.85` (a flat 85% absolute target), recorded as a pre-pilot calibration.

**Edge case 4b**: P < B + 0.30 but other quality signals improved (e.g., agent matching accuracy). REQ-DEL-015 triggers a tuning artifact; do not auto-rollback. Allow user to decide whether to accept the result, tune triggers, or follow up.

**Edge case 4c**: Sample contamination — the 50 post-rollout turns include unrelated environment changes (e.g., user changed habits, new codebase areas). Methodology requires turn-source diversity (at least 5 different sessions, 3 different SPEC types). Document this discipline in M0 baseline artifact.

---

## Scenario 5 — Cross-Reference Consistency (REQ-DEL-016, REQ-DEL-017, AC-5, AC-6)

**Given** the rollout includes CLAUDE.md §1 changes, `.claude/rules/moai/development/agent-authoring.md` updates, and CLAUDE.local.md §16 deferral
**When** a manual cross-reference review is performed
**Then** the reviewer SHALL verify:
  - CLAUDE.md §1 "Delegation Triggers" sub-heading exists with 5 entries.
  - agent-authoring.md has a "Description Trigger Format" section that links to CLAUDE.md §1.
  - `.claude/rules/moai/workflow/spec-workflow.md` references the new triggers if it discusses delegation patterns (verify in plan-auditor pass).
  - CLAUDE.local.md §16 has been updated to defer to CLAUDE.md §1 (no duplicated trigger logic in the local file).
  - CLAUDE.md §14 (Parallel Execution Safeguards) does not contradict the new §1 triggers.
**And** any contradictions SHALL be filed as a blocker on this SPEC's completion.

**Edge case 5a**: CLAUDE.md §16 (which is currently CLAUDE.local.md only — section number may differ in shared CLAUDE.md) discusses self-check protocol. Need to verify the section anchoring is correct after promotion. Cross-check with the auto-loaded zone-registry.md to ensure HARD entry IDs match anchors.

---

## Scenario 6 — Trigger Activation in Live Sessions (REQ-DEL-003 through REQ-DEL-008)

**Given** the SPEC has rolled out
**When** a user submits a task that requires reading 12 files in different domains
**Then** the orchestrator SHALL recognize the file-exploration trigger (REQ-DEL-004) and delegate to `Explore` (or a domain manager subagent) at the start of the task.
**And** the delegation SHALL occur without the user explicitly saying "Use the Explore subagent."
**And** the delegation SHALL be observable in the session transcript as an `Agent()` call.

**Edge case 6a**: A task requires reading exactly 10 files. The threshold is "exceeds 10" per REQ-DEL-004 phrasing. 10 itself does not trigger. 11 triggers. Boundary semantics documented.

**Edge case 6b**: A task initially looks single-file but expands during execution to 11 files. The trigger fires retroactively — the orchestrator delegates the remaining work even mid-task. (This is a soft expectation; depends on Claude's runtime decision-making capacity.)

**Edge case 6c**: Three independent work units are presented but two finish trivially (1-line each) — only one is genuinely substantial. The independence trigger (REQ-DEL-005) fired at task-start; the substantive unit is delegated. The trivial units remain in the orchestrator. This is correct: independence triggers delegation of at least one unit, not all.

**Edge case 6d**: User says "Use the X subagent" explicitly even though no trigger fired. The orchestrator obeys (user-explicit always wins). This is not counted toward the auto-delegation rate.

---

## Scenario 7 — Exception List Honored (REQ-DEL-018, AC-5)

**Given** a task is "fix typo in README.md line 23"
**When** the orchestrator processes the task
**Then** no delegation trigger SHALL fire (matches the explicit exception "typo fixes, single-line changes").
**And** the orchestrator SHALL execute directly via Edit tool without spawning Agent().

**Edge case 7a**: User submits "fix typo in README.md and update CHANGELOG.md and bump version" — three distinct units. The independence trigger (REQ-DEL-005) fires because there are 3 units. But each individual unit is trivially small. The triggers' rationale is to delegate substantive work; the orchestrator may handle this directly with TodoWrite or delegate to manager-git. Resolution: when triggers fire AND units are individually trivial, the orchestrator MAY handle directly per the existing CLAUDE.md §7 Rule 5 exception clause "Single-line typos or formatting fixes." Document this trump rule in plan phase.

---

## Scenario 8 — Body Content Untouched (REQ-DEL-020, AC-2, AC-10)

**Given** the SPEC's diff for any of the 24 agent files
**When** the diff is reviewed
**Then** the changes SHALL be confined to:
  - The `description:` frontmatter field
  - Optionally a small leading capability summary in the body if it duplicates description text (capability summary is part of description, not body)
**And** no changes SHALL appear in:
  - Workflow Steps section
  - Scope Boundaries section
  - Delegation Protocol section
  - Adaptive Behavior section
  - Tools field
  - Model field
  - Hooks field
  - Skills field
**And** the diff size per agent SHALL be small (~10-30 lines per agent).

**Edge case 8a**: An agent's body has a paragraph that duplicates the description (a common artifact). Cleanup of that duplication is acceptable as a side effect of description rewrite, but is not the SPEC's primary change. Document any such cleanups in the M3 before-after artifact.

**Edge case 8b**: An agent's body has stale references to the legacy description format (e.g., "as the description says, used for X"). Updating these stale references is in-scope per the cross-reference principle but should be minimal. Document in M3 artifact.

---

## Scenario 9 — Multi-Language Sections Consistent (OQ-5, REQ-DEL-019)

**Given** an agent file has EN/KO/JA/ZH keyword sections (e.g., expert-security.md lines 7-10)
**When** the trigger-format rewrite is applied
**Then** all four sections SHALL be updated with localized trigger keywords (where appropriate).
**And** the `Use PROACTIVELY when:` phrase itself remains in English (per existing convention).
**And** the `NOT for:` clause is in English (per existing convention).

**Edge case 9a**: A localization change introduces a translation error (e.g., Chinese trigger doesn't match the English meaning). The audit (M4) is English-only; a separate manual review by Korean/Japanese/Chinese speakers is recommended for first-pass quality. Translation review is not blocking for AC-2 but documented as recommended.

---

## Scenario 10 — Tuning Artifact When Target Missed (REQ-DEL-015)

**Given** the pilot measurement shows P < B + 0.30 (target missed)
**And** at least 14 days have elapsed since rollout
**When** the SPEC owner reviews the pilot result
**Then** the owner SHALL author `.moai/reports/auto-deleg-tuning-{DATE}/regression.md` containing:
  - The 50-turn sample with annotations: which triggers should have fired but didn't?
  - Hypothesis: too high threshold? too narrow scope? trigger phrasing not picked up by Claude?
  - Proposed tuning (e.g., "lower file-exploration threshold from 10 to 8" or "rephrase trigger to be more concrete").
  - Decision: accept the pilot result, retry with tuned triggers, or roll back to pre-SPEC state.
**And** the SPEC status SHALL be updated to `partially_approved` if accepted, or `superseded` if rolled back.

**Edge case 10a**: Tuning is accepted. The tuned triggers form SPEC-AUTO-DELEG-002 (follow-up SPEC), keeping SPEC-AUTO-DELEG-001 as the base.

---

## Scenario 11 — Audit Test Counts Match Inventory (OQ-3, AC-4)

**Given** the agent inventory artifact `.moai/reports/auto-deleg-count-{DATE}/inventory.md` from M0 documents the official agent count (24 or 23)
**When** the audit test runs
**Then** the test SHALL assert: number of `.claude/agents/moai/*.md` files equals the documented count.
**And** if the count drifts (e.g., new agent added without SPEC-AUTO-DELEG-002 update), the test SHALL fail with a clear error.

---

## Quality Gates (TRUST 5 Mapping)

- **Tested**: Scenarios 2, 3, 11 are automated audits (M4, M5). Scenarios 4, 10 are semi-automated measurements (M0, M6). Scenarios 1, 5, 6, 7, 8, 9 are manual review checkpoints with documented criteria.
- **Readable**: New CLAUDE.md sub-heading "Delegation Triggers" groups the 5 entries visually. agent-authoring.md "Description Trigger Format" section is a single source of truth for the format.
- **Unified**: All 24 agents conform to the same description structure (REQ-DEL-009). Audit script enforces (M4).
- **Secured**: NOT-for clauses bound delegation; over-delegation risk (R1 in plan.md) is mitigated by NOT-for + the sparse trigger set (5 triggers, not 50). Privacy: baseline measurement uses session transcripts; no user content is published in artifacts (artifact contains turn IDs and binary auto-delegated yes/no flags only).
- **Trackable**: HISTORY in spec.md; baseline + pilot artifacts in `.moai/reports/`; audit test outcomes in CI logs; zone-registry.md updated with 5 new HARD entries.

---

## Definition of Done

- [ ] M0 — OQs resolved; baseline measurement artifact filed
- [ ] M1 — CLAUDE.md §1 has 5 new [HARD] delegation trigger entries
- [ ] M2 — agent-authoring.md has "Description Trigger Format" section
- [ ] M3 — All 24 (or 23, confirmed) agent descriptions rewritten in trigger-centric form
- [ ] M4 — agent_description_audit_test.go passes
- [ ] M5 — agent_routing_graph_test.go passes (acyclic, all targets resolved)
- [ ] M6 — Pilot measurement performed; AC-3 satisfied OR tuning artifact filed
- [ ] All Scenarios 1-11 satisfied (or documented as edge cases)
- [ ] plan-auditor verdict on this SPEC: PASS
- [ ] User approval (AP-4) on pilot outcome

---

**Status**: draft — pending plan-auditor review and pilot rollout
