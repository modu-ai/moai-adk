# SPEC-CORE-BEHAV-001: Agent Core Behaviors and SPEC Workflow Gates

## Meta

- **Status**: Draft
- **Wave**: 1 (parallel with SKILL-ENHANCE-001, TELEMETRY-001)
- **Created**: 2026-04-11
- **Origin**: addyosmani/agent-skills "Six Core Operating Behaviors" (P2.1) + SPEC Gates (P2.2)
- **Blocked By**: SPEC-EVO-001

## Objective

Consolidate six cross-cutting agent behaviors into a single constitutional section, and add explicit human review gate markers to workflow skills (plan/run/sync). Currently these behaviors are scattered across CLAUDE.md Section 7 and agent-common-protocol.md without being enumerated as a cohesive behavioral framework.

## Background

addyosmani/agent-skills defines six "Core Operating Behaviors" in its meta-skill that apply regardless of which specific skill is active. moai-adk has equivalent rules scattered across:
- CLAUDE.md Section 7 Rules 1-5
- agent-common-protocol.md (time estimation, tool usage)
- moai-constitution.md (parallel execution, quality gates)

Missing as explicit HARD rules: **Surface Assumptions**, **Push Back When Warranted**, **Scope Discipline**.

Additionally, workflow skills (plan/run/sync) have implicit phase transitions but lack explicit "HUMAN GATE" markers that clearly state what must be approved before proceeding.

## Requirements (EARS Format)

### R1: Agent Core Behaviors Section [UBIQ]

moai-constitution.md SHALL contain a `## Agent Core Behaviors` section with six numbered HARD rules.

**The Six Behaviors:**

1. **Surface Assumptions** — Before implementing anything non-trivial, list assumptions explicitly and wait for confirmation.
   - Cross-ref: CLAUDE.md Section 7 Rule 5 (Context-First Discovery)
   - Template: `ASSUMPTIONS: 1. [...] 2. [...] -> Correct me now or I proceed with these.`

2. **Manage Confusion Actively** — When encountering inconsistencies, STOP, name the confusion, present options, wait.
   - Cross-ref: CLAUDE.md Section 7 Rule 5 (trigger conditions)
   - Anti-pattern: Silently picking one interpretation

3. **Push Back When Warranted** — Point out issues directly, quantify downsides, propose alternatives.
   - NEW behavior (not currently in any moai rule)
   - Anti-pattern: Sycophantic agreement ("Of course!")

4. **Enforce Simplicity** — Actively resist overcomplexity. Ask: "Can this be done in fewer lines?"
   - Cross-ref: TRUST 5 Readable
   - Anti-pattern: Building 1000 lines when 100 would suffice

5. **Maintain Scope Discipline** — Touch only what was asked. No drive-by refactors.
   - Cross-ref: CLAUDE.md Section 7 Rule 2 (Multi-File Decomposition)
   - Anti-pattern: "Cleaning up" adjacent code

6. **Verify, Don't Assume** — Every task requires evidence. "Seems right" is never sufficient.
   - Cross-ref: CLAUDE.md Section 7 Rule 3 (Post-Implementation Review)
   - Anti-pattern: Claiming tests pass without showing output

**Acceptance Criteria:**
- [ ] New section added to `internal/template/templates/.claude/rules/moai/core/moai-constitution.md`
- [ ] All 6 behaviors are HARD rules
- [ ] Each behavior has: name, one-sentence rule, cross-reference, anti-pattern example
- [ ] Section placed after "Lessons Protocol" and before "URL Verification"
- [ ] `make build` succeeds; local copy updated

### R2: HUMAN GATE Markers in Workflow Skills [EVENT]

WHEN a workflow skill defines a phase transition that requires user approval, the skill SHALL include a `### HUMAN GATE` subsection.

**Gate Format:**
```markdown
### HUMAN GATE: [Gate Name]

**Previous phase output:** [Required artifact from previous phase]
**Approval question:** [Specific question to ask the user]
**Cannot proceed until:**
- [ ] [Specific condition 1]
- [ ] [Specific condition 2]
```

**Target Gates:**

**plan.md:**
- Gate 1 (existing Decision Point 1, line ~305): Before SPEC creation
- Gate 2 (existing Phase 3.6): SPEC quality validation before execution mode selection

**run.md:**
- Gate 1 (existing Decision Point 1, line ~257): Plan approval before Phase 2 implementation
- Gate 2 (new): After Phase 2.8 evaluation, before git operations

**sync.md:**
- Gate 1 (existing Phase 0, line ~125): Pre-sync quality verification
- Gate 2 (existing Phase 1.6, line ~524): User approval before document generation

**Acceptance Criteria:**
- [ ] 6 HUMAN GATE markers added across plan.md, run.md, sync.md
- [ ] Each gate has previous-phase-output, approval-question, cannot-proceed-until
- [ ] Existing Decision Points are enhanced (not replaced) with gate format
- [ ] Gates are wrapped in `<!-- moai:evolvable-start id="gate-XXX" -->` markers
- [ ] `make build` succeeds; local copies updated

### R3: Cross-Reference Consistency [UBIQ]

CLAUDE.md Section 7 SHALL reference the new Agent Core Behaviors section in moai-constitution.md.

**Acceptance Criteria:**
- [ ] CLAUDE.md Section 7 intro paragraph references `moai-constitution.md Agent Core Behaviors`
- [ ] No duplication of content between CLAUDE.md and constitution.md
- [ ] agent-common-protocol.md references constitution.md for behavioral rules

## Modified Files

### Templates (must `make build`)
- `internal/template/templates/.claude/rules/moai/core/moai-constitution.md`: Add Agent Core Behaviors section
- `internal/template/templates/.claude/skills/moai/workflows/plan.md`: Add HUMAN GATE markers
- `internal/template/templates/.claude/skills/moai/workflows/run.md`: Add HUMAN GATE markers
- `internal/template/templates/.claude/skills/moai/workflows/sync.md`: Add HUMAN GATE markers
- `internal/template/templates/CLAUDE.md`: Add cross-reference in Section 7

### Local Copies (sync after make build)
- `.claude/rules/moai/core/moai-constitution.md`
- `.claude/skills/moai/workflows/{plan,run,sync}.md` (if separate from templates)
- `CLAUDE.md`

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Push Back behavior causes excessive questioning | Workflow slowdown | Constrain to "clear problems" only; accept human override |
| HUMAN GATE markers slow autonomous workflows | Reduced throughput in /moai run | Gates are advisory in auto mode; enforced in interactive mode |
| CLAUDE.md character limit (40K) exceeded | Truncation | Only add cross-reference, not full content |

## Dependencies

- SPEC-EVO-001: Evolvable zone markers for HUMAN GATE sections

## Non-Goals

- Modifying existing CLAUDE.md Section 7 rules (only adding cross-references)
- Adding gates to agency workflow skills (agency has its own GAN Loop contract)
- Making gates blocking in headless/auto mode
