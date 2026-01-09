---
name: moai:all-is-well
description: "One-click automation - From SPEC generation to documentation sync"
argument-hint: '"feature description" [--branch] [--pr]'
allowed-tools: Task, AskUserQuestion, TodoWrite, Skill
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git branch --show-current

## Essential Files

@.moai/config/sections/ralph.yaml
@.moai/config/sections/git-strategy.yaml

---

# /moai:all-is-well - One-Click Development Automation

Execute the complete MoAI development workflow with a single command.

## Command Purpose

Automates the full "Plan -> Run -> Sync" workflow:
1. Creates SPEC from description (`/moai:1-plan`)
2. Implements with TDD (`/moai:2-run`)
3. Synchronizes documentation (`/moai:3-sync`)

Feature Description: $ARGUMENTS

## Usage Examples

Basic usage (uses git-strategy settings):
```
/moai:all-is-well "User authentication with JWT tokens"
```

With branch creation override:
```
/moai:all-is-well "Shopping cart feature" --branch
```

With PR creation override:
```
/moai:all-is-well "Payment integration" --pr
```

## Command Options

- `--branch`: Override git-strategy to create a feature branch
- `--pr`: Override git-strategy to create a pull request after sync
- Default behavior follows `.moai/config/sections/git-strategy.yaml` settings

## Workflow Execution

### Phase 1: SPEC Generation

Agent: manager-spec (via /moai:1-plan)

Actions:
- Analyze feature description
- Generate SPEC document in EARS format
- Create acceptance criteria
- Present SPEC for user approval

Checkpoint: User must approve SPEC before proceeding

### Phase 2: TDD Implementation

Agent: manager-tdd (via /moai:2-run)

Actions:
- Create execution plan from SPEC
- Execute RED-GREEN-REFACTOR cycle
- Achieve minimum 85% test coverage
- Validate with TRUST 5 framework

Checkpoint: Quality gate validation before proceeding

### Phase 3: Documentation Sync

Agent: manager-docs (via /moai:3-sync)

Actions:
- Update documentation to match implementation
- Create or update README sections
- Generate API documentation if applicable
- Create PR if configured

Checkpoint: Final review and completion summary

## Git Strategy Integration

This command respects your git-strategy.yaml configuration:

Manual Mode (default):
- No automatic branch creation
- No automatic PR creation
- All changes on current branch

Personal Mode:
- Auto-creates feature branch
- Commits automatically
- PR creation optional

Team Mode:
- Auto-creates feature branch
- Commits automatically
- Auto-creates draft PR

Override with `--branch` or `--pr` flags when needed.

## Ralph Engine Integration

When Ralph Engine is enabled (ralph.yaml):
- LSP diagnostics run after each file change
- AST-grep security scanning is active
- Feedback loop ensures zero-error completion

## Error Recovery

If any phase fails:
- Current progress is preserved
- User is notified of the failure point
- Recovery options are presented:
  - Retry the failed phase
  - Skip to next phase (with warning)
  - Abort and preserve work

## Execution Flow

```
START: /moai:all-is-well "feature description"

PARSE: Extract description and flags from $ARGUMENTS

PHASE 1: Execute /moai:1-plan
  -> Generate SPEC document
  -> User approval checkpoint
  -> IF rejected: EXIT with SPEC for review

PHASE 2: Execute /moai:2-run SPEC-XXX
  -> TDD implementation cycle
  -> Quality validation
  -> IF critical issues: HALT for user decision

PHASE 3: Execute /moai:3-sync SPEC-XXX
  -> Documentation synchronization
  -> PR creation (if configured)
  -> Final summary

END: Complete workflow summary with next steps
```

## Success Criteria

- SPEC document created and approved
- Implementation complete with 85%+ coverage
- All tests passing
- Documentation synchronized
- Git operations complete (per configuration)

## Output Format

Phase completion reports use Markdown formatting:

```markdown
## All-Is-Well Workflow Complete

### Summary
- SPEC: SPEC-XXX created and approved
- Implementation: 12 files, 88% coverage
- Tests: 24/24 passing
- Documentation: Updated

### Git Status
- Branch: feature/SPEC-XXX (or current branch)
- Commits: 3 commits created
- PR: #123 created (if applicable)

### Next Steps
1. Review the implementation
2. Run manual testing if needed
3. Merge when ready
```

---

## Implementation Notes

Tool Usage: This command orchestrates through Task() delegation only.

User Interaction: All AskUserQuestion calls happen at command level before delegation.

Context Propagation: Each phase receives context from previous phases.

Interruption Recovery: Loop state preserved in `.moai/cache/` for resume capability.

---

Version: 1.0.0
Pattern: Sequential Phase Orchestration
Integration: Ralph Engine, Git Strategy, TRUST 5
