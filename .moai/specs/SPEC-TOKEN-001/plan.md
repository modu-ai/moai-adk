---
id: SPEC-TOKEN-001
phase: plan
---

# Implementation Plan

## Approach

Simple YAML frontmatter edits across 17 agent definition files. Each edit removes skill lines from the `skills:` list. No agent body changes. One workflow file gets a new section for JIT language detection.

## Task Breakdown

### Task 1: Expert Agent Skill Reduction (5 files)

| Agent | Remove | Keep |
|-------|--------|------|
| expert-frontend | 23 skills (all lang, most foundation, libraries, platforms, workflows, tools, design) | moai-foundation-core, moai-domain-frontend, moai-domain-uiux, moai-workflow-testing |
| expert-backend | 23 skills (all 11 lang, most foundation, platforms, workflows) | moai-foundation-core, moai-domain-backend, moai-domain-database, moai-workflow-testing |
| expert-testing | 9 skills (all 5 lang, 2 foundation, 2 workflow) | moai-foundation-core, moai-foundation-quality, moai-workflow-testing |
| expert-debug | 8 skills (all 5 lang, 1 foundation, 1 tool) | moai-foundation-core, moai-foundation-quality, moai-workflow-loop |
| expert-devops | 7 skills (foundation-philosopher, platforms, framework, workflow, jit-docs) | moai-foundation-core, moai-platform-deployment, moai-workflow-project |
| expert-performance | 6 skills (foundation-philosopher/context, 5 workflow/lang) | moai-foundation-core, moai-foundation-quality, moai-workflow-testing |
| expert-security | 4 skills (foundation-philosopher/claude, workflow-testing, tool) | moai-foundation-core, moai-foundation-quality, moai-platform-auth |
| expert-refactoring | 3 skills (foundation-claude, foundation-quality, workflow-testing) | moai-foundation-core, moai-tool-ast-grep, moai-workflow-testing |

### Task 2: Manager Agent Skill Reduction (8 files)

| Agent | Remove | Keep |
|-------|--------|------|
| manager-spec | 9 skills (foundation-claude/context/philosopher, platforms, langs) | moai-foundation-core, moai-foundation-thinking, moai-workflow-spec, moai-workflow-project |
| manager-strategy | 5 skills (foundation-claude/context, workflow-thinking/project/spec) | moai-foundation-core, moai-foundation-thinking, moai-workflow-spec, moai-workflow-worktree |
| manager-project | 5 skills (foundation-claude/context/philosopher, workflow-spec/templates) | moai-foundation-core, moai-foundation-thinking, moai-workflow-project, moai-workflow-templates |
| manager-docs | 6 skills (foundation-claude/philosopher/quality/context, workflow-templates/jit-docs) | moai-foundation-core, moai-workflow-project, moai-workflow-jit-docs |
| manager-ddd | 6 skills (foundation-claude/philosopher/quality/context, workflow-tdd, mx-tag) | moai-foundation-core, moai-workflow-ddd, moai-workflow-testing |
| manager-tdd | 4 skills (foundation-claude/quality, workflow-ddd, mx-tag) | moai-foundation-core, moai-workflow-tdd, moai-workflow-testing |
| manager-quality | 4 skills (foundation-claude/context, workflow-testing, workflow-loop) | moai-foundation-core, moai-foundation-quality, moai-tool-ast-grep |
| manager-git | 3 skills (foundation-claude/quality, workflow-testing) | moai-foundation-core, moai-workflow-project, moai-workflow-worktree |

### Task 3: Builder Agent Skill Reduction (1 file)

| Agent | Remove | Keep |
|-------|--------|------|
| builder-skill | 1 skill (workflow-project) | moai-foundation-core, moai-foundation-claude, moai-workflow-templates |

### Task 4: Add JIT Language Detection to run.md

Add a new section "Phase 1.05: JIT Language Skill Detection" to workflows/run.md that instructs the orchestrator to:
1. Detect project language from root indicator files
2. Include appropriate `Skill("moai-lang-{lang}")` instruction in agent spawn prompts

### Task 5: Regenerate embedded templates

Run `make build` to regenerate internal/template/embedded.go after all changes.

### Task 6: Run tests

Run `go test ./...` to verify no regressions.

## Execution Order

Tasks 1-3 can run in parallel (independent file edits).
Task 4 depends on Tasks 1-3 (context needed).
Task 5 depends on all previous tasks.
Task 6 depends on Task 5.

## Risk Mitigation

- All changes are YAML frontmatter only (no logic changes)
- Agent body content unchanged (behavior preserved)
- JIT loading ensures language knowledge still available
- `make build` + `go test` verify integrity
