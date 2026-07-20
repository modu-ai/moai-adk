---
description: >
  Generates project documentation from codebase analysis or user input.
  Creates product.md, structure.md, and tech.md in .moai/project/ directory,
  plus architecture maps in .moai/project/codemaps/ directory.
  Supports new and existing project types with LSP server detection.
  Use when initializing projects or generating project documentation.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "project, documentation, initialization, codebase-analysis, setup"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["project", "init", "documentation", "setup", "initialize"]
  agents: ["manager-docs", "Explore"]
  phases: ["project"]
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai project Phase <N> <enter|exit> -->

# Workflow: project - Project Documentation Generation

Purpose: Generate project documentation through smart questions and codebase analysis. Creates product.md, structure.md, and tech.md in .moai/project/ directory, plus architecture documentation in .moai/project/codemaps/ directory.

This workflow is also triggered automatically when project documentation does not exist and the user requests other workflows (plan, run, sync, etc.). See SKILL.md Step 2.5 for the auto-detection mechanism.

---

## Phase Routing Table

| Phase | Sub-skill | Description |
|---|---|---|
| Mode flag / Scope boundary | `project/mode-detection.md` | --mode compatibility, NO SPEC Generation rule |
| Phase 1: Project Type Detection | `project/mode-detection.md` | Auto-detect existing vs. new project |
| Phase 2: Deep Interview (New) | `project/mode-detection.md` | Two-stage interview: clarity-scored Stage A (Vision-Domain / Technology-Constraints / Scope, variable length up to `project.max_rounds`) + mandatory Stage B extended-axes round |
| Phase 3: Codebase Analysis | `project/codebase-analysis.md` | Explore subagent analysis (existing projects) |
| Phase 4: Deep Interview (Existing) | `project/codebase-analysis.md` | Two-stage interview: clarity-scored Stage A (Ownership-Goal / Constraints / Scope-Priority, variable length up to `project.max_rounds`) + mandatory Stage B extended-axes round |
| Phase 5: User Confirmation | `project/codebase-analysis.md` | Present analysis summary, get proceed/cancel |
| Phase 6: Documentation Generation | `project/doc-generation.md` | manager-docs for product.md/structure.md/tech.md |
| Phase 7: Independent Audit | `project/doc-generation.md` | plan-auditor conditional audit + retry loop |
| Phase 8: harness-spec.yaml Emission | `project/doc-generation.md` | Emit `.moai/project/harness-spec.yaml` from `interview.md` answers (8-field schema, no user interaction, overwrite) |
| Phase 9: Codemaps Generation | `project/doc-generation.md` | Explore + manager-docs for codemaps/ |
| Phase 10: Dev Environment Check | `project/doc-generation.md` | LSP server detection + optional install |
| Phase 11: MCP Server Provisioning | `project/doc-generation.md` | Detect stack, select MCP servers from `mcp-matrix.yaml`, orchestrator approval, additive `.mcp.json` write |
| Phase 12: Dev Methodology Config | `project/doc-generation.md` | Auto-set development_mode in quality.yaml |
| Phase 13: DB Detection | `project/doc-generation.md` | Grep/Glob DB keyword detection, db-detection.json |
| Phase 14: Completion | `project/doc-generation.md` | Summary report + 3-branch next-steps AskUserQuestion |
| Phase 15: Harness Generation Entry | `project/meta-harness.md` | Redirect to the v4 harness Builder (Context-First Discovery + orchestrator-direct 4-phase Builder) |
| Phase 16: 5-Layer Activation | `project/meta-harness.md` | Install CLAUDE.md marker + main.md router, post-generation smoke gate |

---

## Invocation Flow

```
/moai project
  └─ Mode Detection (mode-detection.md)
       ├─ New Project → Phase 2 interview → Phase 6 (skip Phase 3/2)
       └─ Existing Project → Phase 3 analysis
                              └─ codebase-analysis.md
                                   ├─ Phase 4: Two-stage interview (Stage A + Stage B)
                                   └─ Phase 5: User Confirmation
                                        └─ doc-generation.md
                                             ├─ Phase 6: Doc generation
                                             ├─ Phase 7: Audit (conditional)
                                             ├─ Phase 8: harness-spec.yaml emission
                                             ├─ Phase 9: Codemaps
                                             ├─ Phase 10: LSP check
                                             ├─ Phase 11: MCP provisioning
                                             ├─ Phase 12: Dev mode config
                                             ├─ Phase 13: DB detection
                                             └─ Phase 14: Completion
                                                  └─ [optional] Phase 15-Phase 16
                                                       └─ meta-harness.md
                                                            ├─ Phase 15: v4 Builder entry + generation
                                                            └─ Phase 16: 5-Layer Activation
```

---

## Detection Keywords Reference

Full DB engine keywords, dependency manifest files (16 languages), and ORM/ODM lists used by Phase 13 are defined in `project/doc-generation.md` §Detection Keywords Reference section.

For convenience, the DB engine categories are: Relational/SQL (PostgreSQL, MySQL, MariaDB, SQLite, Oracle, SQL Server, CockroachDB, Supabase, Neon, Planetscale), NoSQL Document (MongoDB, Firestore, Firebase, Couchbase), NoSQL Key-Value (Redis, DynamoDB, Cassandra, ScyllaDB, Riak), Search/Analytics (Elasticsearch, ClickHouse, Snowflake, InfluxDB).

---

Version: 2.5.0
Last Updated: 2026-02-21
Provenance: the DB-detection hint policy, the project-harness generation policy, the workflow-split policy
