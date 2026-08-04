---
id: SPEC-PROJECT-NAVIGATOR-003
title: "Project Navigator — tree-sitter auto-derivation into /moai codemaps (16-language AST-based capability rows)"
version: "0.0.0"
status: draft
created: 2026-08-05
updated: 2026-08-05
author: manager-spec
priority: P3
phase: "v3.3 target"
module: project-navigator
lifecycle: spec-anchored
tags: "navigator, tree-sitter, codemaps, ast, auto-derivation, 16-language"
related_specs: [SPEC-PROJECT-NAVIGATOR-001, SPEC-PROJECT-NAVIGATOR-002]
---

# SPEC-PROJECT-NAVIGATOR-003 — tree-sitter auto-derivation (STUB)

> **STUB — not authored at plan-phase.** This entry reserves the SPEC ID and records the boundary decision. It will be fully authored after SPEC-PROJECT-NAVIGATOR-001 and -002 land, because tree-sitter integration builds on top of 001's capability-map format and 002's audit surface.

## Problem (placeholder)

The user-approved Project Navigator scope names a P3 — auto-deriving the capability-map's file/symbol/status columns from tree-sitter AST analysis, integrated into `/moai codemaps`. Today, 001's `capability-map.md` is hand-derived from the SPEC registry + git log; this SPEC enriches it with mechanically-extracted symbol-level rows (research grounding: Aider repo map, Codebase-Memory — see 001's `research.md`).

## Constraints (already decided, non-negotiable)

- **16-language neutrality**: the tree-sitter grammar set MUST cover all 16 supported languages (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift). No Go-only bias (CLAUDE.local.md §15).
- **Integration, not reinvention**: this SPEC extends the existing `/moai codemaps` workflow (`.claude/skills/moai/workflows/codemaps.md`) and consumes the existing `codemaps-extract.js` fan-out where present; it does NOT replace codemaps.
- **Template-First**: any new grammar assets or skill references ship via `internal/template/templates/` + `make build` + §25 neutrality.

## Boundary vs 001 / 002

- **001 owns**: the artifact set + regeneration procedure (hand-derived rows from SPEC registry).
- **002 owns**: the audit algorithm over 001's artifact set.
- **003 owns**: MECHANICALLY enriching 001's capability-map rows by extracting file/symbol/status from the AST via tree-sitter, integrated into `/moai codemaps` so regeneration produces enriched rows.

## Why this is a separate SPEC (Tier L)

- 16-language grammar coverage is genuinely large; each language is a non-trivial verification surface.
- Different cost surface from 001/002: tree-sitter carries a binary dependency per language, a parsing-compatibility surface, and CI matrix implications.
- Different cadence: can ship after 001/002 without blocking them.
- **Tier classification: Tier L** (full artifact set: spec + plan + acceptance + research + design + progress).

## Out of Scope (for this stub)

### Out of Scope — Navigator artifact generation

- Generating the base artifact set is owned by SPEC-PROJECT-NAVIGATOR-001.

### Out of Scope — drift audit

- The audit algorithm is owned by SPEC-PROJECT-NAVIGATOR-002.
