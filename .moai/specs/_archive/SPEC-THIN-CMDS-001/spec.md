# SPEC-THIN-CMDS-001: Verify and Optimize Slash Commands

## Meta

- **Status**: Draft
- **Wave**: 2 (parallel with REFLECT-001)
- **Created**: 2026-04-12
- **Origin**: addyosmani/agent-skills P3.1 (Thin Command Pattern)
- **Blocked By**: SPEC-EVO-001

## Objective

Verify that all moai slash commands are thin routing wrappers (6-20 LOC) and not containers of workflow logic. Exploration confirmed most commands are already thin — this SPEC focuses on **verification, consistency auditing, and minor improvements** rather than major refactoring.

## Background

Initial analysis assumed moai commands contained significant workflow logic that should be moved to skills (following addyosmani/agent-skills' P3.1 pattern). Codebase exploration confirmed:

**Finding**: All 20 moai commands are already 6-7 LOC thin wrappers that call `Skill("moai")` with forwarded arguments.

**Current structure:**
```
internal/template/templates/.claude/commands/moai/
  fix.md.tmpl, run.md.tmpl, plan.md.tmpl, sync.md.tmpl,
  review.md.tmpl, clean.md.tmpl, project.md.tmpl,
  loop.md.tmpl, coverage.md.tmpl, codemaps.md.tmpl,
  feedback.md.tmpl, mx.md.tmpl, e2e.md.tmpl
```

This SPEC reduces to a quality audit: verify consistency, detect drift from the thin-wrapper pattern, and document the pattern for future command authors.

## Requirements (EARS Format)

### R1: Command LOC Audit [UBIQ]

ALL moai and agency slash command files SHALL be under 20 lines of content (excluding frontmatter).

**Acceptance Criteria:**
- [ ] Automated audit script verifies each command file LOC
- [ ] `internal/template/templates/.claude/commands/moai/*.md.tmpl`: all < 20 LOC body
- [ ] `internal/template/templates/.claude/commands/agency/*.md.tmpl`: all < 20 LOC body
- [ ] Commands exceeding threshold are flagged and documented
- [ ] Current state (post-exploration): expected to pass without changes

### R2: Frontmatter Consistency [UBIQ]

ALL slash command files SHALL use consistent YAML frontmatter structure.

**Required Fields:**
```yaml
---
description: [One-sentence what this command does]
argument-hint: "[Optional arg description]"
allowed-tools: [Comma-separated tool list]
---
```

**Acceptance Criteria:**
- [ ] Every command has `description` field (1-sentence, present tense)
- [ ] Every command has `argument-hint` field (or explicit `""` if none)
- [ ] Every command has `allowed-tools` field as CSV string (not YAML array — per CLAUDE.local.md Section 18)
- [ ] No command uses deprecated fields (check for model, disallowedTools)
- [ ] Field ordering is consistent across all commands

### R3: Skill Invocation Pattern [UBIQ]

ALL moai slash commands SHALL invoke a skill via the canonical pattern.

**Canonical Pattern:**
```markdown
---
description: [What]
argument-hint: "[Args]"
allowed-tools: Skill, Read, Grep, Glob, Bash
---

Use Skill("moai") with arguments: [subcommand] $ARGUMENTS
```

**Acceptance Criteria:**
- [ ] Every moai command invokes `Skill("moai")` (not direct agent call)
- [ ] Argument forwarding uses `$ARGUMENTS` placeholder
- [ ] Subcommand name matches the file name (plan.md → plan, run.md → run)
- [ ] No command contains multi-paragraph instructions (those belong in skills)

### R4: Thin Command Pattern Documentation [OPT]

The system SHOULD document the thin command pattern for future command authors.

**Acceptance Criteria:**
- [ ] New section in `.claude/rules/moai/development/coding-standards.md`: "Thin Command Pattern"
- [ ] Section explains: commands are routing-only, logic belongs in skills
- [ ] Includes template example
- [ ] References this SPEC as source
- [ ] Wrapped in `<!-- moai:evolvable-start id="thin-command-pattern" -->` markers

### R5: Audit Test [UBIQ]

The system SHALL include an automated test that enforces the thin command pattern.

**Acceptance Criteria:**
- [ ] New file `internal/template/commands_audit_test.go`
- [ ] Test reads all `*.md.tmpl` in commands directories
- [ ] Test verifies each command: LOC < 20 body, valid frontmatter, Skill invocation pattern
- [ ] Test fails if any command drifts from the pattern
- [ ] Test runs as part of `go test ./internal/template/...`

## Modified Files

### Templates (minor or no changes expected)
- `internal/template/templates/.claude/commands/moai/*.md.tmpl`: Audit only; fix drifts if any
- `internal/template/templates/.claude/commands/agency/*.md.tmpl`: Same
- `internal/template/templates/.claude/rules/moai/development/coding-standards.md`: Add Thin Command Pattern section

### Go Code
- `internal/template/commands_audit_test.go`: NEW - Pattern enforcement test

### Audit Output
- `.moai/reports/command-audit.md`: Generated audit report with current state

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Existing commands pass but some edge case breaks | Low | Exploration already confirmed 20 commands are thin |
| Test too strict for legitimate variations | CI failures | Review failures case-by-case; adjust test, not commands |
| Audit reveals many drifts | Larger refactor than expected | Stop and escalate; do not silently rewrite commands |

## Dependencies

- SPEC-EVO-001: Evolvable zone markers (for R4 documentation section)

## Non-Goals

- Rewriting commands that are already thin
- Moving workflow logic (nothing significant to move per exploration)
- Adding new commands (separate SPEC if needed)
- Changing Skill tool invocation mechanism

## Context Note

**This SPEC is intentionally minimal** because exploration revealed the pattern is already followed. The primary value is institutionalizing the pattern via test enforcement (R5) and documentation (R4), preventing future drift.

If exploration during implementation reveals significant drifts (>3 commands failing R1-R3), this SPEC should be split into:
- SPEC-THIN-CMDS-001: Audit + minor fixes (this SPEC, reduced scope)
- SPEC-THIN-CMDS-002: Refactor detected drift commands (new SPEC)
