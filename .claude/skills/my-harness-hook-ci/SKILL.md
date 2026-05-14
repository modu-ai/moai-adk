---
name: my-harness-hook-ci
description: >
  Hook/CI domain knowledge for moai-adk-go covering shell hook patterns, PostToolUse/SessionStart
  events, GitHub Actions workflows, CI pipeline structure, and release automation.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-05-14"
  modularized: "false"
  tags: "hooks, CI, GitHub Actions, PostToolUse, SessionStart, release, GoReleaser, workflow"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["hook", "PostToolUse", "SessionStart", "CI", "GitHub Actions", "workflow", "release", "GoReleaser", "handle-", "settings.json"]
  agents:
    - "my-harness-hook-ci-specialist"
    - "expert-devops"
  phases:
    - "run"
    - "sync"
---

# Hook/CI Domain Knowledge

Domain-specific knowledge for hooks and CI/CD in moai-adk-go. Supplements `expert-devops` with project-specific patterns.

## Quick Reference

### Hook System Architecture

```
Claude Code hook event
  -> .claude/hooks/moai/handle-{event}.sh (shell wrapper)
  -> moai hook {event} (Go binary handler)
  -> internal/hook/{handler}.go (Go handler logic)
```

### Hook Events (27 total)

**Session lifecycle**: SessionStart, SessionEnd, Compact, PostCompact, InstructionsLoaded
**Tool lifecycle**: PreTool, PostTool, PostToolFailure
**Agent lifecycle**: SubagentStart, SubagentStop, Stop, StopFailure
**State changes**: ConfigChange, CwdChanged, FileChanged
**Permissions**: PermissionRequest, PermissionDenied
**Interaction**: Notification, Elicitation, ElicitationResult, UserPromptSubmit
**Team**: TeammateIdle, TaskCreated, TaskCompleted, AgentHook
**Worktree**: WorktreeCreate, WorktreeRemove

### CI Pipeline Overview

```
PR/Push
  -> ci.yml (Lint + Test ubuntu/macos/windows + Build 5 platforms + CodeQL)
  -> release-drafter.yml (Auto-draft changelog)
  -> spec-lint.yml (SPEC document validation)
  -> docs-i18n-check.yml (4-locale doc sync)
Tag push
  -> release.yml (GoReleaser: 5-platform binaries + GitHub Release)
```

### Key Commands

```bash
# Run hook handler manually
echo '{}' | moai hook session-start

# Test specific hook
echo '{"toolName":"Write"}' | moai hook post-tool

# Verify all hook scripts exist
ls .claude/hooks/moai/handle-*.sh | wc -l  # Should be 27

# CI local check (approximation)
go vet ./... && golangci-lint run && go test ./...
```

## Implementation Guide

### Hook Wrapper Pattern

Every hook follows the thin-wrapper pattern:

```bash
#!/bin/bash
# .claude/hooks/moai/handle-{event}.sh

# Read stdin JSON from Claude Code
INPUT=$(cat)

# Call moai binary with hook subcommand
moai hook {event} <<< "$INPUT"
```

Rules:
- Always quote `$CLAUDE_PROJECT_DIR` in settings.json hook commands
- Use full path: `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-{event}.sh"`
- Set appropriate timeout (default: 5 seconds, max: 10 for post-processing)

### Agent Hook Configuration

Agent-scoped hooks defined in agent YAML frontmatter:

```yaml
hooks:
  PreToolUse:
    - matcher: "Write|Edit|MultiEdit"
      hooks:
        - type: command
          command: "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-agent-hook.sh\" {action}"
          timeout: 5
  PostToolUse:
    - matcher: "Write|Edit|MultiEdit"
      hooks:
        - type: command
          command: "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-agent-hook.sh\" {action}"
          timeout: 10
  SubagentStop:
    hooks:
      - type: command
        command: "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-agent-hook.sh\" {action}"
        timeout: 10
```

### GitHub Actions Workflow Patterns

Key workflows in `.github/workflows/`:

| Workflow | Trigger | Key Features |
|----------|---------|--------------|
| `ci.yml` | push/PR to main | Matrix: ubuntu/macos/windows, Go 1.24, lint + test + build |
| `release.yml` | tag `v*` push | GoReleaser: linux/darwin/windows, amd64/arm64 |
| `release-drafter.yml` | PR merge to main | Auto-label + draft changelog |
| `auto-merge.yml` | Dependabot CI pass | Auto-squash-merge |
| `codeql.yml` | push/PR | Go security analysis |
| `spec-lint.yml` | PR | SPEC document frontmatter validation |
| `spec-status-auto-sync.yml` | schedule | SPEC status drift detection |
| `docs-i18n-check.yml` | PR touching docs-site | 4-locale sync verification |

### Release Process

```bash
# Use release script only (never manual tag push)
./scripts/release.sh vX.Y.Z "Release description"

# Hotfix
./scripts/release.sh vX.Y.Z --hotfix
```

Release automation chain:
1. `scripts/release.sh` creates tag + pushes
2. `release.yml` triggers GoReleaser
3. GoReleaser builds 5 platforms (linux/darwin/windows, amd64/arm64)
4. GitHub Release created with binary assets
5. Release Drafter updates draft changelog

### Hook Handler Testing

Hook handlers are tested via Go unit tests in `internal/hook/`:

```go
func TestHookHandler(t *testing.T) {
    // Hook handlers read JSON from stdin
    input := `{"eventType":"SessionStart",...}`

    // Test handler logic directly
    result, err := handler.Process(input)
    if err != nil {
        t.Errorf("handler failed: %v", err)
    }
}
```

## Cross-References

- `.claude/rules/moai/core/agent-hooks.md`: Agent hook lifecycle and action mapping
- `moai-foundation-cc` skill: Claude Code hooks configuration
- CLAUDE.local.md Section 7: Hook Development Guidelines
- `.github/workflows/ci.yml`: CI pipeline configuration
- `.github/workflows/release.yml`: Release automation
- `internal/hook/`: Go hook handler implementations
