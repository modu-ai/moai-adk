# SPEC-MX-001: Implementation Plan

| Field       | Value           |
|-------------|-----------------|
| SPEC ID     | SPEC-MX-001     |
| Title       | @MX TAG System Implementation Plan |
| Created     | 2026-02-20      |
| Author      | manager-spec    |

---

## 1. Implementation Strategy

### 1.1 Approach

This is a **template-only** implementation. No Go source code changes are required. All deliverables are Markdown files (skills, agent definitions, rules, configuration templates) that live within `internal/template/templates/`.

The implementation follows the DDD methodology (ANALYZE-PRESERVE-IMPROVE) because we are augmenting existing agent definitions and workflow rules:

- **ANALYZE**: Understand current manager-ddd and manager-tdd skill structures, existing workflow rules, and template deployment patterns.
- **PRESERVE**: Ensure existing agent behaviors are not broken by the new skill addition.
- **IMPROVE**: Add the moai-workflow-mx-tag skill, update agent definitions, and create the configuration template.

### 1.2 Risk Assessment

| Risk                                       | Probability | Impact | Mitigation                                        |
|--------------------------------------------|-------------|--------|---------------------------------------------------|
| Skill file exceeds 500-line limit          | Medium      | Low    | Use progressive disclosure; split into modules/    |
| Agent context window overflow with new skill| Low         | Medium | Skill uses progressive disclosure (Level 2 only when triggered) |
| Existing agent tests break after skill addition | Low     | High   | Only add skill name to frontmatter skills list    |
| @MX prefix conflicts with user codebase    | Very Low    | Medium | Prefix is unique; .moai/config/sections/mx.yaml exclude patterns available |
| Token overhead of 3-Pass scan on large codebases | Medium | Medium | Pass 2 reads only P1 files; batch across sessions |

---

## 2. Milestones

### MILE-1: Core Tag System (Primary Goal)

**Scope**: Define all 4 tag types, syntax grammar, lifecycle state machine, configuration file, and per-file limits.

**Deliverables**:
- `.claude/skills/moai/workflows/mx-tag.md` -- Main skill file with tag definitions, syntax, lifecycle rules
- `.moai/config/sections/mx.yaml` template -- Default configuration file
- `.claude/rules/moai/workflow/mx-tag-protocol.md` -- Mandatory workflow rule

**Technical Approach**:
1. Create the skill file following MoAI skill authoring schema (YAML frontmatter + Markdown body)
2. Define all 4 tag types with EARS format requirements embedded in the skill instructions
3. Include the formal tag syntax grammar
4. Define the lifecycle state machine with transition rules
5. Create .moai/config/sections/mx.yaml template with default values
6. Create workflow rule file that is loaded by all agents during run phase

**Acceptance Criteria**: AC-NOTE-001, AC-WARN-001, AC-ANCHOR-001, AC-TODO-001, AC-SPEC-001, AC-AUTHOR-001, AC-LIFE-001 through AC-LIFE-004, AC-CONFIG-001, AC-EDGE-001, AC-EDGE-002, AC-EDGE-005, AC-EDGE-006, AC-EDGE-007

### MILE-2: Workflow Integration (Secondary Goal)

**Scope**: Integrate @MX protocol into TDD and DDD workflow cycles. Update agent definitions. Define report format.

**Deliverables**:
- Updated `.claude/agents/moai/manager-ddd.md` -- Add moai-workflow-mx-tag to skills list
- Updated `.claude/agents/moai/manager-tdd.md` -- Add moai-workflow-mx-tag to skills list
- Report format specification within the skill file

**Technical Approach**:
1. Add `moai-workflow-mx-tag` to the `skills:` frontmatter of manager-ddd.md
2. Add `moai-workflow-mx-tag` to the `skills:` frontmatter of manager-tdd.md
3. Define phase-specific tag protocols (RED/GREEN/REFACTOR for TDD, ANALYZE/PRESERVE/IMPROVE for DDD)
4. Define the agent autonomous report format
5. Document edge cases for team environment, broken SPEC links, and stale tags

**Acceptance Criteria**: AC-TDD-001 through AC-TDD-003, AC-DDD-001 through AC-DDD-003, AC-REPORT-001, AC-EDGE-003, AC-EDGE-004, AC-EDGE-008, AC-EDGE-009, AC-EDGE-010

### MILE-3: 3-Pass Fast Tagging Algorithm (Final Goal)

**Scope**: Define the legacy codebase bootstrap algorithm for DDD ANALYZE phase.

**Deliverables**:
- 3-Pass algorithm documentation within the skill file (or as a module file if content exceeds limits)

**Technical Approach**:
1. Document Pass 1 (Grep scan), Pass 2 (selective Read), Pass 3 (batch Edit)
2. Define priority queue: P1 (fan_in>=5), P2 (goroutine/complexity>=15), P3 (magic/no-godoc), P4 (no-test)
3. Include performance targets (~7 minutes for 100-file codebase)
4. Document batching strategy for large codebases (>500 files)

**Acceptance Criteria**: AC-SCAN-001

### MILE-4: Documentation and Verification (Optional Goal)

**Scope**: Verify the complete system by running a dry-run on the moai-adk-go codebase itself.

