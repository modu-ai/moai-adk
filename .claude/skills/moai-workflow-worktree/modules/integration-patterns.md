# Integration Patterns Module

Purpose: Overview of integration patterns for moai-worktree with MoAI-ADK workflow, development tools, and external systems.

Version: 2.0.0
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

Integration Points:
- MoAI-ADK Workflow: Seamless integration with /moai:1-plan, /moai:2-run, /moai:3-sync
- Development Tools: IDEs, editors, terminal emulators, and development servers
- Git Workflows: Branch management, CI/CD pipelines, and code review processes
- Team Collaboration: Shared worktrees, code sharing, and coordination patterns

Core Integration Pattern:

The basic integration workflow follows three phases. During the Plan Phase, the /moai:1-plan command auto-creates a worktree for the SPEC. During the Development Phase, use the moai-worktree go command to navigate to the isolated environment and run /moai:2-run for DDD implementation. During the Sync Phase, use moai-worktree sync to update the worktree and /moai:3-sync for documentation synchronization.

---

## Module Architecture

This integration patterns module is organized into focused sub-modules for progressive disclosure:

MoAI-ADK Integration: Refer to moai-adk-integration.md
- Plan Phase integration with /moai:1-plan
- Development Phase integration with /moai:2-run
- Sync Phase integration with /moai:3-sync
- Automated cleanup workflows

Tools and External Integration: Refer to tools-integration.md
- IDE integration including VS Code and JetBrains
- Terminal and shell integration
- CI/CD pipeline integration
- Monitoring and analytics

---

## Integration Overview

### MoAI-ADK Workflow Integration

The worktree system integrates deeply with the MoAI Plan-Run-Sync workflow:

Plan Phase (/moai:1-plan):
- Automatic worktree creation after SPEC generation
- Template-based environment setup
- Branch naming conventions applied automatically
- Worktree-specific configuration files created

Development Phase (/moai:2-run):
- DDD execution in isolated worktree context
- Independent dependency management
- Automatic registry updates for access tracking
- Development server isolation per worktree

Sync Phase (/moai:3-sync):
- Worktree synchronization with base branch
- Conflict detection and resolution
- Documentation updates from worktree changes
- PR creation workflow integration

Cleanup Phase:
- Automatic cleanup of merged worktrees
- Archive support for completed work
- Registry maintenance and integrity checks

Detailed Reference: Refer to moai-adk-integration.md for complete workflow patterns

---

### Development Tools Integration

IDE and Editor Integration:
- VS Code multi-root workspace generation
- Dynamic workspace updates as worktrees change
- Worktree-specific IDE settings and configurations
- Task and debug configuration per worktree

Terminal Integration:
- Shell prompt customization for worktree awareness
- Tab completion for worktree commands
- Navigation aliases and functions
- Context preservation between terminal sessions

Git Hooks Integration:
- Post-checkout hooks for worktree detection
- Pre-push hooks for validation
- Automatic environment loading

Detailed Reference: Refer to tools-integration.md for tool configuration patterns

---

### External System Integration

CI/CD Pipeline Integration:
- GitHub Actions workflow templates
- Parallel testing across worktrees
- SPEC-aware build processes
- Automated deployment from worktrees

Monitoring and Analytics:
- Usage metrics collection
- Performance tracking
- Disk usage optimization
- Team activity monitoring

Detailed Reference: Refer to tools-integration.md for external system patterns

---

## Quick Decision Guide

For automatic SPEC worktree creation, integrate with /moai:1-plan by configuring auto-creation in worktree settings. Refer to moai-adk-integration.md for implementation details.

For IDE integration with worktrees, use VS Code multi-root workspace generation with dynamic worktree folder updates. Refer to tools-integration.md for configuration.

For CI/CD pipeline integration, implement GitHub Actions workflows with SPEC-aware testing patterns. Refer to tools-integration.md for workflow templates.

For team collaboration with shared registries, configure team registry settings with developer-specific worktree prefixes. Refer to moai-adk-integration.md for coordination patterns.

---

## Sub-Module References

MoAI-ADK Integration (moai-adk-integration.md):
- Complete /moai:1-plan integration patterns
- DDD-aware /moai:2-run integration
- Sync Phase automation with /moai:3-sync
- Post-PR cleanup workflows
- Team collaboration patterns

Tools Integration (tools-integration.md):
- VS Code workspace generation
- Terminal shell integration
- Git hooks configuration
- CI/CD pipeline templates
- Monitoring and analytics setup
- Resource management patterns

---

Version: 2.0.0
Last Updated: 2026-01-06
Module: Integration patterns overview with progressive disclosure to sub-modules
