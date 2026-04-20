---
name: moai-workflow-codemaps
description: >
  Scan codebase and generate architecture documentation in .moai/project/codemaps/ directory.
  Creates module maps, dependency graphs, and entry point references.
  Supports full regeneration and area-specific focus.
  Use when generating architecture documentation or visualizing codebase structure.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "codemaps, architecture, documentation, visualization, codebase-analysis"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4000

# MoAI Extension: Triggers
triggers:
  keywords: ["codemaps", "architecture", "codebase map", "module map", "dependency graph"]
  agents: ["Explore", "manager-docs"]
  phases: ["codemaps"]
---

# Workflow: Codemaps - Architecture Documentation Generation

Purpose: Scan the codebase and generate architecture documentation in `.moai/project/codemaps/` directory. Creates human-readable and AI-consumable maps of modules, dependencies, entry points, and interaction patterns.

Flow: Explore Codebase -> Analyze Architecture -> Generate Maps -> Verify -> Report

## Supported Flags

- --force (alias --regenerate): Regenerate all codemaps even if they already exist
- --area AREA: Focus on a specific area (e.g., --area api, --area auth, --area cli)
- --format FORMAT: Output format (default: markdown). Options: markdown, mermaid, json
- --depth N: Maximum directory depth for exploration (default: 4)

## Phase 1: Codebase Exploration

[HARD] Delegate codebase exploration to the Explore subagent.

Exploration Objectives passed to Explore agent:

- Directory Structure: Map all top-level and significant subdirectories
- Module Boundaries: Identify package/module boundaries and their responsibilities
- Entry Points: Find main entry points (main.go, index.ts, app.py, etc.)
- Public APIs: List exported functions, types, and interfaces
- Dependency Graph: Map inter-module dependencies (imports, requires)
- External Dependencies: Catalog third-party dependencies with purposes
- Configuration Files: Identify build, deployment, and config files

If --area flag: Limit exploration to the specified area and its dependencies.

Expected Output from Explore agent:

- Module inventory with purpose descriptions
- Dependency adjacency list (who imports whom)
- Entry point catalog
- Technology stack summary
- Architecture pattern identification (MVC, Clean, Hexagonal, etc.)

## Phase 2: Architecture Analysis

[HARD] Delegate architecture analysis to the manager-docs subagent.

Pass to manager-docs:

- Explore agent results from Phase 1
- Existing .moai/project/codemaps/ content (if --force not set, for incremental updates)
- Output format preference (from --format flag)

Analysis Tasks:

- Classify modules by layer (presentation, business, data, infrastructure)
- Identify high fan-in modules (potential @MX:ANCHOR candidates)
- Detect circular dependencies
- Map request/data flow paths
- Identify domain boundaries

## Phase 3: Map Generation

[HARD] Delegate map generation to the manager-docs subagent.

Output Files in `.moai/project/codemaps/` directory:

- `overview.md`: High-level architecture summary with module descriptions
- `modules.md`: Detailed module catalog with responsibilities and dependencies
- `dependencies.md`: Dependency graph (text and/or mermaid diagram)
- `entry-points.md`: Entry point catalog with invocation paths
- `data-flow.md`: Key data flow paths through the system

If --area flag: Generate only area-specific maps:
- `.moai/project/codemaps/{area}/overview.md`
- `.moai/project/codemaps/{area}/modules.md`
- `.moai/project/codemaps/{area}/dependencies.md`

If --format mermaid: Include mermaid diagrams in documentation.
If --format json: Generate machine-readable JSON alongside markdown.

## Phase 4: Verification

- Verify all referenced files and modules actually exist
- Check that dependency relationships are bidirectionally consistent
- Validate entry points are reachable
- Compare with existing .moai/project/codemaps/ to highlight changes (if not --force)

## Phase 5: Report

Display completion summary in user's conversation_language:

- Files generated: List of created/updated codemaps
- Architecture highlights: Key patterns and notable findings
- Potential issues: Circular dependencies, orphaned modules, high coupling

Next Steps (AskUserQuestion):

- Write SPEC for improvements (Recommended): Create a SPEC to address any architectural issues found. Useful if circular dependencies or high coupling were detected.
- Generate project documentation: Run /moai project to create or update product.md, structure.md, tech.md alongside the new codemaps.
- Review codemaps manually: Open the generated files in .moai/project/codemaps/ directory for manual review and editing.

## Task Tracking

[HARD] Task management tools mandatory:
- Each map file creation tracked as a pending task via TaskCreate
- Before each generation: change to in_progress via TaskUpdate
- After each generation: change to completed via TaskUpdate

## Agent Chain Summary

- Phase 1: Explore subagent (codebase exploration, read-only)
- Phase 2-3: manager-docs subagent (analysis and generation)
- Phase 4: MoAI orchestrator (verification checks)

## Execution Summary

1. Parse arguments (extract flags: --force, --area, --format, --depth)
2. Check for existing .moai/project/codemaps/ directory content
3. Delegate codebase exploration to Explore subagent
4. Delegate architecture analysis and map generation to manager-docs subagent
5. Verify generated maps for consistency
6. TaskCreate/TaskUpdate for all generated files
7. Report results with next step options

---

Version: 1.0.0
