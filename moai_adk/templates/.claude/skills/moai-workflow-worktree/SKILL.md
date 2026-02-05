---
name: moai-workflow-worktree
description: >
  Git worktree management for parallel SPEC development with isolated workspaces,
  automatic branch registration, and seamless MoAI-ADK integration.
  Use when setting up parallel development environments, creating isolated SPEC
  workspaces, managing git worktrees, or working on multiple features simultaneously.
  Do NOT use for regular git operations like commit or merge
  (use manager-git agent instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.1.0"
  category: "workflow"
  status: "active"
  updated: "2026-01-08"
  modularized: "true"
  tags: "git, worktree, parallel, development, spec, isolation"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["worktree", "git worktree", "parallel development", "isolated workspace", "multiple SPECs", "branch isolation", "feature branch"]
  phases: ["plan", "run"]
  agents: ["manager-git", "manager-spec", "manager-project"]
---

# MoAI Worktree Management

Git worktree management system for parallel SPEC development with isolated workspaces, automatic registration, and seamless MoAI-ADK integration.

Core Philosophy: Each SPEC deserves its own isolated workspace to enable true parallel development without context switching overhead.

## Quick Reference (30 seconds)

What is MoAI Worktree Management?
A specialized Git worktree system that creates isolated development environments for each SPEC, enabling parallel development without conflicts.

Key Features:
- Isolated Workspaces: Each SPEC gets its own worktree with independent Git state
- Automatic Registration: Worktree registry tracks all active workspaces
- Parallel Development: Multiple SPECs can be developed simultaneously
- Seamless Integration: Works with /moai:1-plan, /moai:2-run, /moai:3-sync workflow
- Smart Synchronization: Automatic sync with base branch when needed
- Cleanup Automation: Automatic cleanup of merged worktrees

Quick Access:
- CLI commands: Refer to Worktree Commands Module at modules/worktree-commands.md
- Management patterns: Refer to Worktree Management Module at modules/worktree-management.md
- Parallel workflow: Refer to Parallel Development Module at modules/parallel-development.md
- Integration guide: Refer to Integration Patterns Module at modules/integration-patterns.md
- Troubleshooting: Refer to Troubleshooting Module at modules/troubleshooting.md

Use Cases:
- Multiple SPECs development in parallel
- Isolated testing environments
- Feature branch isolation
- Code review workflows
- Experimental feature development

---

## Implementation Guide (5 minutes)

### 1. Core Architecture - Worktree Management System

Purpose: Create isolated Git worktrees for parallel SPEC development.

Key Components:

1. Worktree Registry - Central registry tracking all worktrees
2. Manager Layer - Core worktree operations including create, switch, remove, and sync
3. CLI Interface - User-friendly command interface
4. Models - Data structures for worktree metadata
5. Integration Layer - MoAI-ADK workflow integration

Registry Structure:

The registry file stores worktree metadata in JSON format. Each worktree entry contains an identifier, file path, branch name, creation timestamp, last sync time, status (active or merged), and base branch reference. The config section defines the worktree root directory, auto-sync preference, and cleanup behavior for merged branches.

File Structure:

The worktree system creates a dedicated directory structure inside the project's .moai directory. At the worktree root ({repo}/.moai/worktrees/{ProjectName}/), you will find the central registry JSON file and individual directories for each SPEC. Each SPEC directory contains a .git file for worktree metadata and a complete copy of all project files.

Detailed Reference: Refer to Worktree Management Module at modules/worktree-management.md

---

### 2. CLI Commands - Complete Command Interface

Purpose: Provide intuitive CLI commands for worktree management.

Core Commands:

To create a new worktree for a SPEC, use the new command followed by the SPEC ID and description. To list all worktrees, use the list command. To switch to a specific worktree, use the switch command with the SPEC ID. To get the worktree path for shell integration, use the go command with eval. To sync a worktree with its base branch, use the sync command. To remove a worktree, use the remove command. To clean up merged worktrees, use the clean command. To show worktree status, use the status command. For configuration management, use the config command with get or set subcommands.

Command Categories:

1. Creation: The new command creates an isolated worktree
2. Navigation: The list, switch, and go commands enable browsing and navigating
3. Management: The sync, remove, and clean commands maintain worktrees
4. Status: The status command checks worktree state
5. Configuration: The config command manages settings

Shell Integration:

For switching to a worktree directory, two approaches work well. The switch command directly changes to the worktree directory. The go command outputs a cd command that can be evaluated by the shell, which is the recommended pattern for shell scripts and automation.

Detailed Reference: Refer to Worktree Commands Module at modules/worktree-commands.md

---

### 3. Parallel Development Workflow - Isolated SPEC Development

Purpose: Enable true parallel development without context switching.

Workflow Integration:

During the Plan Phase using /moai:1-plan, the SPEC is created and the worktree new command sets up automatic worktree isolation.

During the Development Phase, the isolated worktree environment provides independent Git state with zero context switching overhead.

During the Sync Phase using /moai:3-sync, the worktree sync command ensures clean integration with conflict resolution support.

During the Cleanup Phase, the worktree clean command provides automatic cleanup with registry maintenance.

Parallel Development Benefits:

1. Context Isolation: Each SPEC has its own Git state, files, and environment
2. Zero Switching Cost: Instant switching between worktrees
3. Independent Development: Work on multiple SPECs simultaneously
4. Safe Experimentation: Isolated environment for experimental features
5. Clean Integration: Automatic sync and conflict resolution

Example Workflow:

First, create a worktree for SPEC-001 with a description like "User Authentication" and switch to that directory. Then run /moai:2-run SPEC-001 to develop in isolation. Next, navigate back to the main repository and create another worktree for SPEC-002 with description "Payment Integration". Switch to that worktree and run /moai:2-run SPEC-002 for parallel development. When needed, switch between worktrees and continue development. Finally, sync both worktrees when ready for integration.

Detailed Reference: Refer to Parallel Development Module at modules/parallel-development.md

---

### 4. Integration Patterns - MoAI-ADK Workflow Integration

Purpose: Seamless integration with MoAI-ADK Plan-Run-Sync workflow.

Integration Points:

During Plan Phase Integration with /moai:1-plan, after SPEC creation, create the worktree using the new command with the SPEC ID. The output provides guidance for switching to the worktree using either the switch command or the shell eval pattern with the go command.

During Development Phase with /moai:2-run, worktree isolation provides a clean development environment with independent Git state preventing conflicts and automatic registry tracking.

During Sync Phase with /moai:3-sync, before PR creation run the sync command for the SPEC. After PR merge, run the clean command with the merged-only flag to remove completed worktrees.

Auto-Detection Patterns:

The system detects worktree environments by checking for the registry file in the parent directory. When detected, the SPEC ID is extracted from the current directory name. The status command with sync-check option automatically identifies worktrees that need synchronization.

Configuration Integration:

The MoAI configuration supports worktree settings including auto_create for automatic worktree creation, auto_sync for automatic synchronization, cleanup_merged for automatic cleanup of merged branches, and worktree_root for specifying the worktree directory location with project name substitution.

Detailed Reference: Refer to Integration Patterns Module at modules/integration-patterns.md

---

## Advanced Implementation (10+ minutes)

### Multi-Developer Worktree Coordination

Shared Worktree Registry:

Configure team worktree settings by setting the registry type to team mode and specifying a shared registry path accessible to all team members. For developer-specific worktrees within the shared environment, use the developer flag when creating worktrees to prefix entries with the developer name. The list command with all-developers flag shows worktrees from all team members, and the status command with team-overview provides a consolidated team view.

### Advanced Synchronization Strategies

Selective Sync Patterns:

The sync command supports selective synchronization with include and exclude patterns to sync only specific directories or files. For conflict resolution, choose between auto-resolve for simple conflicts, interactive resolution for manual conflict handling, or abort to cancel the sync operation.

### Worktree Templates and Presets

Custom Worktree Templates:

Create worktrees with specific setups using the template flag. A frontend template might include npm install and eslint setup with pre-commit hooks. A backend template might include virtual environment creation, activation, and dependency installation. Configure custom templates through the config command by setting template-specific setup commands.

### Performance Optimization

Optimized Worktree Operations:

For faster worktree creation, use the shallow flag with a depth value for shallow clones. The background flag enables background synchronization. The parallel flag with all option enables parallel operations across all worktrees. Enable caching through configuration with cache enable and cache TTL settings for faster repeated operations.

---

## Works Well With

Commands:
- moai:1-plan - SPEC creation with automatic worktree setup
- moai:2-run - Development in isolated worktree environment
- moai:3-sync - Integration with automatic worktree sync
- moai:9-feedback - Worktree workflow improvements

Skills:
- moai-foundation-core - Parallel development patterns
- moai-workflow-project - Project management integration
- moai-workflow-spec - SPEC-driven development
- moai-git-strategy - Git workflow optimization

Tools:
- Git worktree - Native Git worktree functionality
- Rich CLI - Formatted terminal output
- Click framework - Command-line interface framework

---

## Quick Decision Guide

For new SPEC development, use the worktree isolation pattern with auto-setup. The primary approach is worktree isolation and the supporting pattern is integration with /moai:1-plan.

For parallel development across multiple SPECs, use multiple worktrees with shell integration. The primary approach is maintaining multiple worktrees and the supporting pattern is fast switching between them.

For team coordination in shared environments, use shared registry with developer prefixes. The primary approach is the shared registry pattern and the supporting pattern is conflict resolution.

For code review workflows, use isolated review worktrees. The primary approach is worktree isolation for reviews and the supporting pattern is clean sync after review completion.

For experimental features, use temporary worktrees with auto-cleanup. The primary approach is creating temporary worktrees and the supporting pattern is safe experimentation with automatic removal.

Module Deep Dives:
- Worktree Commands: Refer to modules/worktree-commands.md for complete CLI reference
- Worktree Management: Refer to modules/worktree-management.md for core architecture
- Parallel Development: Refer to modules/parallel-development.md for workflow patterns
- Integration Patterns: Refer to modules/integration-patterns.md for MoAI-ADK integration
- Troubleshooting: Refer to modules/troubleshooting.md for problem resolution

Full Examples: Refer to examples.md
External Resources: Refer to reference.md