**Deliverables**:
- Verification report documenting the @MX system behavior on a real codebase
- Any adjustments to thresholds or limits based on verification findings

**Technical Approach**:
1. Run the 3-Pass scan on `internal/` directory
2. Verify tag counts do not exceed per-file limits
3. Verify fan-in analysis produces reasonable results
4. Adjust .moai/config/sections/mx.yaml defaults if needed

---

## 3. File Change Summary

### New Files

| File Path (relative to `internal/template/templates/`)     | Description                    |
|------------------------------------------------------------|--------------------------------|
| `.claude/skills/moai/workflows/mx-tag.md`                  | Core @MX TAG skill definition  |
| `.claude/rules/moai/workflow/mx-tag-protocol.md`           | @MX protocol workflow rule     |
| `.moai/config/sections/mx.yaml`                                                 | Default configuration template |

### Modified Files

| File Path (relative to `internal/template/templates/`)     | Change Description                            |
|------------------------------------------------------------|-----------------------------------------------|
| `.claude/agents/moai/manager-ddd.md`                       | Add `moai-workflow-mx-tag` to skills list     |
| `.claude/agents/moai/manager-tdd.md`                       | Add `moai-workflow-mx-tag` to skills list     |

### Unchanged Files

All other agent definitions, skill files, and configuration files remain unchanged.

---

## 4. Technical Decisions

### TD-001: Template-Only Implementation

**Decision**: Implement entirely within the template system (no Go binary changes).

**Rationale**: The @MX TAG system is an agent-level protocol -- it defines how agents annotate code using existing tools (Grep, Read, Edit). No new CLI commands, hook handlers, or binary capabilities are needed. This keeps the implementation lightweight and avoids the `make build` / embedded.go regeneration cycle for the core protocol definition.

### TD-002: Single Skill File vs. Module Directory

**Decision**: Start with a single skill file (`mx-tag.md`). If content exceeds 500 lines, split into `mx-tag/SKILL.md` + `mx-tag/modules/` directory structure.

**Rationale**: The 500-line limit is a soft target. The @MX protocol is self-contained and may fit within a single well-structured file. Splitting prematurely adds complexity without benefit.

### TD-003: Workflow Rule File for Universal Loading

**Decision**: Create a separate `.claude/rules/moai/workflow/mx-tag-protocol.md` rule file in addition to the skill.

**Rationale**: Skills are loaded only by agents that list them in frontmatter. A workflow rule file is loaded for all agents working on files in the project, ensuring that even agents not explicitly configured with the mx-tag skill (e.g., expert-backend, expert-refactoring) can see and respect existing @MX tags.

### TD-004: Grep-Based Fan-In Over AST Analysis

**Decision**: Use Grep-based reference counting for fan-in analysis instead of AST-grep.

**Rationale**: AST-grep provides higher precision but requires the ast-grep binary to be installed. Grep is universally available and provides sufficient accuracy for threshold-based decisions. False positives from name collisions are acceptable because all ANCHOR tags are reviewed in reports.

### TD-005: mx.yaml inside .moai/config/sections/

**Decision**: Place mx.yaml at `.moai/config/sections/mx.yaml` (inside .moai/ config directory).

**Rationale**: MX tags are a MoAI-specific feature. Consolidating all MoAI configuration under `.moai/config/sections/` provides consistent config management alongside quality.yaml, language.yaml, workflow.yaml, etc. Updated from original decision to place at project root.

---

## 5. Dependency Graph

```
MILE-1 (Core Tag System)
    |
    v
MILE-2 (Workflow Integration)  <-- depends on MILE-1 (tag definitions)
    |
    v
MILE-3 (3-Pass Algorithm)      <-- depends on MILE-1 (tag syntax)
    |
    v
MILE-4 (Verification)          <-- depends on MILE-1 + MILE-2 + MILE-3
```

MILE-2 and MILE-3 can be developed in parallel once MILE-1 is complete.

---

## 6. Build and Deployment

After implementation, the following build steps are required:

1. `make build` -- Regenerate `internal/template/embedded.go` to include the new template files
2. `go test ./internal/template/...` -- Verify template deployment integrity
3. Test `moai init` on a fresh project to verify the new files are deployed correctly
4. Verify `moai update` correctly merges the new files for existing projects

---

## 7. Post-Implementation Verification

### Verification Checklist

- [ ] `mx-tag.md` skill file follows MoAI skill authoring schema
- [ ] YAML frontmatter passes validation (name, description, metadata, triggers)
- [ ] manager-ddd.md skills list includes `moai-workflow-mx-tag`
- [ ] manager-tdd.md skills list includes `moai-workflow-mx-tag`
- [ ] `.moai/config/sections/mx.yaml` template deploys correctly via `moai init`
- [ ] `mx-tag-protocol.md` rule file loads correctly for all agents
- [ ] All 4 tag types have complete syntax definitions
- [ ] Lifecycle state machine covers all transitions
- [ ] 3-Pass algorithm is documented with performance targets
- [ ] Report format is fully specified
- [ ] All 10 edge cases are documented with resolution strategies
- [ ] `make build` succeeds with new template files
- [ ] `go test ./internal/template/...` passes
