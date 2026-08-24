# Progress — SPEC-V3R6-GRAPH-FRESHNESS-001

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set: Tier L 5-file set (spec.md, plan.md, acceptance.md, design.md, research.md) + progress.md — Tier L justified by 5 milestones spanning 4 subsystems (internal/graph, internal/cli, internal/navigator/astx, internal/hook/quality + CI + MCP) and 2 cross-cutting conventions; the per-layer metric and cache-anchoring decisions need design.md, and the graft-analysis provenance + anchor re-verification live in research.md so the run phase never re-derives them.
- Requirements: 20 REQ (GEARS), all pattern-annotated; 22 AC (Given-When-Then) + §D.5 mutant-coverage table; REQ↔AC traceability 100% both directions (acceptance.md §D.2).
- Frontmatter: 12 canonical fields + `tier: L` + `era: V3R6`; SPEC ID regex-validated (`PASS`) before write; ID uniqueness verified against the catalog (research.md §3).
- Split decision: single SPEC — 20 REQ / 22 AC sit inside the Tier L ceilings (25/25) with slack; M1-M3 and M4-M5 share the provenance/cache substrate (M2's cache feeds M4/M5), so a split would duplicate that substrate in both halves.
- Evidence base: `.moai/reports/graft/graft-analysis-20260824.md` (read, not re-derived) + anchor re-verification at `baa100ce5` (research.md §2) — drift re-measured 740 commits; mx-index provenance absence and fresh-worktree artifact absence are directly measured facts.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
