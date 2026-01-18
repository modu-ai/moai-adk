# Parallel Development Module

Purpose: Advanced patterns and workflows for parallel SPEC development using isolated worktrees, enabling true concurrent development without context switching overhead.

Version: 2.0.0
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

Parallel Development Benefits:
- Context Isolation: Each SPEC has independent Git state, files, and environment
- Zero Switching Cost: Instant switching between worktrees without loading/unloading
- Concurrent Development: Multiple SPECs developed simultaneously by single developer
- Safe Experimentation: Isolated environments prevent conflicts and contamination
- Clean Integration: Automatic sync and conflict resolution maintain code integrity

Core Workflow:

To set up parallel development, create worktrees for each SPEC using moai-worktree new with the SPEC ID and description. For parallel execution, navigate to each worktree using moai-worktree go and run /moai:2-run in separate terminals or background processes. For integration, use moai-worktree sync --all to synchronize all worktrees, then clean up merged worktrees with moai-worktree clean --merged-only.

---

## Module Architecture

This parallel development module is organized into focused sub-modules:

Parallel Workflows: Refer to parallel-workflows.md
- Independent SPEC development patterns
- Sequential feature development with overlap
- Experiment-production parallel patterns
- Multi-developer coordination

Advanced Use Cases: Refer to parallel-advanced.md
- Feature flag development patterns
- Microservices parallel development
- Performance optimization strategies
- IDE and shell integration

---

## Worktree Isolation Model

### Isolation Layers

Each worktree provides complete isolation across multiple layers:

Git State Isolation:
- Independent branch with separate commit history
- No interference with other worktree branches
- Changes remain isolated until explicit merge

File System Isolation:
- Complete project copy in dedicated directory
- Independent modifications without affecting others
- Separate working directory state

Dependency Isolation:
- Independent node_modules per worktree
- Separate Python virtual environments
- Independent build artifacts and caches

Configuration Isolation:
- Worktree-specific .env files
- Independent IDE settings per worktree
- Separate tool configurations

Process Isolation:
- Independent development servers
- Separate test runners
- Non-conflicting ports

### Directory Structure

Parallel development creates this structure:

Main Repository (project_root/):
- Standard Git repository with .git directory
- Source code, docs, and configuration files
- Central worktree registry file

Worktree Root ({repo}/.moai/worktrees/{project-name}/):
- Contains all worktrees for the project
- Each SPEC has dedicated subdirectory
- Independent environment per worktree

Individual Worktree (SPEC-001/):
- .git file linking to main repository
- Complete project file copy
- Worktree-specific configuration files
- Independent dependency directories

---

## Development Patterns Overview

### Independent SPEC Development

Multiple unrelated features developed simultaneously without interference.

Pattern Characteristics:
- Features have no code dependencies
- Each developer works in isolation
- Integration happens at completion

Best Practices:
- Create worktrees at planning phase
- Sync periodically to stay current with main
- Clean up immediately after merge

Detailed Reference: Refer to parallel-workflows.md for implementation details

### Sequential with Overlap

Features with dependencies developed in sequence with preparation overlap.

Pattern Characteristics:
- Later features depend on earlier ones
- Preparation work begins before dependencies complete
- Integration follows dependency chain

Best Practices:
- Create dependent worktrees early for setup
- Sync after dependency completion
- Maintain clear dependency chain

Detailed Reference: Refer to parallel-workflows.md for implementation details

### Experiment-Production Parallel

Experimental features alongside stable production work.

Pattern Characteristics:
- Production fixes in stable worktree
- Experiments in separate worktree
- No risk to production stability

Best Practices:
- Clear naming convention for experiment vs production
- Regular backup of experimental work
- Careful merge strategy for experiments

Detailed Reference: Refer to parallel-workflows.md for implementation details

---

## Multi-Developer Coordination

### Team Worktree Patterns

Coordinate parallel development across team members.

Shared Registry:
- Central registry accessible to all developers
- Developer-specific worktree prefixes
- Team visibility into all active work

Coordination Strategies:
- Define integration points early
- Maintain shared contracts between related SPECs
- Regular sync to shared branches

Detailed Reference: Refer to parallel-workflows.md for team patterns

---

## Quick Decision Guide

For multiple unrelated features, use independent SPEC development pattern. Create worktrees at planning, sync periodically, clean after merge. Refer to parallel-workflows.md for details.

For features with dependencies, use sequential with overlap pattern. Create dependent worktrees early, sync after dependency completion, maintain dependency chain. Refer to parallel-workflows.md for details.

For experimental work, use experiment-production parallel pattern. Clear naming, regular backup, careful merge. Refer to parallel-workflows.md for details.

For team coordination, use shared registry with developer prefixes. Define integration points, maintain contracts, regular sync. Refer to parallel-workflows.md for details.

For advanced patterns including feature flags, microservices, and performance optimization, refer to parallel-advanced.md.

---

## Sub-Module References

Parallel Workflows (parallel-workflows.md):
- Independent SPEC development pattern implementation
- Sequential feature development with overlap
- Experiment-production parallel patterns
- Multi-developer coordination strategies
- CI/CD pipeline integration for parallel work

Advanced Use Cases (parallel-advanced.md):
- Feature flag development patterns
- Microservices architecture with parallel worktrees
- Performance and resource optimization
- IDE integration for multi-worktree development
- Shell integration and quick functions

---

Version: 2.0.0
Last Updated: 2026-01-06
Module: Parallel development overview with progressive disclosure to sub-modules
