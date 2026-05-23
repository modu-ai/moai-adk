---
id: SPEC-V3R6-CLI-AUDIT-001
title: "Progress Tracker — moai CLI inventory + dead command + integration analysis"
version: "0.1.0"
status: in-progress
created: 2026-05-23
updated: 2026-05-23
author: manager-develop
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
tags: "cli, audit, research, sprint-2, sprint-7-baseline, progress"
---

# Progress Tracker — SPEC-V3R6-CLI-AUDIT-001

## Milestone Tracker

| Milestone | Status | Commit SHA | Files touched | AC link | Notes |
|-----------|--------|-----------|---------------|---------|-------|
| M1 — Subcommand Inventory | PASS | b5ea8d936 | audit-2026-05-23.md (NEW §1) + progress.md (NEW) | AC-CLA-001 | Cobra command tree walked from cmd/moai/main.go root → 113 total subcommands |
| M2 — Dead-Command Classification | PASS | TBD | audit-2026-05-23.md (APPEND §2) + progress.md (UPDATE) | AC-CLA-002 | 116 classification rows (active/internal-only/dead-suspect). 10 dead-suspect candidates (Claude Code forward-compat hook events). harness-observe* preliminary suspect REFUTED (4 hooks reference). db-schema-sync confirmed internal-only per SPEC-DB-SYNC-001. |
| M3 — Integration Map | TBD | TBD | audit-2026-05-23.md (APPEND §3) + progress.md (UPDATE) | AC-CLA-003 | TBD |
| M4 — Sprint 7 Baseline + Methodology + Frontmatter Sync | TBD | TBD | audit-2026-05-23.md (APPEND §4 + §5) + 4 SPEC artifacts (frontmatter sync) + progress.md (UPDATE) | AC-CLA-004 + AC-CLA-005 + AC-CLA-006 | TBD |

## AC Verification Matrix (populated at M4)

| AC | Status | Command | Expected | Actual |
|----|--------|---------|----------|--------|
| AC-CLA-001 | TBD | grep -cE '^\| `moai ' audit-*.md | ≥40 | TBD |
| AC-CLA-002 | TBD | grep -cE '\| (active\|internal-only\|dead-suspect) \|' audit-*.md | ≥40 | TBD |
| AC-CLA-003 | TBD | awk-bounded count §3.2 matrix rows | ≥11 | TBD |
| AC-CLA-004 | TBD | grep -c '^### §4\.[1-4]' audit-*.md | =4 | TBD |
| AC-CLA-005 | TBD | git diff --name-only main (9 protected path classes) | =0 | TBD |
| AC-CLA-006 | TBD | grep -cE 'Generated:\|Git SHA:\|moai version:' audit-*.md | ≥3 | TBD |

## Sync-phase Evidence (populated at /moai sync)

(Populated by manager-docs after sync phase; left blank intentionally during run-phase.)

---

Version: 0.1.0 (run-phase initial — M1 in progress)
